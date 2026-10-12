#!/bin/bash

# =============================================================================
# 数据库每日备份脚本（supervisord 托管，容器内运行）
#
# 功能：
#   - 每天 03:00（容器时区 Asia/Shanghai）自动 mysqldump 备份 training 库
#   - 备份以 gzip 压缩存放于 /app/db_backup（宿主机卷 training_db_backup）
#   - 保留最近 14 份，更早的自动清理
#   - 失败自动重试（5 分钟间隔，最多 3 次），失败不影响下一轮
#
# 手动触发（立即备份一次，不进入循环）：
#   docker exec training /app/backup.sh now
# =============================================================================
BACKUP_DIR="/app/db_backup"
DB_NAME="training"          # 与 config.yaml 的 mysql.database 保持一致
KEEP=14                     # 保留最近 N 份备份
BACKUP_HOUR=3               # 每天备份时刻（小时，24 小时制）
RETRY_INTERVAL=300          # 备份失败重试间隔（秒）
RETRY_MAX=3                 # 最多重试次数
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"

log() { echo "[Backup] $(date '+%Y-%m-%d %H:%M:%S') $*"; }

# 等待 MySQL 就绪（最多 5 分钟），就绪返回 0
wait_for_mysql() {
    for i in $(seq 1 300); do
        if mysqladmin ping -uroot -p"$MYSQL_ROOT_PASSWORD" --silent >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    return 1
}

# 执行一次备份（含重试），成功返回 0
do_backup() {
    local attempt=1
    while true; do
        if ! wait_for_mysql; then
            log "MariaDB 未就绪，跳过本次备份"
            return 1
        fi

        local ts file tmp
        ts=$(date '+%Y%m%d-%H%M%S')
        file="$BACKUP_DIR/training-${ts}.sql.gz"
        tmp="${file}.tmp"

        # --single-transaction: InnoDB 一致性快照，不锁表
        # --quick: 逐行读取，防大表撑爆内存；--routines/--triggers: 保留存储过程与触发器
        if mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
            --single-transaction --quick --routines --triggers \
            --default-character-set=utf8mb4 "$DB_NAME" 2>/tmp/mysqldump.err | gzip > "$tmp"; then
            # 校验产物非空（mysqldump 失败时可能产出空 gzip）
            if [ -s "$tmp" ]; then
                mv "$tmp" "$file"
                log "备份成功: $file ($(du -h "$file" | cut -f1))"
                prune_old
                return 0
            fi
            rm -f "$tmp"
        fi

        rm -f "$tmp"
        log "第 ${attempt} 次备份失败: $(tail -1 /tmp/mysqldump.err 2>/dev/null)"
        if [ "$attempt" -ge "$RETRY_MAX" ]; then
            log "已达最大重试次数，等待下一轮备份时刻"
            return 1
        fi
        attempt=$((attempt + 1))
        sleep "$RETRY_INTERVAL"
    done
}

# 清理超出保留份数的旧备份
prune_old() {
    local old
    old=$(ls -1t "$BACKUP_DIR"/training-*.sql.gz 2>/dev/null | tail -n +$((KEEP + 1)))
    if [ -n "$old" ]; then
        echo "$old" | xargs rm -f
        log "已清理 $(( $(echo "$old" | wc -l) )) 份过期备份（保留最近 $KEEP 份）"
    fi
}

# 距下次备份时刻（每天 BACKUP_HOUR 点）的秒数
seconds_until_next() {
    local target now
    target=$(date -d "today ${BACKUP_HOUR}:00" +%s)
    now=$(date +%s)
    if [ "$target" -le "$now" ]; then
        target=$((target + 86400))
    fi
    echo $((target - now))
}

mkdir -p "$BACKUP_DIR"

# 手动模式：立即备份一次后退出
if [ "${1:-}" = "now" ]; then
    do_backup
    exit $?
fi

# 常驻模式：每天固定时刻备份
log "已启动，每天 ${BACKUP_HOUR}:00 备份 ${DB_NAME} 库到 ${BACKUP_DIR}（保留 ${KEEP} 份）"
while true; do
    sleep "$(seconds_until_next)"
    do_backup
done

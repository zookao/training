#!/bin/bash
set -e

# =============================================================================
# 容器入口脚本：首次启动初始化 MariaDB 数据目录 + 设置 root 密码
# 后续启动直接交由 supervisord 管理
# =============================================================================

MYSQL_DATA_DIR="/app/db"
MYSQL_RUN_DIR="/run/mysqld"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"

# 确保 socket 目录 + 数据目录存在且权限正确（volume 首次挂载可能是 root:root）
mkdir -p "$MYSQL_RUN_DIR" "$MYSQL_DATA_DIR"
chown mysql:mysql "$MYSQL_RUN_DIR" "$MYSQL_DATA_DIR"

# ---- 首次启动：数据目录为空，初始化 MariaDB ----
if [ ! -d "$MYSQL_DATA_DIR/mysql" ]; then
    echo "[entrypoint] 首次启动，初始化 MariaDB 数据目录..."

    # 初始化数据目录（mysql_install_db 是 mariadb-install-db 的兼容别名）
    mysql_install_db --user=mysql --datadir="$MYSQL_DATA_DIR" >/tmp/mysql_install.log 2>&1

    # 启动临时 mysqld（仅本地 socket，不监听网络端口）
    echo "[entrypoint] 启动临时 mysqld 设置 root 密码..."
    mysqld --skip-networking --user=mysql --datadir="$MYSQL_DATA_DIR" &
    temp_pid=$!

    # 等待 mysqld 就绪（最多 60 秒）
    for i in $(seq 1 60); do
        if mysqladmin ping >/dev/null 2>&1; then
            break
        fi
        if [ "$i" -eq 60 ]; then
            echo "[entrypoint] MariaDB 启动超时，查看日志:"
            cat /tmp/mysql_install.log
            exit 1
        fi
        sleep 1
    done

    # 设置 root 密码 + 安全清理
    mysql <<-EOSQL
        ALTER USER 'root'@'localhost' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}';
        DELETE FROM mysql.user WHERE User='';
        DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost','127.0.0.1','::1');
        DROP DATABASE IF EXISTS test;
        FLUSH PRIVILEGES;
EOSQL

    echo "[entrypoint] root 密码已设置"

    # 关闭临时 mysqld
    mysqladmin -uroot -p"$MYSQL_ROOT_PASSWORD" shutdown
    wait "$temp_pid" 2>/dev/null || true
fi

# ---- 启动 supervisord（管理 mysqld + training + nginx）----
echo "[entrypoint] 启动 supervisord..."
exec supervisord -c /etc/supervisor/supervisord.conf

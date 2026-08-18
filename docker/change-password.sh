#!/bin/bash
set -e

# =============================================================================
# 在线修改数据库密码（不丢数据）
#
# 原理: ALTER USER 在线改密码 → 重建容器传新 MYSQL_PASSWORD
#       数据卷 training_db 保留，数据不丢
#
# 用法:
#   ./change-password.sh <新密码>
#   MYSQL_PASSWORD=旧密码 ./change-password.sh <新密码>
#
# 注意: 修改 JWT_SIGNING_KEY 会使已登录 token 失效（需用户重新登录）
#       本脚本只改数据库密码，不改 JWT 密钥
# =============================================================================

CONTAINER_NAME="training"
NEW_PASSWORD="$1"

if [ -z "$NEW_PASSWORD" ]; then
    echo "用法: $0 <新密码>"
    echo "示例: $0 MyNewStrong@2026"
    exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "错误: 容器 ${CONTAINER_NAME} 未运行"
    exit 1
fi

# 旧密码：优先从环境变量读取，否则使用 run.sh 的默认值
OLD_PASSWORD="${MYSQL_PASSWORD:-root}"

echo "[change-password] 在线修改 MariaDB root 密码..."
docker exec "$CONTAINER_NAME" mysql -uroot -p"$OLD_PASSWORD" -e \
    "ALTER USER 'root'@'localhost' IDENTIFIED BY '${NEW_PASSWORD}'; FLUSH PRIVILEGES;"

echo "[change-password] 密码已更新，重建容器应用新配置（数据卷保留，不丢数据）..."
docker stop "$CONTAINER_NAME"
docker rm "$CONTAINER_NAME"

# 持久化新密码到 .db.env，下次 run.sh 重启时自动读取（避免用错旧默认密码）
ENV_FILE="$(dirname "$0")/.db.env"
if [ -f "$ENV_FILE" ]; then
    sed -i "s|^MYSQL_PASSWORD=.*|MYSQL_PASSWORD='${NEW_PASSWORD}'|" "$ENV_FILE"
else
    echo "MYSQL_PASSWORD='${NEW_PASSWORD}'" > "$ENV_FILE"
fi
chmod 600 "$ENV_FILE"

# 用新密码重新启动（run.sh 也会从 .db.env 读到新密码）
MYSQL_PASSWORD="$NEW_PASSWORD" "$(dirname "$0")/run.sh"

echo "[change-password] 完成！新密码: $NEW_PASSWORD"

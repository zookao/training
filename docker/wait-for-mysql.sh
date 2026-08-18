#!/bin/bash
# =============================================================================
# 等待 MariaDB 就绪后启动 Go 后端
# 由 supervisord 调用，exec 替换进程以便 supervisord 直接管理 training 的 PID
# =============================================================================

MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"

echo "[wait-for-mysql] 等待 MariaDB 就绪..."
for i in $(seq 1 60); do
    if mysql -h 127.0.0.1 -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SELECT 1" >/dev/null 2>&1; then
        echo "[wait-for-mysql] MariaDB 已就绪，启动后端..."
        cd /app
        exec ./training
    fi
    sleep 1
done

echo "[wait-for-mysql] MariaDB 60 秒内未就绪，退出"
exit 1

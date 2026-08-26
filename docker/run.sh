#!/bin/bash
set -e

# =============================================================================
# 一键部署脚本（在服务器执行）
# 自动加载镜像（如未加载）→ 创建数据卷 → 启动容器
# 架构：传参 amd64/arm64 则用参数，不传则跟随宿主机架构
#
# 用法:
#   ./run.sh                          # 架构跟随宿主机，使用默认密码启动
#   ./run.sh amd64                   # 指定 amd64
#   ./run.sh arm64                   # 指定 arm64
#   MYSQL_PASSWORD=YourPass ./run.sh  # 指定数据库密码（架构跟随宿主机）
#   JWT_SIGNING_KEY=xxx ./run.sh      # 指定 JWT 密钥
# =============================================================================

# 架构：第一个参数为 amd64/arm64 则使用，否则按宿主机架构
ARCH=""
case "$1" in
    amd64|arm64) ARCH="$1" ;;
esac
if [ -z "$ARCH" ]; then
    case "$(uname -m)" in
        x86_64|amd64) ARCH=amd64 ;;
        aarch64|arm64) ARCH=arm64 ;;
        *) echo "[run] 不支持的架构: $(uname -m)"; exit 1 ;;
    esac
fi
IMAGE_NAME="training:${ARCH}"
CONTAINER_NAME="training"
TARBALL="training-${ARCH}.tar.gz"

ENV_FILE="$(dirname "$0")/.db.env"

# 读取密码：命令行/环境变量优先 → .db.env 持久化文件（改密码后自动保存）→ 默认值
# 这样改过密码后重启容器，无需手动传参，run.sh 自动用上次的新密码
load_env() {
    local key="$1" default="$2"
    if [ -n "${!key}" ]; then echo "${!key}"; return; fi          # 1. 环境变量优先
    if [ -f "$ENV_FILE" ]; then                                    # 2. 持久化文件
        local val
        val=$(grep -E "^${key}=" "$ENV_FILE" | tail -1 | cut -d= -f2- | tr -d "'\"")
        if [ -n "$val" ]; then echo "$val"; return; fi
    fi
    echo "$default"                                                # 3. 默认值
}
MYSQL_PASSWORD=$(load_env MYSQL_PASSWORD root)
JWT_SIGNING_KEY=$(load_env JWT_SIGNING_KEY training-secret-key-2026)

# ---- 自动加载镜像 ----
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
    # 在常见位置查找 tarball
    for f in "../${TARBALL}" "./${TARBALL}" "../${TARBALL}"; do
        if [ -f "$f" ]; then
            echo "[run] 加载镜像: $f"
            docker load < "$f"
            break
        fi
    done
    if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
        echo "[run] 错误: 镜像 ${IMAGE_NAME} 不存在"
        echo "[run] 请先执行: docker load < ${TARBALL}"
        exit 1
    fi
fi

# ---- 停止并删除旧容器 ----
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[run] 停止并删除旧容器..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
fi

# ---- 启动容器 ----
echo "[run] 启动容器..."
docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    -p 80:80 -p 81:81 \
    -e MYSQL_ROOT_PASSWORD="$MYSQL_PASSWORD" \
    -e MYSQL_PASSWORD="$MYSQL_PASSWORD" \
    -e JWT_SIGNING_KEY="$JWT_SIGNING_KEY" \
    -v training_upload:/app/upload \
    -v training_db:/app/db \
    -v training_log:/var/log/supervisor \
    "$IMAGE_NAME"

# 持久化当前密码到 .db.env，下次重启自动读取（避免改密码后重启用错默认密码）
cat > "$ENV_FILE" <<EOF
MYSQL_PASSWORD='${MYSQL_PASSWORD}'
JWT_SIGNING_KEY='${JWT_SIGNING_KEY}'
EOF
chmod 600 "$ENV_FILE"

echo ""
echo "========================================"
echo "  部署完成！"
echo "========================================"
echo "  管理后台:  http://<服务器IP>:80"
echo "  学员前台:  http://<服务器IP>:81"
echo "  默认账号:  admin / admin123"
echo "  数据库密码: $MYSQL_PASSWORD"
echo "========================================"
echo ""
echo "常用命令:"
echo "  查看日志:   docker logs -f training"
echo "  服务日志:   docker exec training tail -f /var/log/supervisor/training.log"
echo "  日志卷:     training_log → /var/log/supervisor（宿主机可直接查看）"
echo "  查看服务状态: docker exec training supervisorctl status"
echo "  停止:       docker stop training"
echo "  重启:       docker restart training"
echo "  修改数据库密码:   ./change-password.sh <新密码>"

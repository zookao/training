#!/bin/bash
set -e

# =============================================================================
# 构建脚本（在开发机执行）
# 构建 Docker 镜像并导出为 tar.gz，供离线拷贝到目标服务器
#
# 用法:
#   ./docker/build.sh              # 默认 arm64 构建并导出
#   ./docker/build.sh amd64         # 构建 amd64（x86_64，如 CentOS/Intel Xeon）
#   ./docker/build.sh arm64        # 构建 arm64（银河麒麟 V10）
#   ./docker/build.sh amd64 --no-export  # 仅构建不导出
# =============================================================================

cd "$(dirname "$0")/.."

ARCH=arm64
EXPORT=true
for arg in "$@"; do
    case "$arg" in
        amd64|arm64) ARCH="$arg" ;;
        --no-export) EXPORT=false ;;
    esac
done

IMAGE_NAME="training:${ARCH}"
TARBALL="training-${ARCH}.tar.gz"

# 检查 buildx 是否可用
if ! docker buildx version >/dev/null 2>&1; then
    echo "[build] docker buildx 不可用，请升级 Docker 或安装 buildx 插件"
    exit 1
fi

echo "[build] 构建 ${IMAGE_NAME} (linux/${ARCH})..."
echo "[build]   前端: node:20 原生构建"
echo "[build]   后端: golang:1.26 交叉编译 GOARCH=${ARCH}"
echo "[build]   运行时: ubuntu:22.04 ${ARCH}"
echo ""

docker buildx build --platform "linux/${ARCH}" --load \
    -t "$IMAGE_NAME" -f docker/Dockerfile .

echo ""
echo "[build] 镜像构建完成: ${IMAGE_NAME}"

if [ "$EXPORT" = true ]; then
    echo "[build] 导出镜像到 ${TARBALL}..."
    docker save "$IMAGE_NAME" | gzip > "$TARBALL"
    SIZE=$(du -h "$TARBALL" | cut -f1)
    echo "[build] 导出完成: ${TARBALL} (${SIZE})"
fi

echo ""
echo "========================================"
echo "  下一步："
echo "========================================"
echo "  1. 将 ${TARBALL} 和 docker/ 目录拷贝到服务器"
echo "  2. 在服务器执行:"
echo "       docker load < ${TARBALL}"
echo "       cd docker && ./run.sh"
echo "========================================"

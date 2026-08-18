#!/bin/bash
set -e

# =============================================================================
# 构建脚本（在开发机执行）
# 构建 arm64 Docker 镜像并导出为 tar.gz，供离线拷贝到银河麒麟 V10 服务器
#
# 用法:
#   ./docker/build.sh           # 构建并导出
#   ./docker/build.sh --no-export  # 仅构建不导出
# =============================================================================

cd "$(dirname "$0")/.."

IMAGE_NAME="training:arm64"
TARBALL="training-arm64.tar.gz"
EXPORT=true

if [ "$1" = "--no-export" ]; then
    EXPORT=false
fi

# 检查 buildx 是否可用
if ! docker buildx version >/dev/null 2>&1; then
    echo "[build] docker buildx 不可用，请升级 Docker 或安装 buildx 插件"
    exit 1
fi

echo "[build] 构建 ${IMAGE_NAME} (linux/arm64)..."
echo "[build]   前端: node:20 原生构建"
echo "[build]   后端: golang:1.26 交叉编译 GOARCH=arm64"
echo "[build]   运行时: ubuntu:22.04 arm64"
echo ""

docker buildx build --platform linux/arm64 --load \
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

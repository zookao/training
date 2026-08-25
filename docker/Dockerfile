# =============================================================================
# 培训系统 Docker 离线部署镜像（目标平台：linux/arm64，银河麒麟 V10）
# 多阶段构建：前端(宿主平台原生) → 后端(交叉编译) → 运行时(arm64)
#
# 构建命令（在开发机执行）:
#   docker buildx build --platform linux/arm64 --load \
#     -t training:arm64 -f docker/Dockerfile .
# =============================================================================

# ===== Stage 1: 构建前端（在宿主平台执行，避免 QEMU 模拟拖慢 npm） =====
FROM --platform=$BUILDPLATFORM node:20-bookworm AS frontend-builder
WORKDIR /build

# 先拷 package.json 利用 Docker layer 缓存
COPY back/package.json back/package-lock.json ./back/
COPY front/package.json front/package-lock.json ./front/

# 安装依赖（两个前端项目并行）
RUN cd back && npm ci && cd ../front && npm ci

# 拷贝源码并构建
COPY back/ ./back/
COPY front/ ./front/
RUN cd back && npm run build && cd ../front && npm run build

# ===== Stage 2: 构建后端（在宿主平台执行，Go 交叉编译到 arm64） =====
FROM --platform=$BUILDPLATFORM golang:1.26 AS backend-builder
ARG TARGETARCH
WORKDIR /build

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./
# CGO_ENABLED=0 纯静态二进制，GOARCH 交叉编译到目标架构（无需 QEMU，原生速度）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o training .

# ===== Stage 3: 运行时（默认目标平台 arm64，全部依赖构建时装入，服务器 load 后无需联网） =====
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

# 安装运行时依赖：MariaDB + Nginx + ffmpeg + LibreOffice(headless) + CJK 字体 + supervisor
RUN tee /etc/apt/sources.list <<'EOF'
deb http://mirrors.aliyun.com/ubuntu-ports/ jammy main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu-ports/ jammy-updates main restricted universe multiverse
EOF
RUN apt-get update && apt-get install -y --no-install-recommends \
    mariadb-server \
    nginx \
    ffmpeg \
    libreoffice \
    fonts-noto-cjk \
    fonts-wqy-zenhei \
    supervisor \
    logrotate \
    tzdata \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 容器时区设为北京时间（默认 UTC 会影响日志时间戳与日志轮转午夜计算）
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

# 创建应用目录（MariaDB 数据目录 /app/db 由 volume 持久化）
RUN mkdir -p /app/upload /app/.chunks /app/db /var/log/supervisor /run/mysqld \
    && chown mysql:mysql /app/db /run/mysqld

# 拷贝后端二进制 + 配置
COPY --from=backend-builder /build/training /app/training
COPY docker/config.yaml /app/config.yaml
RUN chmod +x /app/training

# 拷贝前端构建产物
COPY --from=frontend-builder /build/back/dist /app/back-dist
COPY --from=frontend-builder /build/front/dist /app/front-dist

# Nginx 配置（替换默认站点）
COPY docker/nginx.conf /etc/nginx/conf.d/training.conf
RUN rm -f /etc/nginx/sites-enabled/default

# supervisord 配置
COPY docker/supervisord.conf /etc/supervisor/supervisord.conf

# 日志按天轮转配置 + 轮转脚本
COPY docker/logrotate-supervisor /etc/logrotate.d/supervisor
COPY docker/logrotate.sh /app/logrotate.sh

# 启动脚本
COPY docker/entrypoint.sh /app/entrypoint.sh
COPY docker/wait-for-mysql.sh /app/wait-for-mysql.sh
RUN chmod +x /app/entrypoint.sh /app/wait-for-mysql.sh /app/logrotate.sh

EXPOSE 80 81

HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
    CMD curl -sf http://localhost:8000/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]

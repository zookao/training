# Docker 离线部署（银河麒麟 V10 / ARM64）

单容器 all-in-one 部署方案：MariaDB + Go 后端 + Nginx + ffmpeg + LibreOffice + CJK 字体，全部装入一个 Docker 镜像，服务器 `docker load` 后一条命令启动。

## 架构

```
┌─────────────────── Docker 容器 (linux/arm64) ───────────────────┐
│                                                                  │
│  Nginx:80  ──┐         ┌── Nginx:81 (学员前台 front/dist)       │
│  (管理后台    │  /api   │                                          │
│   back/dist) └─反代──→ Go:8000 ←──→ MariaDB:3306                  │
│               /upload                                            │
│                                                                  │
│  supervisord 托管: mysqld(10) → training(20) → nginx(30)         │
│                                                                  │
│  持久化卷: training_upload→/app/upload                           │
│           training_db→/app/db                                    │
└──────────────────────────────────────────────────────────────────┘
```

## 文件说明

| 文件 | 用途 |
|------|------|
| `Dockerfile` | 多阶段构建：node:20 构建前端 → golang:1.26 交叉编译 → ubuntu:22.04 运行时 |
| `supervisord.conf` | 进程管理：mysqld + training + nginx |
| `entrypoint.sh` | 容器入口：首次启动初始化 MariaDB + 设置 root 密码 |
| `wait-for-mysql.sh` | 等 MariaDB 就绪后启动 Go 后端 |
| `nginx.conf` | Nginx 配置：端口 80 管理后台 + 端口 81 学员前台 |
| `config.yaml` | 后端运行时配置（敏感字段由环境变量覆盖） |
| `build.sh` | **开发机执行**：构建镜像 + 导出 tar.gz |
| `run.sh` | **服务器执行**：一键启动容器 |
| `change-password.sh` | **服务器执行**：在线修改数据库密码（不丢数据） |

---

## 部署步骤

### 第一步：在开发机（Mac）构建镜像

```bash
# 进入项目根目录
cd /path/to/training

# 构建并导出（M 系列 Mac 原生快，Intel Mac 走 QEMU 较慢）
./docker/build.sh
```

构建完成后生成 `training-arm64.tar.gz`（约 1-2 GB）。

> **Intel Mac 加速提示**：Dockerfile 已用 `--platform=$BUILDPLATFORM` 让前端和 Go 在宿主原生平台构建，仅运行时阶段的 `apt-get` 走 QEMU，比全 QEMU 快很多。

### 第二步：拷贝到服务器

将以下文件拷贝到银河麒麟 V10 服务器同一目录：

```
training-arm64.tar.gz
docker/run.sh
docker/change-password.sh
```

### 第三步：服务器一键启动

```bash
# 加载镜像
docker load < training-arm64.tar.gz

# 启动（使用默认密码，首次部署建议改强密码）
cd docker && ./run.sh

# 或指定密码
MYSQL_PASSWORD='YourStrong@Pass' ./run.sh
```

### 完成

- 管理后台：`http://<服务器IP>:80` （admin / admin123）
- 学员前台：`http://<服务器IP>:81`

---

## 麒麟 V10 离线安装 Docker

如果服务器没有 Docker，在另一台**同架构 (aarch64) 联网**的麒麟机上下载 rpm 包：

```bash
# 在联网的麒麟机器上下载（不安装）
sudo dnf install --downloadonly --downloaddir=./docker-rpms \
    docker-ce docker-ce-cli containerd.io

# 拷到离线服务器后安装
sudo dnf localinstall -y ./docker-rpms/*.rpm
sudo systemctl enable --now docker
```

---

## 运维命令

```bash
# 查看容器日志
docker logs -f training

# 查看各服务状态
docker exec training supervisorctl status

# 查看后端日志
docker exec training tail -f /var/log/supervisor/training.log

# 重启单个服务
docker exec training supervisorctl restart training

# 停止 / 重启容器
docker stop training
docker restart training
```

## 修改数据库密码（不丢数据）

```bash
cd docker
./change-password.sh 'NewStrong@Pass'
```

脚本会：`ALTER USER` 在线改密码 → 重建容器传新密码 → 数据卷保留不丢。

## 数据持久化

| Docker Volume | 容器路径 | 用途 |
|---------------|----------|------|
| `training_upload` | `/app/upload` | 上传的视频、课件、图片 |
| `training_db` | `/app/db` | MariaDB 数据库文件 |
| `training_log` | `/var/log/supervisor` | supervisord 各进程日志（mysqld/training/nginx） |

删除容器重建数据不丢。**删除 volume 才会丢数据**：
```bash
# 危险！会删除所有数据
docker volume rm training_upload training_db
```

## 备份与迁移

```bash
# 备份数据库
docker exec training mysqldump -uroot -p"$MYSQL_PASSWORD" training > backup.sql

# 备份上传文件
docker run --rm -v training_upload:/data -v "$PWD":/backup alpine tar czf /backup/upload.tar.gz -C /data .

# 恢复到新服务器：先 run.sh 启动，再恢复
docker exec -i training mysql -uroot -p"$MYSQL_PASSWORD" training < backup.sql
docker run --rm -v training_upload:/data -v "$PWD":/backup alpine tar xzf /backup/upload.tar.gz -C /data
```

## 注意事项

1. **密码一致性**：`MYSQL_ROOT_PASSWORD`（初始化 MariaDB）必须与 `MYSQL_PASSWORD`（覆盖后端 config）相同，`run.sh` 已自动处理。
2. **首次启动较慢**：MariaDB 初始化 + Go 后端自动建库建表 + 种子数据，约 10-30 秒。
3. **改 JWT 密钥**：会使已登录用户 token 失效，需重新登录。本脚本默认不改 JWT 密钥。
4. **端口冲突**：如 80/81 被占用，修改 `run.sh` 中的 `-p` 参数。
5. **镜像不可直接 `docker build`**：需用 `docker buildx build --platform linux/arm64`（build.sh 已封装）。

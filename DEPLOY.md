# 安装mysql
# 安装nginx
# 安装libreoffice（如果安装的机器上字体不全，需要安装字体包）
```
apt install libreoffice
或
sudo yum install -y libreoffice-headless libreoffice-writer libreoffice-calc libreoffice-impress

soffice --version

sudo apt-get install -y fonts-noto-cjk fonts-wqy-zenhei fonts-wqy-microhei
或
sudo dnf install -y google-noto-sans-cjk-ttc-fonts google-noto-serif-cjk-ttc-fonts
sudo fc-cache -fv
```
# 安装ffmpeg（视频时长获取 + 封面自动截取，启动时硬性检测）
```
apt install ffmpeg        # Debian/Ubuntu
或
yum install ffmpeg -y     # CentOS/RHEL（需 epel 或 rpmfusion）
ffmpeg -version           # 验证
ffprobe -version          # ffprobe 需同时存在（通常一起安装）
```
> ffmpeg 和 ffprobe 也可以用便携版：在 config.yaml 中配置 `ffmpeg.path` 指向 ffmpeg 二进制路径，ffprobe 需在同一目录。
# 构建后端二进制文件
```
cd api
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o training
```
# 构建管理后台
```
cd back
npm run build
```
# 构建学员前台
```
cd front
npm run build
```
# 后端配置文件（配置好数据库连接信息）
```
cp config.yaml 二进制文件所在目录
cd 到二进制文件所在目录 && ./training
``` 
8、nginx
参考docker的nginx.conf
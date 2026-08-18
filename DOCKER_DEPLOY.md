# 关闭slinux
```
sudo setenforce 0

```

# 备份podman的扩展（和docker冲突，如果podman已安装，需要这一步骤）
```
mv /usr/bin/runc /usr/bin/runc.back
```

# 安装docker
```
tar -zxf docker-20.10.24.tgz
mkdir -p /opt/docker/bin
cp docker/* /usr/bin/
docker version
```

# 配置docker的service
```
`vim /etc/systemd/system/docker.service`

```
[Unit]
Description=Docker Application Container Engine
After=network-online.target
Wants=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=notify
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
Restart=always
LimitNOFILE=infinity
LimitNPROC=infinity
```
自启动
```
systemctl daemon-reload
systemctl enable --now docker
docker -v
```

# 启动镜像
```
cd /Users/zookao/goCode/self/training
docker load < training-arm64.tar.gz
chmod +x run.sh && ./run.sh
docker logs -f training
```

# 开放端口
```
sudo firewall-cmd --zone=public --list-ports
sudo firewall-cmd --permanent --zone=public --add-port=80/tcp
sudo firewall-cmd --permanent --zone=public --add-port=81/tcp
sudo firewall-cmd --reload
```

# 修改数据库密码
./change-password.sh 'MyNewStrongPwd@2026'
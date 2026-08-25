#!/bin/bash
# 每天北京时间午夜轮转 supervisor 日志（由 supervisord 拉起，崩溃自动重启）
# logrotate 用 copytruncate 清空原文件，supervisor 无需重开文件句柄、进程无需重启
set -e
while true; do
    now=$(TZ=Asia/Shanghai date +%s)
    midnight=$(TZ=Asia/Shanghai date -d "tomorrow 00:00:00" +%s)
    sleep $((midnight - now))
    /usr/sbin/logrotate -s /tmp/logrotate-supervisor.state /etc/logrotate.d/supervisor
done

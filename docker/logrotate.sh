#!/bin/bash
# 按大小轮转 supervisor 日志（单文件超 70M 触发），每分钟检查一次
# logrotate 用 copytruncate 清空原文件，supervisor 无需重开文件句柄、进程无需重启
set -e
while true; do
    /usr/sbin/logrotate -s /tmp/logrotate-supervisor.state /etc/logrotate.d/supervisor
    sleep 60
done

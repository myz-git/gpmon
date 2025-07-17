#!/bin/bash

# GPMon 日志轮转设置脚本
# 用法: sudo ./setup-logrotate.sh

set -e

echo "Setting up logrotate for GPMon..."

# 检查是否以root权限运行
if [[ $EUID -ne 0 ]]; then
   echo "此脚本需要root权限运行，请使用 sudo ./setup-logrotate.sh"
   exit 1
fi

# 配置文件内容
LOGROTATE_CONFIG="/etc/logrotate.d/gpmon"

# 创建logrotate配置文件
cat > $LOGROTATE_CONFIG << 'EOF'
/workspace/gpmon/log/*.log {
    #daily                    # 每天轮转一次
   # missingok               # 如果日志文件不存在，不报错
    #rotate 7                # 保留7个轮转的日志文件
   # compress                # 压缩旧的日志文件
   # delaycompress          # 延迟压缩，下次轮转时才压缩
    #notifempty             # 空文件不轮转
    #create 0644 root root  # 创建新日志文件的权限和所有者
    #sharedscripts          # 多个日志文件共享脚本
    #copytruncate           # 复制后截断原文件，适用于持续写入的程序
    #maxage 30              # 删除30天前的日志文件
    #size 10M              # 当日志文件超过10M时立即轮转

    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 root root
    sharedscripts
    copytruncate
    maxage 30
    size 10M


    postrotate
        echo "Log rotation completed at $(date)" >> /var/log/gpmon-rotation.log
    endscript
}
EOF

# 设置正确的权限
chmod 644 $LOGROTATE_CONFIG

# 测试配置文件语法
echo "测试logrotate配置..."
logrotate -d $LOGROTATE_CONFIG

# 创建日志目录（如果不存在）
mkdir -p /workspace/gpmon/log

# 显示当前配置
echo ""
echo "配置已完成！"
echo "配置文件位置: $LOGROTATE_CONFIG"
echo ""
echo "配置说明:"
echo "- 每天轮转一次日志"
echo "- 保留7天的日志文件"
echo "- 文件超过100M时立即轮转"
echo "- 自动压缩旧日志文件"
echo "- 自动删除30天前的日志"
echo ""
echo "手动测试轮转: sudo logrotate -f $LOGROTATE_CONFIG"
echo "查看轮转状态: sudo logrotate -d $LOGROTATE_CONFIG"
echo "" 
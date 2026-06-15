#!/bin/bash

# GPMon 邮件工具编译脚本（兼容入口，实际调用 build.sh --mail）
# 用法: ./scripts/build-mail-tool.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TARGET_FILE="$PROJECT_ROOT/send_mail_cli"

"$SCRIPT_DIR/build.sh" --mail

echo ""
echo "测试编译结果..."
if "$TARGET_FILE" -help >/dev/null 2>&1; then
    echo "工具运行正常"
    echo ""
    echo "测试邮件配置..."
    "$TARGET_FILE" -test-config || true
else
    echo "警告: 工具编译成功但运行可能有问题"
fi

echo ""
echo "=== 邮件工具编译完成 ==="
echo "  测试邮件配置: ./send_mail_cli -test-config"
echo "  发送测试邮件: ./scripts/send-daily-report.sh --test-mail"
echo "  查看系统状态: ./scripts/db-maintenance.sh --status"
echo "  发送日报:     ./scripts/send-daily-report.sh --send"
echo ""
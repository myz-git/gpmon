#!/bin/bash

# GPMon 邮件工具编译脚本
# 用法: ./build-mail-tool.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SOURCE_FILE="$PROJECT_ROOT/send_mail_cli.go"
TARGET_FILE="$PROJECT_ROOT/send_mail_cli"

echo "编译 GPMon 邮件发送工具..."

# 检查Go是否安装
if ! command -v go >/dev/null 2>&1; then
    echo "❌ 错误: 未找到Go编译器"
    echo "请先安装Go语言环境："
    echo "  https://golang.org/dl/"
    exit 1
fi

# 检查源文件是否存在
if [ ! -f "$SOURCE_FILE" ]; then
    echo "❌ 错误: 找不到源文件: $SOURCE_FILE"
    exit 1
fi

echo "Go版本: $(go version)"
echo "源文件: $SOURCE_FILE"
echo "目标文件: $TARGET_FILE"

# 切换到项目根目录进行编译
cd "$PROJECT_ROOT"

# 设置Go模块环境
export GO111MODULE=on

# 检查依赖
echo ""
echo "检查依赖..."
go mod tidy 2>/dev/null || echo "警告: go mod tidy 失败，可能需要初始化模块"

# 编译
echo ""
echo "开始编译..."
if go build -o "$TARGET_FILE" "$SOURCE_FILE"; then
    echo "✅ 编译成功!"
    echo ""
    echo "可执行文件: $TARGET_FILE"
    echo "文件大小: $(du -h "$TARGET_FILE" | cut -f1)"
    
    # 设置执行权限
    chmod +x "$TARGET_FILE"
    echo "已设置执行权限"
    
    # 测试编译结果
    echo ""
    echo "测试编译结果..."
    if "$TARGET_FILE" -help >/dev/null 2>&1; then
        echo "✅ 工具运行正常"
        
        # 测试邮件配置
        echo ""
        echo "测试邮件配置..."
        "$TARGET_FILE" -test-config
        
    else
        echo "⚠️  工具编译成功但运行可能有问题"
    fi
    
else
    echo "❌ 编译失败"
    exit 1
fi

echo ""
echo "=== 编译完成 ==="
echo ""
echo "现在您可以使用以下命令:"
echo "  测试邮件配置: ./send_mail_cli -test-config"
echo "  发送测试邮件: ./scripts/gpmon-maintenance.sh --test-mail"
echo "  预览报告: ./scripts/gpmon-maintenance.sh --status"
echo "  发送日报: ./scripts/gpmon-maintenance.sh --send-report"
echo ""
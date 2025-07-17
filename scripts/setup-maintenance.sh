#!/bin/bash

# GPMon 运维环境设置脚本
# 功能：一键设置定时任务和运维环境
# 用法: sudo ./setup-maintenance.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 显示帮助信息
show_help() {
    echo "GPMon 运维环境设置脚本"
    echo ""
    echo "用法: sudo $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --setup              设置所有运维环境"
    echo "  --setup-cron         仅设置定时任务"
    echo "  --setup-logs         仅设置日志管理"
    echo "  --remove             移除所有配置"
    echo "  --status             查看配置状态"
    echo "  -h, --help           显示此帮助信息"
    echo ""
    echo "说明:"
    echo "  此脚本需要root权限运行"
    echo "  会设置以下组件:"
    echo "  - 定时任务（邮件日报和数据库维护）"
    echo "  - 日志轮转配置"
    echo "  - 系统日志目录"
    echo ""
}

# 检查root权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo "❌ 此脚本需要root权限运行"
        echo "请使用: sudo $0 $*"
        exit 1
    fi
}

# 设置定时任务
setup_cron() {
    echo "=== 设置定时任务 ==="
    
    local cron_config="/etc/cron.d/gpmon-maintenance"
    
    cat > "$cron_config" << EOF
# GPMon 运维定时任务
# 
# 每天早上8点: 发送监控日报
0 8 * * * root $SCRIPT_DIR/send-daily-report.sh --send >/dev/null 2>&1

# 每天凌晨2点: 执行数据库维护
0 2 * * * root $SCRIPT_DIR/db-maintenance.sh --full --force >/dev/null 2>&1
EOF

    chmod 644 "$cron_config"
    echo "✅ 定时任务配置完成: $cron_config"
    
    # 重启cron服务
    if command -v systemctl >/dev/null 2>&1; then
        systemctl reload crond 2>/dev/null || systemctl reload cron 2>/dev/null || true
        echo "✅ cron服务已重载"
    elif command -v service >/dev/null 2>&1; then
        service crond reload 2>/dev/null || service cron reload 2>/dev/null || true
        echo "✅ cron服务已重载"
    fi
}

# 设置日志管理
setup_logs() {
    echo "=== 设置日志管理 ==="
    
    # 创建日志目录
    mkdir -p /var/log
    mkdir -p "$PROJECT_ROOT/log"
    
    # 设置日志文件权限
    touch /var/log/gpmon-maintenance.log
    touch /var/log/gpmon-daily-report.log
    chmod 644 /var/log/gpmon-maintenance.log
    chmod 644 /var/log/gpmon-daily-report.log
    
    echo "✅ 日志目录和文件已创建"
    echo "  维护日志: /var/log/gpmon-maintenance.log"
    echo "  报告日志: /var/log/gpmon-daily-report.log"
    echo "  应用日志: $PROJECT_ROOT/log/"
}

# 移除配置
remove_config() {
    echo "=== 移除配置 ==="
    
    # 移除定时任务
    if [ -f "/etc/cron.d/gpmon-maintenance" ]; then
        rm -f /etc/cron.d/gpmon-maintenance
        echo "✅ 已删除定时任务配置"
    fi
    
    # 移除logrotate配置
    if [ -f "/etc/logrotate.d/gpmon" ]; then
        rm -f /etc/logrotate.d/gpmon
        echo "✅ 已删除logrotate配置"
    fi
    
    # 重启cron服务
    if command -v systemctl >/dev/null 2>&1; then
        systemctl reload crond 2>/dev/null || systemctl reload cron 2>/dev/null || true
    elif command -v service >/dev/null 2>&1; then
        service crond reload 2>/dev/null || service cron reload 2>/dev/null || true
    fi
    
    echo "✅ 配置移除完成"
}

# 查看配置状态
show_status() {
    echo "=== 配置状态 ==="
    echo ""
    
    # 检查定时任务
    echo "⏰ 定时任务:"
    if [ -f "/etc/cron.d/gpmon-maintenance" ]; then
        echo "  状态: ✅ 已配置"
        echo "  配置文件: /etc/cron.d/gpmon-maintenance"
        echo "  内容:"
        sed 's/^/    /' /etc/cron.d/gpmon-maintenance
    else
        echo "  状态: ❌ 未配置"
    fi
    
    echo ""
    
    # 检查logrotate
    echo "📄 日志轮转:"
    if [ -f "/etc/logrotate.d/gpmon" ]; then
        echo "  状态: ✅ 已配置"
        echo "  配置文件: /etc/logrotate.d/gpmon"
    else
        echo "  状态: ❌ 未配置"
        echo "  建议运行: sudo $SCRIPT_DIR/setup-logrotate.sh"
    fi
    
    echo ""
    
    # 检查脚本文件
    echo "📝 脚本文件:"
    local scripts=("send-daily-report.sh" "db-maintenance.sh" "build-mail-tool.sh" "setup-logrotate.sh")
    for script in "${scripts[@]}"; do
        if [ -f "$SCRIPT_DIR/$script" ]; then
            if [ -x "$SCRIPT_DIR/$script" ]; then
                echo "  $script: ✅ 存在且可执行"
            else
                echo "  $script: ⚠️  存在但不可执行"
            fi
        else
            echo "  $script: ❌ 不存在"
        fi
    done
    
    echo ""
    
    # 检查邮件工具
    echo "📧 邮件工具:"
    if [ -f "$PROJECT_ROOT/send_mail_cli" ]; then
        if [ -x "$PROJECT_ROOT/send_mail_cli" ]; then
            echo "  状态: ✅ 已编译且可执行"
            echo "  路径: $PROJECT_ROOT/send_mail_cli"
        else
            echo "  状态: ⚠️  已编译但不可执行"
        fi
    else
        echo "  状态: ❌ 未编译"
        echo "  建议运行: $SCRIPT_DIR/build-mail-tool.sh"
    fi
    
    echo ""
    
    # 检查日志文件
    echo "📋 日志文件:"
    local log_files=("/var/log/gpmon-maintenance.log" "/var/log/gpmon-daily-report.log")
    for log_file in "${log_files[@]}"; do
        if [ -f "$log_file" ]; then
            local size=$(du -h "$log_file" | cut -f1)
            echo "  $(basename "$log_file"): ✅ 存在 ($size)"
        else
            echo "  $(basename "$log_file"): ❌ 不存在"
        fi
    done
}

# 完整设置
setup_all() {
    echo "=== 设置 GPMon 运维环境 ==="
    echo ""
    
    # 设置脚本执行权限
    chmod +x "$SCRIPT_DIR"/*.sh
    echo "✅ 脚本执行权限已设置"
    
    # 设置定时任务
    setup_cron
    
    echo ""
    
    # 设置日志管理
    setup_logs
    
    echo ""
    echo "🎉 GPMon 运维环境设置完成！"
    echo ""
    echo "定时任务:"
    echo "  - 每天早上8点: 自动发送监控日报"
    echo "  - 每天凌晨2点: 自动执行数据库维护"
    echo ""
    echo "手动命令:"
    echo "  查看数据库状态: $SCRIPT_DIR/db-maintenance.sh --status"
    echo "  发送日报: $SCRIPT_DIR/send-daily-report.sh --send"
    echo "  预览报告: $SCRIPT_DIR/send-daily-report.sh --preview"
    echo "  执行维护: $SCRIPT_DIR/db-maintenance.sh --full"
    echo "  测试邮件: $SCRIPT_DIR/send-daily-report.sh --test-mail"
    echo ""
    echo "注意事项:"
    echo "  - 日志轮转需要单独设置: sudo $SCRIPT_DIR/setup-logrotate.sh"
    echo "  - 邮件工具需要先编译: $SCRIPT_DIR/build-mail-tool.sh"
    echo ""
    echo "日志文件:"
    echo "  - 维护日志: /var/log/gpmon-maintenance.log"
    echo "  - 报告日志: /var/log/gpmon-daily-report.log"
    echo "  - 应用日志: $PROJECT_ROOT/log/"
    echo ""
}

# 主程序
main() {
    local action=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --setup)
                action="setup"
                shift
                ;;
            --setup-cron)
                action="setup-cron"
                shift
                ;;
            --setup-logs)
                action="setup-logs"
                shift
                ;;
            --remove)
                action="remove"
                shift
                ;;
            --status)
                action="status"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定操作，默认完整设置
    if [ -z "$action" ]; then
        action="setup"
    fi
    
    # 检查权限（status操作除外）
    if [ "$action" != "status" ]; then
        check_root
    fi
    
    # 执行对应操作
    case "$action" in
        setup)
            setup_all
            ;;
        setup-cron)
            setup_cron
            ;;
        setup-logs)
            setup_logs
            ;;
        remove)
            remove_config
            ;;
        status)
            show_status
            ;;
        *)
            echo "错误: 未知操作 '$action'"
            exit 1
            ;;
    esac
}

# 运行主程序
main "$@"
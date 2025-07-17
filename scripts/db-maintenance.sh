#!/bin/bash

# GPMon 数据库维护脚本
# 功能：数据库清理、备份、优化等维护操作
# 用法: ./db-maintenance.sh [选项]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DB_PATH="$PROJECT_ROOT/messages.db"
BACKUP_DIR="$PROJECT_ROOT/backup"
MAINTENANCE_LOG="/var/log/gpmon-maintenance.log"

# 默认配置
DATA_RETENTION_DAYS=30    # 数据保留天数
BACKUP_RETENTION_DAYS=30  # 备份保留天数

# 显示帮助信息
show_help() {
    echo "GPMon 数据库维护脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --status             显示数据库状态"
    echo "  --clean              清理过期数据"
    echo "  --backup             备份数据库"
    echo "  --optimize           优化数据库"
    echo "  --full               完整维护(清理+备份+优化)"
    echo "  --list-backups       列出备份文件"
    echo "  --test-backups       测试备份完整性"
    echo "  --cleanup-backups    清理过期备份"
    echo "  --data-days <天数>   数据保留天数 (默认: 90天)"
    echo "  --backup-days <天数> 备份保留天数 (默认: 30天)"
    echo "  --force              强制执行，不询问确认"
    echo "  -h, --help           显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --status          # 查看数据库状态"
    echo "  $0 --full            # 执行完整维护"
    echo "  $0 --clean --force   # 强制清理数据"
    echo "  $0 --backup          # 仅备份数据库"
    echo ""
}

# 记录日志
log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message"
    echo "[$timestamp] $message" >> "$MAINTENANCE_LOG"
}

# 检查依赖
check_dependencies() {
    # 检查数据库文件
    if [ ! -f "$DB_PATH" ]; then
        echo "错误: 数据库文件不存在: $DB_PATH"
        exit 1
    fi
    
    # 检查sqlite3命令
    if ! command -v sqlite3 >/dev/null 2>&1; then
        echo "错误: sqlite3 命令未找到"
        exit 1
    fi
    
    # 检查gzip命令
    if ! command -v gzip >/dev/null 2>&1; then
        echo "错误: gzip 命令未找到"
        exit 1
    fi
    
    # 创建必要目录
    mkdir -p "$BACKUP_DIR" "$(dirname "$MAINTENANCE_LOG")"
}

# 显示数据库状态
show_status() {
    echo "=== 数据库状态 ==="
    echo ""
    
    # 数据库基本信息
    echo "📊 数据库信息:"
    echo "  路径: $DB_PATH"
    echo "  大小: $(du -h "$DB_PATH" | cut -f1)"
    
    # 历史记录统计
    local total_records=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his;" 2>/dev/null || echo "0")
    echo "  历史记录: $total_records 条"
    
    if [ "$total_records" -gt 0 ]; then
        local latest=$(sqlite3 "$DB_PATH" "SELECT MAX(chk_time) FROM check_his;" 2>/dev/null)
        echo "  最新记录: $latest"
        
        local last_7_days=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his WHERE chk_time >= datetime('now', '-7 days');" 2>/dev/null || echo "0")
        local last_30_days=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his WHERE chk_time >= datetime('now', '-30 days');" 2>/dev/null || echo "0")
        local old_records=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his WHERE chk_time < datetime('now', '-${DATA_RETENTION_DAYS} days');" 2>/dev/null || echo "0")
        
        echo "  最近7天: $last_7_days 条"
        echo "  最近30天: $last_30_days 条"
        echo "  可清理记录: $old_records 条 (${DATA_RETENTION_DAYS}天前)"
    fi
    
    echo ""
    
    # 当前监控状态
    echo "📈 当前监控状态:"
    local current_total=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result;" 2>/dev/null || echo "0")
    local current_ok=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='OK';" 2>/dev/null || echo "0")
    local current_error=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='ERROR';" 2>/dev/null || echo "0")
    local current_warning=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='WARNING';" 2>/dev/null || echo "0")
    
    echo "  总监控项: $current_total"
    echo "  正常: $current_ok"
    echo "  错误: $current_error"
    echo "  警告: $current_warning"
    
    if [ $current_total -gt 0 ]; then
        local ok_percent=$(awk "BEGIN {printf \"%.1f\", $current_ok*100/$current_total}")
        echo "  健康率: ${ok_percent}%"
    fi
    
    echo ""
    
    # 备份状态
    echo "💾 备份信息:"
    echo "  备份目录: $BACKUP_DIR"
    if [ -d "$BACKUP_DIR" ]; then
        local backup_count=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" 2>/dev/null | wc -l)
        echo "  备份数量: $backup_count"
        
        if [ $backup_count -gt 0 ]; then
            local backup_size=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)
            echo "  备份总大小: $backup_size"
            
            local latest_backup=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" -type f 2>/dev/null | sort | tail -1)
            if [ -n "$latest_backup" ]; then
                echo "  最新备份: $(basename "$latest_backup")"
                local backup_mtime=$(stat -c %y "$latest_backup" 2>/dev/null | cut -d. -f1)
                echo "  备份时间: $backup_mtime"
            fi
            
            local old_backups=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" -mtime +$BACKUP_RETENTION_DAYS -type f 2>/dev/null | wc -l)
            echo "  可清理备份: $old_backups 个 (${BACKUP_RETENTION_DAYS}天前)"
        fi
    else
        echo "  状态: 目录不存在"
    fi
    
    echo ""
    echo "配置参数:"
    echo "  数据保留: ${DATA_RETENTION_DAYS}天"
    echo "  备份保留: ${BACKUP_RETENTION_DAYS}天"
    echo "  维护日志: $MAINTENANCE_LOG"
}

# 清理过期数据
clean_database() {
    echo "=== 清理过期数据 ==="
    
    local delete_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his WHERE chk_time < datetime('now', '-${DATA_RETENTION_DAYS} days');" 2>/dev/null || echo "0")
    
    if [ "$delete_count" -eq 0 ]; then
        echo "没有需要清理的过期数据"
        return 0
    fi
    
    echo "将要删除 $delete_count 条记录 (${DATA_RETENTION_DAYS}天前的数据)"
    
    if [ "$FORCE_MODE" != "true" ]; then
        read -p "确认清理过期数据? (y/N): " -r confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            echo "操作已取消"
            return 0
        fi
    fi
    
    local start_time=$(date +%s)
    
    sqlite3 "$DB_PATH" "DELETE FROM check_his WHERE chk_time < datetime('now', '-${DATA_RETENTION_DAYS} days');"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $? -eq 0 ]; then
        echo "✅ 数据清理完成! 删除了 $delete_count 条记录，耗时: ${duration}秒"
        log_message "数据清理完成: 删除 $delete_count 条记录，耗时: ${duration}秒"
        
        local remaining_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his;" 2>/dev/null || echo "0")
        echo "当前剩余记录数: $remaining_count"
        return 0
    else
        echo "❌ 数据清理失败"
        log_message "错误: 数据清理失败"
        return 1
    fi
}

# 备份数据库
backup_database() {
    echo "=== 备份数据库 ==="
    
    mkdir -p "$BACKUP_DIR"
    
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="$BACKUP_DIR/messages_backup_${timestamp}.db"
    local compressed_file="${backup_file}.gz"
    
    echo "开始备份..."
    echo "源文件: $DB_PATH"
    echo "备份文件: $compressed_file"
    
    local start_time=$(date +%s)
    
    # 备份数据库
    if ! sqlite3 "$DB_PATH" ".backup '$backup_file'"; then
        echo "❌ 数据库备份失败"
        log_message "错误: 数据库备份失败"
        return 1
    fi
    
    # 压缩备份文件
    if ! gzip -9 "$backup_file"; then
        echo "❌ 压缩失败"
        log_message "错误: 备份压缩失败"
        rm -f "$backup_file"
        return 1
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # 获取大小信息
    local original_size=$(stat -c%s "$DB_PATH")
    local compressed_size=$(stat -c%s "$compressed_file")
    local compression_ratio=$(awk "BEGIN {printf \"%.1f\", $compressed_size*100/$original_size}")
    
    echo "✅ 备份完成!"
    echo "原始大小: $(numfmt --to=iec $original_size)"
    echo "压缩大小: $(numfmt --to=iec $compressed_size)"
    echo "压缩率: ${compression_ratio}%"
    echo "耗时: ${duration}秒"
    
    log_message "数据库备份完成: $(basename "$compressed_file"), 大小: $(numfmt --to=iec $compressed_size), 耗时: ${duration}秒"
    
    return 0
}

# 优化数据库
optimize_database() {
    echo "=== 优化数据库 ==="
    
    echo "正在分析数据库..."
    sqlite3 "$DB_PATH" "ANALYZE;"
    
    echo "正在优化数据库结构..."
    local before_size=$(stat -c%s "$DB_PATH")
    
    local start_time=$(date +%s)
    sqlite3 "$DB_PATH" "VACUUM;"
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    local after_size=$(stat -c%s "$DB_PATH")
    local saved_space=$((before_size - after_size))
    
    if [ $saved_space -gt 0 ]; then
        echo "✅ 数据库优化完成"
        echo "优化前大小: $(numfmt --to=iec $before_size)"
        echo "优化后大小: $(numfmt --to=iec $after_size)"
        echo "节省空间: $(numfmt --to=iec $saved_space)"
        echo "耗时: ${duration}秒"
        log_message "数据库优化完成，节省空间: $(numfmt --to=iec $saved_space)，耗时: ${duration}秒"
    else
        echo "✅ 数据库优化完成（无空间节省）"
        echo "数据库大小: $(numfmt --to=iec $after_size)"
        echo "耗时: ${duration}秒"
        log_message "数据库优化完成，耗时: ${duration}秒"
    fi
    
    return 0
}

# 清理过期备份
cleanup_old_backups() {
    echo "=== 清理过期备份 ==="
    
    local old_backups=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" -type f -mtime +$BACKUP_RETENTION_DAYS 2>/dev/null)
    
    if [ -z "$old_backups" ]; then
        echo "没有需要清理的过期备份"
        return 0
    fi
    
    local count=0
    local size=0
    while IFS= read -r file; do
        if [ -f "$file" ]; then
            local file_size=$(stat -c%s "$file" 2>/dev/null || echo "0")
            size=$((size + file_size))
            count=$((count + 1))
            echo "删除: $(basename "$file")"
            rm -f "$file"
        fi
    done <<< "$old_backups"
    
    if [ $count -gt 0 ]; then
        echo "✅ 清理完成: 删除 $count 个文件，释放 $(numfmt --to=iec $size) 空间"
        log_message "清理过期备份: 删除 $count 个文件，释放 $(numfmt --to=iec $size)"
    fi
    
    return 0
}

# 列出备份文件
list_backups() {
    echo "=== 备份文件列表 ==="
    
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "备份目录不存在"
        return 1
    fi
    
    local backup_files=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" -type f 2>/dev/null | sort -r)
    
    if [ -z "$backup_files" ]; then
        echo "没有找到备份文件"
        return 0
    fi
    
    echo ""
    printf "%-35s | %8s | %s\n" "文件名" "大小" "修改时间"
    echo "-------------------------------------------------------------------"
    
    while IFS= read -r file; do
        if [ -f "$file" ]; then
            local size=$(du -h "$file" | cut -f1)
            local mtime=$(stat -c %y "$file" 2>/dev/null | cut -d. -f1)
            local basename=$(basename "$file")
            printf "%-35s | %8s | %s\n" "$basename" "$size" "$mtime"
        fi
    done <<< "$backup_files"
    
    echo ""
    local file_count=$(echo "$backup_files" | wc -l)
    local total_size=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)
    echo "总计: $file_count 个文件，占用: $total_size"
}

# 测试备份完整性
test_backups() {
    echo "=== 测试备份完整性 ==="
    
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "备份目录不存在"
        return 1
    fi
    
    local backup_files=$(find "$BACKUP_DIR" -name "messages_backup_*.db.gz" -type f 2>/dev/null | sort -r | head -5)
    
    if [ -z "$backup_files" ]; then
        echo "没有找到备份文件"
        return 0
    fi
    
    echo "测试最近5个备份文件:"
    echo ""
    
    local success_count=0
    local total_count=0
    
    while IFS= read -r file; do
        if [ -f "$file" ]; then
            total_count=$((total_count + 1))
            local basename=$(basename "$file")
            echo -n "测试 $basename ... "
            
            if gzip -t "$file" 2>/dev/null; then
                echo "✓ 完整"
                success_count=$((success_count + 1))
            else
                echo "✗ 损坏"
            fi
        fi
    done <<< "$backup_files"
    
    echo ""
    echo "测试结果: $success_count/$total_count 个文件完整"
    
    if [ $success_count -eq $total_count ]; then
        log_message "备份完整性测试: 全部 $total_count 个文件完整"
        return 0
    else
        log_message "备份完整性测试: $success_count/$total_count 个文件完整，有文件损坏"
        return 1
    fi
}

# 完整维护
full_maintenance() {
    echo "=== 执行完整维护 ==="
    
    local start_time=$(date +%s)
    local success_count=0
    local total_tasks=3
    
    # 1. 清理数据库
    echo ""
    echo "步骤 1/$total_tasks: 清理过期数据"
    if clean_database; then
        success_count=$((success_count + 1))
    fi
    
    # 2. 备份数据库
    echo ""
    echo "步骤 2/$total_tasks: 备份数据库"
    if backup_database; then
        success_count=$((success_count + 1))
    fi
    
    # 3. 优化数据库
    echo ""
    echo "步骤 3/$total_tasks: 优化数据库"
    if optimize_database; then
        success_count=$((success_count + 1))
    fi
    
    # 清理过期备份
    echo ""
    cleanup_old_backups
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    echo ""
    echo "=== 完整维护完成 ==="
    echo "成功任务: $success_count/$total_tasks"
    echo "总耗时: ${total_duration}秒"
    
    log_message "完整维护完成: 成功 $success_count/$total_tasks 任务，总耗时: ${total_duration}秒"
    
    if [ $success_count -eq $total_tasks ]; then
        return 0
    else
        return 1
    fi
}

# 主程序
main() {
    local action=""
    FORCE_MODE=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --status)
                action="status"
                shift
                ;;
            --clean)
                action="clean"
                shift
                ;;
            --backup)
                action="backup"
                shift
                ;;
            --optimize)
                action="optimize"
                shift
                ;;
            --full)
                action="full"
                shift
                ;;
            --list-backups)
                action="list-backups"
                shift
                ;;
            --test-backups)
                action="test-backups"
                shift
                ;;
            --cleanup-backups)
                action="cleanup-backups"
                shift
                ;;
            --data-days)
                DATA_RETENTION_DAYS="$2"
                shift 2
                ;;
            --backup-days)
                BACKUP_RETENTION_DAYS="$2"
                shift 2
                ;;
            --force)
                FORCE_MODE=true
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
    
    # 如果没有指定操作，显示帮助
    if [ -z "$action" ]; then
        show_help
        exit 1
    fi
    
    # 验证参数
    if [[ ! "$DATA_RETENTION_DAYS" =~ ^[0-9]+$ ]] || [ "$DATA_RETENTION_DAYS" -lt 1 ]; then
        echo "错误: 数据保留天数必须是正整数"
        exit 1
    fi
    
    if [[ ! "$BACKUP_RETENTION_DAYS" =~ ^[0-9]+$ ]] || [ "$BACKUP_RETENTION_DAYS" -lt 1 ]; then
        echo "错误: 备份保留天数必须是正整数"
        exit 1
    fi
    
    # 检查依赖
    check_dependencies
    
    # 执行对应操作
    case "$action" in
        status)
            show_status
            ;;
        clean)
            clean_database
            ;;
        backup)
            backup_database
            ;;
        optimize)
            optimize_database
            ;;
        full)
            full_maintenance
            ;;
        list-backups)
            list_backups
            ;;
        test-backups)
            test_backups
            ;;
        cleanup-backups)
            cleanup_old_backups
            ;;
        *)
            echo "错误: 未知操作 '$action'"
            exit 1
            ;;
    esac
}

# 运行主程序
main "$@"
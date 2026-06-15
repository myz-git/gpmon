#!/bin/bash

# GPMon 邮件报告脚本
# 功能：生成并发送数据库监控日报
# 用法: ./send-daily-report.sh [选项]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DB_PATH="$PROJECT_ROOT/messages.db"
MAIL_TOOL="$PROJECT_ROOT/send_mail_cli"
REPORT_LOG="/var/log/gpmon-daily-report.log"

# 显示帮助信息
show_help() {
    echo "GPMon 邮件报告脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --send           发送监控日报邮件"
    echo "  --test-mail      测试邮件发送功能"
    echo "  --preview        预览报告内容（不发送）"
    echo "  -h, --help       显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --send        # 生成并发送日报"
    echo "  $0 --test-mail   # 测试邮件配置"
    echo "  $0 --preview     # 预览报告内容"
    echo ""
}

# 记录日志
log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message"
    echo "[$timestamp] $message" >> "$REPORT_LOG"
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
    
    # 检查邮件工具
    if [ ! -f "$MAIL_TOOL" ]; then
        echo "错误: 邮件工具不存在: $MAIL_TOOL"
        echo "请先运行: $SCRIPT_DIR/build.sh --mail"
        exit 1
    fi
    
    # 创建日志目录
    mkdir -p "$(dirname "$REPORT_LOG")"
}

# 生成报告内容
generate_report() {
    local report_date=$(date '+%Y-%m-%d')
    local report_time=$(date '+%H:%M:%S')
    
    # 获取统计数据
    local total_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result;" 2>/dev/null || echo "0")
    local ok_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='OK';" 2>/dev/null || echo "0")
    local error_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='ERROR';" 2>/dev/null || echo "0")
    local warning_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='WARNING';" 2>/dev/null || echo "0")
    
    # 计算百分比
    local ok_percent=0
    local error_percent=0
    local warning_percent=0
    
    if [ $total_count -gt 0 ]; then
        ok_percent=$(awk "BEGIN {printf \"%.1f\", $ok_count*100/$total_count}")
        error_percent=$(awk "BEGIN {printf \"%.1f\", $error_count*100/$total_count}")
        warning_percent=$(awk "BEGIN {printf \"%.1f\", $warning_count*100/$total_count}")
    fi
    
    # 生成文本报告
    cat << EOF
========================================
GPMon 数据库监控日报
========================================
报告时间: $report_date $report_time

[监控统计] 总体状况
----------------------------------------
总监控项: $total_count
正常数量: $ok_count ($ok_percent%)
错误数量: $error_count ($error_percent%)
警告数量: $warning_count ($warning_percent%)

[数据统计] 数据库类型统计
----------------------------------------
EOF

    # 添加数据库类型统计
    if [ $total_count -gt 0 ]; then
        echo "类型      总数  正常  错误  警告  健康率"
        echo "--------------------------------------"
        sqlite3 "$DB_PATH" "
        SELECT 
            printf('%-8s %4d %4d %4d %4d %5.1f%%', 
                   dbtype, 
                   COUNT(*),
                   SUM(CASE WHEN chk_result='OK' THEN 1 ELSE 0 END),
                   SUM(CASE WHEN chk_result='ERROR' THEN 1 ELSE 0 END),
                   SUM(CASE WHEN chk_result='WARNING' THEN 1 ELSE 0 END),
                   CASE WHEN COUNT(*) > 0 THEN 
                       SUM(CASE WHEN chk_result='OK' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) 
                   ELSE 0 END
            )
        FROM check_result 
        GROUP BY dbtype 
        ORDER BY dbtype;
        " 2>/dev/null
    fi
    
    # 添加异常详情
    echo ""
    echo "[异常详情] 错误和警告"
    echo "----------------------------------------"
    
    local error_details=$(sqlite3 "$DB_PATH" "
    SELECT 
        printf('[%s] %s:%d (%s/%s) - %s', 
               chk_result, ip, port, dbtype, dbname, chk_nm)
    FROM check_result 
    WHERE chk_result IN ('ERROR', 'WARNING')
    ORDER BY chk_result DESC, chk_time DESC
    LIMIT 20;
    " 2>/dev/null)
    
    if [ -n "$error_details" ]; then
        echo "$error_details"
    else
        echo "[正常] 当前没有错误或警告!"
    fi
    
    # 添加系统信息
    echo ""
    echo "[系统信息] 数据库状态"
    echo "----------------------------------------"
    echo "数据库大小: $(du -h "$DB_PATH" | cut -f1)"
    
    local hist_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his;" 2>/dev/null || echo "0")
    echo "历史记录: $hist_count 条"
    
    if [ -d "$PROJECT_ROOT/backup" ]; then
        local backup_count=$(find "$PROJECT_ROOT/backup" -name "messages_backup_*.db.gz" 2>/dev/null | wc -l)
        echo "备份数量: $backup_count 个"
    fi
    
    echo ""
    echo "========================================"
    echo "报告生成完成 - $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================"
}

# 生成HTML报告（美化版本，固定2列布局）
generate_html_report() {
    local report_date=$(date '+%Y-%m-%d')
    local report_time=$(date '+%H:%M:%S')
    local temp_file="/tmp/gpmon_html_temp_$(date +%s).tmp"
    
    # 获取统计数据
    local total_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result;" 2>/dev/null || echo "0")
    local ok_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='OK';" 2>/dev/null || echo "0")
    local error_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='ERROR';" 2>/dev/null || echo "0")
    local warning_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='WARNING';" 2>/dev/null || echo "0")
    
    # 计算百分比
    local ok_percent=0
    local error_percent=0
    local warning_percent=0
    
    if [ $total_count -gt 0 ]; then
        ok_percent=$(awk "BEGIN {printf \"%.1f\", $ok_count*100/$total_count}")
        error_percent=$(awk "BEGIN {printf \"%.1f\", $error_count*100/$total_count}")
        warning_percent=$(awk "BEGIN {printf \"%.1f\", $warning_count*100/$total_count}")
    fi
    
    # 生成HTML报告头部
    cat > "$temp_file" << EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GPMon 监控日报 - $report_date</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f7fa;
            margin: 0;
            padding: 20px;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
            position: relative;
        }
        
        .header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: url('data:image/svg+xml;charset=UTF-8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="25" cy="25" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="75" cy="75" r="3" fill="rgba(255,255,255,0.1)"/><circle cx="50" cy="15" r="1.5" fill="rgba(255,255,255,0.1)"/><circle cx="15" cy="85" r="2.5" fill="rgba(255,255,255,0.1)"/></svg>');
            opacity: 0.3;
        }
        
        .header h1 {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 8px;
            position: relative;
            z-index: 1;
        }
        
        .header .subtitle {
            font-size: 16px;
            opacity: 0.9;
            position: relative;
            z-index: 1;
        }
        
        .content {
            padding: 30px;
        }
        
        .section {
            margin-bottom: 35px;
        }
        
        .section-title {
            background: linear-gradient(90deg, #f8f9fa 0%, #e9ecef 100%);
            padding: 15px 20px;
            border-left: 5px solid #007bff;
            font-weight: 600;
            font-size: 18px;
            color: #2c3e50;
            margin-bottom: 20px;
            border-radius: 0 8px 8px 0;
            box-shadow: 0 2px 8px rgba(0, 123, 255, 0.1);
        }
        
        /* 固定2列统计卡片布局 */
        .stats-container {
            margin-bottom: 25px;
        }
        
        .stats-row {
            display: flex;
            gap: 15px;
            margin-bottom: 15px;
            width: 100%;
        }
        
        .stat-card {
            flex: 1;
            background: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
            box-shadow: 0 3px 15px rgba(0, 0, 0, 0.08);
            border-top: 4px solid;
            transition: transform 0.2s ease;
            min-width: 0; /* 防止文字溢出 */
        }
        
        .stat-card:hover {
            transform: translateY(-2px);
        }
        
        .stat-card.total {
            border-top-color: #6c757d;
        }
        
        .stat-card.success {
            border-top-color: #28a745;
        }
        
        .stat-card.error {
            border-top-color: #dc3545;
        }
        
        .stat-card.warning {
            border-top-color: #ffc107;
        }
        
        .stat-number {
            font-size: 32px;
            font-weight: 700;
            margin-bottom: 5px;
            word-wrap: break-word;
        }
        
        .stat-label {
            font-size: 14px;
            color: #6c757d;
            font-weight: 500;
        }
        
        .stat-percentage {
            font-size: 12px;
            margin-top: 5px;
            padding: 3px 8px;
            border-radius: 12px;
            display: inline-block;
        }
        
        .table-container {
            background: white;
            border-radius: 10px;
            overflow: hidden;
            box-shadow: 0 3px 15px rgba(0, 0, 0, 0.08);
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
        }
        
        th {
            background: linear-gradient(135deg, #495057 0%, #343a40 100%);
            color: white;
            padding: 15px 12px;
            text-align: left;
            font-weight: 600;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        td {
            padding: 12px;
            border-bottom: 1px solid #e9ecef;
            vertical-align: middle;
        }
        
        tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        
        tr:hover {
            background-color: #e3f2fd;
        }
        
        .status-ok {
            color: #28a745;
            font-weight: 600;
        }
        
        .status-error {
            color: #dc3545;
            font-weight: 600;
        }
        
        .status-warning {
            color: #ffc107;
            font-weight: 600;
        }
        
        .alert-box {
            background: linear-gradient(135deg, #fff3cd 0%, #ffeaa7 100%);
            border: 1px solid #ffeaa7;
            border-radius: 10px;
            padding: 20px;
            margin: 15px 0;
            box-shadow: 0 3px 10px rgba(255, 193, 7, 0.2);
        }
        
        .alert-box ul {
            margin: 0;
            padding-left: 20px;
        }
        
        .alert-box li {
            margin-bottom: 8px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            background: rgba(255, 255, 255, 0.7);
            padding: 5px 10px;
            border-radius: 5px;
            border-left: 3px solid #ffc107;
        }
        
        .success-box {
            background: linear-gradient(135deg, #d4edda 0%, #c3e6cb 100%);
            border: 1px solid #c3e6cb;
            border-radius: 10px;
            padding: 20px;
            margin: 15px 0;
            color: #155724;
            text-align: center;
            box-shadow: 0 3px 10px rgba(40, 167, 69, 0.2);
        }
        
        .success-box strong {
            font-size: 18px;
            display: block;
            margin-bottom: 8px;
        }
        
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        
        .info-item {
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
            border-left: 4px solid #17a2b8;
        }
        
        .info-label {
            font-weight: 600;
            color: #495057;
            font-size: 14px;
            margin-bottom: 5px;
        }
        
        .info-value {
            font-size: 18px;
            color: #2c3e50;
            font-weight: 500;
        }
        
        .footer {
            background: #f8f9fa;
            padding: 25px;
            text-align: center;
            color: #6c757d;
            font-size: 13px;
            border-top: 1px solid #e9ecef;
        }
        
        .footer p {
            margin: 5px 0;
        }
        
        .badge {
            display: inline-block;
            padding: 4px 8px;
            font-size: 11px;
            font-weight: 600;
            border-radius: 4px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .badge-success {
            background-color: #d4edda;
            color: #155724;
        }
        
        .badge-danger {
            background-color: #f8d7da;
            color: #721c24;
        }
        
        .badge-warning {
            background-color: #fff3cd;
            color: #856404;
        }
        
        .health-bar {
            width: 100%;
            height: 6px;
            background-color: #e9ecef;
            border-radius: 3px;
            overflow: hidden;
            margin-top: 5px;
        }
        
        .health-fill {
            height: 100%;
            background: linear-gradient(90deg, #28a745 0%, #20c997 100%);
            border-radius: 3px;
            transition: width 0.3s ease;
        }
        
        /* 移动端适配 - 在小屏幕上仍保持2列布局 */
        @media (max-width: 600px) {
            .container {
                margin: 10px;
                border-radius: 8px;
            }
            
            .content {
                padding: 20px;
            }
            
            .stats-row {
                gap: 10px;
                margin-bottom: 10px;
            }
            
            .stat-card {
                padding: 15px;
            }
            
            .stat-number {
                font-size: 24px;
            }
            
            .stat-label {
                font-size: 12px;
            }
            
            th, td {
                padding: 8px;
                font-size: 12px;
            }
        }
        
        /* 超小屏幕适配 */
        @media (max-width: 480px) {
            .stat-number {
                font-size: 20px;
            }
            
            .stat-label {
                font-size: 11px;
            }
            
            .stat-card {
                padding: 12px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>GPMon 数据库监控日报</h1>
            <div class="subtitle">报告时间: $report_date $report_time</div>
        </div>
        
        <div class="content">
            <div class="section">
                <div class="section-title">监控统计总览</div>
                <div class="stats-container">
                    <!-- 第一行：总监控项 和 正常 -->
                    <div class="stats-row">
                        <div class="stat-card total">
                            <div class="stat-number">$total_count</div>
                            <div class="stat-label">总监控项</div>
                        </div>
                        <div class="stat-card success">
                            <div class="stat-number status-ok">$ok_count</div>
                            <div class="stat-label">正常</div>
                            <div class="stat-percentage badge-success">$ok_percent%</div>
                        </div>
                    </div>
                    
                    <!-- 第二行：错误 和 警告 -->
                    <div class="stats-row">
                        <div class="stat-card error">
                            <div class="stat-number status-error">$error_count</div>
                            <div class="stat-label">错误</div>
                            <div class="stat-percentage badge-danger">$error_percent%</div>
                        </div>
                        <div class="stat-card warning">
                            <div class="stat-number status-warning">$warning_count</div>
                            <div class="stat-label">警告</div>
                            <div class="stat-percentage badge-warning">$warning_percent%</div>
                        </div>
                    </div>
                </div>
            </div>
EOF

    # 添加数据库类型统计表（如果有数据）
    if [ $total_count -gt 0 ]; then
        cat >> "$temp_file" << EOF
            <div class="section">
                <div class="section-title">数据库类型详细统计</div>
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>数据库类型</th>
                                <th>总数</th>
                                <th>正常</th>
                                <th>错误</th>
                                <th>警告</th>
                                <th>健康率</th>
                                <th>状态</th>
                            </tr>
                        </thead>
                        <tbody>
EOF
        
        # 获取数据库类型统计
        local db_types=$(sqlite3 "$DB_PATH" "SELECT DISTINCT dbtype FROM check_result ORDER BY dbtype;" 2>/dev/null)
        
        if [ -n "$db_types" ]; then
            while IFS= read -r dbtype; do
                if [ -n "$dbtype" ]; then
                    local stats=$(sqlite3 "$DB_PATH" "
                    SELECT 
                        COUNT(*) || '|' ||
                        SUM(CASE WHEN chk_result='OK' THEN 1 ELSE 0 END) || '|' ||
                        SUM(CASE WHEN chk_result='ERROR' THEN 1 ELSE 0 END) || '|' ||
                        SUM(CASE WHEN chk_result='WARNING' THEN 1 ELSE 0 END) || '|' ||
                        CASE WHEN COUNT(*) > 0 THEN 
                            SUM(CASE WHEN chk_result='OK' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) 
                        ELSE 0 END
                    FROM check_result 
                    WHERE dbtype='$dbtype';" 2>/dev/null)
                    
                    if [ -n "$stats" ]; then
                        local total=$(echo "$stats" | cut -d'|' -f1)
                        local ok_count=$(echo "$stats" | cut -d'|' -f2)
                        local error_count=$(echo "$stats" | cut -d'|' -f3)
                        local warning_count=$(echo "$stats" | cut -d'|' -f4)
                        local health_rate=$(echo "$stats" | cut -d'|' -f5)
                        local health_rate_rounded=$(printf "%.1f" "$health_rate")
                        
                        # 确定状态标识
                        local status_badge=""
                        if [ "$error_count" -gt 0 ]; then
                            status_badge="<span class=\"badge badge-danger\">异常</span>"
                        elif [ "$warning_count" -gt 0 ]; then
                            status_badge="<span class=\"badge badge-warning\">警告</span>"
                        else
                            status_badge="<span class=\"badge badge-success\">正常</span>"
                        fi
                        
                        cat >> "$temp_file" << EOF
                            <tr>
                                <td><strong>$dbtype</strong></td>
                                <td>$total</td>
                                <td class="status-ok">$ok_count</td>
                                <td class="status-error">$error_count</td>
                                <td class="status-warning">$warning_count</td>
                                <td>
                                    <strong>${health_rate_rounded}%</strong>
                                    <div class="health-bar">
                                        <div class="health-fill" style="width: ${health_rate_rounded}%"></div>
                                    </div>
                                </td>
                                <td>$status_badge</td>
                            </tr>
EOF
                    fi
                fi
            done <<< "$db_types"
        fi
        
        cat >> "$temp_file" << EOF
                        </tbody>
                    </table>
                </div>
            </div>
EOF
    fi
    
    # 添加异常详情
    cat >> "$temp_file" << EOF
            <div class="section">
                <div class="section-title">异常详情监控</div>
EOF

    local error_details=$(sqlite3 "$DB_PATH" "
    SELECT 
        printf('[%s] %s:%d (%s/%s) - %s', 
               chk_result, ip, port, dbtype, dbname, chk_nm)
    FROM check_result 
    WHERE chk_result IN ('ERROR', 'WARNING')
    ORDER BY chk_result DESC, chk_time DESC
    LIMIT 10;
    " 2>/dev/null)
    
    if [ -n "$error_details" ]; then
        echo "                <div class=\"alert-box\">" >> "$temp_file"
        echo "                    <h4 style=\"margin-bottom: 15px; color: #856404;\">发现异常项目</h4>" >> "$temp_file"
        echo "                    <ul>" >> "$temp_file"
        while IFS= read -r line; do
            if [ -n "$line" ]; then
                echo "                        <li>$line</li>" >> "$temp_file"
            fi
        done <<< "$error_details"
        echo "                    </ul>" >> "$temp_file"
        echo "                </div>" >> "$temp_file"
    else
        echo "                <div class=\"success-box\">" >> "$temp_file"
        echo "                    <strong>系统运行正常</strong>" >> "$temp_file"
        echo "                    <p>所有监控项目均处于正常状态，未发现任何错误或警告！</p>" >> "$temp_file"
        echo "                </div>" >> "$temp_file"
    fi

    # 添加系统信息
    local hist_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_his;" 2>/dev/null || echo "0")
    local backup_count=0
    if [ -d "$PROJECT_ROOT/backup" ]; then
        backup_count=$(find "$PROJECT_ROOT/backup" -name "messages_backup_*.db.gz" 2>/dev/null | wc -l)
    fi
    local db_size=$(du -h "$DB_PATH" | cut -f1)

    cat >> "$temp_file" << EOF
            </div>
            
            <div class="section">
                <div class="section-title">系统状态信息</div>
                <div class="info-grid">
                    <div class="info-item">
                        <div class="info-label">数据库大小</div>
                        <div class="info-value">$db_size</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">历史记录</div>
                        <div class="info-value">$hist_count 条</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">备份文件</div>
                        <div class="info-value">$backup_count 个</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">报告状态</div>
                        <div class="info-value"><span class="badge badge-success">已生成</span></div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="footer">
            <p><strong>GPMon 数据库监控系统</strong></p>
            <p>此报告由系统自动生成 | 生成时间: $(date '+%Y-%m-%d %H:%M:%S')</p>
            <p>如有疑问，请联系系统管理员</p>
        </div>
    </div>
</body>
</html>
EOF

    # 输出生成的HTML内容
    cat "$temp_file"
    
    # 清理临时文件
    rm -f "$temp_file"
}

# 发送监控日报
send_daily_report() {
    echo "=== 发送监控日报 ==="
    
    # 生成报告
    local report_date=$(date '+%Y-%m-%d')
    local temp_dir="/tmp/gpmon-report"
    mkdir -p "$temp_dir"
    
    local text_file="$temp_dir/report_${report_date}.txt"
    local html_file="$temp_dir/report_${report_date}.html"
    
    echo "正在生成报告..."
    
    # 生成文本报告
    echo "生成文本报告..."
    if ! generate_report > "$text_file"; then
        echo "❌ 文本报告生成失败"
        return 1
    fi
    
    # 生成HTML报告
    echo "生成HTML报告..."
    if ! generate_html_report > "$html_file"; then
        echo "❌ HTML报告生成失败"
        return 1
    fi
    
    echo "报告生成完成:"
    echo "  文本报告: $text_file ($(du -h "$text_file" | cut -f1))"
    echo "  HTML报告: $html_file ($(du -h "$html_file" | cut -f1))"
    
    # 验证文件生成
    if [ ! -f "$text_file" ] || [ ! -s "$text_file" ]; then
        echo "❌ 文本报告生成失败或文件为空"
        return 1
    fi
    
    if [ ! -f "$html_file" ] || [ ! -s "$html_file" ]; then
        echo "❌ HTML报告生成失败或文件为空"
        return 1
    fi
    
    # 构建邮件主题
    local error_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='ERROR';" 2>/dev/null || echo "0")
    local warning_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM check_result WHERE chk_result='WARNING';" 2>/dev/null || echo "0")
    
    local subject="GPMon 监控日报 - $report_date"
    if [ $error_count -gt 0 ]; then
        subject="$subject [警告: ${error_count}个错误]"
    elif [ $warning_count -gt 0 ]; then
        subject="$subject [注意: ${warning_count}个警告]"
    else
        subject="$subject [正常]"
    fi
    
    echo "正在发送邮件..."
    echo "主题: $subject"
    
    # 发送邮件
    if "$MAIL_TOOL" -subject "$subject" -text-file "$text_file" -html-file "$html_file"; then
        echo "✅ 邮件发送成功"
        log_message "邮件发送成功: 主题='$subject'"
        
        # 清理旧的临时文件（保留3天）
        find "$temp_dir" -name "report_*.txt" -mtime +3 -delete 2>/dev/null || true
        find "$temp_dir" -name "report_*.html" -mtime +3 -delete 2>/dev/null || true
        
        return 0
    else
        echo "❌ 邮件发送失败"
        log_message "邮件发送失败: 主题='$subject'"
        
        # 保留失败的报告文件用于调试
        echo "失败的报告文件已保留:"
        echo "  文本报告: $text_file"
        echo "  HTML报告: $html_file"
        
        return 1
    fi
}

# 测试邮件功能
test_mail() {
    echo "=== 测试邮件功能 ==="
    
    if "$MAIL_TOOL" -test-config; then
        echo "✅ 邮件配置测试成功"
        log_message "邮件配置测试成功"
        return 0
    else
        echo "❌ 邮件配置测试失败"
        log_message "邮件配置测试失败"
        return 1
    fi
}

# 预览报告
preview_report() {
    echo "=== 预览报告内容 ==="
    echo ""
    generate_report
}

# 主程序
main() {
    local action=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --send)
                action="send"
                shift
                ;;
            --test-mail)
                action="test-mail"
                shift
                ;;
            --preview)
                action="preview"
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
    
    # 如果没有指定操作，默认发送报告
    if [ -z "$action" ]; then
        action="send"
    fi
    
    # 检查依赖（预览模式除外）
    if [ "$action" != "preview" ]; then
        check_dependencies
    else
        # 预览模式只检查数据库
        if [ ! -f "$DB_PATH" ]; then
            echo "错误: 数据库文件不存在: $DB_PATH"
            exit 1
        fi
    fi
    
    # 执行对应操作
    case "$action" in
        send)
            send_daily_report
            ;;
        test-mail)
            test_mail
            ;;
        preview)
            preview_report
            ;;
        *)
            echo "错误: 未知操作 '$action'"
            exit 1
            ;;
    esac
}

# 运行主程序
main "$@" 
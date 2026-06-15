#!/bin/bash
# GPMon 统一编译脚本
# 用法:
#   ./scripts/build.sh              # 编译全部组件
#   ./scripts/build.sh --mail       # 仅编译邮件工具
#   ./scripts/build.sh --server     # 仅编译 gRPC 服务
#   ./scripts/build.sh --clients    # 仅编译监控客户端
#   ./scripts/build.sh --skip-db2   # 跳过 DB2 客户端

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
CGO_ENABLED="${CGO_ENABLED:-1}"

BUILD_SERVER=false
BUILD_CLIENTS=false
BUILD_MAIL=false
BUILD_ALL=true
SKIP_DB2=false

show_help() {
    cat <<'EOF'
GPMon 统一编译脚本

用法: ./scripts/build.sh [选项]

选项:
  --all           编译全部组件（默认）
  --server        仅编译 gRPC 服务 (startgpmon，不需要 DB2/Oracle 驱动)
  --clients       仅编译监控客户端 (orasvc/mysqlsvc/db2svc)
  --mail          仅编译邮件工具 (send_mail_cli)
  --skip-db2      跳过 DB2 客户端编译
  -h, --help      显示帮助

环境变量:
  GOOS            目标系统，默认 linux
  GOARCH          目标架构，默认 amd64
  CGO_ENABLED     是否启用 CGO，默认 1（Oracle/DB2/SQLite 需要）

输出目录:
  所有可执行文件生成在项目根目录:
    startgpmon, orasvc, mysqlsvc, db2svc, send_mail_cli

示例:
  ./scripts/build.sh
  ./scripts/build.sh --mail
  GOOS=linux GOARCH=amd64 ./scripts/build.sh --clients --skip-db2
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --all)
            BUILD_ALL=true
            ;;
        --server)
            BUILD_ALL=false
            BUILD_SERVER=true
            ;;
        --clients)
            BUILD_ALL=false
            BUILD_CLIENTS=true
            ;;
        --mail)
            BUILD_ALL=false
            BUILD_MAIL=true
            ;;
        --skip-db2)
            SKIP_DB2=true
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
    shift
done

if $BUILD_ALL; then
    BUILD_SERVER=true
    BUILD_CLIENTS=true
    BUILD_MAIL=true
fi

if ! command -v go >/dev/null 2>&1; then
    echo "错误: 未找到 Go 编译器，请先安装 Go 1.20+"
    exit 1
fi

export GO111MODULE=on
export GOOS GOARCH CGO_ENABLED

cd "$PROJECT_ROOT"
mkdir -p "$PROJECT_ROOT/log"

load_native_env() {
    if [ -f "$PROJECT_ROOT/install_cfg/bash_profile" ]; then
        # shellcheck source=/dev/null
        . "$PROJECT_ROOT/install_cfg/bash_profile"
    fi
}

echo "=== GPMon 编译 ==="
echo "Go 版本 : $(go version)"
echo "目标平台: ${GOOS}/${GOARCH}"
echo "CGO     : ${CGO_ENABLED}"
echo "项目目录: ${PROJECT_ROOT}"
echo ""

build_bin() {
    local label="$1"
    local workdir="$2"
    local source="$3"
    local output="$4"
    local tags="${5:-}"

    echo ">> 编译 ${label} ..."
    (
        cd "$workdir"
        if [ -n "$tags" ]; then
            go build -tags "$tags" -o "$PROJECT_ROOT/$output" "$source"
        else
            go build -o "$PROJECT_ROOT/$output" "$source"
        fi
    )
    chmod +x "$PROJECT_ROOT/$output"
    echo "   完成: $PROJECT_ROOT/$output ($(du -h "$PROJECT_ROOT/$output" | cut -f1))"
    echo ""
}

if $BUILD_SERVER; then
    build_bin "gRPC 服务" "$PROJECT_ROOT/mon-server" "main.go" "startgpmon"
fi

if $BUILD_CLIENTS; then
    load_native_env
    build_bin "Oracle 监控客户端" "$PROJECT_ROOT/mon-client-ora" "startora.go" "orasvc" "oracle"
    build_bin "MySQL 监控客户端" "$PROJECT_ROOT/mon-client-mysql" "startmysql.go" "mysqlsvc" "mysql"
    if ! $SKIP_DB2; then
        if [ -z "${IBM_DB_HOME:-}" ] || [ ! -f "${IBM_DB_HOME}/include/sqlcli1.h" ]; then
            echo "错误: 未找到 DB2 clidriver 头文件 (sqlcli1.h)"
            echo "请先安装 clidriver 并配置 IBM_DB_HOME，参考 install_cfg/bash_profile"
            echo "若不需要 DB2 监控，可使用: ./scripts/build.sh --skip-db2"
            exit 1
        fi
        build_bin "DB2 监控客户端" "$PROJECT_ROOT/mon-client-db2" "startdb2.go" "db2svc" "db2"
    else
        echo ">> 跳过 DB2 客户端 (--skip-db2)"
        echo ""
    fi
fi

if $BUILD_MAIL; then
    echo ">> 编译邮件工具 ..."
    go build -o "$PROJECT_ROOT/send_mail_cli" "$PROJECT_ROOT/send_mail_cli.go"
    chmod +x "$PROJECT_ROOT/send_mail_cli"
    echo "   完成: $PROJECT_ROOT/send_mail_cli ($(du -h "$PROJECT_ROOT/send_mail_cli" | cut -f1))"
    echo ""
fi

echo "=== 编译完成 ==="
ls -lh "$PROJECT_ROOT"/startgpmon "$PROJECT_ROOT"/orasvc "$PROJECT_ROOT"/mysqlsvc \
    "$PROJECT_ROOT"/db2svc "$PROJECT_ROOT"/send_mail_cli 2>/dev/null || true

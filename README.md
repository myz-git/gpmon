# GPMon 数据库监控系统

基于 gRPC 的数据库监控微服务，支持 Oracle、MySQL、DB2 等数据库。通过 Supervisor 管理常驻进程，配合运维脚本实现日报邮件、数据库维护与日志轮转。

## 目录

- [项目结构](#项目结构)
- [快速开始（生产部署）](#快速开始生产部署)
- [环境准备](#环境准备)
- [编译](#编译)
- [编译产物打包](#编译产物打包)
- [部署与运行](#部署与运行)
- [运维脚本](#运维脚本)
- [数据库配置](#数据库配置)
- [本地开发调试](#本地开发调试)
- [Proto 编译](#proto-编译)
- [常见问题](#常见问题)
- [Git 操作参考](#git-操作参考)

---

## 项目结构

```
gpmon/
├── mon-server/              # gRPC 服务端
│   └── main.go
├── mon-client-ora/          # Oracle 监控客户端
│   └── startora.go
├── mon-client-mysql/          # MySQL 监控客户端
│   └── startmysql.go
├── mon-client-db2/            # DB2 监控客户端
│   └── startdb2.go
├── grpc/                      # gRPC 通信层
│   ├── server.go
│   ├── client.go
│   └── proto/
├── db/                        # 数据库访问层
├── utils/                     # 工具（邮件发送等）
├── scripts/                   # 运维脚本
│   ├── build.sh               # 统一编译脚本
│   ├── build-mail-tool.sh     # 编译邮件工具（兼容入口）
│   ├── send-daily-report.sh   # 日报发送
│   ├── db-maintenance.sh      # 数据库维护
│   ├── setup-maintenance.sh   # 定时任务安装
│   └── setup-logrotate.sh     # 日志轮转安装
├── install_cfg/               # 部署配置模板
│   ├── bash_profile           # 环境变量参考
│   └── supervisord.d/         # Supervisor 配置模板
├── send_mail_cli.go           # 邮件工具源码
├── messages.db                # SQLite 配置库
└── log/                       # 应用日志目录
```

### 编译产物（均在项目根目录）

| 可执行文件 | 来源 | 说明 |
|-----------|------|------|
| `startgpmon` | mon-server | gRPC 服务端，监听 5051 |
| `orasvc` | mon-client-ora | Oracle 监控客户端 |
| `mysqlsvc` | mon-client-mysql | MySQL 监控客户端 |
| `db2svc` | mon-client-db2 | DB2 监控客户端 |
| `send_mail_cli` | send_mail_cli.go | 邮件发送工具 |

---

## 快速开始（生产部署）

以下假设项目部署在 `/workspace/gpmon`，按需修改路径。

```bash
# 1. 获取代码
cd /workspace/gpmon
git pull

# 2. 配置环境变量（Oracle Instant Client + DB2 clidriver）
cp install_cfg/bash_profile ~/.bash_profile   # 按实际路径修改后 source
source ~/.bash_profile

# 3. 修复脚本换行符（Windows 上传时可能需要）
sed -i 's/\r$//' scripts/*.sh install_cfg/bash_profile install_cfg/supervisord.d/*.ini
chmod +x scripts/*.sh

# 4. 加载编译环境并编译全部组件
source <(sed 's/\r$//' install_cfg/bash_profile)   # 按实际路径修改后使用
./scripts/build.sh

# 5. 配置 Supervisor 并启动监控服务
cp install_cfg/supervisord.d/*.ini /etc/supervisord.d/
# 修改 ini 中的服务器 IP 和路径
supervisorctl update
supervisorctl start all

# 6. 安装运维定时任务（日报 + 数据库维护）
sudo ./scripts/setup-maintenance.sh --setup

# 7. 安装日志轮转（可选）
sudo ./scripts/setup-logrotate.sh

# 8. 验证
./scripts/db-maintenance.sh --status
./scripts/send-daily-report.sh --test-mail
```

---

## 环境准备

### Go 语言

- 版本：Go 1.20+
- Linux 安装参考：https://go.dev/dl/

```bash
export PATH=$PATH:/usr/local/go/bin
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

### Oracle Instant Client

下载：https://www.oracle.com/database/technologies/instant-client/

```bash
# Linux 示例
unzip instantclient-basic-linux.x64-*.zip -d /instantclient
export TNS_ADMIN=/instantclient
export LD_LIBRARY_PATH=/instantclient:$LD_LIBRARY_PATH
echo "/instantclient" >> /etc/ld.so.conf && ldconfig
```

Windows 解压后设置 `TNS_ADMIN` 和 `LD_LIBRARY_PATH`（或 PATH）指向 instantclient 目录。

编译和运行 Oracle 相关组件需要 `CGO_ENABLED=1`。

### DB2 clidriver

参考：https://github.com/ibmdb/go_ibm_db

```bash
go install github.com/ibmdb/go_ibm_db/installer@v0.4.3
cd $(go env GOPATH)/pkg/mod/github.com/ibmdb/go_ibm_db@v0.4.3/installer
go run setup.go
cp -r $(go env GOPATH)/pkg/mod/github.com/ibmdb/clidriver /workspace/gpmon/local/

export IBM_DB_HOME=/workspace/gpmon/local/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib
export LD_LIBRARY_PATH=$IBM_DB_HOME/lib:$LD_LIBRARY_PATH
```

完整环境变量模板见 `install_cfg/bash_profile`。

### 依赖与 vendor

项目已包含 `go.mod` / `go.sum`，**不要**删除后重新 `go mod init`。

| 场景 | 说明 |
|------|------|
| 有网络 | 服务器直接 `go build`，按 `go.mod` 拉取依赖 |
| 离线/版本锁定 | 开发机执行 `go mod vendor`，将 `vendor/` 目录一并部署到服务器 |
| 版本漂移 | 无 `vendor/` 时，服务器可能拉到与开发机不同的依赖版本（如 `go_ibm_db` v0.5.x） |

开发机刷新 vendor：

```bash
go mod tidy
go mod vendor
```

仓库通过 `.gitattributes` 约定 `*.sh`、`install_cfg/**` 使用 LF 换行，避免 Linux 上出现 `$'\r': command not found`。

| 数据库 | Go 驱动 |
|--------|---------|
| SQLite（配置库） | github.com/mattn/go-sqlite3 |
| Oracle | github.com/godror/godror |
| MySQL | github.com/go-sql-driver/mysql |
| DB2 | github.com/ibmdb/go_ibm_db |

---

## 编译

统一使用 `scripts/build.sh`，所有可执行文件输出到项目根目录。

```bash
cd /workspace/gpmon

# 编译全部（推荐）
./scripts/build.sh

# 仅编译监控服务
./scripts/build.sh --server

# 仅编译客户端（Oracle + MySQL + DB2）
./scripts/build.sh --clients

# 不编译 DB2 客户端
./scripts/build.sh --clients --skip-db2

# 仅编译邮件工具
./scripts/build.sh --mail
# 或
./scripts/build-mail-tool.sh
```

### 编译说明

各组件按需编译，**gRPC 服务端不依赖 DB2/Oracle 客户端库**：

| 组件 | 编译标签 | 额外依赖 |
|------|----------|----------|
| startgpmon | 无 | SQLite（CGO） |
| orasvc | `oracle` | Oracle Instant Client |
| mysqlsvc | `mysql` | 无（纯 Go 驱动） |
| db2svc | `db2` | DB2 clidriver（`IBM_DB_HOME`） |
| send_mail_cli | 无 | SQLite（CGO） |

DB2 编译前需配置环境变量（参考 `install_cfg/bash_profile`）：

```bash
export IBM_DB_HOME=/workspace/gpmon/local/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib
export LD_LIBRARY_PATH=$IBM_DB_HOME/lib:$LD_LIBRARY_PATH
```

若服务器未安装 DB2 clidriver，可跳过 DB2 客户端编译，不影响其他组件。

编译客户端前，`build.sh` 会自动加载 `install_cfg/bash_profile` 中的 Oracle/DB2 环境变量。仅编服务端（`--server`）时不需要 DB2 头文件。

### 编译选项速查

| 命令 | 产出 |
|------|------|
| `./scripts/build.sh` | 全部 5 个可执行文件 |
| `./scripts/build.sh --server` | 仅 `startgpmon` |
| `./scripts/build.sh --clients` | `orasvc` + `mysqlsvc` + `db2svc` |
| `./scripts/build.sh --clients --skip-db2` | `orasvc` + `mysqlsvc` |
| `./scripts/build.sh --mail` | 仅 `send_mail_cli` |

Windows 本地编译示例（仅邮件工具等无需 Oracle/DB2 的组件）：

```bat
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
scripts\build.sh --mail
```

### 更新部署

```bash
cd /workspace/gpmon
git pull
sed -i 's/\r$//' scripts/*.sh install_cfg/bash_profile 2>/dev/null
supervisorctl stop all
./scripts/build.sh
supervisorctl start all
```

---

## 编译产物打包

编译完成后，可在项目根目录将可执行文件打成 zip，便于分发到其他同类 Linux 服务器。

### 仅打包可执行文件

```bash
cd /workspace/gpmon

zip -j gpmon-bin-$(date +%Y%m%d).zip \
  startgpmon orasvc mysqlsvc db2svc send_mail_cli
```

### 连同配置一起打包（适合首次部署）

```bash
zip -r gpmon-deploy-$(date +%Y%m%d).zip \
  startgpmon orasvc mysqlsvc db2svc send_mail_cli \
  install_cfg/ messages.db
```

### 解压与权限

```bash
unzip gpmon-bin-20250615.zip -d /workspace/gpmon/
chmod +x /workspace/gpmon/{startgpmon,orasvc,mysqlsvc,db2svc,send_mail_cli}
```

未安装 zip 时可用 `tar czvf gpmon-bin-$(date +%Y%m%d).tar.gz ...` 代替。

**注意**：二进制需在目标操作系统上编译（Linux amd64 编出的文件不能在 Windows 上运行）。

---

## 部署与运行
### Supervisor 配置

配置模板位于 `install_cfg/supervisord.d/`：

| 文件 | 程序 | 命令示例 |
|------|------|----------|
| gpmon.ini | gRPC 服务 | `/workspace/gpmon/startgpmon` |
| orasvc.ini | Oracle 客户端 | `/workspace/gpmon/orasvc <server_ip>` |
| mysqlsvc.ini | MySQL 客户端 | `/workspace/gpmon/mysqlsvc <server_ip>` |
| db2svc.ini | DB2 客户端 | `/workspace/gpmon/db2svc <server_ip>` |

部署步骤：

```bash
cp install_cfg/supervisord.d/*.ini /etc/supervisord.d/

# 修改以下内容：
# 1. command 中的路径和服务器 IP
# 2. environment 中的 LD_LIBRARY_PATH（Oracle + DB2 库路径）

supervisorctl update
supervisorctl status
```

**重要**：Supervisor 不会继承 shell 环境变量，必须在 ini 中显式设置：

```ini
environment=LD_LIBRARY_PATH="/workspace/gpmon/local/clidriver/lib:/instantclient"
```

### 日志位置

```
/workspace/gpmon/log/gpmon.log          # 服务端日志
/workspace/gpmon/log/orasvc.error.log   # Oracle 客户端错误日志
/workspace/gpmon/log/mysqlsvc.error.log
/workspace/gpmon/log/db2svc.error.log
```

### 防火墙

监控客户端需能访问服务端 gRPC 端口 **5051**。

---

## 运维脚本

```
scripts/
├── build.sh               # 统一编译
├── build-mail-tool.sh     # 编译邮件工具
├── send-daily-report.sh   # 日报生成与发送
├── db-maintenance.sh      # 数据库维护
├── setup-maintenance.sh   # 安装定时任务
└── setup-logrotate.sh     # 安装日志轮转
```

### 一键安装运维环境

```bash
sudo ./scripts/setup-maintenance.sh --setup    # 定时任务 + 日志目录
sudo ./scripts/setup-logrotate.sh              # 日志轮转
```

安装后的定时任务（`/etc/cron.d/gpmon-maintenance`）：

```
# 每天 08:00 发送监控日报
0 8 * * * root /workspace/gpmon/scripts/send-daily-report.sh --send

# 每天 02:00 数据库维护（清理 + 备份）
0 2 * * * root /workspace/gpmon/scripts/db-maintenance.sh --full --force
```

### 日常操作

```bash
# 查看系统状态
./scripts/db-maintenance.sh --status

# 测试邮件
./scripts/send-daily-report.sh --test-mail

# 手动发送日报
./scripts/send-daily-report.sh --send

# 手动维护
./scripts/db-maintenance.sh --full --force
```

### 运维日志

```
/var/log/gpmon-maintenance.log     # 维护脚本日志
/var/log/gpmon-daily-report.log    # 日报脚本日志
/workspace/gpmon/log/*.log         # 应用日志（logrotate 管理）
```

---

## 数据库配置

使用 SQLite 数据库 `messages.db`，主要表：

| 表名 | 说明 |
|------|------|
| client_info | 被监控数据库连接配置 |
| dbmonsql | 监控 SQL 配置 |
| check_map | 数据库与监控项关联 |
| mail_cfg | 邮件发送配置 |
| check_result | 当前监控结果 |
| check_his | 监控历史记录 |

### mail_cfg 邮件配置表

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | INTEGER | 自增主键 | 配置记录 ID |
| sender | VARCHAR(30) | 是 | 发件人邮箱 |
| recipient | VARCHAR(100) | 是 | 收件人，多个地址用逗号/分号分隔 |
| cc | VARCHAR(100) | 否 | 抄送，多个地址用逗号/分号分隔 |
| smtp_server | VARCHAR(30) | 是 | SMTP 服务器 |
| smtp_port | INTEGER | 是 | SMTP 端口 |
| smtp_user | VARCHAR(30) | 是 | SMTP 用户名 |
| smtp_password | VARCHAR(30) | 是 | SMTP 密码 |
| isenable | INTEGER | 默认 0 | 1=启用，程序读取第一条启用记录 |

告警邮件（`utils/mail.go`）和日报邮件（`send_mail_cli`）均支持 `recipient`、`cc` 字段填写多个地址，逗号或分号分隔。

### 添加监控数据库

```sql
-- 添加数据库连接
INSERT INTO client_info VALUES(11,'1.1.1.191',1521,'ORACLE','racdb','jason','oracle',1);

-- 关联监控项
INSERT INTO check_map VALUES(11, 101, 1);

-- 启用
UPDATE client_info SET isenable=1 WHERE id=11;
UPDATE check_map SET isenable=1 WHERE client_id=11;
```

### Oracle 监控用户权限

```sql
GRANT CONNECT, SELECT_CATALOG_ROLE TO eamon IDENTIFIED BY "password";
```

---

## 本地开发调试

### 启动 gRPC 服务

```bash
cd mon-server
CGO_ENABLED=1 go run main.go
```

### 启动监控客户端

```bash
# 另开终端，<server_ip> 为 gRPC 服务地址
cd mon-client-ora && go run startora.go <server_ip>
cd mon-client-mysql && go run startmysql.go <server_ip>
cd mon-client-db2 && go run startdb2.go <server_ip>
```

---

## Proto 编译

仅在修改 `grpc/proto/dbstatus.proto` 后需要重新生成：

```bash
cd gpmon/
protoc -I grpc -I . \
  --go_out=grpc --go_opt=paths=source_relative \
  --go-grpc_out=grpc --go-grpc_opt=paths=source_relative \
  grpc/proto/dbstatus.proto
```

生成文件：`grpc/proto/dbstatus.pb.go`、`grpc/proto/dbstatus_grpc.pb.go`

---

## 常见问题

### 脚本执行报错 `/bin/bash^M` 或 `$'\r': command not found`

Windows 换行符（CRLF）导致。处理：

```bash
sed -i 's/\r$//' scripts/*.sh install_cfg/bash_profile install_cfg/supervisord.d/*.ini
```

### 编 startgpmon 时报 sqlcli1.h: No such file or directory

旧版代码会将 DB2 驱动编入服务端。确认 `db/db2db.go` 首行为 `//go:build db2`，并使用 `./scripts/build.sh --server` 单独编译服务端。

### 全量编译没有生成 db2svc

常见原因：

1. `install_cfg/bash_profile` 加载失败（CRLF）——编译在客户端阶段中断
2. 未配置 `IBM_DB_HOME` 或缺少 `include/sqlcli1.h`——脚本在 DB2 检查处退出
3. 使用了 `--server` 或 `--skip-db2`

检查：

```bash
echo "IBM_DB_HOME=$IBM_DB_HOME"
ls $IBM_DB_HOME/include/sqlcli1.h
```

### git pull 提示本地修改冲突

手动改过与远程相同的文件时：

```bash
git checkout -- db/db2db.go db/mysqldb.go db/oradb.go
rm -f db/checkitem.go scripts/build.sh   # 若为未跟踪的冲突文件
git pull
```

### godror 编译错误（undefined: VersionInfo 等）

```bash
go env -w CGO_ENABLED=1
go get github.com/godror/godror@latest
```

### DPI-1047: Cannot locate Oracle Client library

检查 Instant Client 是否安装，环境变量 `TNS_ADMIN`、`LD_LIBRARY_PATH` 是否指向正确目录。Supervisor 需在 ini 中配置 `environment=LD_LIBRARY_PATH=...`。

### libdb2.so.1: cannot open shared object file

Supervisor 启动时未加载 DB2 库路径，在 ini 中添加：

```ini
environment=LD_LIBRARY_PATH="/workspace/gpmon/local/clidriver/lib:/instantclient"
```

### 客户端连接被拒绝（dial tcp :5051: connection refused）

确认 gRPC 服务已启动，防火墙放行 5051，客户端 IP 参数正确。Supervisor 下手动 `go run` 正常但 supervisor 失败，通常是环境变量未配置。

### go mod tidy 超时

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### supervisor.sock no such file

修改 `/etc/supervisord.conf`，将 socket 路径从 `/tmp` 改为 `/var/log`。

### 启动后无日志也无报错

检查是否已添加 `check_map` 关联监控项；仅有 `client_info` 记录不够。

### sqlite3 终端中文乱码（Windows）

```bat
chcp 65001
```

---

## Git 操作参考

```bash
# 获取更新
git pull

# 服务器 SSH 密钥未配置时，参考 GitHub Settings → SSH keys
# 或改用 HTTPS：git remote set-url origin https://github.com/myz-git/gpmon.git

# 提交
git add .
git commit -m "描述"
git push

# 强制覆盖本地（不保留本地修改）
git reset --hard && git pull

# 保留本地修改后拉取
git stash push -u -m "backup" && git pull && git stash drop
```

---

## 模块说明

| 路径 | 说明 |
|------|------|
| db/checkitem.go | 监控检查项公共类型 |
| db/oradb.go | Oracle 监控（`//go:build oracle`） |
| db/mysqldb.go | MySQL 监控（`//go:build mysql`） |
| db/db2db.go | DB2 监控（`//go:build db2`） |
| db/serverdb.go | SQLite 交互、邮件发送状态管理 |
| grpc/server.go | gRPC 服务端实现，调用 `utils.SendEmail` 发告警 |
| mon-client-*/start*.go | 各数据库监控客户端入口 |
| mon-server/main.go | gRPC 服务入口，监听 5051 |
| utils/mail.go | 告警邮件发送，支持多收件人/抄送 |
| send_mail_cli.go | 日报 HTML 邮件发送工具 |

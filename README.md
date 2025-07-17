# 环境准备

   gpmon项目是基于gRPC协议微服务框架的数据库监控系统,  被监控的数据库有Oracle,Mysql,DB2,PostgreSQL等,以注册服务的形式来启用某一类数据库的监控;

## 项目结构

```
GPMON
├── db  
│   ├── db2db.go
│   ├── mysqldb.go
│   ├── oradb.go
│   └── serverdb.go
├── go.mod
├── go.sum
├── grpc
│   ├── client.go
│   ├── proto
│   │   ├── dbstatus_grpc.pb.go
│   │   ├── dbstatus.pb.go
│   │   └── dbstatus.proto
│   └── server.go
├── log
├── messages.db
├── mon-client-db2
│   └── startdb2.go
├── mon-client-mysql
│   └── startmysql.go
├── mon-client-ora
│   └── startora.go
├── mon-server
│   └── main.go
├── README.md
└── utils
    └── mail.go

```

## 文件说明

### db/oradb.go|db2db.go|mysqldb.go

 实现各数据库的告警项执行

### db/serverdb.go:

它包含了与SQLite数据库交互的功能，例如初始化数据库连接、创建表、插入消息、获取客户端信息、更新客户端状态，以及管理邮件发送状态。

### grpc/server.go:

这个文件定义了gRPC服务器端的实现，包括处理数据库状态更新的服务和客户端信息获取的服务

### mon-client-ora/startora.go:

这个文件是一个Oracle监控服务入口点，使用gRPC与服务器通信，检索客户端配置信息，执行数据库检查，并将检查结果发送回服务器。

其他startmysql.go,startdb2.go 功能相同;

### mon-server/main.go:

服务器的主程序，gRPC服务器的入口点,它监听TCP端口5051，并使用 grpc/server.go 中的逻辑处理来自客户端的请求，目前注册了两个服务：数据库状态服务和客户信息服务。

### grpc/proto/dbstatus.proto:

gRPC的消息和服务定义。
这个文件定义了与gRPC服务相关的协议，其中包括了用于数据库状态和客户端信息的服务和消息类型

### utils/mail.go

这个文件定义了发送电子邮件的功能，它通过检查ShouldSendEmail函数来决定是否应该发送邮件，并在邮件发送后更新数据库中的状态。

### grpc/client.go (可选):

如果您的客户端gRPC逻辑相对复杂，并且您想将其与 mon-client/main.go 分开，可以在这里放置客户端的gRPC逻辑。

grpc/client.go 和 grpc/server.go 包含具体的gRPC通信逻辑。例如，client.go 可能有一个函数，它接受一个数据库状态，并发送给gRPC服务器；而server.go可能有一个函数，它启动gRPC服务器，等待客户端消息，然后执行相应的操作。

README.md 包含项目的文档，描述了如何设置、构建和运行项目。

## 安装Oracle数据库Go驱动

```
cd mon
go mod init mon
go get github.com/godror/godror

下载oracle客户端, 
https://www.oracle.com/database/technologies/instant-client/winx64-64-downloads.html
#for linux:
#https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
解压到D:\Oracle\instantclient_19_20
设置环境变量:
TNS_ADMIN 指向  D:\Oracle\instantclient_19_20
LD_LIBRARY_PATH 同样指向  D:\Oracle\instantclient_19_20

注1:
执行时报:
Received status: ERROR, Details: ORA-00000: DPI-1047: Cannot locate a 64-bit Oracle Client library: "The specified module could not be found". 
需要设置LD_LIBRARY_PATH, 然后编译 db.go时,SET CGO_ENABLED=1 启用交叉编译

注2:
执行时报:
godror WARNING: discrepancy between DBTIMEZONE ("+00:00"=0) and SYSTIMESTAMP ("+08:00"=800) 
可以在DSN中加入timezone=UTC ,如:
const DSN = `user="username" password="password" connectString="ip:1521/oradb timezone=UTC"`
```

### 安装 SQLite3 的 Go 驱动

```
go get github.com/mattn/go-sqlite3
```



## 安装Db2驱动

### For WIN:

https://github.com/ibmdb/go_ibm_db

```

go get github.com/ibmdb/go_ibm_db
go install github.com/ibmdb/go_ibm_db/installer@v0.4.3

cd  C:\Program Files\Go\pkg\mod\github.com\ibmdb\go_ibm_db@v0.4.3\installer
SET GOOS=windows
SET GOARCH=amd64
go run setup.go
setenvwin.bat
注: 会自动下载db2cli64.dll到 C:\Program Files\Go\pkg\mod\github.com\ibmdb\clidriver\bin\

配置环境变量:
IBM_DB_HOME=C:\Users\uname\go\src\github.com\ibmdb\clidriver
PATH=%PATH%;%IBM_DB_HOME%\bin

```

**执行测试验证**:

```
cd gpmon
go run db2test.go
Connected to DB2 version: DB2 v9.7.0.0

如果报找不到db2cli64.dll , 检查环境变量是否设置
看是否能执行db2cli  
```



## 安装MYSQL驱动

```
go get -u github.com/go-sql-driver/mysql
```



# 数据库设计

````
使用SQLITE3数据库的messages.db ,含如下表:
client_info 数据库连接配置表,记录需要监控的数据库连接地址及凭证信息
dbmonsql 监控SQL配置表
check_map 数据库和监控MAP表
mail_cfg  邮件配置表
check_result 数据库监控检查表(只保留当前数据),记录每一个数据库每一项监控的检查结果以及邮件发送时间
check_his  数据库监控检查历史表, 类似check_result,它会保留所有监控记录
详细见<<数据库设计.txt>>
````

## 目标Oracle库监控用户及权限

```
grant connect , SELECT_CATALOG_ROLE to eamon identified by "Welcome1#123";
```

## 添加监控数据库

client_info 中插入数据库;

```
INSERT INTO client_info VALUES(11,'1.1.1.191',1521,'ORACLE','racdb','jason','oracle',1);
INSERT INTO client_info VALUES(12,'1.1.1.100',1521,'ORACLE','oradb','jason','oracle',1);
```

check_map中插入数据库和监控的关联;

```
sqlite> insert into check_map values(11,101,1);
sqlite> insert into check_map values(12,101,1);
sqlite> select * from v_map;
```

控制数据库是否启用:

```
update client_info set isenable=1 where id=NN;
update check_map set isenable=1 where client_id=NN;
```



# proto文件

建立 gpmon/grpc/proto/dbstatus.proto 文件

```
//grpc/proto/dbstatus.proto
syntax = "proto3";
option go_package = "./proto";
package proto;

import "google/protobuf/timestamp.proto";

service DatabaseStatusService {
    rpc SendStatus (DatabaseStatus) returns (DatabaseStatusResponse);
}

service ClientInfoService {
    rpc GetClientInfo (ClientInfoRequest) returns (ClientInfoResponse);
}

message DatabaseStatus {
    string ip = 1;
    int32 port = 2; // 确保与ClientInfo中的port类型匹配
    string dbtype = 3;
    string dbnm = 4;
    string checkNm = 5;  // 检查名称
    string checkResult = 6; // 检查结果
    string details = 7;
    google.protobuf.Timestamp timestamp = 8;
}

message DatabaseStatusResponse {
    string message = 1;
}

message ClientInfoRequest {
    string DbType = 1; // 使用dbtype作为请求参数
}

message ClientInfoResponse {
    repeated ClientInfo clientInfos = 1;
}

message ClientInfo {
    string Ip = 1;
    int32 Port = 2;
    string DbType = 3;
    string DbName = 4;
    string DbUser = 5;
    string UserPwd = 6;
    bool IsEnable = 7;
    int32 id = 8; 
}

```

## 编译proto文件

参见"安装 protoc和protoc-gen-go.md"内容,将google目录放在gpmon\grpc\下

```
cd gpmon/
protoc -I grpc -I . --go_out=grpc --go_opt=paths=source_relative --go-grpc_out=grpc --go-grpc_opt=paths=source_relative grpc/proto/dbstatus.proto
```

会在gpmon/grpc/proto目录(和dbstatus.proto同级) 生成两个GO文件:  dbstatus.pb.go dbstatus_grpc.pb.go

#还一种, 在gpmon/grpc下执行,但是只生成grpc/proto/dbstatus.pb.go 一个文件, 不知道是否可用

```
#protoc --go_out=plugins=grpc:. proto/dbstatus.proto
```



# 本地测试运行

## SERVER端启动 gRPC 服务器

```
cd  mon-server
SET CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=amd64
go run main.go
```

## 启动 gRPC 客户端

在另一个终端中，启动oracle监控,注意防火墙

```
cd mon-client-ora
go run startora.go 1.1.1.1
#1.1.1.1 为serverip
```

分别在其他终端窗口，启动MYSQL及DB2监控

```
cd mon-client-db2
go run startdb2.go 1.1.1.1
cd mon-client-mysql
go run startmysql.go 1.1.1.1
```

# LINUX端部署程序

## Linux 安装go

```
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
#GOPROXY=https://mirrors.aliyun.com/goproxy/

验证网络:
GO111MODULE=on GOPROXY=https://goproxy.cn,direct go list -m -json -versions golang.org/x/text@latest

```

## Linux安装oracle instantclient

```
#for linux:
#https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
解压到  /instantclient
设置环境变量:
vi .bash_profile
export TNS_ADMIN=/instantclient
export LD_LIBRARY_PATH=/instantclient

ldconfig 
```

## Linux安装MYSQL驱动

```
go get -u github.com/go-sql-driver/mysql
```

### Linux安装 DB2 驱动

https://github.com/ibmdb/go_ibm_db#how-to-install-in-linuxmac

```
go get github.com/ibmdb/go_ibm_db
go install github.com/ibmdb/go_ibm_db/installer@v0.4.3
cd /root/go/pkg/mod/github.com/ibmdb/go_ibm_db@v0.4.3/installer
go run setup.go
cd /root/go/pkg/mod/github.com/ibmdb
cp -r clidriver /workspace/gpmon/local/


vi  ~/.bash_profile 
export TNS_ADMIN=/instantclient
export LD_LIBRARY_PATH=/instantclient

export IBM_DB_HOME=/workspace/gpmon/local/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib 
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$IBM_DB_HOME/lib

source  ~/.bash_profile
echo $LD_LIBRARY_PATH

```



## 部署源程序gpmon

```
mkdir /workspace/
上传gpmon.zip 解压
cd /workspace/gpmon
mkdir log
chmod -R 755 /gpmon

rm go.mod
rm go.sum
go mod init mon
go mod tidy
```

## Linux端编译程序

```
#获取最新代码
cd  /workspace/gpmon/
git status
git pull

supervisorctl stop all

#编译
cd /workspace/gpmon/mon-server
GOOS=linux GOARCH=amd64 go build -o startgpmon
mv startgpmon ../

cd /workspace/gpmon/mon-client-ora
GOOS=linux GOARCH=amd64 go build -o orasvc
mv orasvc ../

cd /workspace/gpmon/mon-client-db2
GOOS=linux GOARCH=amd64 go build -o db2svc
mv db2svc ../

cd /workspace/gpmon/mon-client-mysql
GOOS=linux GOARCH=amd64 go build -o mysqlsvc
mv mysqlsvc ../




单独上传更新messages.db
mv messages.db messages.db.bk
ftp messages.db to gpmon/

#启动服务
supervisorctl start all

#查看日志
log/gpmon.log 等
```

## 部署SUPERVISOR

```
yum reinstall  gcc install openssl-devel bzip2-devel expat-devel gdbm-devel readline-devel zlib-devel libffi-devel  --downloadonl --downloaddir=/opt/soft/pkg

yum install supervisor --downloadonl --downloaddir=/opt/soft/pkg
#yum reinstall python-meld3  --downloadonl --downloaddir=/opt/soft/pkg

#修改/etc/supervisord.conf
将里面指向/tmp 改为/var/log
grep "/tmp" /etc/supervisord.conf |grep -v "^;"

#修改 /usr/bin/supervisord及/usr/bin/supervisorctl  第一行python 为python2
vi /usr/bin/supervisorctl
vi /usr/bin/supervisord
python改为python2

配置/etc/supervisord.d/gpmon.ini
详细文件见:
/etc/supervisord.d/db2svc.ini  gpmon.ini  mysqlsvc.ini  orasvc.ini

需要修改db2svc.ini,mysqlsvc.ini orasvc.ini中的IP地址,替换位gpmon服务器地址;
command=/workspace/gpmon/orasvc 1.1.1.1

#配置完成后, 加载主配置文件
#supervisord -c /etc/supervisord.conf
systemctl stop supervisord.service
systemctl start supervisord.service
systemctl status supervisord.service

#更新配置, 每次修改都需要更新
supervisorctl update
supervisorctl status

systemctl enable supervisord
systemctl start supervisord


```

### 常见supervisor错误:

**问题:**

手动执行 \gpmon\go run main.go 时正常;
但使用 supervisorctl start  时马上退出:
BACKOFF   Exited too quickly
查看 err.log显示 :
/workspace/gpmon/startgpmon: error while loading shared libraries: libdb2.so.1: cannot open shared object file: No such file or directory

**解决:**  在所有ini中增加如下指定,

environment=LD_LIBRARY_PATH="/workspace/gpmon/local/clidriver/lib:/instantclient"

```
vi /etc/supervisord.d/gpmon.ini
[program:gpmon]
; 命令执行的目录
;directory=/workspace/gpmon
directory=/
; 运行程序的命令,要用绝对路径
command=/workspace/gpmon/startgpmon
environment=LD_LIBRARY_PATH="/workspace/gpmon/local/clidriver/lib:/instantclient"


supervisorctl update
supervisorctl status

```



**其它报错问题**

```
4.1 报错 unix:///var/run/supervisor.sock no such file
方案1:
（1）先停止supervisor
systemctl sotp supervisor.service
1
（2）然后查看是否有supervisord进程没有结束
ps ax |grep supervisor
1
如果有进程没结束就手动杀死
kill -9 [pid]
1
（3）使用以下命令，重新加载配置文件。整个服务会自动启动
supervisord -c /etc/supervisor/supervisord.conf
#或
/usr/bin/python2 /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
1
2
3
方案2：
（1）问题原因是文件被系统删除，可以手动创建sock文件。然后重新加载配置
sudo touch /var/run/supervisor.sock
sudo chmod 777 /var/run/supervisor.sock
supervisord -c /etc/supervisor/supervisord.conf
1
2
3
（2）然后查看supervisord的进程，如果有重复进程需要进行查杀，否则cpu占用率可能会比较高。
ps ax |grep supervisor
#如果有多余进程没结束就手动杀死
kill -9 [pid]
1
2
3
4.2 报错 Error: Another program is already listening on a port that one of our HTTP servers is configured to use. Shut this program down first before starting supervisord
（1）尝试关闭配置文件中包含的所有服务

supervisorctl status
supervisorctl stop [programName]
1
2
（2）如果关闭不掉，试用命令ps ax |grep supervisor 查看pid，杀掉所有查看到的进程。

ps ax |grep supervisor
kill -9 [pid]
1
2
（3）如果还是报错，取消socket的软连接。注意，下方的路径是在supervisord.conf中声明的。

unlink /var/run/supervisor.sock
unlink /tmp/supervisor.sock 
supervisord -c /etc/supervisor/supervisord.conf
1
2
3
4.3 cpu占用率过高的问题
(1) 使用以下命令查看是否有重复启动的配置文件

ps -ax |grep supervisor
1
(2) 使用kill命令杀掉多余的进程，只保留以一个

ps ax |grep supervisor
#如果有多余进程没结束就手动杀死
kill -9 [pid]
1
2
3
(3) 如果还是占用率过高，使用以下命令编辑配置文件，把http服务注释掉

vim /etc/supervisor/supervisor.conf
1
4.4 报错 [line 57]: ‘json module not found, using jsonujson module not found, using jsonujson module not found, using json\n’
此问题不是json模块找不到，是配置文件格式不对，仔细检查或重写配置文件就能搞定。

https://blog.csdn.net/dorlolo/article/details/119336687
```





# 用户端部署GPMON:

```
--client
systemctl stop firewalld
systemctl disable firewalld

unzip instantclient-basic-linux.x64-19.20.0.0.0dbru.zip 
mv instantclient_19_20/  /opt/instantclient
chmod -R 775 /opt/instantclient

cp install_cfg/bash_profile  ~/.bash_profile 

. ~/.bash_profile 
echo "/opt/instantclient" >> /etc/ld.so.conf
ldconfig

--server
systemctl stop firewalld
systemctl disable firewalld

```

# 运维脚本

```
gpmon/
├── send_mail_cli.go              # 邮件工具源码
├── send_mail_cli                 # 邮件工具可执行文件
├── scripts/ # 关键运维脚本
│ ├── build-mail-tool.sh # ✉️ 编译邮件工具
│ ├── db-maintenance.sh # 🛠️ 数据库维护脚本
│ ├── send-daily-report.sh # 📧 日报发送脚本
│ ├── setup-logrotate.sh # 🔄 日志轮转安装脚本
│ └── setup-maintenance.sh # ⚙️ 定时任务&环境安装脚本
```

### 一键部署流程

```
cd /workspace/gpmon

# 1. 编译邮件工具
./scripts/build-mail-tool.sh

# 2. 设置日志轮转（独立功能）
./scripts/setup-logrotate.sh

# 3. 设置运维定时任务
./scripts/setup-maintenance.sh   # 安装 / 更新定时任务、环境变量

# 4. 验证
# 测试邮件配置
./scripts/send-daily-report.sh --test-mail

# 手动触发数据库维护
./scripts/db-maintenance.sh --full --force

# 手动发送日报
./scripts/send-daily-report.sh --send
```

### 脚本功能说明

#### 1. db-maintenance.sh - 核心运维脚本

主要功能：

- 数据库维护（清理过期数据、备份、优化）

- 监控日报生成和发送

- 系统状态查看

- 备份管理



###  使用logrotate定期清理日志

#### setup-logrotate.sh - 日志轮转设置

功能： 配置系统logrotate服务自动管理GPMon日志文件

使用方法：

```
# 设置日志轮转（需要root权限）
sudo ./scripts/setup-logrotate.sh

# 手动测试日志轮转
sudo logrotate -f /etc/logrotate.d/gpmon

# 查看轮转配置
sudo logrotate -d /etc/logrotate.d/gpmon
```

轮转策略：

- 每天轮转一次

- 保留7天的日志文件

- 文件超过10M立即轮转

- 自动压缩旧日志

- 自动删除30天前的日志

#### build-mail-tool.sh - 邮件工具编译

功能： 编译邮件发送工具，支持HTML邮件发送

使用方法：

```
# 编译邮件工具
./scripts/build-mail-tool.sh

# 编译完成后会自动测试邮件配置
```

### 自动化定时任务

设置完成后，系统会自动配置以下定时任务：

```
cat /etc/cron.d/gpmon*

# 每天早上8点: 发送监控状态日报
0 8 * * * root /workspace/gpmon/scripts/send-daily-report.sh --send-report

# 每天凌晨2点: 执行维护任务（数据清理+备份）
0 2 * * * root /workspace/gpmon/scripts/db-maintenance.sh --daily-tasks --force
```

### 日常运维操作

#### 查看系统状态

```
./scripts/db-maintenance.sh --status
```

显示内容包括：

- 📊 数据库信息（大小、记录数、可清理数据）

- 📈 当前监控状态（正常/错误/警告统计）

- 💾 备份信息（备份数量、大小、最新备份时间）

- 📧 邮件工具状态

- ⏰ 定时任务配置状态

#### 手动维护操作

```
# 完整维护（推荐）
./scripts/db-maintenance.sh --daily-tasks

# 单独操作
./scripts/db-maintenance.sh --clean-db      # 只清理数据
./scripts/db-maintenance.sh --backup-db     # 只备份数据库
```

#### 备份管理

```
# 查看备份列表
./scripts/gpmon-maintenance.sh --list-backups

# 测试备份完整性
./scripts/gpmon-maintenance.sh --test-backups

# 清理过期备份
./scripts/gpmon-maintenance.sh --cleanup-backups
```

### 日志文件位置

```
# 运维日志
/var/log/gpmon-maintenance.log

# 邮件报告日志
/var/log/gpmon-daily-report.log

# 应用日志（由logrotate管理）
/workspace/gpmon/log/*.log

# 日志轮转日志
/var/log/gpmon-rotation.log
```



### 重要提醒

- 首次部署：必须按顺序执行编译邮件工具 → 设置日志轮转 → 设置运维任务
- 权限要求：setup-logrotate.sh 和 gpmon-maintenance.sh --setup 需要sudo权限
- 邮件配置：确保数据库中 mail_cfg 表有正确的邮件服务器配置
- 磁盘空间：定期检查 backup 目录的磁盘空间
- 功能分离：日志轮转由系统logrotate服务管理，数据库维护由定时任务管理




# GIT

## --初始化一个新的Git仓库

```
cd D:\Workspace\gpmon
git init
```

## --添加文件到Git仓库

```
#WIN环境设置 CRLF(\r\n)上传后自动转为LF(\n)
git config --global core.autocrlf true

#
git config --global user.email "yongzhi.m@gmail.com"
git config --global user.name "Jason Ma"
#去除M
git config --global core.whitespace cr-at-eol

#添加本地文件
git add .
```



## 获取更新

```
git pull
git status
红色表示未commit , 绿色表示未push
```

## 获取更新--强制覆盖本地

```
git reset --hard
git pull
不保留本地的修改，直接覆盖
```

## 获取更新--保存本地(放入缓存中)

```
git stash
git pull 
git stash pop
解析：
git stash: 将改动藏起来
git pull:用新代码覆盖本地代码
git stash pop: 将刚藏起来的改动恢复
这样操作的效果是在最新的仓库代码的基础仍保留本地的改动
```

## 提交更改

```
git add .
git commit -m "release 3.1 , 增加通过数据库和监控MAP表check_map来控制数据库监控项配置"
```

## 上传更改

```
git push
git status
```



## 其它操作

### 创建分支（可选）

Git允许您创建分支来隔离功能开发或修复错误。例如，如果您想添加一个新功能，可以创建一个新分支

```
git checkout -b new-feature-branch
```

### 查看更改历史

```
git log
```

### 回退更改

如果需要，您可以回退到先前的提交。首先，使用`git log`找到您想要回退到的提交的哈希值，然后执行

```
git checkout [commit_hash]
替换[commit_hash]为您从git log中获得的实际哈希值。
```



# Issue

## issue1: godror错误

```
go run main.go
# github.com/godror/godror
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:558:19: undefined: VersionInfo
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:559:19: undefined: VersionInfo
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:560:10: undefined: StartupMode
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:561:11: undefined: ShutdownMode
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:563:31: undefined: Event
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:563:42: undefined: SubscriptionOption
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:563:64: undefined: Subscription
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:564:31: undefined: ObjectType
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:565:59: undefined: Data
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:566:28: undefined: DirectLob
C:\Program Files\Go\pkg\mod\github.com\godror\godror@v0.40.3\orahlp.go:566:28: too many errors

解决:
确保godror正确安装
go get github.com/godror/godror@latest
设置变量
go env -w CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=amd64
go run main.go
```

## issue2: gcc: error: unrecognized

```
D:\Workspace\gpmon\mon-client>go run main.go
# runtime/cgo
gcc: error: x86_64: No such file or directory
gcc: error: unrecognized command line option '-arch'; did you mean '-march='?

解决:
SET GOOS=windows
// SET GOARCH=amd64
go run main.go
```



## issue3:  ORA-00000: DPI-1047

```
Details: ORA-00000: DPI-1047: Cannot locate a 64-bit Oracle Client library: "libclntsh.so: cannot open shared object file: No such file or directory". See https://oracle.github.io/odpi/doc/installation.html#linux for help
解决:
检查instantclient是否安装
检查环境变量~/.bash_profile及/etc/profile 中TNS_ADMIN及LD_LIBRARY_PATH是否设置
echo $TNS_ADMIN
/instantclient
cd $LD_LIBRARY_PATH
/instantclient



```

## issue4: go mod tidy问题

```
go mod tidy时报
go mod ...github.com/godror/godror i/o timeout dial tcp: lookup...
解决:
检查网络是否通畅
ping goproxy.cn
或者用ip代替 goproxy.cn
如果外网不通,检查 /etc/resolv.conf
# Generated by NetworkManager
nameserver 198.18.0.2
nameserver 192.168.31.1
可能要重启下虚机
```

## issue5: sqlite3 终端中文乱码问题

```
chcp 65001        （将编码方式改为UTF-8）
chcp 936            （将编码方式改回GBK）
```

## issue6: transport: Error while dialing: dial tcp

```
客户端程序执行显示
Failed to retrieve configurations: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp 1.1.1.9:5051: connect: connection refused"

解决:
手动执行 go run 试试, 疑似环境变量问题

```

## issue7: unix:///tmp/supervisor.sock no such file

```
supervisorctl 提示:
unix:///tmp/supervisor.sock no such file

原因为tmp/下文件被自动清理


解决:
#修改/etc/supervisord.conf
将里面指向/tmp 改为/var/log
grep "/tmp" /etc/supervisord.conf |grep -v "^;"
```

### issu8: 启动后无日志输出也没报错

添加数据库client_info后,没有添加check_map;



# TMP

```

```


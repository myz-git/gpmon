# 环境准备

## 文件结构

```
mon/
│
├── cmd/
│   ├── mon-client/
│   │   └── main.go         // 客户端主程序
│   │
│   └── mon-server/
│       └── main.go         // 服务器端主程序
│
├── db/
│   └── db.go               // 数据库相关操作
│
├── grpc/
│   ├── proto/
│   │   └── dbstatus.proto  // 定义gRPC消息和服务
│   │
│   ├── server.go           // 服务器端的gRPC处理逻辑
│   │
│   └── client.go           // 客户端的gRPC处理逻辑（可选，如果客户端逻辑较复杂的话）
│
└── go.mod                  // Go模块依赖文件
│
└── README.md           # 项目文档

cmd/mon-client/main.go:
这个是客户端的主程序，它主要负责从数据库检索状态，并通过gRPC将状态发送到服务器。

cmd/mon-server/main.go:
服务器的主程序，它监听指定端口，并使用 grpc/server.go 中的逻辑处理来自客户端的请求。

db/db.go:
此文件包含与数据库交互的函数，例如检查数据库状态。

grpc/proto/dbstatus.proto:
gRPC的消息和服务定义。

grpc/server.go:
定义了gRPC服务器逻辑的文件。例如，当服务器收到客户端的状态消息时，它应该怎么做。

grpc/client.go (可选):
如果您的客户端gRPC逻辑相对复杂，并且您想将其与 cmd/mon-client/main.go 分开，可以在这里放置客户端的gRPC逻辑。

cmd/mon-client/main.go 和 cmd/mon-server/main.go 是两个应用的主入口。他们主要负责读取配置、初始化服务和启动服务。
grpc/client.go 和 grpc/server.go 包含具体的gRPC通信逻辑。例如，client.go 可能有一个函数，它接受一个数据库状态，并发送给gRPC服务器；而server.go可能有一个函数，它启动gRPC服务器，等待客户端消息，然后执行相应的操作。
db/db.go 包含与数据库交互的逻辑，例如检查数据库状态的函数。

README.md 包含项目的文档，描述了如何设置、构建和运行项目。
```



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

### 初始化数据库并创建所需的表

我们可以创建一个 `database` 包来处理数据库相关的操作。

**mon/db/serverdb.go**:

```
package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

// InitDB initializes the SQLite3 database and creates tables if they don't exist.
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTable(db)

	return db
}

func createTable(db *sql.DB) {
	createDatabaseStatusTableSQL := `
		CREATE TABLE IF NOT EXISTS database_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(createDatabaseStatusTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// InsertDatabaseStatus inserts the status into the database_status table.
func InsertDatabaseStatus(db *sql.DB, status string) error {
	query := `
		INSERT INTO database_status (status)
		VALUES (?)
	`

	_, err := db.Exec(query, status)
	return err
}

```

# 数据库

```
sqlite3 messages.db
.schema cli%
CREATE TABLE client_info (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip varchar(20),
    port int,
    dbtype varchar(10),
    dbname varchar(20),
    dbuser varchar(10),
    userpwd varchar(10),
    isenable int,
    isok int DEFAULT 1,
    ismail int DEFAULT 0,
    updatetm DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (ip, port, dbtype, dbname)
);

insert into client_info (ip,port,dbtype,dbname,dbuser,userpwd,isenable) values('120.27.245.208',1521,'ORACLE','oradb','jason','oracle',1);
insert into client_info (ip,port,dbtype,dbname,dbuser,userpwd,isenable) values('1.1.1.112',1521,'ORACLE','oradb','jason','oracle',0);
.head on
.mod col
sqlite> select * from client_info;
id  ip              port  dbtype  dbname  dbuser  userpwd  isenable  isok  ismail  updatetm
--  --------------  ----  ------  ------  ------  -------  --------  ----  ------  -------------------
1   120.27.245.208  1521  ORACLE  oradb   jason   oracle   1         1     0       2023-10-27 01:36:43
2   1.1.1.112       1521  ORACLE  oradb   jason   oracle   0         1     0       2023-10-27 01:37:30

.mod insert
select * from client_info;
```





# 程序代码

## grpc/proto/dbstatus.proto

```
syntax = "proto3";
//新版本需要的配置
option go_package = "./proto";
package proto;

import "google/protobuf/timestamp.proto";

service DatabaseStatusService {
    rpc SendStatus (DatabaseStatus) returns (DatabaseStatusResponse);
}

message DatabaseStatus {
    string status = 1;
    string details = 2;
    google.protobuf.Timestamp timestamp = 3;
}

message DatabaseStatusResponse {
    string message = 1;
}


```

## 编译proto文件

参见"安装 protoc和protoc-gen-go.md"内容,将google目录放在mon\grpc\下

```
cd mon/
protoc -I grpc -I . --go_out=grpc --go_opt=paths=source_relative --go-grpc_out=grpc --go-grpc_opt=paths=source_relative grpc/proto/dbstatus.proto
```

会在mon/grpc/proto目录(和dbstatus.proto同级) 生成两个GO文件:  dbstatus.pb.go dbstatus_grpc.pb.go

还一种, 在mon/grpc下执行:

```
protoc --go_out=plugins=grpc:. proto/dbstatus.proto
```

但是只生成grpc/proto/dbstatus.pb.go 一个文件, 不知道是否可用

## grpc/server.go

```
package grpc

import (
	"context"
	"log"
	"mon/grpc/proto"
	"mon/utils"
)

// Server 实现 proto.DatabaseStatusServiceServer 接口。
type Server struct {
	proto.UnimplementedDatabaseStatusServiceServer // 这是gRPC新版本推荐的做法，确保向后兼容。
}

// SendStatus 是 gRPC 接口的实现，用于接收数据库状态。
func (s *Server) SendStatus(ctx context.Context, status *proto.DatabaseStatus) (*proto.DatabaseStatusResponse, error) {
	// 这里你可以根据收到的状态做相应的处理。
	// log.Printf("Received database status: %s at %s", status.Status, status.Timestamp.AsTime().String())
	log.Printf("Received database status: %s ", status.Status)

	// 分析消息，如果是ERROR，发送邮件
	if status.Status == "ERROR" {
		go utils.SendEmail("Mon: Oracle Database Error", status.Details)
	}

	// 在这个示例中，我们只是简单地回应了一个消息，但你可以根据需要进行适当的操作。
	response := &proto.DatabaseStatusResponse{
		Message: "Received status successfully",
	}

	return response, nil
}


```



## grpc/client.go

```
package grpc

import (
	"context"

	pb "mon/grpc/proto"

	"google.golang.org/grpc"
)

// SetupClient 设置gRPC客户端并连接到服务器
func SetupClient(serverAddr string) (pb.DatabaseStatusServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure()) // 使用 WithInsecure() 为了简化示例，实际生产中不推荐
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewDatabaseStatusServiceClient(conn)

	return client, conn, nil
}

// SendStatus 发送数据库状态到服务器
func SendStatus(client pb.DatabaseStatusServiceClient, status, details string) (*pb.DatabaseStatusResponse, error) {
	return client.SendStatus(context.Background(), &pb.DatabaseStatus{
		Status:  status,
		Details: details,
	})
}

```



## cmd/mon-client/main.go

 该文件是客户端的主入口

1. 检索数据库状态。
2. 使用gRPC连接到服务器。
3. 将数据库状态发送到服务器

```
package main

import (
	"context"
	"log"
	"time"

	"mon/db"
	"mon/grpc"
	pb "mon/grpc/proto" // Assuming the generated Go code from the proto is here

	"google.golang.org/grpc"
)

const (
	serverAddress = "1.1.1.1:5051" // 请填写你的mon-server的地址和端口
)

func main() {
	// 1. 从数据库获取状态
	status, details, err := db.CheckDatabaseStatus()
	if err != nil {
		log.Fatalf("Error checking database status: %v", err)
	}

	// 2. 创建gRPC客户端连接
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure()) // 使用 WithInsecure() 为了简化示例，实际生产中不推荐
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDatabaseStatusServiceClient(conn)

	// 打印发送的消息日志
	log.Printf("Sending database status to server: %v", &pb.DatabaseStatus{Status: status, Details: details, Timestamp: ptypes.TimestampNow()})

	// 3. 通过gRPC发送数据库状态
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.SendStatus(ctx, &pb.DatabaseStatus{Status: status, Details: details, Timestamp: ptypes.TimestampNow()})
	if err != nil {
		log.Fatalf("could not send status: %v", err)
	}

	log.Printf("Server response: %s", response.Message)
}

```

## cmd/mon-server/main.go

```
package main

import (
	"log"
	monGrpc "mon/grpc" // 别名用于避免命名冲突
	"mon/grpc/proto"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":5051") // 你可以根据需要更改端口
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// 注册服务时使用正确的处理器
	serverHandler := &monGrpc.Server{}
	// monGrpc.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	proto.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)

	log.Println("Server started on :5051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

```

## db.go

检查数据库状态的程序 

```
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/godror/godror"
)

func checkDatabaseStatus(dsn string) (string, error) {
	db, err := sql.Open("godror", dsn)
	if err != nil {
		return "", fmt.Errorf("cannot connect to database: %v", err)
	}
	defer db.Close()

	// 做一个简单的查询来检查数据库状态
	_, err = db.Exec("SELECT 1 FROM DUAL")
	if err != nil {
		return "ERROR", err
	}

	return "OK", nil
}

func main() {
	dsn := "your_oracle_dsn_here" // 你需要填写适当的DSN字符串
	status, err := checkDatabaseStatus(dsn)
	if err != nil {
		log.Fatalf("Error checking database status: %v", err)
	}
	fmt.Println("Database status:", status)
}

```

## utils/mail.go

```
go get gopkg.in/gomail.v2
```

```
// utils/mail.go
package utils

import (
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(subject, message string) error {
	m := gomail.NewMessage()

	// 设置发件人、收件人、主题和内容
	m.SetHeader("From", "myz@dbhome.cc")
	m.SetHeader("To", "mayz@vastdata.com.cn")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// 设置SMTP服务器的地址、端口和登录凭据
	d := gomail.NewDialer("smtp.qiye.aliyun.com", 587, "myz@dbhome.cc", "Myz_123456")

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
		return err
	}
	return nil
}

```



# 运行

## SERVER端启动 gRPC 服务器

```
cd  mon\cmd\mon-server
SET CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=amd64
go run main.go
```

## 启动 gRPC 客户端

在另一个终端中，进入到 `mon/cmd/mon-client` 目录并运行

```
cd mon/cmd/mon-client
go run main.go
go run main.go -dbtype ORACLE -dbnm mydb

```



# LINUX端部署程序

## 安装oracle instantclient

```
#for linux:
#https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
解压到  /instantclient
设置环境变量:
vi .bash_profile
export TNS_ADMIN=/instantclient
export LD_LIBRARY_PATH=/instantclient
```

## linux 安装go

```
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
#GOPROXY=https://mirrors.aliyun.com/goproxy/

验证网络:
GO111MODULE=on GOPROXY=https://goproxy.cn,direct go list -m -json -versions golang.org/x/text@latest


```

## 部署源程序mon

```
上传mon.zip 解压
cd mon
rm go.mod
rm go.sum
go mod init mon
go mod tidy
```

## Oracle库

```
grant connect , SELECT_CATALOG_ROLE to eamon identified by "Welc";
```



## 运行

## 编译





## 部署:

```
--client
systemctl stop firewalld
systemctl disable firewalld

unzip instantclient-basic-linux.x64-19.20.0.0.0dbru.zip 
mv instantclient_19_20/  /opt/instantclient
chmod -R 775 /opt/instantclient

cat >> ~/.bash_profile <Eof1
export TNS_ADMIN=/opt/instantclient
export LD_LIBRARY_PATH=/opt/instantclient
export ORACLE_HOME=/opt/instantclient
export PATH=$PATH:/opt/instantclient
export LANG=C.UTF-8
Eof1
. ~/.bash_profile 
echo "/opt/instantclient" >> /etc/ld.so.conf
ldconfig

--server
systemctl stop firewalld
systemctl disable firewalld

```



## issue1:

```
Details: ORA-00000: DPI-1047: Cannot locate a 64-bit Oracle Client library: "libclntsh.so: cannot open shared object file: No such file or directory". See https://oracle.github.io/odpi/doc/installation.html#linux for help
检查instantclient是否安装, 环境变量 TNS_ADMIN及LD_LIBRARY_PATH是否设置
```

## issue2:

```
go mod tidy时报
go mod ...github.com/godror/godror i/o timeout dial tcp: lookup...
检查网络是否通畅
ping goproxy.cn
或者用ip代替 goproxy.cn
如果外网不通,检查 /etc/resolv.conf
# Generated by NetworkManager
nameserver 198.18.0.2
nameserver 192.168.31.1
可能要重启下虚机
```




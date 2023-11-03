// mon-server/main.go

package main

import (
	monGrpc "gpmon/grpc" // 别名用于避免命名冲突
	"gpmon/grpc/proto"
	"io"
	"log"
	"net"
	"os"
	"path"
	"runtime"

	"google.golang.org/grpc"
)

func main() {
	/*** 获取项目根路径 ***/
	_, filename, _, _ := runtime.Caller(0)
	wd := path.Dir(path.Dir(filename))
	// log.Printf("wd:  %s", wd)
	/*** End ***/

	/*** 设定log 同时输出到控制台及log文件中 ***/
	f := wd + "/log/" + "server.log"
	logFile, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	/*** End ***/

	lis, err := net.Listen("tcp", ":5051") // 你可以根据需要更改端口
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Server started on :5051")
	grpcServer := grpc.NewServer()

	// 注册服务时使用正确的处理器
	serverHandler := &monGrpc.Server{}
	clientHandler := &monGrpc.ClientInfoServer{}
	// monGrpc.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	// proto.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	proto.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	log.Println("DatabaseStatusService registered")
	proto.RegisterClientInfoServiceServer(grpcServer, clientHandler)
	log.Println("ClientInfoService registered")

	log.Println("Starting gRPC server...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// mon-server/main.go

package main

import (
	monGrpc "gpmon/grpc" // 别名用于避免命名冲突
	"gpmon/grpc/proto"
	"log"
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
	clientHandler := &monGrpc.ClientInfoServer{}
	// monGrpc.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	// proto.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	proto.RegisterDatabaseStatusServiceServer(grpcServer, serverHandler)
	proto.RegisterClientInfoServiceServer(grpcServer, clientHandler)

	log.Println("Server started on :5051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

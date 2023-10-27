package grpc

import (
	"context"

	pb "gpmon/grpc/proto"

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

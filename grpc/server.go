// grpc/server.go
package grpc

import (
	"context"
	"gpmon/db"
	"gpmon/grpc/proto"
	"gpmon/utils"
	"log"
	"time"
)

type Server struct {
	proto.UnimplementedDatabaseStatusServiceServer
}

type ClientInfoServer struct {
	proto.UnimplementedClientInfoServiceServer
}

func (s *Server) SendStatus(ctx context.Context, status *proto.DatabaseStatus) (*proto.DatabaseStatusResponse, error) {
	// Adjust timestamp to consider local timezone
	localTimestamp := status.Timestamp.AsTime().In(time.Local)
	log.Printf("%s,%s,%s,%s,%s", status.Ip, status.Dbtype, status.Dbnm, status.Status, status.Details)

	if status.Status == "ERROR" {
		emailContent := "Alert! Database Error detected.\n"
		emailContent += "IP Address: " + status.Ip + "\n"
		emailContent += "DB Type: " + status.Dbtype + "\n"
		emailContent += "DB Name: " + status.Dbnm + "\n"
		emailContent += "Details: " + status.Details + "\n"
		emailContent += "Timestamp: " + localTimestamp.String()
		log.Printf("emailContent: ", emailContent)
		go utils.SendEmail("Database Monitoring Alert", emailContent)
	}
	err := db.InsertMessage(status.Status, status.Details, status.Ip, status.Dbtype, status.Dbnm, status.Timestamp.AsTime())
	if err != nil {
		log.Printf("Failed to insert message into the database: %v", err)
		return &proto.DatabaseStatusResponse{
			Message: "Failed to insert message into the database.",
		}, err
	}

	return &proto.DatabaseStatusResponse{
		Message: "Message received and saved successfully.",
	}, nil
}

func (c *ClientInfoServer) GetClientInfo(ctx context.Context, req *proto.ClientInfoRequest) (*proto.ClientInfoResponse, error) {
	clients, err := db.GetClientInfos(req.DbType) // Adjusted function name
	if err != nil {
		log.Printf("Failed to retrieve client info from the database: %v", err)
		return nil, err
	}

	var clientInfos []*proto.ClientInfo
	for _, client := range clients {
		clientInfo := &proto.ClientInfo{
			Ip:       client.IP,
			Port:     int32(client.Port),
			DbType:   client.DbType,
			DbName:   client.DbName,
			DbUser:   client.DbUser,
			UserPwd:  client.UserPwd,
			IsEnable: true,
		}
		clientInfos = append(clientInfos, clientInfo)
	}

	return &proto.ClientInfoResponse{ClientInfos: clientInfos}, nil
}

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
	const base_format = "2006-01-02 15:04:05"
	localTimestamp := status.Timestamp.AsTime().In(time.Local)
	log.Printf(":==>%s,%v,%s,%s,%s,%s", status.Ip, status.Port, status.Dbtype, status.Dbnm, status.CheckNm, status.CheckResult)

	// If the status is ERROR, and it's time to send an email (not in cooldown period), send an email
	if status.CheckResult != "OK" && db.ShouldSendEmail(status.Ip, status.Port, status.Dbtype, status.Dbnm, status.CheckNm) {
		emailContent := "Alert! " + status.CheckNm + " " + status.CheckResult + " detected.\n"
		emailContent += "IP Address: " + status.Ip + "\n"
		emailContent += "DB Type: " + status.Dbtype + "\n"
		emailContent += "DB Name: " + status.Dbnm + "\n"
		emailContent += "Details: " + status.Details + "\n"
		emailContent += "Check Time: " + localTimestamp.Local().Format(base_format)
		// log.Printf(emailContent)
		// Assuming the email sending function will return an error if unsuccessful

		if err := utils.SendEmail(status.Ip, status.Dbtype, status.Dbnm, "Database Monitoring Alert", emailContent); err != nil {
			log.Println("邮件发送失败:", err)
		} else {
			//邮件发送成功 更新发送时间
			db.UpdateMailTm(status.Ip, status.Port, status.Dbtype, status.Dbnm, status.CheckNm)
			log.Printf("已发送邮件: %s: %s: %s", status.Ip, status.Dbnm, status.CheckNm)
		}
	}

	err := db.InsertCheckhis(status.Ip, status.Port, status.Dbtype, status.Dbnm, status.CheckNm, status.CheckResult, status.Details)
	if err != nil {
		log.Printf("Failed to insert message into the database: %v", err)
		return &proto.DatabaseStatusResponse{
			Message: "Failed to insert message into the database.",
		}, err
	}

	return &proto.DatabaseStatusResponse{Message: "received and processed"}, nil
}

// GetClientInfo retrieves the client information from the database.
func (c *ClientInfoServer) GetClientInfo(ctx context.Context, req *proto.ClientInfoRequest) (*proto.ClientInfoResponse, error) {
	clients, err := db.GetClientInfos(req.DbType)
	if err != nil {
		log.Printf("Failed to retrieve client info from the database: %v", err)
		return nil, err
	}

	clientInfos := make([]*proto.ClientInfo, 0, len(clients))
	for _, client := range clients {
		clientInfos = append(clientInfos, &proto.ClientInfo{
			Id:      client.ID,
			Ip:      client.IP,
			Port:    client.Port,
			DbType:  client.DbType,
			DbName:  client.DbName,
			DbUser:  client.DbUser,
			UserPwd: client.UserPwd,
			// IsEnable: client.IsEnable == 1,
		})
	}

	return &proto.ClientInfoResponse{ClientInfos: clientInfos}, nil
}

// mon-client/main.go
package main

import (
	"context"
	"fmt"
	"gpmon/db"
	"gpmon/grpc/proto"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getClientInfos retrieves all clients' database configurations from the gRPC server using dbType.
func getClientInfos(serverIP, dbTypeReq string) ([]*proto.ClientInfo, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:5051", serverIP), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := proto.NewClientInfoServiceClient(conn)
	response, err := c.GetClientInfo(context.Background(), &proto.ClientInfoRequest{DbType: dbTypeReq})
	if err != nil {
		return nil, err
	}

	return response.ClientInfos, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <server IP>", os.Args[0])
	}

	serverIP := os.Args[1]
	dbTypeReq := "ORACLE" // Sample DbType to request. Adjust as needed.

	// 1. Retrieve configurations from the gRPC server using dbType
	clientInfos, err := getClientInfos(serverIP, dbTypeReq)
	if err != nil {
		log.Fatalf("Failed to retrieve configurations: %v", err)
	}

	for _, clientInfo := range clientInfos {
		DSN := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s" timezone=UTC`,
			clientInfo.DbUser, clientInfo.UserPwd, clientInfo.Ip, clientInfo.Port, clientInfo.DbName)

		// 2. Perform Database Check
		status, details, err := db.CheckDatabaseStatus(DSN)
		if err != nil {
			log.Printf("Failed to get database status for IP %s. Error: %v", clientInfo.Ip, err)
			// Update CLIENT_INFO with ERROR status
			err = db.UpdateClientInfoOnError(clientInfo.Ip, clientInfo.DbName, clientInfo.DbType)
			if err != nil {
				log.Printf("Failed to update client info for IP %s with ERROR status. Error: %v", clientInfo.Ip, err)
			}
			continue
		}

		// Assume the database status check returns a status of "OK" or "ERROR"
		if status == "OK" {
			// Update CLIENT_INFO with OK status
			err = db.UpdateClientInfoOnSuccess(clientInfo.Ip, clientInfo.DbName, clientInfo.DbType)
			if err != nil {
				log.Printf("Failed to update client info for IP %s with OK status. Error: %v", clientInfo.Ip, err)
			}
		} else {
			// Update CLIENT_INFO with ERROR status
			err = db.UpdateClientInfoOnError(clientInfo.Ip, clientInfo.DbName, clientInfo.DbType)
			if err != nil {
				log.Printf("Failed to update client info for IP %s with ERROR status. Error: %v", clientInfo.Ip, err)
			}
		}

		// 3. Prepare message to send
		msg := &proto.DatabaseStatus{
			Status:    status,
			Details:   details,
			Ip:        clientInfo.Ip,
			Dbtype:    clientInfo.DbType,
			Dbnm:      clientInfo.DbName,
			Timestamp: timestamppb.Now(),
		}

		// Adjust the timestamp to consider local timezone
		localNow := time.Now().In(time.Local)
		msg.Timestamp = timestamppb.New(localNow)

		// 4. Send the message to the gRPC server
		conn, err := grpc.Dial(fmt.Sprintf("%s:5051", serverIP), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := proto.NewDatabaseStatusServiceClient(conn)

		response, err := c.SendStatus(context.Background(), msg)
		if err != nil {
			log.Printf("could not send status for IP %s: %v", clientInfo.Ip, err)
			continue
		}
		log.Printf("Response from server for IP %s: %s", clientInfo.Ip, response.Message)
	}
}

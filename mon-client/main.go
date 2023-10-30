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

func performCheck(serverIP, dbTypeReq string, check db.CheckItem) {
	clientInfos, err := getClientInfos(serverIP, dbTypeReq)
	if err != nil {
		log.Printf("Failed to retrieve configurations: %v", err)
		return
	}

	for _, clientInfo := range clientInfos {
		DSN := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s" timezone=UTC`,
			clientInfo.DbUser, clientInfo.UserPwd, clientInfo.Ip, clientInfo.Port, clientInfo.DbName)

		// Execute the SQL check based on check.CheckSQL
		// Perform Database Check
		status, details, checkLvl, err := db.ExecuteCheck(DSN, check)
		if err != nil {
			log.Printf("Failed to perform check '%s' for IP %s. Error: %v", check.CheckName, clientInfo.Ip, err)
			// 根据故障级别处理错误
			if checkLvl == "ERROR" {
				// 处理错误级别为 ERROR 的情况
			} else if checkLvl == "WARNING" {
				// 处理错误级别为 WARNING 的情况
			}
			continue
		}
		// Prepare message to send
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

		// Send the message to the gRPC server
		conn, err := grpc.Dial(fmt.Sprintf("%s:5051", serverIP), grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to server: %v", err)
			continue
		}
		defer conn.Close()
		c := proto.NewDatabaseStatusServiceClient(conn)

		response, err := c.SendStatus(context.Background(), msg)
		if err != nil {
			log.Printf("Failed to send status for IP %s: %v", clientInfo.Ip, err)
			continue
		}
		log.Printf("Response from server for IP %s: %s", clientInfo.Ip, response.Message)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <server IP>", os.Args[0])
	}

	serverIP := os.Args[1]
	dbTypeReq := "ORACLE" // Sample DbType to request. Adjust as needed.

	// Retrieve all enabled checks
	checks, err := db.GetEnabledChecks()
	if err != nil {
		log.Fatalf("Failed to get enabled checks: %v", err)
	}

	// Setup tickers for each check based on its frequency
	tickers := make(map[int]*time.Ticker)
	for _, check := range checks {
		tickers[check.ID] = time.NewTicker(time.Duration(check.Frequency) * time.Minute)
		defer tickers[check.ID].Stop()
	}

	// Perform the initial check before starting the loop
	for _, check := range checks {
		performCheck(serverIP, dbTypeReq, check)
	}

	// Start an infinite loop for each check
	for {
		for _, check := range checks {
			ticker := tickers[check.ID]
			select {
			case <-ticker.C:
				performCheck(serverIP, dbTypeReq, check)
			default:
				// Non-blocking select to allow multiple tickers
			}
		}
		time.Sleep(1 * time.Second) // Sleep to prevent a busy loop
	}
}

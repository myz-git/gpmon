// gpmon/mon-client-db2/startdb2.go
package main

import (
	"context"
	"fmt"
	"gpmon/db"
	"gpmon/grpc/proto"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	_ "github.com/ibmdb/go_ibm_db"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func getClientInfos_db2(serverIP, dbTypeReq string) ([]*proto.ClientInfo, error) {
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

func performCheck_db2(serverIP string, clientInfo *proto.ClientInfo, check db.DB2CheckItem) {

	// DSN := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s" timezone=UTC`,
	// 	clientInfo.DbUser, clientInfo.UserPwd, clientInfo.Ip, clientInfo.Port, clientInfo.DbName)
	DSN := fmt.Sprintf("HOSTNAME=%s;PORT=%d;DATABASE=%s;UID=%s;PWD=%s;",
		clientInfo.Ip, clientInfo.Port, clientInfo.DbName, clientInfo.DbUser, clientInfo.UserPwd)

	// Execute the SQL check based on check.CheckSQL
	status, details, _ := db.ExecuteCheckDB2(DSN, check)

	err := db.InsertCheckResult(clientInfo.Ip, clientInfo.Port, clientInfo.DbType, clientInfo.DbName, check.CheckName, status, details)
	if err != nil {
		log.Printf("Failed check '%s' for IP %s", check.CheckName, clientInfo.Ip)
	}

	// Prepare message to send
	msg := &proto.DatabaseStatus{
		Ip:          clientInfo.Ip,
		Port:        clientInfo.Port,
		Dbtype:      clientInfo.DbType,
		Dbnm:        clientInfo.DbName,
		CheckNm:     check.CheckName,
		CheckResult: status,
		Details:     details,
		Timestamp:   timestamppb.Now(),
	}

	// Adjust the timestamp to consider local timezone
	localNow := time.Now().In(time.Local)
	msg.Timestamp = timestamppb.New(localNow)

	// Send the message to the gRPC server
	conn, err := grpc.Dial(fmt.Sprintf("%s:5051", serverIP), grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)

	}
	defer conn.Close()
	c := proto.NewDatabaseStatusServiceClient(conn)

	response, err := c.SendStatus(context.Background(), msg)
	if err != nil {
		log.Printf("Failed to send status for IP %s: %v", clientInfo.Ip, err)

	}

	// Insert check result into check_results table with OK status
	// err = db.InsertCheckResult(clientInfo.Ip, int(clientInfo.Port), clientInfo.DbType, clientInfo.DbName, check.CheckName, status, details)
	// if err != nil {
	// 	log.Printf("Failed to insert check result for IP %s: %v", clientInfo.Ip, err)
	// }
	log.Printf("Response from server for IP %s: %s: %s", clientInfo.Ip, check.CheckName, response.Message)
}

func main() {
	/*** 获取项目根路径 ***/
	_, filename, _, _ := runtime.Caller(0)
	wd := path.Dir(path.Dir(filename))
	// log.Printf("wd:  %s", wd)
	/*** End ***/

	/*** 设定log 同时输出到控制台及log文件中 ***/
	f := wd + "/log/" + "db2svc.log"
	logFile, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	/*** End ***/

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <server IP>", os.Args[0])
	}

	serverIP := os.Args[1]
	dbTypeReq := "DB2" // Sample DbType to request. Adjust as needed.

	// Retrieve all enabled checks
	checks, err := db.GetEnabledChecksDB2(dbTypeReq)
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
	clientInfos, err := getClientInfos_db2(serverIP, dbTypeReq)
	if err != nil {
		log.Fatalf("Failed to retrieve configurations: %v", err)
	}
	for _, clientInfo := range clientInfos {
		for _, check := range checks {
			// log.Printf("client_main-> init-> performCheck,%s", clientInfo.Ip)
			performCheck_db2(serverIP, clientInfo, check)
		}
	}

	// Start an infinite loop for each check
	for {
		for _, check := range checks {
			ticker := tickers[check.ID]
			select {
			case <-ticker.C:
				for _, clientInfo := range clientInfos {
					// log.Printf("client_main-> for-> performCheck,%s", clientInfo.Ip)
					performCheck_db2(serverIP, clientInfo, check)
				}
			default:
				// Non-blocking select to allow multiple tickers
			}
		}
		time.Sleep(1 * time.Second) // Sleep to prevent a busy loop
	}
}

// gpmon/mon-client-mysql/startmysql.go
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

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func getMySQLClientInfos(serverIP, dbTypeReq string) ([]*proto.ClientInfo, error) {
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

func performMySQLCheck(serverIP string, clientInfo *proto.ClientInfo, check db.MYSQLCheckItem) {

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		clientInfo.DbUser, clientInfo.UserPwd, clientInfo.Ip, clientInfo.Port, clientInfo.DbName)

	// Execute the SQL check based on check.CheckSQL
	// log.Printf("check:==>%v", check.CheckName)

	status, details, err := db.ExecuteCheckMYSQL(DSN, check)

	if err != nil {
		log.Printf("Failed check '%s' for IP %s with error: %v", check.CheckName, clientInfo.Ip, err)
		// You may want to insert a failed check result here and continue or return
	}

	err = db.InsertCheckResult(clientInfo.Ip, clientInfo.Port, clientInfo.DbType, clientInfo.DbName, check.CheckName, status, details)
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
	f := wd + "/log/" + "mysqlsvc.log"
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
	dbTypeReq := "MYSQL"
	// Retrieve all enabled checks
	checks, err := db.GetEnabledChecksMYSQL(dbTypeReq)
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
	clientInfos, err := getMySQLClientInfos(serverIP, dbTypeReq)
	if err != nil {
		log.Fatalf("Failed to retrieve configurations: %v", err)
	}
	for _, clientInfo := range clientInfos {
		for _, check := range checks {
			// log.Printf("client_main-> init-> performCheck,%s", clientInfo.Ip)
			performMySQLCheck(serverIP, clientInfo, check)
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
					performMySQLCheck(serverIP, clientInfo, check)
				}
			default:
				// Non-blocking select to allow multiple tickers
			}
		}
		time.Sleep(1 * time.Second) // Sleep to prevent a busy loop
	}
}

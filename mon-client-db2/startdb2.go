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

func performCheck_db2(serverIP string, clientInfo *proto.ClientInfo, check db.CheckItem) {
	DSN := fmt.Sprintf("HOSTNAME=%s;PORT=%d;DATABASE=%s;UID=%s;PWD=%s;",
		clientInfo.Ip, clientInfo.Port, clientInfo.DbName, clientInfo.DbUser, clientInfo.UserPwd)

	var status, details string
	var err error

	// 最大重试次数
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 尝试执行数据库检查
		status, details, err = db.ExecuteCheckDB2(DSN, check)

		if err != nil {
			log.Printf("Attempt %d: Failed check '%s' for IP %s with error: %v", attempt, check.CheckName, clientInfo.Ip, err)

			if attempt < maxRetries {
				// 如果不是最后一次尝试，等待一段时间后重试
				time.Sleep(5 * time.Second)
				continue
			} else {
				// 连续两次失败，处理失败情况
				break
			}
		} else {
			// 成功，跳出循环
			break
		}
	}

	if err != nil {
		// 所有尝试都失败了，可以在这里处理错误
		// 可以记录错误、发送警告消息等
	}

	err = db.InsertCheckResult(clientInfo.Ip, clientInfo.Port, clientInfo.DbType, clientInfo.DbName, check.CheckName, status, details)
	if err != nil {
		log.Printf("Failed to insert check result for IP %s: %v", clientInfo.Ip, err)
	}

	// 准备要发送的消息
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

	// 考虑本地时区调整时间戳
	localNow := time.Now().In(time.Local)
	msg.Timestamp = timestamppb.New(localNow)

	// 发送消息到gRPC服务器
	conn, err := grpc.Dial(fmt.Sprintf("%s:5051", serverIP), grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()
	c := proto.NewDatabaseStatusServiceClient(conn)

	response, err := c.SendStatus(context.Background(), msg)
	if err != nil {
		log.Printf("Failed to send status for IP %s: %v", clientInfo.Ip, err)
	} else {
		log.Printf("Response from ID[%v] %s:%v: %v: %s: %s", clientInfo.Id, clientInfo.Ip, clientInfo.Port, clientInfo.DbName, check.CheckName, response.Message)
	}
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
		// 获取客户端配置的相关检查项ID列表
		checkIDs, err := db.GetClientChecks(clientInfo.Id)

		if err != nil {
			log.Printf("Failed to get checks for client %v: %v", clientInfo.Id, err)
			continue
		}
		// fmt.Printf("===>clientInfo: %v\n", clientInfo)
		// 对于每个检查项ID，找到对应的检查配置并执行检查
		for _, checkID := range checkIDs {
			check, err := db.GetCheckItemByID(checkID) // 假设你有这样一个函数来获取检查项
			// fmt.Printf("=========>初始检查checkid: %v,clientInfo: %v %v %v\n", check, clientInfo.Id, clientInfo.DbType, clientInfo.DbName)
			// fmt.Printf("======>checkID: %v\n", checkID)
			if err != nil {
				log.Printf("Failed to get check item for ID %v: %v", checkID, err)
				continue
			}
			performCheck_db2(serverIP, clientInfo, check)
		}
	}

	// Start an infinite loop for each check
	for _, clientInfo := range clientInfos {
		go func(client *proto.ClientInfo) {
			// 为每个客户端创建一个独立的定时器集合
			clientTickers := make(map[int]*time.Ticker)

			for {
				// 获取客户端对应的检查项
				checkIDs, err := db.GetClientChecks(client.Id)
				if err != nil {
					log.Printf("Failed to get checks for client %v: %v", client.Id, err)
					continue
				}

				for _, checkID := range checkIDs {
					check, err := db.GetCheckItemByID(checkID)
					if err != nil {
						log.Printf("Failed to get check item for ID %v: %v", checkID, err)
						continue
					}

					// 为每个检查项创建或获取定时器
					ticker, ok := clientTickers[check.ID]
					if !ok {
						ticker = time.NewTicker(time.Duration(check.Frequency) * time.Minute)
						clientTickers[check.ID] = ticker
						defer ticker.Stop()
					}

					select {
					case <-ticker.C:
						// fmt.Printf("=========>定时检查checkID: %v ,clientInfo:  %v %v %v\n", checkID, client.Id, client.DbType, client.DbName)
						performCheck_db2(serverIP, client, check)
					default:

						// 非阻塞 select
					}
				}

				time.Sleep(10 * time.Second) // 防止忙循环
			}
		}(clientInfo)
	}

	// 防止主程序退出
	select {}
}

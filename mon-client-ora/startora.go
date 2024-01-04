// gpmon/mon-client-ora/startora.go
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

func performCheck(serverIP string, clientInfo *proto.ClientInfo, check db.CheckItem) {

	DSN := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s" timezone=UTC`,
		clientInfo.DbUser, clientInfo.UserPwd, clientInfo.Ip, clientInfo.Port, clientInfo.DbName)

	// Execute the SQL check based on check.CheckSQL
	status, details, _ := db.ExecuteCheck(DSN, check)

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

	log.Printf("Response from server for IP %s: %s: %s", clientInfo.Ip, check.CheckName, response.Message)
}

func main() {
	/*** 获取项目根路径 ***/
	_, filename, _, _ := runtime.Caller(0)
	wd := path.Dir(path.Dir(filename))
	// log.Printf("wd:  %s", wd)
	/*** End ***/

	/*** 设定log 同时输出到控制台及log文件中 ***/
	f := wd + "/log/" + "orasvc.log"
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
	dbTypeReq := "ORACLE" // Sample DbType to request. Adjust as needed.

	// Retrieve all enabled checks
	checks, err := db.GetEnabledChecks(dbTypeReq)
	if err != nil {
		log.Fatalf("Failed to get enabled checks: %v", err)
	}

	// Setup tickers for each check based on its frequency
	tickers := make(map[int]*time.Ticker)
	for _, check := range checks {
		tickers[check.ID] = time.NewTicker(time.Duration(check.Frequency) * time.Minute)
		defer tickers[check.ID].Stop()
	}

	// 任务启动时,初始化对所有CHECK执行一项
	clientInfos, err := getClientInfos(serverIP, dbTypeReq)
	if err != nil {
		log.Fatalf("Failed to retrieve configurations: %v", err)
	}
	for _, clientInfo := range clientInfos {
		// 获取客户端配置的相关检查项ID列表
		checkIDs, err := db.GetClientChecks(clientInfo.Id)
		if err != nil {
			log.Printf("Failed to get check IDs for client %d: %v", clientInfo.Id, err)
			continue
		}
		// 对于每个检查项ID，找到对应的检查配置并执行检查
		for _, checkID := range checkIDs {
			check, err := db.GetCheckItemByID(checkID)
			if err != nil {
				log.Printf("Failed to get check %d: %v", checkID, err)
				continue
			}
			performCheck(serverIP, clientInfo, check)
		}
	}

	// 开始定时任务循环,执行频率根据check.freq
	for {
		for _, clientInfo := range clientInfos {
			checkIDs, err := db.GetClientChecks(clientInfo.Id)
			if err != nil {
				log.Printf("Failed to get check IDs for client %d: %v", clientInfo.Id, err)
				continue
			}

			for _, checkID := range checkIDs {
				check, err := db.GetCheckItemByID(checkID)
				if err != nil {
					log.Printf("Failed to get check %d: %v", checkID, err)
					continue
				}

				// 如果是时间执行检查，则进行检查
				ticker := tickers[check.ID]
				select {
				case <-ticker.C:
					performCheck(serverIP, clientInfo, check)
				default:
					// 非阻塞 select
				}
			}
		}
		time.Sleep(1 * time.Second) // 防止忙等
	}
}

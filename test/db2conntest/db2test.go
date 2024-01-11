package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/ibmdb/go_ibm_db"
)

type DBConfig struct {
	Hostname string `json:"hostname"`
	Database string `json:"database"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// 读取配置文件
	var config DBConfig
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%s;UID=%s;PWD=%s",
		config.Hostname, config.Database, config.Port, config.Username, config.Password)

	connStr = "HOSTNAME=1.1.1.96;DATABASE=myzdb;PORT=60000;UID=db2inst1;PWD=oracle"

	// 连接数据库
	db, err := sql.Open("go_ibm_db", connStr)
	if err != nil {
		fmt.Println("Error opening DB connection:", err)
		return
	}
	defer db.Close()

	// 执行查询
	var version string
	err = db.QueryRow("SELECT SERVICE_LEVEL FROM TABLE(SYSPROC.ENV_GET_INST_INFO()) AS INSTANCEINFO").Scan(&version)
	if err != nil {
		fmt.Println("Error on query:", err)
		return
	}

	fmt.Println("Connected to DB2 version:", version)
}

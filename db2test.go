package main

import (
	"database/sql"
	"fmt"

	_ "github.com/ibmdb/go_ibm_db"
)

func main() {
	connStr := "HOSTNAME=1.1.1.97;DATABASE=myzdb;PORT=60006;UID=db2inst1;PWD=db2inst1"
	db, err := sql.Open("go_ibm_db", connStr)
	if err != nil {
		fmt.Println("Error opening DB connection:", err)
		return
	}
	defer db.Close()

	// 尝试执行一个简单的查询来验证连接
	var version string
	err = db.QueryRow("SELECT SERVICE_LEVEL FROM TABLE(SYSPROC.ENV_GET_INST_INFO()) AS INSTANCEINFO").Scan(&version)
	if err != nil {
		fmt.Println("Error on query:", err)
		return
	}

	fmt.Println("Connected to DB2 version:", version)
}

// gpmon/db/serverdb.go
package db

import (
	"database/sql"
	"log"
	"path"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type ClientConfig struct {
	IP      string
	Port    int32
	DbType  string
	DbName  string
	DbUser  string
	UserPwd string
}

func init() {
	/*** 获取项目根路径 ***/
	_, filename, _, _ := runtime.Caller(0)
	wd := path.Dir(path.Dir(filename))
	// log.Printf("wd:  %s", wd)
	/*** End ***/

	/*** 设定dbfile路径 ***/
	dbf := wd + "/messages.db"

	var err error
	db, err = sql.Open("sqlite3", dbf)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	/*** End ***/

	// Create the table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS check_his (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip VARCHAR(20),
		port INT,
		dbtype VARCHAR(10),
		dbname VARCHAR(20),
		chk_nm VARCHAR(30),
		chk_result varchar(10),
		chk_details text,
		chk_time DATETIME DEFAULT (datetime('now', 'localtime'))
	)
	`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// InsertCheckResult inserts a new check result into the check_results table.
func InsertCheckResult(ip string, port int32, dbtype, dbname, checkname, checkResult, checkdatails string) error {
	statement := `
	INSERT INTO check_result (ip, port, dbtype, dbname, chk_nm, chk_result, chk_details)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(ip, port, dbtype, dbname, chk_nm) DO UPDATE SET
	chk_result=excluded.chk_result,
	chk_details=excluded.chk_details,
	chk_time=datetime('now', 'localtime');`

	_, err := db.Exec(statement, ip, port, dbtype, dbname, checkname, checkResult, checkdatails)
	return err
}

// GetClientInfos retrieves all the active client configurations for a given DB type.
func GetClientInfos(targetDbType string) ([]ClientConfig, error) {
	rows, err := db.Query("SELECT ip, port, dbtype, dbname, dbuser, userpwd FROM client_info WHERE isenable=1 and dbtype = ?", targetDbType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []ClientConfig
	for rows.Next() {
		var config ClientConfig
		err := rows.Scan(&config.IP, &config.Port, &config.DbType, &config.DbName, &config.DbUser, &config.UserPwd)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func ShouldSendEmail(ip string, port int32, dbType string, dbName string, checkNm string) bool {
	var lastEmailSent time.Time
	err := db.QueryRow(`
        SELECT mailtm FROM check_result 
		where ip=? and port=? and dbtype=? and dbname=? and chk_nm=?`, ip, port, dbType, dbName, checkNm).Scan(&lastEmailSent)

	if err == sql.ErrNoRows {
		//当没有记录,即邮件发送时间为空,则返回TRUE,可以发送
		return true
	}

	// 设定邮件发送的冷却期为24小时
	return time.Since(lastEmailSent) > 24*time.Hour
}

// 更新邮件发送时间的函数
func UpdateMailTm(ip string, port int32, dbType string, dbName string, checkNm string) error {
	// log.Printf("更新邮件发送时间: %s,%v,%s,%s,%s ", ip, port, dbType, dbName, checkNm)
	_, err := db.Exec(`
        UPDATE check_result 
        SET mailtm     = datetime('now', 'localtime')
        where ip=? and port=? and dbtype=? and dbname=? and chk_nm=?`, ip, port, dbType, dbName, checkNm)
	// log.Printf("更新邮件发送ERR: %s", err)
	return err
}

// InsertMessage inserts a new message into the database.
func InsertCheckhis(ip string, port int32, dbtype string, dbnm string, checkname string, checkresult string, details string) error {
	_, err := db.Exec(`
	INSERT INTO check_his (ip,port, dbtype, dbname, chk_nm,chk_result,chk_details)
	VALUES (?, ?, ?, ?, ?,?, ?)
	`, ip, port, dbtype, dbnm, checkname, checkresult, details)
	return err
}

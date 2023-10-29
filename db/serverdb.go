// db/serverdb.go
package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "D:\\Workspace\\gpmon\\messages.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status varchar(10),
		details TEXT,
		ip varchar(20),
		dbtype varchar(10),
		dbnm varchar(20),
		timestamp  DATETIME DEFAULT (datetime('now','localtime'))
	  )
	`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// InsertMessage inserts a new message into the database.
func InsertMessage(status, details, ip, dbtype, dbnm string) error {
	_, err := db.Exec(`
	INSERT INTO messages (status, details, ip, dbtype, dbnm) 
	VALUES (?, ?, ?, ?, ?)
	`, status, details, ip, dbtype, dbnm)
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

type ClientConfig struct {
	IP      string
	Port    int
	DbType  string
	DbName  string
	DbUser  string
	UserPwd string
}

func UpdateClientInfoOnError(ip string, dbName string, dbType string) error {
	_, err := db.Exec(`
        UPDATE client_info 
        SET status = 'ERROR', updatetm = datetime(CURRENT_TIMESTAMP, 'localtime')
        WHERE ip = ? AND dbname = ? AND dbtype = ?`, ip, dbName, dbType)
	return err
}

func UpdateClientInfoOnSuccess(ip string, dbName string, dbType string) error {
	_, err := db.Exec(`
        UPDATE client_info 
        SET status = 'OK', updatetm = datetime(CURRENT_TIMESTAMP, 'localtime'),ismail=0,last_email_sent=''
        WHERE ip = ? AND dbname = ? AND dbtype = ?`, ip, dbName, dbType)
	//当数据库检查正常时,设置邮件为未发送
	return err
}

// UpdateClientInfoIsMail updates the 'ismail' field in the 'client_info' table.
func UpdateClientInfoIsMail(ip string, dbType string, dbName string, isMail int) error {
	_, err := db.Exec(`
		UPDATE client_info 
		SET ismail = ?
		WHERE ip = ? AND dbname = ? AND dbtype = ?`, isMail, ip, dbName, dbType)
	return err
}

func ShouldSendEmail(ip string, dbType string, dbName string) bool {
	var lastEmailSent time.Time
	err := db.QueryRow(`
        SELECT last_email_sent FROM client_info 
        WHERE ip = ? AND dbname = ? AND dbtype = ?`, ip, dbName, dbType).Scan(&lastEmailSent)

	if err != nil {
		// 处理错误，可能是因为记录不存在
	}

	// 设定邮件发送的冷却期为24小时
	return time.Since(lastEmailSent) > 24*time.Hour
}

// 更新邮件发送时间的函数
func UpdateLastEmailSent(ip string, dbType string, dbName string) error {
	_, err := db.Exec(`
        UPDATE client_info 
        SET last_email_sent = datetime(CURRENT_TIMESTAMP, 'localtime')
        WHERE ip = ? AND dbname = ? AND dbtype = ?`, ip, dbName, dbType)
	return err
}

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
	db, err = sql.Open("sqlite3", "./messages.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT,
		details TEXT,
		ip TEXT,
		dbtype TEXT,
		dbnm TEXT,
		timestamp DATETIME
	)
	`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// InsertMessage inserts a new message into the database.
func InsertMessage(status, details, ip, dbtype, dbnm string, timestamp time.Time) error {
	_, err := db.Exec(`
	INSERT INTO messages (status, details, ip, dbtype, dbnm, timestamp) 
	VALUES (?, ?, ?, ?, ?, ?)
	`, status, details, ip, dbtype, dbnm, timestamp)
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

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

func CheckDatabaseStatus(DSN string) (status string, details string, err error) {
	DSN = DSN + "  as sysdba"
	db, err := sql.Open("godror", DSN)
	if err != nil {
		return "", "", fmt.Errorf("cannot connect to database: %v", err)
	}
	defer db.Close()

	// 做一个简单的查询来检查数据库状态
	_, err = db.Exec("SELECT 1 FROM DUAL")
	if err != nil {
		return "ERROR", err.Error(), nil
	}

	return "OK", "Database is healthy", nil
}
func UpdateClientInfoOnError(ip, dbType string) error {
	_, err := db.Exec(`
        UPDATE client_info 
        SET isok = 0, updatetm = CURRENT_TIMESTAMP
        WHERE ip = ? AND dbtype = ?`, ip, dbType)
	return err
}

func UpdateClientInfoOnSuccess(ip, dbType string) error {
	_, err := db.Exec(`
        UPDATE client_info 
        SET isok = 1, updatetm = CURRENT_TIMESTAMP
        WHERE ip = ? AND dbtype = ?`, ip, dbType)
	return err
}

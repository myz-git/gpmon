// gpmon/db/oradb.go
package db

import (
	"database/sql"

	_ "github.com/godror/godror"
)

type MYSQLCheckItem struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string // 添加检查级别字段
	Frequency int
	IsEnable  int
}

// GetEnabledChecks retrieves all enabled checks from the dbmonsql table.
func GetEnabledChecksMYSQL(dbType string) ([]MYSQLCheckItem, error) {
	var checks []MYSQLCheckItem
	rows, err := db.Query("SELECT id, checknm, checksql, checklvl, freq, isenable FROM dbmonsql WHERE  dbtype=? and isenable = 1", dbType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var check MYSQLCheckItem
		if err := rows.Scan(&check.ID, &check.CheckName, &check.CheckSQL, &check.CheckLvl, &check.Frequency, &check.IsEnable); err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, nil
}

// ExecuteCheck executes a single SQL check against the database and returns the result along with the check level.
func ExecuteCheckMYSQL(DSN string, check MYSQLCheckItem) (status string, details string, err error) {
	db, err := sql.Open("godror", DSN)
	if err != nil {
		return "ERROR", "Cannot connect to database", err
	}
	defer db.Close()

	var result int
	err = db.QueryRow(check.CheckSQL).Scan(&result)
	// log.Printf("ExecuteCheck db.QueryRow checksql:===%s ,result:===%v", check.CheckSQL, result)
	if err != nil {
		if err == sql.ErrNoRows {
			// No result means the check failed, return the level of the check.
			return check.CheckLvl, "No rows returned", nil
		}
		// log.Printf("ExecuteCheck db.QueryRow err:===%s", err)
		// log.Printf("ExecuteCheck db.QueryRow err.Error:===%s", err.Error())
		return "ERROR", err.Error(), err
		//这里err和err.Error() 内容一样
	}
	return "OK", "Check successful", nil
}

// / CheckResult 表示数据库检查的结果
type MYSQLCheckResult struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string
	Status    string
	Error     error
}

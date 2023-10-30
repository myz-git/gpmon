// db/oradb.go
package db

import (
	"database/sql"

	_ "github.com/godror/godror"
)

// db/oradb.go

// ... [其他引入的包]

type CheckItem struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string // 添加检查级别字段
	Frequency int
	IsEnable  int
}

// GetEnabledChecks retrieves all enabled checks from the dbmonsql table.
func GetEnabledChecks() ([]CheckItem, error) {
	var checks []CheckItem
	rows, err := db.Query("SELECT id, checknm, checksql, checklvl, freq, isenable FROM dbmonsql WHERE isenable = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var check CheckItem
		if err := rows.Scan(&check.ID, &check.CheckName, &check.CheckSQL, &check.CheckLvl, &check.Frequency, &check.IsEnable); err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, nil
}

// PerformChecks executes the SQL checks against the database.
func PerformChecks(DSN string, checks []CheckItem) ([]CheckResult, error) {
	var results []CheckResult
	for _, check := range checks {
		status, details, checklvl, err := ExecuteCheck(DSN, check)
		results = append(results, CheckResult{
			ID:        check.ID,
			CheckName: check.CheckName,
			Status:    status,
			Details:   details,
			CheckLvl:  checklvl,
			Err:       err,
		})
	}
	return results, nil
}

// executeCheck executes a single SQL check against the database and returns the status.
func ExecuteCheck(DSN string, check CheckItem) (status string, details string, checkLvl string, err error) {
	db, err := sql.Open("godror", DSN)
	if err != nil {
		return "ERROR", "Cannot connect to database", check.CheckLvl, err
	}
	defer db.Close()

	// 假设我们的检查都是通过查询单个值来确认状态的
	// 您可能需要根据实际的SQL检查来调整这里的逻辑
	var result int
	err = db.QueryRow(check.CheckSQL).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			// 检查失败，没有记录，返回checklvl
			return "ERROR", "No records found", check.CheckLvl, nil
		} else {
			// SQL执行错误
			return "ERROR", err.Error(), check.CheckLvl, err
		}
	}

	// 检查成功
	return "OK", "Check successful", "OK", nil
}

// CheckResult represents the result of a single database check.
type CheckResult struct {
	ID        int
	CheckName string
	Status    string
	Details   string
	CheckLvl  string
	Err       error
}

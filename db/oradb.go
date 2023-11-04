// gpmon/db/oradb.go
package db

import (
	"database/sql"

	_ "github.com/godror/godror"
)

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

// PerformChecks 根据dbmonsql表中的检查项执行数据库健康检查
// func PerformChecks(DSN string, checks []CheckItem) ([]CheckResult, error) {
// 	var results []CheckResult
// 	db, err := sql.Open("godror", DSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot open database: %v", err)
// 	}
// 	defer db.Close()

// 	for _, check := range checks {
// 		result, err := executeSQLCheck(db, check.CheckSQL)
// 		results = append(results, CheckResult{
// 			ID:        check.ID,
// 			CheckName: check.CheckName,
// 			CheckSQL:  check.CheckSQL,
// 			CheckLvl:  check.CheckLvl,
// 			Status:    result,
// 			Error:     err,
// 		})
// 	}
// 	return results, nil
// }

// executeCheck executes a single SQL check against the database and returns the status.
// func ExecuteChecks(db *sql.DB, checks []CheckItem) ([]CheckResult, error) {
// 	var results []CheckResult
// 	for _, check := range checks {
// 		result, err := executeSQLCheck(db, check.CheckSQL)
// 		results = append(results, CheckResult{
// 			ID:        check.ID,
// 			CheckName: check.CheckName,
// 			CheckSQL:  check.CheckSQL,
// 			CheckLvl:  check.CheckLvl,
// 			Status:    result,
// 			Error:     err,
// 		})
// 	}
// 	return results, nil
// }

// executeSQLCheck 执行单个SQL检查并返回结果
// func executeSQLCheck(db *sql.DB, sqlQuery string) (string, error) {
// 	var result string
// 	err := db.QueryRow(sqlQuery).Scan(&result)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return "WARNING", fmt.Errorf("no rows returned for sql: %s", sqlQuery)
// 		}
// 		return "ERROR", fmt.Errorf("error executing sql: %s, error: %v", sqlQuery, err)
// 	}
// 	return "OK", nil
// }

// ExecuteCheck executes a single SQL check against the database and returns the result along with the check level.
func ExecuteCheck(DSN string, check CheckItem) (status string, details string, err error) {
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
type CheckResult struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string
	Status    string
	Error     error
}

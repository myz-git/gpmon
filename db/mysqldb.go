// gpmon/db/mysqldb.go
package db

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
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
// ExecuteCheckMYSQL executes a single SQL check against the database and returns the result along with the check level.
func ExecuteCheckMYSQL(DSN string, check CheckItem) (status string, details string, err error) {
	mysqlDB, err := sql.Open("mysql", DSN)
	if err != nil {
		return "ERROR", "Cannot connect to MySQL database", err
	}
	defer mysqlDB.Close()

	if strings.ToLower(check.CheckSQL) == "show slave status" {
		// Special handling for "SHOW SLAVE STATUS"
		rows, err := mysqlDB.Query(check.CheckSQL)
		if err != nil {
			return "ERROR", "Failed to execute SHOW SLAVE STATUS", err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return "ERROR", "Failed to get columns from SHOW SLAVE STATUS", err
		}

		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				return "ERROR", "Failed to scan SHOW SLAVE STATUS", err
			}

			var secondsBehindMasterStr string
			for i, colName := range columns {
				if colName == "Seconds_Behind_Master" {
					secondsBehindMasterStr = string(values[i])
					break
				}
			}

			if secondsBehindMasterStr == "" {
				// Handle the case where Seconds_Behind_Master is NULL or missing
				return "WARNING", "Replication Seconds_Behind_Master is NULL or missing", nil
			}

			secondsBehindMaster, err := strconv.ParseInt(secondsBehindMasterStr, 10, 64)
			if err != nil {
				// Handle the case where Seconds_Behind_Master cannot be converted to an integer
				return "ERROR", "Failed to parse Seconds_Behind_Master as integer", err
			}

			// If Seconds_Behind_Master is more than 3600 seconds (1 hour), return WARNING
			if secondsBehindMaster > 3600 {
				return "WARNING", "Replication delay is more than 1 hour", nil
			} else {
				return "OK", "Check successful", nil
			}
		} else {
			return "ERROR", "No rows returned by SHOW SLAVE STATUS", nil
		}
	} else {
		// For other checks, execute the provided SQL
		var result interface{}
		err = mysqlDB.QueryRow(check.CheckSQL).Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				return check.CheckLvl, "No rows returned", nil
			}
			return "ERROR", err.Error(), err
		}
		return "OK", "Check successful", nil
	}
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

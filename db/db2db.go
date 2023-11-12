// gpmon/db/db2db.go
package db

import (
	"database/sql"

	_ "github.com/ibmdb/go_ibm_db" // DB2 driver
)

// DB2CheckItem represents a check item specifically for DB2.
type DB2CheckItem struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string // Check level field
	Frequency int
	IsEnable  int
}

// GetEnabledChecksDB2 retrieves all enabled checks from the DB2-specific table.
func GetEnabledChecksDB2(dbType string) ([]DB2CheckItem, error) {
	var checks []DB2CheckItem
	// Adjust SQL syntax if necessary for DB2
	rows, err := db.Query("SELECT id, checknm, checksql, checklvl, freq, isenable FROM dbmonsql WHERE  dbtype=? and isenable = 1", dbType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var check DB2CheckItem
		if err := rows.Scan(&check.ID, &check.CheckName, &check.CheckSQL, &check.CheckLvl, &check.Frequency, &check.IsEnable); err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, nil
}

// ExecuteCheckDB2 executes a single SQL check against the DB2 database and returns the result along with the check level.
func ExecuteCheckDB2(DSN string, check DB2CheckItem) (status string, details string, err error) {
	db, err := sql.Open("go_ibm_db", DSN)
	if err != nil {
		return "ERROR", "Cannot connect to database", err
	}
	defer db.Close()

	// Assuming the result is a string that holds the status of the HADR delay
	var result string
	err = db.QueryRow(check.CheckSQL).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows are returned, we consider it as a failed check
			return check.CheckLvl, "No rows returned", nil
		}
		// If there's an error executing the SQL, we return it as an ERROR status
		return "ERROR", err.Error(), err
	}

	// If the query was successful, return the result which should be either 'OK' or 'WARNING'
	return result, "Check successful", nil
}

// DB2CheckResult represents the result of a database check specifically for DB2.
type DB2CheckResult struct {
	ID        int
	CheckName string
	CheckSQL  string
	CheckLvl  string
	Status    string
	Error     error
}

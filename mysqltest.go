package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Replace with your own database information
	user := "root"
	password := "111111"
	host := "1.1.1.201:3306" // or "your_host:your_port"
	dbname := "mysql"

	// Construct the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, dbname)

	// Open a SQL connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Try to ping the database
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	} else {
		log.Println("Successfully connected to the database!")
	}

	// Optionally, you could run a simple query
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Fatalf("Error running query: %v", err)
	} else {
		log.Printf("Database version: %s\n", version)
	}
}

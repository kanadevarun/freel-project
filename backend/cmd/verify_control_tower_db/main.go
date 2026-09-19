package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.Close()

	fmt.Println("=== ALL RELEVANT MARIADB TABLES ===")
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("%v", err)
		}
		lower := strings.ToLower(tableName)
		if strings.Contains(lower, "contract") ||
			strings.Contains(lower, "work") ||
			strings.Contains(lower, "auto") ||
			strings.Contains(lower, "plan") ||
			strings.Contains(lower, "govern") ||
			strings.Contains(lower, "event") ||
			strings.Contains(lower, "decis") ||
			strings.Contains(lower, "approv") ||
			strings.Contains(lower, "alert") ||
			strings.Contains(lower, "notif") {
			var count int
			_ = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
			fmt.Printf(" - %-32s : %5d rows\n", tableName, count)
		}
	}
}

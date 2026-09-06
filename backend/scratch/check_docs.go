package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_URL")
	}
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Println("DB connect error:", err)
		return
	}
	defer db.Close()

	alters := []string{
		"ALTER TABLE shipment_documents MODIFY shipment_id BIGINT NULL",
		"ALTER TABLE shipment_documents ADD COLUMN customer_id BIGINT NULL AFTER shipment_id",
		"ALTER TABLE shipment_documents ADD COLUMN lead_id BIGINT NULL AFTER customer_id",
		"ALTER TABLE shipment_documents ADD COLUMN booking_id BIGINT NULL AFTER lead_id",
		"ALTER TABLE shipment_documents ADD COLUMN original_file_name VARCHAR(500) NULL AFTER file_name",
		"ALTER TABLE shipment_documents ADD COLUMN file_path VARCHAR(1000) NULL AFTER s3_key",
		"ALTER TABLE shipment_documents ADD COLUMN mime_type VARCHAR(100) NULL AFTER file_type",
		"ALTER TABLE shipment_documents ADD COLUMN file_size BIGINT DEFAULT 0 AFTER mime_type",
	}
	for _, alter := range alters {
		_, err := db.Exec(alter)
		if err != nil {
			fmt.Println("Alter error:", alter, err)
		}
	}

	type Col struct {
		Field   string  `db:"Field"`
		Type    string  `db:"Type"`
		Null    string  `db:"Null"`
		Key     string  `db:"Key"`
		Default *string `db:"Default"`
		Extra   string  `db:"Extra"`
	}

	var cols []Col
	err = db.Select(&cols, "DESCRIBE shipment_documents")
	if err != nil {
		fmt.Println("Describe error:", err)
		return
	}

	fmt.Println("Columns in shipment_documents:")
	for _, c := range cols {
		fmt.Printf(" - %s (%s, Nullable: %s)\n", c.Field, c.Type, c.Null)
	}
}

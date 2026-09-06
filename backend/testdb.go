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
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Println("DB connect error:", err)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			r.id AS rate_id,
			r.carrier_name,
			COALESCE(r.carrier_code, '') AS carrier_code,
			r.rate_type,
			r.version_number,
			r.origin_port,
			r.destination_port,
			r.transport_mode,
			COALESCE(r.equipment_type, '') AS equipment_type,
			r.currency,
			r.base_amount,
			DATE_FORMAT(r.valid_from, '%Y-%m-%d') AS valid_from,
			DATE_FORMAT(r.valid_until, '%Y-%m-%d') AS valid_until,
			r.status,
			COALESCE(DATEDIFF(r.valid_until, CURDATE()), 999) AS days_remaining,
			CASE 
				WHEN r.status = 'EXPIRED' OR (r.valid_until IS NOT NULL AND DATEDIFF(r.valid_until, CURDATE()) < 0) THEN 'EXPIRED'
				WHEN r.status = 'EXPIRING_SOON' AND DATEDIFF(r.valid_until, CURDATE()) <= 7 THEN 'EXPIRING_7D'
				WHEN r.status = 'EXPIRING_SOON' OR (r.valid_until IS NOT NULL AND DATEDIFF(r.valid_until, CURDATE()) <= 30) THEN 'EXPIRING_30D'
				WHEN r.status = 'SUPERSEDED' THEN 'SUPERSEDED'
				ELSE 'ACTIVE'
			END AS attention_bucket,
			(SELECT COUNT(DISTINCT qrs.quotation_id) FROM quotation_rate_selections qrs WHERE qrs.org_id = r.org_id AND qrs.rate_id = r.id AND qrs.is_active = TRUE) AS affected_quotes,
			COALESCE(c.contract_code, '') AS contract_code
		FROM rates r
		LEFT JOIN rate_contracts c ON r.org_id = c.org_id AND r.contract_id = c.id
		WHERE r.org_id = 1
		  AND (
		      r.status IN ('EXPIRING_SOON', 'EXPIRED', 'SUPERSEDED')
		      OR (r.valid_until IS NOT NULL AND DATEDIFF(r.valid_until, CURDATE()) <= 30)
		  )
		ORDER BY 
			CASE 
				WHEN r.status = 'EXPIRED' THEN 1
				WHEN r.status = 'EXPIRING_SOON' THEN 2
				WHEN r.status = 'SUPERSEDED' THEN 3
				ELSE 4
			END,
			days_remaining ASC
		LIMIT 100
	`
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("Query error:", err)
	} else {
		fmt.Println("Query succeeded!")
	}
}

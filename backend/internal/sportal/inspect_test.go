package sportal

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

func TestInspectDB(t *testing.T) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC"
	}
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Skipf("skipping live DB inspect test: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	var tables []string
	err = db.SelectContext(ctx, &tables, "SHOW TABLES")
	if err != nil {
		t.Fatalf("show tables failed: %v", err)
	}
	fmt.Printf("Total tables: %d\n", len(tables))
	for _, tbl := range tables {
		fmt.Printf("TABLE: %s\n", tbl)
	}

	type SubUsage struct {
		ID           int64   `db:"id"`
		OrgID        int64   `db:"org_id"`
		MetricName   string  `db:"metric_name"`
		CurrentUsage int     `db:"current_usage"`
		LimitAmount  *int    `db:"limit_amount"`
		PeriodStart  *string `db:"period_start"`
		PeriodEnd    *string `db:"period_end"`
	}
	var usages []SubUsage
	_ = db.SelectContext(ctx, &usages, "SELECT id, org_id, metric_name, current_usage, limit_amount, CAST(period_start AS CHAR) as period_start, CAST(period_end AS CHAR) as period_end FROM subscription_usage")
	fmt.Printf("\n--- subscription_usage (%d rows) ---\n", len(usages))
	for _, u := range usages {
		lim := "nil"
		if u.LimitAmount != nil {
			lim = fmt.Sprintf("%d", *u.LimitAmount)
		}
		fmt.Printf("Org %d: %s = %d / %s\n", u.OrgID, u.MetricName, u.CurrentUsage, lim)
	}
}

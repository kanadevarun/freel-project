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

	fmt.Println("=== ORGANIZATIONS ===")
	var orgs []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	if err := db.Select(&orgs, "SELECT id, name FROM organizations"); err != nil {
		fmt.Println("Error orgs:", err)
	} else {
		for _, o := range orgs {
			fmt.Printf("Org ID: %d, Name: %s\n", o.ID, o.Name)
		}
	}

	fmt.Println("\n=== USERS & ORG_MEMBERS ===")
	var members []struct {
		UserID    int64  `db:"user_id"`
		Email     string `db:"email"`
		OrgID     int64  `db:"org_id"`
		OrgName   string `db:"org_name"`
		RoleName  string `db:"role_name"`
		Status    string `db:"status"`
	}
	query := `
		SELECT u.id as user_id, u.email, om.org_id, o.name as org_name, r.name as role_name, om.status
		FROM org_members om
		JOIN users u ON om.user_id = u.id
		JOIN organizations o ON om.org_id = o.id
		JOIN roles r ON om.role_id = r.id
	`
	if err := db.Select(&members, query); err != nil {
		fmt.Println("Error members:", err)
	} else {
		for _, m := range members {
			fmt.Printf("User %d (%s) -> Org %d (%s), Role: %s, Status: %s\n", m.UserID, m.Email, m.OrgID, m.OrgName, m.RoleName, m.Status)
		}
	}

	fmt.Println("\n=== INVITATIONS ===")
	var invites []struct {
		ID       int64  `db:"id"`
		OrgID    int64  `db:"org_id"`
		Email    string `db:"email"`
		RoleID   int64  `db:"role_id"`
	}
	if err := db.Select(&invites, "SELECT id, org_id, email, role_id FROM invitations"); err != nil {
		fmt.Println("Error invites:", err)
	} else {
		for _, i := range invites {
			fmt.Printf("Invite %d -> Org %d, Email: %s, RoleID: %d\n", i.ID, i.OrgID, i.Email, i.RoleID)
		}
	}
}

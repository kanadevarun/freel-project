package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("=== MARIADB WORKFORCE DATABASE VERIFICATION ===")

	tables := []string{
		"workforce_agents",
		"workforce_tasks",
		"workforce_contexts",
		"workforce_messages",
		"workforce_handoffs",
		"ai_agent_outcomes",
		"ai_agent_memories",
		"audit_logs",
	}

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := db.QueryRow(query).Scan(&count); err != nil {
			fmt.Printf(" [!] Table %s: error querying (%v)\n", table, err)
		} else {
			fmt.Printf(" [OK] %-22s : %5d records\n", table, count)
		}
	}

	fmt.Println("\n=== REGISTERED WORKFORCE AGENTS IN DB ===")
	rows, err := db.Query("SELECT agent_id, agent_type, autonomy_level, health_status, is_enabled FROM workforce_agents ORDER BY agent_id")
	if err != nil {
		log.Fatalf("Failed to query agents: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var agentID, agentType, autonomyLevel, healthStatus string
		var isEnabled bool
		if err := rows.Scan(&agentID, &agentType, &autonomyLevel, &healthStatus, &isEnabled); err != nil {
			log.Fatalf("Failed to scan agent row: %v", err)
		}
		fmt.Printf(" - %-18s | Type: %-11s | Autonomy: %-28s | Health: %-8s | Enabled: %v\n",
			agentID, agentType, autonomyLevel, healthStatus, isEnabled)
	}

	fmt.Println("\n=== RECENT WORKFORCE TASKS PERSISTENCE VERIFICATION ===")
	taskRows, err := db.Query("SELECT task_id, org_id, assigned_agent_id, status, priority, created_at FROM workforce_tasks ORDER BY id DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Failed to query recent tasks: %v", err)
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var taskID, agentID, status, priority, createdAt string
		var orgID int64
		if err := taskRows.Scan(&taskID, &orgID, &agentID, &status, &priority, &createdAt); err != nil {
			log.Fatalf("Failed to scan task row: %v", err)
		}
		fmt.Printf(" - Task: %-20s | Org: %d | Agent: %-16s | Status: %-10s | Priority: %-6s | Created: %s\n",
			taskID, orgID, agentID, status, priority, createdAt)
	}
}

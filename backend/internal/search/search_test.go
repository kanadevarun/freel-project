package search_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/search"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestGlobalSearch(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("MySQL ping failed:", err)
		return
	}

	repo := search.NewRepository(db)
	svc := search.NewService(repo)
	ctx := context.Background()

	// 1. Test empty query returns empty results
	emptyRes, err := svc.Search(ctx, 1, "", "", 20)
	if err != nil {
		t.Fatalf("Empty search failed: %v", err)
	}
	if emptyRes.TotalMatches != 0 || len(emptyRes.Groups) != 0 {
		t.Errorf("Expected 0 results for empty query, got %d", emptyRes.TotalMatches)
	}

	// 2. Test searching for general keyword e.g. "MSC" or "BK" or "IN"
	res, err := svc.Search(ctx, 1, "BK", "", 20)
	if err != nil {
		t.Fatalf("Search failed for 'BK': %v", err)
	}
	t.Logf("Found %d matches across %d groups for query 'BK'", res.TotalMatches, len(res.Groups))
	for _, g := range res.Groups {
		t.Logf("Group %s (%s): %d items", g.Category, g.CategoryLabel, len(g.Items))
		for _, item := range g.Items {
			t.Logf("  -> [%s] %s | %s (%s)", item.Category, item.Title, item.Subtitle, item.URL)
		}
	}

	// 3. Test tenant isolation: non-existent org has 0 results
	resIsolated, err := svc.Search(ctx, 999999, "BK", "", 20)
	if err != nil {
		t.Fatalf("Search failed for org 999999: %v", err)
	}
	if resIsolated.TotalMatches != 0 {
		t.Errorf("Tenant isolation violation: Org 999999 got %d matches", resIsolated.TotalMatches)
	}
}

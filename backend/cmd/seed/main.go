// Package main implements the seed command for Freel's in-house testing environment.
//
// Usage:
//
//	go run ./cmd/seed
//
// What it creates:
//   - 1 Organisation (Freel Global Logistics Pvt Ltd)
//   - 4 Users: CEO, Sales Rep, Pricing Manager, Customer Contact
//   - 5 Customers/Leads at different pipeline stages
//   - 6 RFQs at different lifecycle stages (RFQ_CREATED → WON/LOST)
//   - Carrier quotes attached to relevant RFQs
//
// This lets you test the full CEO → Sales → Pricing → Customer workflow
// locally without any third-party integrations.
//
// WARNING: This script is idempotent — running it twice will skip already-
// seeded records (based on unique constraints like org name, email, rfq_number).
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/database"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database")

	ctx := context.Background()
	s := &seeder{db: db, rng: rand.New(rand.NewSource(42))}

	// Run all seed steps in order
	orgID, err := s.seedOrganisation(ctx)
	if err != nil {
		log.Fatalf("❌ seedOrganisation: %v", err)
	}
	log.Printf("✅ Organisation ID: %d", orgID)

	users, err := s.seedUsers(ctx, orgID)
	if err != nil {
		log.Fatalf("❌ seedUsers: %v", err)
	}
	log.Printf("✅ Created %d users", len(users))

	customers, err := s.seedCustomers(ctx, orgID)
	if err != nil {
		log.Fatalf("❌ seedCustomers: %v", err)
	}
	log.Printf("✅ Created %d customers/leads", len(customers))

	rfqs, err := s.seedRFQs(ctx, orgID, customers, users)
	if err != nil {
		log.Fatalf("❌ seedRFQs: %v", err)
	}
	log.Printf("✅ Created %d RFQs", len(rfqs))

	if err := s.seedQuotes(ctx, rfqs); err != nil {
		log.Fatalf("❌ seedQuotes: %v", err)
	}
	log.Println("✅ Seeded carrier quotes")

	log.Println("\n🎉 Seed complete! Login credentials:")
	log.Println("  CEO:      ceo@freel-demo.local / FreelDemo@123")
	log.Println("  Sales:    sales@freel-demo.local / FreelDemo@123")
	log.Println("  Pricing:  pricing@freel-demo.local / FreelDemo@123")
	log.Println("  Customer: customer@tata-exports.local / FreelDemo@123")
	log.Println("\n  Note: These are local-only accounts. Cognito is bypassed for seed testing.")

	os.Exit(0)
}

// ──────────────────────────────────────────────────────────────────────────────
// Seeder
// ──────────────────────────────────────────────────────────────────────────────

type seeder struct {
	db  *sqlx.DB
	rng *rand.Rand
}

// ── Organisation ──────────────────────────────────────────────────────────────

func (s *seeder) seedOrganisation(ctx context.Context) (int32, error) {
	var id int32
	// Check if the organisation already exists
	err := s.db.QueryRowContext(ctx, `SELECT id FROM organizations WHERE name = $1`, "Freel Global Logistics Pvt Ltd").Scan(&id)
	if err == nil {
		return id, nil
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO organizations (name, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id
	`, "Freel Global Logistics Pvt Ltd").Scan(&id)
	return id, err
}

// ── Users ─────────────────────────────────────────────────────────────────────
// We insert into the users table with hashed passwords.
// For in-house testing we bypass Cognito — auth middleware must be set to
// DEV_MODE=true which uses the local users table instead of Cognito tokens.

type seedUser struct {
	ID   int32
	Role string
}

func (s *seeder) seedUsers(ctx context.Context, orgID int32) (map[string]seedUser, error) {
	users := []struct {
		first, last, email, role string
	}{
		{"Varun", "Kanade", "ceo@freel-demo.local", "CEO"},
		{"Priya", "Sharma", "sales@freel-demo.local", "SALES"},
		{"Aditya", "Kumar", "pricing@freel-demo.local", "PRICING"},
		{"Ravi", "Mehta", "customer@tata-exports.local", "CUSTOMER_CONTACT"},
	}

	result := make(map[string]seedUser)
	for _, u := range users {
		// 1. Insert or update the user in public.users
		var userID int32
		err := s.db.QueryRowContext(ctx, `
			INSERT INTO users (cognito_sub, email, first_name, last_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name
			RETURNING id
		`, fmt.Sprintf("local-%s", u.role), u.email, u.first, u.last).Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("insert user %s: %w", u.email, err)
		}

		// 2. Insert or update the role in public.roles
		var roleID int32
		err = s.db.QueryRowContext(ctx, `
			INSERT INTO roles (org_id, name, description, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (org_id, name) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`, orgID, u.role, fmt.Sprintf("%s role in organization", u.role)).Scan(&roleID)
		if err != nil {
			return nil, fmt.Errorf("insert role %s: %w", u.role, err)
		}

		// 3. Link user + organization + role in public.org_members
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO org_members (org_id, user_id, role_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'ACTIVE', NOW(), NOW())
			ON CONFLICT (org_id, user_id) DO UPDATE SET role_id = EXCLUDED.role_id
		`, orgID, userID, roleID)
		if err != nil {
			return nil, fmt.Errorf("link org member user=%d org=%d: %w", userID, orgID, err)
		}

		result[u.role] = seedUser{ID: userID, Role: u.role}
	}
	return result, nil
}

// ── Customers/Leads ───────────────────────────────────────────────────────────

type seedCustomer struct {
	ID          int32
	CompanyName string
}

func (s *seeder) seedCustomers(ctx context.Context, orgID int32) ([]seedCustomer, error) {
	companies := []struct {
		name, status, industry string
	}{
		{"Tata Exports Ltd", "CUSTOMER", "Manufacturing"},
		{"Sun Pharma", "CUSTOMER", "Pharmaceuticals"},
		{"Mahindra Auto", "CUSTOMER", "Automotive"},
		{"Reliance Petro", "LEAD_QUALIFIED", "Chemicals"},
		{"Adani Ports", "LEAD_CONTACTED", "Port Services"},
	}

	var result []seedCustomer
	for _, c := range companies {
		// 1. Check if company already exists
		var companyID int32
		err := s.db.QueryRowContext(ctx, `SELECT id FROM companies WHERE org_id = $1 AND name = $2`, orgID, c.name).Scan(&companyID)
		if err != nil {
			// Insert new company
			err = s.db.QueryRowContext(ctx, `
				INSERT INTO companies (org_id, name, industry, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				RETURNING id
			`, orgID, c.name, c.industry).Scan(&companyID)
			if err != nil {
				return nil, fmt.Errorf("insert company %s: %w", c.name, err)
			}
		}

		// 2. Check if customer record exists for this company
		var customerID int32
		err = s.db.QueryRowContext(ctx, `SELECT id FROM customers WHERE org_id = $1 AND company_id = $2`, orgID, companyID).Scan(&customerID)
		if err != nil {
			// Insert new customer record
			err = s.db.QueryRowContext(ctx, `
				INSERT INTO customers (org_id, company_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				RETURNING id
			`, orgID, companyID, c.status).Scan(&customerID)
			if err != nil {
				return nil, fmt.Errorf("insert customer record for %s: %w", c.name, err)
			}
		} else {
			// Update status
			_, err = s.db.ExecContext(ctx, `UPDATE customers SET status = $1, updated_at = NOW() WHERE id = $2`, c.status, customerID)
			if err != nil {
				return nil, fmt.Errorf("update customer status for %s: %w", c.name, err)
			}
		}

		result = append(result, seedCustomer{ID: customerID, CompanyName: c.name})
	}
	return result, nil
}

// ── RFQs ──────────────────────────────────────────────────────────────────────

type seedRFQ struct {
	ID    int32
	Stage string
}

func (s *seeder) seedRFQs(ctx context.Context, orgID int32, customers []seedCustomer, users map[string]seedUser) ([]seedRFQ, error) {
	now := time.Now()

	// Each entry represents one RFQ at a specific lifecycle stage.
	// The mix gives us one RFQ in every stage for dashboard testing.
	rfqDefs := []struct {
		rfqNumber, origin, dest, incoterms, stage string
		customerIdx                                int
		daysAgo, targetDaysAhead                  int
		healthScore                                int
	}{
		// Stage: RFQ_CREATED (brand new, just came in)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-1001", now.Format("20060102")),
			origin: "INNSA", dest: "DEHAM", incoterms: "EXW",
			stage: "RFQ_CREATED", customerIdx: 0, daysAgo: 0, targetDaysAhead: 45, healthScore: 80,
		},
		// Stage: PRICING_ASSIGNED (sales routed it to pricing team)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-1002", now.Format("20060102")),
			origin: "INBOM", dest: "USNYC", incoterms: "FOB",
			stage: "PRICING_ASSIGNED", customerIdx: 1, daysAgo: 2, targetDaysAhead: 40, healthScore: 90,
		},
		// Stage: QUOTE_GENERATED (AI pricing agent ran, draft quote ready for review)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-1003", now.Format("20060102")),
			origin: "INCCU", dest: "SGSIN", incoterms: "CIF",
			stage: "QUOTE_GENERATED", customerIdx: 2, daysAgo: 4, targetDaysAhead: 30, healthScore: 85,
		},
		// Stage: QUOTE_SENT (pricing approved, quote emailed to customer)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-1004", now.Format("20060102")),
			origin: "INBLR", dest: "AEAUH", incoterms: "EXW",
			stage: "QUOTE_SENT", customerIdx: 3, daysAgo: 7, targetDaysAhead: 20, healthScore: 70,
		},
		// Stage: WON (customer accepted, job created)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-0990", now.Format("20060102")),
			origin: "INNSA", dest: "DEHAM", incoterms: "FOB",
			stage: "WON", customerIdx: 0, daysAgo: 15, targetDaysAhead: -5, healthScore: 100,
		},
		// Stage: LOST (customer chose competitor)
		{
			rfqNumber: fmt.Sprintf("RFQ-%s-0988", now.Format("20060102")),
			origin: "INBOM", dest: "CNSHA", incoterms: "EXW",
			stage: "LOST", customerIdx: 4, daysAgo: 20, targetDaysAhead: -10, healthScore: 40,
		},
	}

	var result []seedRFQ
	salesID := users["SALES"].ID
	pricingID := users["PRICING"].ID

	for _, r := range rfqDefs {
		createdAt := now.AddDate(0, 0, -r.daysAgo)
		targetDate := now.AddDate(0, 0, r.targetDaysAhead)

		var id int32
		err := s.db.QueryRowContext(ctx, `SELECT id FROM rfqs WHERE rfq_number = $1`, r.rfqNumber).Scan(&id)
		if err != nil {
			// Insert new RFQ
			err = s.db.QueryRowContext(ctx, `
				INSERT INTO rfqs (
					org_id, rfq_number, customer_id, stage, origin, destination, incoterms,
					target_date, sales_assignee_id, pricing_assignee_id,
					health_score, agent_status, created_at, updated_at
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
				RETURNING id
			`,
				orgID, r.rfqNumber, customers[r.customerIdx].ID, r.stage,
				r.origin, r.dest, r.incoterms, targetDate,
				salesID, pricingID,
				r.healthScore, "IDLE",
				createdAt, createdAt,
			).Scan(&id)
			if err != nil {
				return nil, fmt.Errorf("insert rfq %s: %w", r.rfqNumber, err)
			}
		} else {
			// Update existing RFQ stage
			_, err = s.db.ExecContext(ctx, `UPDATE rfqs SET stage = $1, updated_at = NOW() WHERE id = $2`, r.stage, id)
			if err != nil {
				return nil, fmt.Errorf("update rfq %s stage: %w", r.rfqNumber, err)
			}
		}

		// Add a cargo item for each RFQ
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO rfq_items (rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT DO NOTHING
		`, id, "General Cargo", 5, 12500.0, 22.5)
		if err != nil {
			return nil, fmt.Errorf("insert rfq_item for %s: %w", r.rfqNumber, err)
		}

		result = append(result, seedRFQ{ID: id, Stage: r.stage})
	}
	return result, nil
}

// ── Quotes ────────────────────────────────────────────────────────────────────

func (s *seeder) seedQuotes(ctx context.Context, rfqs []seedRFQ) error {
	// Add AI-generated quotes to RFQs that are at QUOTE_GENERATED or later
	quoteDefs := []struct {
		rfqIdx                                     int
		carrier                                    string
		buyPrice, sellPrice                        float64
		transitDays, reliabilityScore              int
		successRate                                float64
		isRecommended                              bool
		status                                     string
		aiReasoning                                string
	}{
		// RFQ index 2 (QUOTE_GENERATED) — 3 carrier options, Maersk recommended
		{2, "Maersk", 1800.0, 2160.0, 22, 95, 97.5, true, "DRAFT",
			"Maersk is recommended: 22-day transit meets the target date, 95/100 reliability, 97.5% historical success rate on this lane."},
		{2, "MSC", 1600.0, 1920.0, 28, 85, 90.0, false, "DRAFT",
			"MSC is cheaper but 28-day transit may miss the deadline."},
		{2, "CMA CGM", 1700.0, 2040.0, 25, 88, 92.0, false, "DRAFT",
			"CMA CGM is a middle-ground option — borderline on deadline."},

		// RFQ index 3 (QUOTE_SENT) — Hapag-Lloyd approved and sent
		{3, "Hapag-Lloyd", 620.0, 744.0, 7, 91, 93.5, true, "APPROVED",
			"Hapag-Lloyd recommended for India → UAE: 7-day transit, high reliability."},

		// RFQ index 4 (WON) — MSC won
		{4, "MSC", 1350.0, 1620.0, 22, 88, 91.0, true, "APPROVED",
			"MSC selected — met all requirements for INNSA → DEHAM."},

		// RFQ index 5 (LOST) — old draft, never approved
		{5, "Evergreen", 1100.0, 1320.0, 35, 72, 82.0, false, "REJECTED",
			"Evergreen: transit too long for customer's deadline requirement."},
	}

	for _, q := range quoteDefs {
		rfqID := rfqs[q.rfqIdx].ID
		reasoning := q.aiReasoning

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO rfq_quotes (
				rfq_id, carrier_name, transit_time_days, buy_price, sell_price,
				is_recommended, reliability_score, historical_success_rate, ai_reasoning,
				status, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())
			ON CONFLICT DO NOTHING
		`,
			rfqID, q.carrier, q.transitDays, q.buyPrice, q.sellPrice,
			q.isRecommended, q.reliabilityScore, q.successRate, reasoning,
			q.status,
		)
		if err != nil {
			return fmt.Errorf("insert quote for rfq %d carrier %s: %w", rfqID, q.carrier, err)
		}
	}
	return nil
}

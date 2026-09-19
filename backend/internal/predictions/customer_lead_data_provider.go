package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DeterministicLeadData holds structured telemetry queried from real MariaDB tables
type DeterministicLeadData struct {
	LeadID                 int64                  `json:"lead_id"`
	OrgID                  int64                  `json:"org_id"`
	CompanyName            string                 `json:"company_name"`
	ContactName            string                 `json:"contact_name"`
	Email                  string                 `json:"email"`
	Phone                  string                 `json:"phone"`
	Status                 string                 `json:"status"`
	AIScore                int                    `json:"ai_score"`
	AgeHours               float64                `json:"age_hours"`
	InactivityHours        float64                `json:"inactivity_hours"`
	InteractionCount       int                    `json:"interaction_count"`
	UnansweredInquiries    int                    `json:"unanswered_inquiries"`
	LastInteractionSubject string                 `json:"last_interaction_subject,omitempty"`
	LastInteractionDate    *time.Time             `json:"last_interaction_date,omitempty"`
	LinkedRFQCount         int                    `json:"linked_rfq_count"`
	WonRFQCount            int                    `json:"won_rfq_count"`
	LinkedQuoteCount       int                    `json:"linked_quote_count"`
	IsConverted            bool                   `json:"is_converted"`
	CustomerID             *int64                 `json:"customer_id,omitempty"`
	MissingFields          []string               `json:"missing_fields"`
	ContextMap             map[string]interface{} `json:"context_map"`
}

// DeterministicCustomerData holds structured customer telemetry from real MariaDB tables
type DeterministicCustomerData struct {
	CustomerID          int64                  `json:"customer_id"`
	OrgID               int64                  `json:"org_id"`
	Name                string                 `json:"name"`
	ContactName         string                 `json:"contact_name"`
	ContactEmail        string                 `json:"contact_email"`
	ContactPhone        string                 `json:"contact_phone"`
	Status              string                 `json:"status"`
	HealthScore         int                    `json:"health_score"`
	CreditStatus        string                 `json:"credit_status"`
	TenureDays          int                    `json:"tenure_days"`
	InactivityDays      int                    `json:"inactivity_days"`
	TotalRFQs           int                    `json:"total_rfqs"`
	WonRFQs             int                    `json:"won_rfqs"`
	TotalBookings       int                    `json:"total_bookings"`
	TotalShipments      int                    `json:"total_shipments"`
	PendingTasksCount   int                    `json:"pending_tasks_count"`
	RecentRFQDays       int                    `json:"recent_rfq_days"`
	HasUnresolvedQuotes bool                   `json:"has_unresolved_quotes"`
	ContextMap          map[string]interface{} `json:"context_map"`
}

// CustomerLeadDataProvider fetches real persistent telemetry for leads and customers
type CustomerLeadDataProvider struct {
	db *sql.DB
}

func NewCustomerLeadDataProvider(db *sql.DB) *CustomerLeadDataProvider {
	return &CustomerLeadDataProvider{db: db}
}

// FetchLeadData queries real tables (leads, lead_interactions, rfqs, customer_lead_links) with tenant isolation
func (p *CustomerLeadDataProvider) FetchLeadData(ctx context.Context, orgID int64, leadID int64) (*DeterministicLeadData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	leadQuery := `
		SELECT id, org_id, COALESCE(company_name, ''), COALESCE(contact_name, ''),
		       COALESCE(email, ''), COALESCE(phone, ''), COALESCE(status, 'NEW'),
		       COALESCE(ai_score, 0), created_at, updated_at
		FROM leads
		WHERE id = ? AND org_id = ?`

	var d DeterministicLeadData
	var createdAt, updatedAt time.Time
	err := p.db.QueryRowContext(ctx, leadQuery, leadID, orgID).Scan(
		&d.LeadID, &d.OrgID, &d.CompanyName, &d.ContactName,
		&d.Email, &d.Phone, &d.Status, &d.AIScore, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lead %d not found for organization %d", leadID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying lead %d: %w", leadID, err)
	}

	now := time.Now().UTC()
	d.AgeHours = now.Sub(createdAt).Hours()
	latestActivityTime := updatedAt
	if createdAt.After(latestActivityTime) {
		latestActivityTime = createdAt
	}

	// 1. Query lead interactions
	var lastSubject sql.NullString
	var lastDate sql.NullTime
	var lastDirection sql.NullString

	interactionQuery := `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN direction = 'INBOUND' THEN 1 ELSE 0 END), 0) as in_cnt,
		       COALESCE(SUM(CASE WHEN direction = 'OUTBOUND' THEN 1 ELSE 0 END), 0) as out_cnt
		FROM lead_interactions
		WHERE lead_id = ? AND org_id = ?`

	var inCnt, outCnt int
	_ = p.db.QueryRowContext(ctx, interactionQuery, leadID, orgID).Scan(&d.InteractionCount, &inCnt, &outCnt)

	latestInteractionQuery := `
		SELECT subject, created_at, direction
		FROM lead_interactions
		WHERE lead_id = ? AND org_id = ?
		ORDER BY created_at DESC
		LIMIT 1`
	err = p.db.QueryRowContext(ctx, latestInteractionQuery, leadID, orgID).Scan(&lastSubject, &lastDate, &lastDirection)
	if err == nil && lastDate.Valid {
		d.LastInteractionDate = &lastDate.Time
		if lastDate.Time.After(latestActivityTime) {
			latestActivityTime = lastDate.Time
		}
		if lastSubject.Valid {
			d.LastInteractionSubject = lastSubject.String
		}
		// If the most recent interaction was inbound and not responded to
		if lastDirection.Valid && strings.ToUpper(lastDirection.String) == "INBOUND" {
			d.UnansweredInquiries = 1
		}
	}

	d.InactivityHours = now.Sub(latestActivityTime).Hours()

	// 2. Query linked RFQs
	rfqQuery := `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN status = 'WON' THEN 1 ELSE 0 END), 0)
		FROM rfqs
		WHERE lead_id = ? AND org_id = ?`
	_ = p.db.QueryRowContext(ctx, rfqQuery, leadID, orgID).Scan(&d.LinkedRFQCount, &d.WonRFQCount)

	// 3. Query linked Quotes
	quoteQuery := `
		SELECT COUNT(*)
		FROM rfq_quotes q
		INNER JOIN rfqs r ON q.rfq_id = r.id
		WHERE r.lead_id = ? AND r.org_id = ?`
	_ = p.db.QueryRowContext(ctx, quoteQuery, leadID, orgID).Scan(&d.LinkedQuoteCount)

	// 4. Check if converted to customer
	linkQuery := `
		SELECT customer_id
		FROM customer_lead_links
		WHERE lead_id = ? AND org_id = ?
		LIMIT 1`
	var custID int64
	if err := p.db.QueryRowContext(ctx, linkQuery, leadID, orgID).Scan(&custID); err == nil && custID > 0 {
		d.IsConverted = true
		d.CustomerID = &custID
	}
	if strings.ToUpper(d.Status) == "CONVERTED" {
		d.IsConverted = true
	}

	// 5. Evaluate missing critical fields
	var missing []string
	if strings.TrimSpace(d.Phone) == "" {
		missing = append(missing, "phone")
	}
	if strings.TrimSpace(d.ContactName) == "" {
		missing = append(missing, "contact_name")
	}
	d.MissingFields = missing

	// Construct context map for sidecar
	d.ContextMap = map[string]interface{}{
		"lead_id":                  d.LeadID,
		"company_name":             d.CompanyName,
		"contact_name":             d.ContactName,
		"email":                    d.Email,
		"phone":                    d.Phone,
		"status":                   d.Status,
		"ai_score":                 d.AIScore,
		"age_hours":                d.AgeHours,
		"inactivity_hours":         d.InactivityHours,
		"interaction_count":        d.InteractionCount,
		"unanswered_inquiries":     d.UnansweredInquiries,
		"last_interaction_subject": d.LastInteractionSubject,
		"linked_rfq_count":         d.LinkedRFQCount,
		"won_rfq_count":            d.WonRFQCount,
		"linked_quote_count":       d.LinkedQuoteCount,
		"is_converted":             d.IsConverted,
		"missing_fields":           d.MissingFields,
	}
	if d.CustomerID != nil {
		d.ContextMap["customer_id"] = *d.CustomerID
	}

	return &d, nil
}

// FetchCustomerData queries real tables (customers, rfqs, bookings, shipments, customer_followup_tasks) with tenant isolation
func (p *CustomerLeadDataProvider) FetchCustomerData(ctx context.Context, orgID int64, customerID int64) (*DeterministicCustomerData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	custQuery := `
		SELECT id, org_id, COALESCE(name, ''), COALESCE(contact_name, ''),
		       COALESCE(contact_email, ''), COALESCE(contact_phone, ''),
		       COALESCE(status, 'ACTIVE'), COALESCE(health_score, 80),
		       COALESCE(credit_status, 'GOOD'), created_at, updated_at
		FROM customers
		WHERE id = ? AND org_id = ?`

	var d DeterministicCustomerData
	var createdAt, updatedAt time.Time
	err := p.db.QueryRowContext(ctx, custQuery, customerID, orgID).Scan(
		&d.CustomerID, &d.OrgID, &d.Name, &d.ContactName,
		&d.ContactEmail, &d.ContactPhone, &d.Status, &d.HealthScore,
		&d.CreditStatus, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("customer %d not found for organization %d", customerID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying customer %d: %w", customerID, err)
	}

	now := time.Now().UTC()
	d.TenureDays = int(now.Sub(createdAt).Hours() / 24.0)
	latestActivity := updatedAt
	if createdAt.After(latestActivity) {
		latestActivity = createdAt
	}

	// 1. Query RFQ volume & won RFQs
	rfqQuery := `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN status = 'WON' THEN 1 ELSE 0 END), 0),
		       MAX(created_at)
		FROM rfqs
		WHERE customer_id = ? AND org_id = ?`

	var latestRFQ sql.NullTime
	_ = p.db.QueryRowContext(ctx, rfqQuery, customerID, orgID).Scan(&d.TotalRFQs, &d.WonRFQs, &latestRFQ)
	if latestRFQ.Valid {
		d.RecentRFQDays = int(now.Sub(latestRFQ.Time).Hours() / 24.0)
		if latestRFQ.Time.After(latestActivity) {
			latestActivity = latestRFQ.Time
		}
	} else {
		d.RecentRFQDays = 999
	}

	// 2. Query bookings via RFQs
	bookingQuery := `
		SELECT COUNT(*)
		FROM bookings b
		INNER JOIN rfqs r ON b.rfq_id = r.id
		WHERE r.customer_id = ? AND r.org_id = ?`
	_ = p.db.QueryRowContext(ctx, bookingQuery, customerID, orgID).Scan(&d.TotalBookings)

	// 3. Query shipments via bookings & RFQs
	shipmentQuery := `
		SELECT COUNT(*), MAX(s.created_at)
		FROM shipments s
		INNER JOIN bookings b ON s.booking_id = b.id
		INNER JOIN rfqs r ON b.rfq_id = r.id
		WHERE r.customer_id = ? AND r.org_id = ?`
	var latestShipment sql.NullTime
	_ = p.db.QueryRowContext(ctx, shipmentQuery, customerID, orgID).Scan(&d.TotalShipments, &latestShipment)
	if latestShipment.Valid && latestShipment.Time.After(latestActivity) {
		latestActivity = latestShipment.Time
	}

	// 4. Query pending customer follow-up tasks
	taskQuery := `
		SELECT COUNT(*)
		FROM customer_followup_tasks
		WHERE customer_id = ? AND org_id = ? AND status != 'COMPLETED'`
	_ = p.db.QueryRowContext(ctx, taskQuery, customerID, orgID).Scan(&d.PendingTasksCount)

	// 5. Query unresolved quotations
	quoteQuery := `
		SELECT COUNT(*)
		FROM rfq_quotes q
		INNER JOIN rfqs r ON q.rfq_id = r.id
		WHERE r.customer_id = ? AND r.org_id = ? AND (q.status = 'PENDING' OR r.status = 'SUBMITTED')`
	var unresolvedCount int
	_ = p.db.QueryRowContext(ctx, quoteQuery, customerID, orgID).Scan(&unresolvedCount)
	d.HasUnresolvedQuotes = unresolvedCount > 0

	d.InactivityDays = int(now.Sub(latestActivity).Hours() / 24.0)

	// Construct context map for sidecar
	d.ContextMap = map[string]interface{}{
		"customer_id":           d.CustomerID,
		"name":                  d.Name,
		"contact_name":          d.ContactName,
		"contact_email":         d.ContactEmail,
		"contact_phone":         d.ContactPhone,
		"status":                d.Status,
		"health_score":          d.HealthScore,
		"credit_status":         d.CreditStatus,
		"tenure_days":           d.TenureDays,
		"inactivity_days":       d.InactivityDays,
		"total_rfqs":            d.TotalRFQs,
		"won_rfqs":              d.WonRFQs,
		"total_bookings":        d.TotalBookings,
		"total_shipments":       d.TotalShipments,
		"pending_tasks_count":   d.PendingTasksCount,
		"recent_rfq_days":       d.RecentRFQDays,
		"has_unresolved_quotes": d.HasUnresolvedQuotes,
	}

	return &d, nil
}

package bcontext

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// GetCustomer360Intelligence assembles the comprehensive, organization-isolated customer intelligence model.
func (s *defaultService) GetCustomer360Intelligence(ctx context.Context, orgID int64, customerID int64, correlationID string, requestingUserID int64) (*Customer360Intelligence, error) {
	if orgID <= 0 {
		return nil, errors.New("organization ID is required")
	}
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-cust360-%d-%d", customerID, time.Now().UnixNano())
	}

	// 1. Fetch Primary Customer Record
	var cust struct {
		ID             int64      `db:"id"`
		OrgID          int64      `db:"org_id"`
		Name           string     `db:"name"`
		CustomerCode   *string    `db:"customer_code"`
		TradingName    *string    `db:"trading_name"`
		CustomerType   string     `db:"customer_type"`
		Industry       *string    `db:"industry"`
		Status         string     `db:"status"`
		Country        *string    `db:"country"`
		City           *string    `db:"city"`
		Website        *string    `db:"website"`
		Currency       string     `db:"currency"`
		PaymentTerms   string     `db:"payment_terms"`
		CreditLimit    float64    `db:"credit_limit"`
		CreditStatus   string     `db:"credit_status"`
		HealthScore    int        `db:"health_score"`
		AccountOwnerID *int64     `db:"account_owner_id"`
		CreatedAt      *time.Time `db:"created_at"`
		UpdatedAt      *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &cust, `
		SELECT id, org_id, name, customer_code, trading_name, customer_type, industry, status, 
		       country, city, website, currency, payment_terms, credit_limit, credit_status, 
		       health_score, account_owner_id, created_at, updated_at
		FROM customers
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, customerID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer %d not found in organization", customerID)
		}
		return nil, fmt.Errorf("failed to fetch customer record: %w", err)
	}

	code := fmt.Sprintf("CUST-%d", cust.ID)
	if cust.CustomerCode != nil && *cust.CustomerCode != "" {
		code = *cust.CustomerCode
	}
	custName := cust.Name
	if strings.TrimSpace(custName) == "" {
		custName = code
	}
	tradingName := ""
	if cust.TradingName != nil {
		tradingName = *cust.TradingName
	}
	industry := ""
	if cust.Industry != nil {
		industry = *cust.Industry
	}
	country := ""
	if cust.Country != nil {
		country = *cust.Country
	}
	city := ""
	if cust.City != nil {
		city = *cust.City
	}
	website := ""
	if cust.Website != nil {
		website = *cust.Website
	}

	ownerName := "Unassigned"
	if cust.AccountOwnerID != nil && *cust.AccountOwnerID > 0 {
		var u struct {
			FirstName string `db:"first_name"`
			LastName  string `db:"last_name"`
		}
		if err := s.db.GetContext(ctx, &u, `SELECT first_name, last_name FROM users WHERE id = ? AND org_id = ? LIMIT 1`, *cust.AccountOwnerID, orgID); err == nil {
			ownerName = strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
	}

	// 2. Fetch Authorized Contacts
	var contactRows []struct {
		ID          int64   `db:"id"`
		FirstName   string  `db:"first_name"`
		LastName    string  `db:"last_name"`
		Email       *string `db:"email"`
		Phone       *string `db:"phone"`
		JobTitle    *string `db:"job_title"`
		Department  *string `db:"department"`
		ContactRole string  `db:"contact_role"`
		IsPrimary   bool    `db:"is_primary"`
	}
	_ = s.db.SelectContext(ctx, &contactRows, `
		SELECT id, first_name, last_name, email, phone, job_title, department, contact_role, is_primary
		FROM contacts
		WHERE customer_id = ? AND org_id = ?
		ORDER BY is_primary DESC, id ASC
		LIMIT 20
	`, customerID, orgID)

	contacts := make([]CustomerContactItem, 0, len(contactRows))
	for _, c := range contactRows {
		em := ""
		if c.Email != nil {
			em = *c.Email
		}
		ph := ""
		if c.Phone != nil {
			ph = *c.Phone
		}
		jt := ""
		if c.JobTitle != nil {
			jt = *c.JobTitle
		}
		dept := ""
		if c.Department != nil {
			dept = *c.Department
		}
		contacts = append(contacts, CustomerContactItem{
			ID:          c.ID,
			FirstName:   c.FirstName,
			LastName:    c.LastName,
			Email:       em,
			Phone:       ph,
			JobTitle:    jt,
			Department:  dept,
			ContactRole: c.ContactRole,
			IsPrimary:   c.IsPrimary,
		})
	}

	custSummary := CustomerIdentitySummary{
		ID:             cust.ID,
		OrgID:          cust.OrgID,
		Name:           custName,
		CustomerCode:   code,
		TradingName:    tradingName,
		CustomerType:   cust.CustomerType,
		Industry:       industry,
		Status:         cust.Status,
		Country:        country,
		City:           city,
		Website:        website,
		Currency:       cust.Currency,
		PaymentTerms:   cust.PaymentTerms,
		CreditLimit:    cust.CreditLimit,
		CreditStatus:   cust.CreditStatus,
		HealthScore:    cust.HealthScore,
		AccountOwnerID: cust.AccountOwnerID,
		AccountOwner:   ownerName,
		Contacts:       contacts,
		CreatedAt:      cust.CreatedAt,
		UpdatedAt:      cust.UpdatedAt,
	}

	sourceRefs := []SourceReference{
		{
			EntityType:      EntityTypeCustomer,
			EntityID:        cust.ID,
			ReferenceNumber: code,
			Label:           custName,
			Status:          cust.Status,
			Path:            fmt.Sprintf("/dashboard/customers/%d", cust.ID),
			KeyFields: map[string]interface{}{
				"credit_status": cust.CreditStatus,
				"health_score":  cust.HealthScore,
				"payment_terms": cust.PaymentTerms,
			},
		},
	}

	fieldRefs := []FieldReference{
		{SourceRecord: code, FieldName: "customer_name", FieldValue: custName},
		{SourceRecord: code, FieldName: "status", FieldValue: cust.Status},
		{SourceRecord: code, FieldName: "credit_status", FieldValue: cust.CreditStatus},
		{SourceRecord: code, FieldName: "health_score", FieldValue: strconv.Itoa(cust.HealthScore)},
	}

	observations := make([]CustomerObservation, 0)
	attentionItems := make([]string, 0)
	followUps := make([]string, 0)

	// 3. Commercial: RFQs, Quotes, Bookings, Contracts
	var rfqs []struct {
		ID        int64      `db:"id"`
		Number    *string    `db:"rfq_number"`
		Status    string     `db:"status"`
		Stage     *string    `db:"stage"`
		Origin    *string    `db:"origin"`
		Dest      *string    `db:"destination"`
		CreatedAt *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &rfqs, `
		SELECT id, rfq_number, status, stage, origin, destination, created_at
		FROM rfqs
		WHERE customer_id = ? AND org_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, customerID, orgID)

	totalRFQs := len(rfqs)
	openRFQs := 0
	for _, r := range rfqs {
		st := strings.ToUpper(r.Status)
		if st != "CLOSED" && st != "CANCELLED" && st != "WON" {
			openRFQs++
		}
		num := fmt.Sprintf("RFQ-%d", r.ID)
		if r.Number != nil && *r.Number != "" {
			num = *r.Number
		}
		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeRFQ,
			EntityID:        r.ID,
			ReferenceNumber: num,
			Label:           fmt.Sprintf("RFQ (%s)", r.Status),
			Status:          r.Status,
			Path:            fmt.Sprintf("/dashboard/rfqs?id=%d", r.ID),
		})
	}

	var quotes []struct {
		ID          int64      `db:"id"`
		Number      string     `db:"quotation_number"`
		Status      string     `db:"status"`
		TotalAmount float64    `db:"total_amount"`
		Currency    string     `db:"currency"`
		CreatedAt   *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &quotes, `
		SELECT id, quotation_number, status, total_amount, currency, created_at
		FROM quotations
		WHERE customer_id = ? AND org_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, customerID, orgID)

	totalQuotes := len(quotes)
	openQuotes := 0
	acceptedQuotes := 0
	totalQuotedAmount := 0.0
	acceptedQuotedAmount := 0.0

	for _, q := range quotes {
		totalQuotedAmount += q.TotalAmount
		st := strings.ToUpper(q.Status)
		if st == "ACCEPTED" || st == "APPROVED" || st == "WON" {
			acceptedQuotes++
			acceptedQuotedAmount += q.TotalAmount
		} else if st != "REJECTED" && st != "EXPIRED" && st != "CANCELLED" {
			openQuotes++
		}
		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeQuotation,
			EntityID:        q.ID,
			ReferenceNumber: q.Number,
			Label:           fmt.Sprintf("Quote %s %0.2f", q.Currency, q.TotalAmount),
			Status:          q.Status,
			Path:            fmt.Sprintf("/dashboard/quotes?id=%d", q.ID),
			KeyFields: map[string]interface{}{
				"total_amount": q.TotalAmount,
			},
		})
	}

	var bookings []struct {
		ID            int64      `db:"id"`
		BookingNumber string     `db:"booking_number"`
		Status        string     `db:"status"`
		CarrierName   string     `db:"carrier_name"`
		CreatedAt     *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &bookings, `
		SELECT b.id, b.booking_number, b.status, b.carrier_name, b.created_at
		FROM bookings b
		JOIN rfqs r ON b.rfq_id = r.id
		WHERE r.customer_id = ? AND b.org_id = ?
		ORDER BY b.created_at DESC
		LIMIT 50
	`, customerID, orgID)

	totalBookings := len(bookings)
	activeBookings := 0
	for _, b := range bookings {
		st := strings.ToUpper(b.Status)
		if st != "CANCELLED" && st != "COMPLETED" && st != "REJECTED" {
			activeBookings++
		}
		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeBooking,
			EntityID:        b.ID,
			ReferenceNumber: b.BookingNumber,
			Label:           fmt.Sprintf("Booking (%s)", b.CarrierName),
			Status:          b.Status,
			Path:            "/dashboard/bookings",
		})
	}

	var contracts []struct {
		ID                int64      `db:"id"`
		ContractReference string     `db:"contract_reference"`
		ContractName      string     `db:"contract_name"`
		Status            string     `db:"status"`
		ContractValue     float64    `db:"contract_value"`
		Currency          string     `db:"currency"`
		CreatedAt         *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &contracts, `
		SELECT DISTINCT c.id, c.contract_reference, c.contract_name, c.status, 
		       COALESCE(c.contract_value, 0.0) AS contract_value, c.currency, c.created_at
		FROM contracts c
		LEFT JOIN contract_parties cp ON c.party_id = cp.id
		LEFT JOIN contract_links cl ON c.id = cl.contract_id AND cl.linked_entity_type = 'CUSTOMER'
		WHERE c.org_id = ? AND (cp.customer_id = ? OR cl.linked_entity_id = ?)
		ORDER BY c.created_at DESC
		LIMIT 20
	`, orgID, customerID, customerID)

	totalContracts := len(contracts)
	activeContracts := 0
	for _, c := range contracts {
		if strings.ToUpper(c.Status) == "ACTIVE" {
			activeContracts++
		}
		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeContract,
			EntityID:        c.ID,
			ReferenceNumber: c.ContractReference,
			Label:           c.ContractName,
			Status:          c.Status,
			Path:            "/dashboard/rate-management/contracts",
		})
	}

	var convRate *float64
	convExplanation := "Insufficient quotation data to calculate quote-to-booking conversion rate."
	if totalQuotes > 0 {
		rate := (float64(totalBookings) / float64(totalQuotes)) * 100.0
		if rate > 100.0 {
			rate = 100.0
		}
		convRate = &rate
		convExplanation = fmt.Sprintf("Conversion rate is %.1f%% based on %d booking(s) across %d commercial quotation(s).", rate, totalBookings, totalQuotes)
	}

	var avgQuoteVal *float64
	if totalQuotes > 0 {
		avg := totalQuotedAmount / float64(totalQuotes)
		avgQuoteVal = &avg
	}

	commercialMetrics := CommercialIntelligenceMetrics{
		TotalRFQs:              totalRFQs,
		OpenRFQs:               openRFQs,
		TotalQuotations:        totalQuotes,
		OpenQuotations:         openQuotes,
		AcceptedQuotations:     acceptedQuotes,
		TotalBookings:          totalBookings,
		ActiveBookings:         activeBookings,
		TotalContracts:         totalContracts,
		ActiveContracts:        activeContracts,
		TotalQuotedAmount:      totalQuotedAmount,
		AcceptedQuotedAmount:   acceptedQuotedAmount,
		Currency:               cust.Currency,
		QuoteToBookingConvRate: convRate,
		AvgQuotationValue:      avgQuoteVal,
		ConversionExplanation:  convExplanation,
	}

	if totalRFQs > 0 {
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "total_rfqs", FieldValue: strconv.Itoa(totalRFQs)})
	}
	if totalBookings > 0 {
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "total_bookings", FieldValue: strconv.Itoa(totalBookings)})
	}

	// 4. Operations: Shipments & Exceptions
	var shipments []struct {
		ID          int64      `db:"id"`
		MBL         *string    `db:"mbl_number"`
		Status      string     `db:"status"`
		Origin      string     `db:"origin_port"`
		Destination string     `db:"destination_port"`
		ETD         *time.Time `db:"etd"`
		ETA         *time.Time `db:"eta"`
		CreatedAt   *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &shipments, `
		SELECT DISTINCT s.id, s.mbl_number, s.status, s.origin_port, s.destination_port, s.etd, s.eta, s.created_at
		FROM shipments s
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN bookings b ON s.booking_id = b.id
		LEFT JOIN rfqs rb ON b.rfq_id = rb.id
		WHERE (r.customer_id = ? OR rb.customer_id = ?) AND s.org_id = ?
		ORDER BY s.created_at DESC
		LIMIT 50
	`, customerID, customerID, orgID)

	totalShipments := len(shipments)
	activeShipments := 0
	deliveredShipments := 0
	delayedShipments := 0
	var lastShipmentDate *time.Time

	now := time.Now()
	for _, shp := range shipments {
		st := strings.ToUpper(shp.Status)
		if st == "DELIVERED" || st == "COMPLETED" {
			deliveredShipments++
		} else if st != "CANCELLED" {
			activeShipments++
		}

		if st == "DELAYED" || (shp.ETA != nil && shp.ETA.Before(now) && st != "DELIVERED" && st != "COMPLETED") {
			delayedShipments++
		}

		if shp.CreatedAt != nil {
			if lastShipmentDate == nil || shp.CreatedAt.After(*lastShipmentDate) {
				lastShipmentDate = shp.CreatedAt
			}
		}

		ref := fmt.Sprintf("SHP-%d", shp.ID)
		if shp.MBL != nil && *shp.MBL != "" {
			ref = *shp.MBL
		}
		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeShipment,
			EntityID:        shp.ID,
			ReferenceNumber: ref,
			Label:           fmt.Sprintf("%s → %s (%s)", shp.Origin, shp.Destination, shp.Status),
			Status:          shp.Status,
			Path:            fmt.Sprintf("/dashboard/shipments/%d", shp.ID),
		})
	}

	var excs []struct {
		ID       int64      `db:"id"`
		ShipID   int64      `db:"shipment_id"`
		Type     string     `db:"exception_type"`
		Severity string     `db:"severity"`
		Title    string     `db:"title"`
		Status   string     `db:"status"`
		Resolved bool       `db:"resolved"`
		Created  *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &excs, `
		SELECT se.id, se.shipment_id, se.exception_type, se.severity, se.title, se.status, se.resolved, se.created_at
		FROM shipment_exceptions se
		JOIN shipments s ON se.shipment_id = s.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN bookings b ON s.booking_id = b.id
		LEFT JOIN rfqs rb ON b.rfq_id = rb.id
		WHERE (r.customer_id = ? OR rb.customer_id = ?) AND se.org_id = ?
		ORDER BY se.created_at DESC
		LIMIT 20
	`, customerID, customerID, orgID)

	totalExceptions := len(excs)
	openExceptions := 0
	for _, e := range excs {
		if !e.Resolved && strings.ToUpper(e.Status) != "RESOLVED" && strings.ToUpper(e.Status) != "CLOSED" {
			openExceptions++
			observations = append(observations, CustomerObservation{
				Module:       "OPERATIONS",
				RecordType:   "SHIPMENT_EXCEPTION",
				RecordID:     e.ID,
				Reference:    fmt.Sprintf("EXC-%d (Shipment #%d)", e.ID, e.ShipID),
				Finding:      fmt.Sprintf("%s: %s (Severity: %s)", e.Type, e.Title, e.Severity),
				Significance: "CRITICAL",
				Timestamp:    e.Created,
			})
			attentionItems = append(attentionItems, fmt.Sprintf("Shipment #%d has open exception: %s (%s)", e.ShipID, e.Title, e.Severity))
		}
	}

	deliveryNote := "Insufficient completed shipments to calculate average transit cycle."
	if deliveredShipments > 0 {
		deliveryNote = fmt.Sprintf("%d of %d shipment(s) delivered successfully.", deliveredShipments, totalShipments)
	}

	operationsMetrics := OperationsIntelligenceMetrics{
		TotalShipments:          totalShipments,
		ActiveShipments:         activeShipments,
		DeliveredShipments:      deliveredShipments,
		DelayedShipments:        delayedShipments,
		OpenExceptionsCount:     openExceptions,
		TotalExceptionsCount:    totalExceptions,
		LastShipmentDate:        lastShipmentDate,
		DeliveryPerformanceNote: deliveryNote,
	}

	if totalShipments > 0 {
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "total_shipments", FieldValue: strconv.Itoa(totalShipments)})
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "active_shipments", FieldValue: strconv.Itoa(activeShipments)})
	}

	// 5. Financial: Invoices & Balances
	var invs []struct {
		ID          int64      `db:"id"`
		Number      string     `db:"invoice_number"`
		Status      string     `db:"status"`
		TotalAmount float64    `db:"total_amount"`
		BalanceDue  float64    `db:"balance_due"`
		Currency    string     `db:"currency"`
		DueDate     *time.Time `db:"due_date"`
		CreatedAt   *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &invs, `
		SELECT id, invoice_number, status, total_amount, balance_due, currency, due_date, created_at
		FROM customer_invoices
		WHERE customer_id = ? AND org_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, customerID, orgID)

	totalInvoiced := 0.0
	outstandingBalance := 0.0
	overdueCount := 0
	paidCount := 0

	for _, inv := range invs {
		totalInvoiced += inv.TotalAmount
		outstandingBalance += inv.BalanceDue

		st := strings.ToUpper(inv.Status)
		if st == "PAID" || inv.BalanceDue <= 0.001 {
			paidCount++
		}

		isOverdue := st == "OVERDUE" || (inv.DueDate != nil && inv.DueDate.Before(now) && inv.BalanceDue > 0.01)
		if isOverdue {
			overdueCount++
			observations = append(observations, CustomerObservation{
				Module:       "FINANCIAL",
				RecordType:   "INVOICE",
				RecordID:     inv.ID,
				Reference:    inv.Number,
				Finding:      fmt.Sprintf("Invoice %s has overdue balance of %s %.2f (Total: %.2f)", inv.Number, inv.Currency, inv.BalanceDue, inv.TotalAmount),
				Significance: "ATTENTION",
				Timestamp:    inv.CreatedAt,
			})
			attentionItems = append(attentionItems, fmt.Sprintf("Invoice %s is overdue (%s %.2f outstanding)", inv.Number, inv.Currency, inv.BalanceDue))
		}

		sourceRefs = append(sourceRefs, SourceReference{
			EntityType:      EntityTypeInvoice,
			EntityID:        inv.ID,
			ReferenceNumber: inv.Number,
			Label:           fmt.Sprintf("Invoice %s %0.2f (%s)", inv.Currency, inv.TotalAmount, inv.Status),
			Status:          inv.Status,
			Path:            fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID),
			KeyFields: map[string]interface{}{
				"balance_due": inv.BalanceDue,
			},
		})
	}

	totalPaid := math.Max(0, totalInvoiced-outstandingBalance)
	complianceRate := 100.0
	if len(invs) > 0 {
		complianceRate = (float64(len(invs)-overdueCount) / float64(len(invs))) * 100.0
	}

	financialMetrics := FinancialIntelligenceMetrics{
		TotalInvoicedAmount:   totalInvoiced,
		TotalPaidAmount:       totalPaid,
		OutstandingBalance:    outstandingBalance,
		OverdueInvoicesCount:  overdueCount,
		TotalInvoicesCount:    len(invs),
		PaidInvoicesCount:     paidCount,
		Currency:              cust.Currency,
		PaymentComplianceRate: complianceRate,
	}

	if len(invs) > 0 {
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "total_invoiced", FieldValue: fmt.Sprintf("%s %.2f", cust.Currency, totalInvoiced)})
		fieldRefs = append(fieldRefs, FieldReference{SourceRecord: code, FieldName: "outstanding_balance", FieldValue: fmt.Sprintf("%s %.2f", cust.Currency, outstandingBalance)})
	}

	// 6. Governance & Activity: Approvals, AI Tasks, Audit Logs, Lead Interactions
	var approvals []struct {
		ID        int64      `db:"id"`
		Code      string     `db:"request_code"`
		Title     string     `db:"title"`
		Status    string     `db:"status"`
		Type      string     `db:"type"`
		CreatedAt *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &approvals, `
		SELECT id, request_code, title, status, type, created_at
		FROM approval_requests
		WHERE customer_id = ? AND org_id = ?
		ORDER BY created_at DESC
		LIMIT 20
	`, customerID, orgID)

	openApprovals := 0
	for _, a := range approvals {
		st := strings.ToUpper(a.Status)
		if st == "PENDING" || st == "PENDING_APPROVAL" || st == "IN_REVIEW" {
			openApprovals++
			attentionItems = append(attentionItems, fmt.Sprintf("Pending approval: %s (%s)", a.Title, a.Code))
			observations = append(observations, CustomerObservation{
				Module:       "GOVERNANCE",
				RecordType:   "APPROVAL_REQUEST",
				RecordID:     a.ID,
				Reference:    a.Code,
				Finding:      fmt.Sprintf("Approval pending for %s (%s)", a.Title, a.Type),
				Significance: "ATTENTION",
				Timestamp:    a.CreatedAt,
			})
		}
	}

	var aiTasksCount int
	_ = s.db.GetContext(ctx, &aiTasksCount, `
		SELECT COUNT(*)
		FROM ai_processing_tasks
		WHERE org_id = ? AND entity_type = 'CUSTOMER' AND entity_id = ?
	`, orgID, strconv.FormatInt(customerID, 10))

	var auditLogsCount int
	_ = s.db.GetContext(ctx, &auditLogsCount, `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE org_id = ? AND resource_type = 'CUSTOMER' AND resource_id = ?
	`, orgID, strconv.FormatInt(customerID, 10))

	var interactions []struct {
		ID      int64      `db:"id"`
		Type    string     `db:"interaction_type"`
		Subject *string    `db:"subject"`
		Created *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &interactions, `
		SELECT li.id, li.interaction_type, li.subject, li.created_at
		FROM lead_interactions li
		JOIN customer_lead_links cll ON li.lead_id = cll.lead_id
		WHERE cll.customer_id = ? AND li.org_id = ?
		ORDER BY li.created_at DESC
		LIMIT 10
	`, customerID, orgID)

	var lastInteractionDate *time.Time
	if len(interactions) > 0 && interactions[0].Created != nil {
		lastInteractionDate = interactions[0].Created
	}

	// Calculate Engagement Trend
	recentEvents30d := 0
	recentEvents90d := 0
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	for _, r := range rfqs {
		if r.CreatedAt != nil {
			if r.CreatedAt.After(thirtyDaysAgo) {
				recentEvents30d++
			}
			if r.CreatedAt.After(ninetyDaysAgo) {
				recentEvents90d++
			}
		}
	}
	for _, s := range shipments {
		if s.CreatedAt != nil {
			if s.CreatedAt.After(thirtyDaysAgo) {
				recentEvents30d++
			}
			if s.CreatedAt.After(ninetyDaysAgo) {
				recentEvents90d++
			}
		}
	}
	for _, it := range interactions {
		if it.Created != nil {
			if it.Created.After(thirtyDaysAgo) {
				recentEvents30d++
			}
			if it.Created.After(ninetyDaysAgo) {
				recentEvents90d++
			}
		}
	}

	engagementTrend := "INACTIVE"
	engagementExp := "No operational inquiries, shipments, or communications recorded in the past 90 days."
	if recentEvents30d >= 3 {
		engagementTrend = "HIGH"
		engagementExp = fmt.Sprintf("High engagement with %d business events in the last 30 days.", recentEvents30d)
	} else if recentEvents30d >= 1 {
		engagementTrend = "MODERATE"
		engagementExp = fmt.Sprintf("Moderate engagement with %d active interaction(s) in the last 30 days.", recentEvents30d)
	} else if recentEvents90d >= 1 {
		engagementTrend = "LOW"
		engagementExp = fmt.Sprintf("Low engagement with %d event(s) in the past 90 days, but none in the last 30 days.", recentEvents90d)
	}

	governanceMetrics := GovernanceAndActivityMetrics{
		RecentInteractionsCount: len(interactions),
		LastInteractionDate:     lastInteractionDate,
		OpenApprovalsCount:      openApprovals,
		RecentAITasksCount:      aiTasksCount,
		RecentAuditLogsCount:    auditLogsCount,
		EngagementTrend:         engagementTrend,
		EngagementExplanation:   engagementExp,
	}

	// 7. Grounded AI Summary Synthesis
	var execSummary strings.Builder
	execSummary.WriteString(fmt.Sprintf("%s (%s) is an %s account with health score %d/100 and credit status '%s'. ",
		custName, code, strings.ToLower(cust.Status), cust.HealthScore, cust.CreditStatus))

	if totalShipments > 0 || totalRFQs > 0 {
		execSummary.WriteString(fmt.Sprintf("To date, the customer has %d RFQ(s), %d commercial quotation(s), and %d freight shipment(s) on record. ",
			totalRFQs, totalQuotes, totalShipments))
	} else {
		execSummary.WriteString("No operational freight movements or inquiries are currently active for this account. ")
	}

	if outstandingBalance > 0 {
		execSummary.WriteString(fmt.Sprintf("Commercial invoices show an outstanding balance of %s %.2f with %d overdue invoice(s).",
			cust.Currency, outstandingBalance, overdueCount))
	} else if len(invs) > 0 {
		execSummary.WriteString("All commercial invoices are fully settled with zero overdue balances.")
	}

	commPosition := fmt.Sprintf("%s maintains %d RFQ(s) and %d quotation(s) totaling %s %.2f.",
		custName, totalRFQs, totalQuotes, cust.Currency, totalQuotedAmount)
	if activeContracts > 0 {
		commPosition += fmt.Sprintf(" Account has %d active long-term service contract(s).", activeContracts)
	}

	opsPosition := fmt.Sprintf("Operations show %d shipment(s) total with %d currently active and %d open operational exception(s).",
		totalShipments, activeShipments, openExceptions)
	if delayedShipments > 0 {
		opsPosition += fmt.Sprintf(" Note: %d shipment(s) have schedule variance or delay alerts.", delayedShipments)
	}

	finConcerns := "Accounts receivable are in good order with no overdue invoices."
	if overdueCount > 0 {
		finConcerns = fmt.Sprintf("Attention required: %d invoice(s) totaling %s %.2f are past due.",
			overdueCount, cust.Currency, outstandingBalance)
		followUps = append(followUps, fmt.Sprintf("Finance team: Issue payment reminder for %d overdue invoice(s)", overdueCount))
	}

	if openExceptions > 0 {
		followUps = append(followUps, fmt.Sprintf("Operations team: Review and resolve %d active shipment exception(s)", openExceptions))
	}
	if openApprovals > 0 {
		followUps = append(followUps, fmt.Sprintf("Management: Complete review on %d pending approval request(s)", openApprovals))
	}
	if openRFQs > 0 && openQuotes == 0 {
		followUps = append(followUps, "Commercial team: Provide carrier pricing quotation for open RFQ inquiries")
	}

	confidence := "HIGH"
	if totalRFQs == 0 && totalShipments == 0 && len(invs) == 0 {
		confidence = "MEDIUM"
	}

	aiSummary := CustomerAISummary{
		ExecutiveSummary:    execSummary.String(),
		CommercialPosition:  commPosition,
		OperationalPosition: opsPosition,
		FinancialConcerns:   finConcerns,
		AttentionItems:      attentionItems,
		RecommendedFollowUp: followUps,
		KeyObservations:     observations,
		ConfidenceLevel:     confidence,
		DataFreshness:       time.Now().UTC().Format(time.RFC3339),
		IsInformationalOnly: true,
	}

	return &Customer360Intelligence{
		OrgID:                     orgID,
		CustomerID:                customerID,
		Customer:                  custSummary,
		CommercialMetrics:         commercialMetrics,
		OperationsMetrics:         operationsMetrics,
		FinancialMetrics:          financialMetrics,
		GovernanceMetrics:         governanceMetrics,
		AISummary:                 aiSummary,
		SupportingRecords:         sourceRefs,
		SupportingFieldReferences: fieldRefs,
		CorrelationID:             correlationID,
		DataFreshness:             time.Now().UTC(),
		IsReadOnly:                true,
	}, nil
}

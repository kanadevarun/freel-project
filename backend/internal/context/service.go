package bcontext

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/rbac"
	"github.com/jmoiron/sqlx"
)

// Service defines the contract for read-only business context and intelligence insights.
type Service interface {
	GetBusinessContext(ctx context.Context, req ContextRequest) (*BusinessContext, error)
	GenerateInsight(ctx context.Context, req InsightRequest) (*IntelligenceInsight, error)
	GetCustomer360Intelligence(ctx context.Context, orgID int64, customerID int64, correlationID string, requestingUserID int64) (*Customer360Intelligence, error)
	GetRFQ360PricingIntelligence(ctx context.Context, orgID int64, rfqID int64, correlationID string, requestingUserID int64) (*RFQ360PricingIntelligence, error)
	GetShipment360OperationsIntelligence(ctx context.Context, orgID int64, shipmentID int64, correlationID string, requestingUserID int64) (*Shipment360OperationsIntelligence, error)
	GetOrgOperationsSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*OrgOperationsSummary, error)
	GetInvoice360FinanceIntelligence(ctx context.Context, orgID int64, invoiceID int64, correlationID string, requestingUserID int64) (*Invoice360FinanceIntelligence, error)
	GetOrgFinanceSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*OrgFinanceSummary, error)
	GetContract360ComplianceIntelligence(ctx context.Context, orgID int64, contractID int64, correlationID string, requestingUserID int64) (*Contract360ComplianceIntelligence, error)
	GetContractCoverageForEntity(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, requestingUserID int64) (*ContractCoverageEvaluation, error)
	GetOrgContractComplianceSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*OrgContractComplianceSummary, error)
	GetCrossModuleInsights(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, requestingUserID int64) (*CrossModuleInsightsResult, error)
	GetOrgCrossModuleSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*OrgCrossModuleSummary, error)
}

type defaultService struct {
	db      *sqlx.DB
	rbacSvc rbac.Service
}

// NewService creates a new business context service instance.
func NewService(db *sqlx.DB, rbacSvc rbac.Service) Service {
	return &defaultService{
		db:      db,
		rbacSvc: rbacSvc,
	}
}

// GetBusinessContext retrieves primary and related records strictly scoped to the organization.
func (s *defaultService) GetBusinessContext(ctx context.Context, req ContextRequest) (*BusinessContext, error) {
	if req.OrgID <= 0 {
		return nil, errors.New("invalid organization ID")
	}
	if req.PrimaryID <= 0 {
		return nil, errors.New("invalid primary record ID")
	}

	bCtx := &BusinessContext{
		OrgID:                 req.OrgID,
		RequestingUserID:      req.UserID,
		RequestingUserRole:    req.UserRole,
		PrimaryType:           req.PrimaryType,
		PrimaryID:             req.PrimaryID,
		RelatedRecords:        make(map[EntityType][]RecordSummary),
		OperationalExceptions: make([]ExceptionSummary, 0),
		Milestones:            make([]MilestoneSummary, 0),
		RecentActivity:        make([]ActivitySummary, 0),
		SourceReferences:      make([]SourceReference, 0),
		CorrelationID:         req.CorrelationID,
		DataFreshness:         time.Now().UTC(),
		IsReadOnly:            true,
	}

	switch req.PrimaryType {
	case EntityTypeCustomer:
		if err := s.buildCustomerContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeRFQ:
		if err := s.buildRFQContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeShipment:
		if err := s.buildShipmentContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeInvoice:
		if err := s.buildInvoiceContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeQuotation:
		if err := s.buildQuotationContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeLead:
		if err := s.buildLeadContext(ctx, bCtx); err != nil {
			return nil, err
		}
	case EntityTypeContract:
		if err := s.buildContractContext(ctx, bCtx); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", req.PrimaryType)
	}

	return bCtx, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CUSTOMER CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildCustomerContext(ctx context.Context, bCtx *BusinessContext) error {
	// 1. Primary Customer Record
	var cust struct {
		ID            int64      `db:"id"`
		Name          string     `db:"name"`
		CustomerCode  *string    `db:"customer_code"`
		Status        string     `db:"status"`
		CreditStatus  string     `db:"credit_status"`
		CustomerType  string     `db:"customer_type"`
		Currency      string     `db:"currency"`
		PaymentTerms  string     `db:"payment_terms"`
		CreditLimit   float64    `db:"credit_limit"`
		HealthScore   int        `db:"health_score"`
		Country       *string    `db:"country"`
		City          *string    `db:"city"`
		CreatedAt     *time.Time `db:"created_at"`
		UpdatedAt     *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &cust, `
		SELECT id, name, customer_code, status, credit_status, customer_type, currency, payment_terms, credit_limit, health_score, country, city, created_at, updated_at
		FROM customers
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("customer %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	code := ""
	if cust.CustomerCode != nil {
		code = *cust.CustomerCode
	}
	country := ""
	if cust.Country != nil {
		country = *cust.Country
	}
	city := ""
	if cust.City != nil {
		city = *cust.City
	}

	title := cust.Name
	if strings.TrimSpace(title) == "" {
		if code != "" {
			title = code
		} else {
			title = fmt.Sprintf("Customer #%d", cust.ID)
		}
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeCustomer,
		ID:              cust.ID,
		ReferenceNumber: code,
		Title:           title,
		Status:          cust.Status,
		CreatedAt:       cust.CreatedAt,
		UpdatedAt:       cust.UpdatedAt,
		FinancialValues: map[string]interface{}{
			"credit_limit":  cust.CreditLimit,
			"credit_status": cust.CreditStatus,
			"currency":      cust.Currency,
			"payment_terms": cust.PaymentTerms,
		},
		KeyAttributes: map[string]interface{}{
			"health_score":  cust.HealthScore,
			"customer_type": cust.CustomerType,
			"country":       country,
			"city":          city,
		},
		DeepLink: fmt.Sprintf("/dashboard/customers/%d", cust.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeCustomer,
		EntityID:        cust.ID,
		ReferenceNumber: code,
		Label:           cust.Name,
		Status:          cust.Status,
		Path:            fmt.Sprintf("/dashboard/customers/%d", cust.ID),
		KeyFields: map[string]interface{}{
			"credit_status": cust.CreditStatus,
			"health_score":  cust.HealthScore,
		},
	})

	// 2. Related RFQs
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
		ORDER BY id DESC LIMIT 10
	`, bCtx.PrimaryID, bCtx.OrgID)

	rfqList := make([]RecordSummary, 0, len(rfqs))
	for _, r := range rfqs {
		num := fmt.Sprintf("RFQ-%d", r.ID)
		if r.Number != nil && *r.Number != "" {
			num = *r.Number
		}
		stage := ""
		if r.Stage != nil {
			stage = *r.Stage
		}
		orig := ""
		if r.Origin != nil {
			orig = *r.Origin
		}
		dest := ""
		if r.Dest != nil {
			dest = *r.Dest
		}

		item := RecordSummary{
			EntityType:      EntityTypeRFQ,
			ID:              r.ID,
			ReferenceNumber: num,
			Title:           fmt.Sprintf("%s → %s", orig, dest),
			Status:          r.Status,
			Stage:           stage,
			CreatedAt:       r.CreatedAt,
			DeepLink:        fmt.Sprintf("/dashboard/rfqs?id=%d", r.ID),
		}
		rfqList = append(rfqList, item)
		bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
			EntityType:      EntityTypeRFQ,
			EntityID:        r.ID,
			ReferenceNumber: num,
			Label:           item.Title,
			Status:          r.Status,
			Path:            item.DeepLink,
		})
	}
	bCtx.RelatedRecords[EntityTypeRFQ] = rfqList

	// 3. Related Invoices
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
		ORDER BY id DESC LIMIT 10
	`, bCtx.PrimaryID, bCtx.OrgID)

	invList := make([]RecordSummary, 0, len(invs))
	for _, inv := range invs {
		item := RecordSummary{
			EntityType:      EntityTypeInvoice,
			ID:              inv.ID,
			ReferenceNumber: inv.Number,
			Title:           fmt.Sprintf("Invoice %s", inv.Number),
			Status:          inv.Status,
			CreatedAt:       inv.CreatedAt,
			FinancialValues: map[string]interface{}{
				"total_amount": inv.TotalAmount,
				"balance_due":  inv.BalanceDue,
				"currency":     inv.Currency,
			},
			DeepLink: fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID),
		}
		invList = append(invList, item)
		bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
			EntityType:      EntityTypeInvoice,
			EntityID:        inv.ID,
			ReferenceNumber: inv.Number,
			Label:           fmt.Sprintf("%s %0.2f (%s)", inv.Currency, inv.TotalAmount, inv.Status),
			Status:          inv.Status,
			Path:            item.DeepLink,
			KeyFields: map[string]interface{}{
				"balance_due": inv.BalanceDue,
			},
		})
	}
	bCtx.RelatedRecords[EntityTypeInvoice] = invList

	// 4. Related Quotations
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
		ORDER BY id DESC LIMIT 5
	`, bCtx.PrimaryID, bCtx.OrgID)

	quoteList := make([]RecordSummary, 0, len(quotes))
	for _, q := range quotes {
		item := RecordSummary{
			EntityType:      EntityTypeQuotation,
			ID:              q.ID,
			ReferenceNumber: q.Number,
			Title:           fmt.Sprintf("Quote %s", q.Number),
			Status:          q.Status,
			CreatedAt:       q.CreatedAt,
			FinancialValues: map[string]interface{}{
				"total_amount": q.TotalAmount,
				"currency":     q.Currency,
			},
			DeepLink: fmt.Sprintf("/dashboard/quotes?id=%d", q.ID),
		}
		quoteList = append(quoteList, item)
	}
	bCtx.RelatedRecords[EntityTypeQuotation] = quoteList

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RFQ CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildRFQContext(ctx context.Context, bCtx *BusinessContext) error {
	var rfq struct {
		ID          int64      `db:"id"`
		Number      *string    `db:"rfq_number"`
		CustomerID  int64      `db:"customer_id"`
		LeadID      *int64     `db:"lead_id"`
		Status      string     `db:"status"`
		Stage       *string    `db:"stage"`
		AgentStatus string     `db:"agent_status"`
		Origin      *string    `db:"origin"`
		Destination *string    `db:"destination"`
		Incoterms   *string    `db:"incoterms"`
		HealthScore int        `db:"health_score"`
		TargetDate  *time.Time `db:"target_date"`
		CreatedAt   *time.Time `db:"created_at"`
		UpdatedAt   *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &rfq, `
		SELECT id, rfq_number, customer_id, lead_id, status, stage, agent_status, origin, destination, incoterms, health_score, target_date, created_at, updated_at
		FROM rfqs
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("rfq %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	num := fmt.Sprintf("RFQ-%d", rfq.ID)
	if rfq.Number != nil && *rfq.Number != "" {
		num = *rfq.Number
	}
	orig := ""
	if rfq.Origin != nil {
		orig = *rfq.Origin
	}
	dest := ""
	if rfq.Destination != nil {
		dest = *rfq.Destination
	}
	stage := "DRAFT"
	if rfq.Stage != nil {
		stage = *rfq.Stage
	}
	incoterms := ""
	if rfq.Incoterms != nil {
		incoterms = *rfq.Incoterms
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeRFQ,
		ID:              rfq.ID,
		ReferenceNumber: num,
		Title:           fmt.Sprintf("%s → %s", orig, dest),
		Status:          rfq.Status,
		Stage:           stage,
		CreatedAt:       rfq.CreatedAt,
		UpdatedAt:       rfq.UpdatedAt,
		KeyAttributes: map[string]interface{}{
			"origin":       orig,
			"destination":  dest,
			"incoterms":    incoterms,
			"agent_status": rfq.AgentStatus,
			"health_score": rfq.HealthScore,
			"customer_id":  rfq.CustomerID,
		},
		DeepLink: fmt.Sprintf("/dashboard/rfqs?id=%d", rfq.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeRFQ,
		EntityID:        rfq.ID,
		ReferenceNumber: num,
		Label:           fmt.Sprintf("%s to %s", orig, dest),
		Status:          rfq.Status,
		Path:            fmt.Sprintf("/dashboard/rfqs?id=%d", rfq.ID),
		KeyFields: map[string]interface{}{
			"origin":      orig,
			"destination": dest,
			"incoterms":   incoterms,
		},
	})

	// 1. Related Customer
	if rfq.CustomerID > 0 {
		var custName string
		_ = s.db.GetContext(ctx, &custName, `SELECT name FROM customers WHERE id = ? AND org_id = ? LIMIT 1`, rfq.CustomerID, bCtx.OrgID)
		if custName != "" {
			bCtx.RelatedRecords[EntityTypeCustomer] = []RecordSummary{
				{
					EntityType:      EntityTypeCustomer,
					ID:              rfq.CustomerID,
					ReferenceNumber: fmt.Sprintf("CUST-%d", rfq.CustomerID),
					Title:           custName,
					Status:          "ACTIVE",
					DeepLink:        fmt.Sprintf("/dashboard/customers/%d", rfq.CustomerID),
				},
			}
			bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
				EntityType:      EntityTypeCustomer,
				EntityID:        rfq.CustomerID,
				ReferenceNumber: fmt.Sprintf("CUST-%d", rfq.CustomerID),
				Label:           custName,
				Status:          "ACTIVE",
				Path:            fmt.Sprintf("/dashboard/customers/%d", rfq.CustomerID),
			})
		}
	}

	// 2. Related Quotations
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
		WHERE rfq_id = ? AND org_id = ?
		ORDER BY id DESC LIMIT 5
	`, bCtx.PrimaryID, bCtx.OrgID)

	quoteList := make([]RecordSummary, 0, len(quotes))
	for _, q := range quotes {
		item := RecordSummary{
			EntityType:      EntityTypeQuotation,
			ID:              q.ID,
			ReferenceNumber: q.Number,
			Title:           fmt.Sprintf("Quotation %s", q.Number),
			Status:          q.Status,
			CreatedAt:       q.CreatedAt,
			FinancialValues: map[string]interface{}{
				"total_amount": q.TotalAmount,
				"currency":     q.Currency,
			},
			DeepLink: fmt.Sprintf("/dashboard/quotes?id=%d", q.ID),
		}
		quoteList = append(quoteList, item)
		bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
			EntityType:      EntityTypeQuotation,
			EntityID:        q.ID,
			ReferenceNumber: q.Number,
			Label:           fmt.Sprintf("%s %0.2f (%s)", q.Currency, q.TotalAmount, q.Status),
			Status:          q.Status,
			Path:            item.DeepLink,
			KeyFields: map[string]interface{}{
				"total_amount": q.TotalAmount,
			},
		})
	}
	bCtx.RelatedRecords[EntityTypeQuotation] = quoteList

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SHIPMENT CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildShipmentContext(ctx context.Context, bCtx *BusinessContext) error {
	var shp struct {
		ID            int64      `db:"id"`
		BookingNumber *string    `db:"booking_number"`
		BookingID     *int64     `db:"booking_id"`
		RFQID         *int64     `db:"rfq_id"`
		QuoteID       *int64     `db:"quote_id"`
		CarrierSCAC   string     `db:"carrier_scac"`
		OriginPort    string     `db:"origin_port"`
		DestPort      string     `db:"destination_port"`
		VesselName    *string    `db:"vessel_name"`
		VoyageNumber  *string    `db:"voyage_number"`
		Status        *string    `db:"status"`
		ETD           *time.Time `db:"etd"`
		ETA           *time.Time `db:"eta"`
		CreatedAt     *time.Time `db:"created_at"`
		UpdatedAt     *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &shp, `
		SELECT id, booking_number, booking_id, rfq_id, quote_id, carrier_scac, origin_port, destination_port, vessel_name, voyage_number, status, etd, eta, created_at, updated_at
		FROM shipments
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("shipment %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	bkgNum := fmt.Sprintf("SHP-%d", shp.ID)
	if shp.BookingNumber != nil && *shp.BookingNumber != "" {
		bkgNum = *shp.BookingNumber
	}
	status := "BOOKED"
	if shp.Status != nil && *shp.Status != "" {
		status = *shp.Status
	}
	vessel := ""
	if shp.VesselName != nil {
		vessel = *shp.VesselName
	}
	voyage := ""
	if shp.VoyageNumber != nil {
		voyage = *shp.VoyageNumber
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeShipment,
		ID:              shp.ID,
		ReferenceNumber: bkgNum,
		Title:           fmt.Sprintf("%s → %s via %s", shp.OriginPort, shp.DestPort, shp.CarrierSCAC),
		Status:          status,
		CreatedAt:       shp.CreatedAt,
		UpdatedAt:       shp.UpdatedAt,
		KeyAttributes: map[string]interface{}{
			"carrier_scac":     shp.CarrierSCAC,
			"origin_port":      shp.OriginPort,
			"destination_port": shp.DestPort,
			"vessel_name":      vessel,
			"voyage_number":    voyage,
		},
		DeepLink: fmt.Sprintf("/dashboard/shipments?id=%d", shp.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeShipment,
		EntityID:        shp.ID,
		ReferenceNumber: bkgNum,
		Label:           fmt.Sprintf("%s to %s", shp.OriginPort, shp.DestPort),
		Status:          status,
		Path:            fmt.Sprintf("/dashboard/shipments?id=%d", shp.ID),
		KeyFields: map[string]interface{}{
			"carrier": shp.CarrierSCAC,
			"status":  status,
		},
	})

	// 1. Shipment Milestones
	var milestones []struct {
		ID          int64      `db:"id"`
		Code        string     `db:"milestone_code"`
		Description *string    `db:"description"`
		Status      string     `db:"status"`
		PlannedDate *time.Time `db:"planned_date"`
		ActualDate  *time.Time `db:"actual_date"`
		Location    *string    `db:"location"`
	}
	_ = s.db.SelectContext(ctx, &milestones, `
		SELECT id, milestone_code, description, status, planned_date, actual_date, location
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY id ASC
	`, bCtx.PrimaryID)

	for _, m := range milestones {
		desc := m.Code
		if m.Description != nil && *m.Description != "" {
			desc = *m.Description
		}
		loc := ""
		if m.Location != nil {
			loc = *m.Location
		}
		bCtx.Milestones = append(bCtx.Milestones, MilestoneSummary{
			ID:          m.ID,
			Code:        m.Code,
			Description: desc,
			Status:      m.Status,
			PlannedDate: m.PlannedDate,
			ActualDate:  m.ActualDate,
			Location:    loc,
		})
	}

	// 2. Shipment Exceptions
	var exceptions []struct {
		ID          int64      `db:"id"`
		Type        string     `db:"exception_type"`
		Severity    string     `db:"severity"`
		Title       string     `db:"title"`
		Description *string    `db:"description"`
		Status      string     `db:"status"`
		CreatedAt   *time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &exceptions, `
		SELECT id, exception_type, severity, title, description, status, created_at
		FROM shipment_exceptions
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY id DESC
	`, bCtx.PrimaryID, bCtx.OrgID)

	for _, e := range exceptions {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		bCtx.OperationalExceptions = append(bCtx.OperationalExceptions, ExceptionSummary{
			ID:          e.ID,
			Type:        e.Type,
			Severity:    e.Severity,
			Title:       e.Title,
			Description: desc,
			Status:      e.Status,
			CreatedAt:   e.CreatedAt,
		})
	}

	// 3. Related Customer Invoices
	var invs []struct {
		ID          int64      `db:"id"`
		Number      string     `db:"invoice_number"`
		Status      string     `db:"status"`
		TotalAmount float64    `db:"total_amount"`
		BalanceDue  float64    `db:"balance_due"`
		Currency    string     `db:"currency"`
	}
	_ = s.db.SelectContext(ctx, &invs, `
		SELECT id, invoice_number, status, total_amount, balance_due, currency
		FROM customer_invoices
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY id DESC LIMIT 5
	`, bCtx.PrimaryID, bCtx.OrgID)

	invList := make([]RecordSummary, 0, len(invs))
	for _, inv := range invs {
		invList = append(invList, RecordSummary{
			EntityType:      EntityTypeInvoice,
			ID:              inv.ID,
			ReferenceNumber: inv.Number,
			Title:           fmt.Sprintf("Invoice %s", inv.Number),
			Status:          inv.Status,
			FinancialValues: map[string]interface{}{
				"total_amount": inv.TotalAmount,
				"balance_due":  inv.BalanceDue,
				"currency":     inv.Currency,
			},
			DeepLink: fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID),
		})
	}
	bCtx.RelatedRecords[EntityTypeInvoice] = invList

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// INVOICE CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildInvoiceContext(ctx context.Context, bCtx *BusinessContext) error {
	var inv struct {
		ID             int64      `db:"id"`
		InvoiceNumber  string     `db:"invoice_number"`
		CustomerID     int64      `db:"customer_id"`
		CustomerName   string     `db:"customer_name"`
		ShipmentID     *int64     `db:"shipment_id"`
		ShipmentNumber string     `db:"shipment_number"`
		BookingID      *int64     `db:"booking_id"`
		BookingNumber  string     `db:"booking_number"`
		QuotationID    *int64     `db:"quotation_id"`
		QuoteNumber    string     `db:"quote_number"`
		Route          string     `db:"route"`
		InvoiceDate    time.Time  `db:"invoice_date"`
		DueDate        time.Time  `db:"due_date"`
		Currency       string     `db:"currency"`
		Subtotal       float64    `db:"subtotal"`
		TaxAmount      float64    `db:"tax_amount"`
		TotalAmount    float64    `db:"total_amount"`
		PaidAmount     float64    `db:"paid_amount"`
		BalanceDue     float64    `db:"balance_due"`
		Status         string     `db:"status"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &inv, `
		SELECT id, invoice_number, customer_id, customer_name, shipment_id, shipment_number, booking_id, booking_number, quotation_id, quote_number, route, invoice_date, due_date, currency, subtotal, tax_amount, total_amount, paid_amount, balance_due, status, created_at, updated_at
		FROM customer_invoices
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invoice %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeInvoice,
		ID:              inv.ID,
		ReferenceNumber: inv.InvoiceNumber,
		Title:           fmt.Sprintf("Invoice %s (%s)", inv.InvoiceNumber, inv.CustomerName),
		Status:          inv.Status,
		CreatedAt:       &inv.CreatedAt,
		UpdatedAt:       &inv.UpdatedAt,
		KeyDates: map[string]string{
			"invoice_date": inv.InvoiceDate.Format("2006-01-02"),
			"due_date":     inv.DueDate.Format("2006-01-02"),
		},
		FinancialValues: map[string]interface{}{
			"subtotal":     inv.Subtotal,
			"tax_amount":   inv.TaxAmount,
			"total_amount": inv.TotalAmount,
			"paid_amount":  inv.PaidAmount,
			"balance_due":  inv.BalanceDue,
			"currency":     inv.Currency,
		},
		KeyAttributes: map[string]interface{}{
			"customer_name":   inv.CustomerName,
			"customer_id":     inv.CustomerID,
			"shipment_number": inv.ShipmentNumber,
			"booking_number":  inv.BookingNumber,
			"quote_number":    inv.QuoteNumber,
			"route":           inv.Route,
		},
		DeepLink: fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeInvoice,
		EntityID:        inv.ID,
		ReferenceNumber: inv.InvoiceNumber,
		Label:           fmt.Sprintf("%s %0.2f to %s", inv.Currency, inv.TotalAmount, inv.CustomerName),
		Status:          inv.Status,
		Path:            fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID),
		KeyFields: map[string]interface{}{
			"balance_due":  inv.BalanceDue,
			"total_amount": inv.TotalAmount,
			"due_date":     inv.DueDate.Format("2006-01-02"),
		},
	})

	// 1. Related Customer
	if inv.CustomerID > 0 {
		bCtx.RelatedRecords[EntityTypeCustomer] = []RecordSummary{
			{
				EntityType:      EntityTypeCustomer,
				ID:              inv.CustomerID,
				ReferenceNumber: fmt.Sprintf("CUST-%d", inv.CustomerID),
				Title:           inv.CustomerName,
				Status:          "ACTIVE",
				DeepLink:        fmt.Sprintf("/dashboard/customers/%d", inv.CustomerID),
			},
		}
	}

	// 2. Related Debit Notes (if any)
	var dns []struct {
		ID          int64   `db:"id"`
		Number      string  `db:"debit_note_number"`
		Status      string  `db:"status"`
		TotalAmount float64 `db:"total_amount"`
		Reason      string  `db:"reason"`
	}
	_ = s.db.SelectContext(ctx, &dns, `
		SELECT id, debit_note_number, status, total_amount, reason
		FROM debit_notes
		WHERE (invoice_number = ? OR customer_id = ?) AND org_id = ?
		ORDER BY id DESC LIMIT 5
	`, inv.InvoiceNumber, inv.CustomerID, bCtx.OrgID)

	dnList := make([]RecordSummary, 0, len(dns))
	for _, dn := range dns {
		dnList = append(dnList, RecordSummary{
			EntityType:      EntityTypeInvoice,
			ID:              dn.ID,
			ReferenceNumber: dn.Number,
			Title:           fmt.Sprintf("Debit Note %s (%s)", dn.Number, dn.Reason),
			Status:          dn.Status,
			FinancialValues: map[string]interface{}{
				"total_amount": dn.TotalAmount,
			},
			DeepLink: fmt.Sprintf("/dashboard/debit-notes?id=%d", dn.ID),
		})
	}
	if len(dnList) > 0 {
		bCtx.RelatedRecords[EntityTypeInvoice] = dnList
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// QUOTATION CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildQuotationContext(ctx context.Context, bCtx *BusinessContext) error {
	var q struct {
		ID              int64      `db:"id"`
		QuotationNumber string     `db:"quotation_number"`
		CustomerID      *int64     `db:"customer_id"`
		CustomerName    *string    `db:"customer_name"`
		RFQID           *int64     `db:"rfq_id"`
		RFQNumber       *string    `db:"rfq_number"`
		Status          string     `db:"status"`
		Origin          *string    `db:"origin"`
		Destination     *string    `db:"destination"`
		Currency        string     `db:"currency"`
		TotalAmount     float64    `db:"total_amount"`
		GrossMarginPct  float64    `db:"gross_margin_pct"`
		ValidUntil      *time.Time `db:"valid_until"`
		CreatedAt       *time.Time `db:"created_at"`
		UpdatedAt       *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &q, `
		SELECT id, quotation_number, customer_id, customer_name, rfq_id, rfq_number, status, origin, destination, currency, total_amount, gross_margin_pct, valid_until, created_at, updated_at
		FROM quotations
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("quotation %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	custName := ""
	if q.CustomerName != nil {
		custName = *q.CustomerName
	}
	rfqNum := ""
	if q.RFQNumber != nil {
		rfqNum = *q.RFQNumber
	}
	orig := ""
	if q.Origin != nil {
		orig = *q.Origin
	}
	dest := ""
	if q.Destination != nil {
		dest = *q.Destination
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeQuotation,
		ID:              q.ID,
		ReferenceNumber: q.QuotationNumber,
		Title:           fmt.Sprintf("%s (%s → %s)", q.QuotationNumber, orig, dest),
		Status:          q.Status,
		CreatedAt:       q.CreatedAt,
		UpdatedAt:       q.UpdatedAt,
		FinancialValues: map[string]interface{}{
			"total_amount":     q.TotalAmount,
			"gross_margin_pct": q.GrossMarginPct,
			"currency":         q.Currency,
		},
		KeyAttributes: map[string]interface{}{
			"customer_name": custName,
			"rfq_number":    rfqNum,
			"origin":        orig,
			"destination":   dest,
		},
		DeepLink: fmt.Sprintf("/dashboard/quotes?id=%d", q.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeQuotation,
		EntityID:        q.ID,
		ReferenceNumber: q.QuotationNumber,
		Label:           fmt.Sprintf("%s %0.2f (%s)", q.Currency, q.TotalAmount, q.Status),
		Status:          q.Status,
		Path:            fmt.Sprintf("/dashboard/quotes?id=%d", q.ID),
	})

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LEAD CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildLeadContext(ctx context.Context, bCtx *BusinessContext) error {
	var lead struct {
		ID          int64      `db:"id"`
		CompanyName string     `db:"company_name"`
		ContactName *string    `db:"contact_name"`
		Email       *string    `db:"email"`
		Status      string     `db:"status"`
		CreatedAt   *time.Time `db:"created_at"`
		UpdatedAt   *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &lead, `
		SELECT id, company_name, contact_name, email, status, created_at, updated_at
		FROM leads
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("lead %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	contact := ""
	if lead.ContactName != nil {
		contact = *lead.ContactName
	}
	email := ""
	if lead.Email != nil {
		email = *lead.Email
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeLead,
		ID:              lead.ID,
		ReferenceNumber: fmt.Sprintf("LEAD-%d", lead.ID),
		Title:           lead.CompanyName,
		Status:          lead.Status,
		CreatedAt:       lead.CreatedAt,
		UpdatedAt:       lead.UpdatedAt,
		KeyAttributes: map[string]interface{}{
			"contact_name": contact,
			"email":        email,
		},
		DeepLink: fmt.Sprintf("/dashboard/leads?id=%d", lead.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeLead,
		EntityID:        lead.ID,
		ReferenceNumber: fmt.Sprintf("LEAD-%d", lead.ID),
		Label:           lead.CompanyName,
		Status:          lead.Status,
		Path:            fmt.Sprintf("/dashboard/leads?id=%d", lead.ID),
	})

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CONTRACT CONTEXT
// ─────────────────────────────────────────────────────────────────────────────

func (s *defaultService) buildContractContext(ctx context.Context, bCtx *BusinessContext) error {
	var c struct {
		ID         int64      `db:"id"`
		Title      string     `db:"title"`
		CustomerID *int64     `db:"customer_id"`
		Status     string     `db:"status"`
		StartDate  *time.Time `db:"start_date"`
		EndDate    *time.Time `db:"end_date"`
		Value      float64    `db:"value"`
		CreatedAt  *time.Time `db:"created_at"`
		UpdatedAt  *time.Time `db:"updated_at"`
	}

	err := s.db.GetContext(ctx, &c, `
		SELECT id, title, customer_id, status, start_date, end_date, value, created_at, updated_at
		FROM contracts
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, bCtx.PrimaryID, bCtx.OrgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("contract %d not found in organization", bCtx.PrimaryID)
		}
		return err
	}

	bCtx.PrimaryRecord = &RecordSummary{
		EntityType:      EntityTypeContract,
		ID:              c.ID,
		ReferenceNumber: fmt.Sprintf("CTR-%d", c.ID),
		Title:           c.Title,
		Status:          c.Status,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		FinancialValues: map[string]interface{}{
			"value": c.Value,
		},
		DeepLink: fmt.Sprintf("/dashboard/contracts?id=%d", c.ID),
	}

	bCtx.SourceReferences = append(bCtx.SourceReferences, SourceReference{
		EntityType:      EntityTypeContract,
		EntityID:        c.ID,
		ReferenceNumber: fmt.Sprintf("CTR-%d", c.ID),
		Label:           c.Title,
		Status:          c.Status,
		Path:            fmt.Sprintf("/dashboard/contracts?id=%d", c.ID),
	})

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GROUNDED INTELLIGENCE INSIGHT GENERATOR
// ─────────────────────────────────────────────────────────────────────────────

// GenerateInsight synthesizes deterministic, grounded business intelligence from real records.
func (s *defaultService) GenerateInsight(ctx context.Context, req InsightRequest) (*IntelligenceInsight, error) {
	// Retrieve full verified context
	bCtx, err := s.GetBusinessContext(ctx, ContextRequest{
		OrgID:         req.OrgID,
		UserID:        req.UserID,
		UserRole:      req.UserRole,
		PrimaryType:   req.EntityType,
		PrimaryID:     req.EntityID,
		CorrelationID: req.CorrelationID,
	})
	if err != nil {
		return nil, err
	}

	insight := &IntelligenceInsight{
		IsInformationalOnly:       true,
		SupportingRecords:         bCtx.SourceReferences,
		SupportingFieldReferences: make([]FieldReference, 0),
		KeyHighlights:             make([]string, 0),
		Warnings:                  make([]string, 0),
		RecommendedFollowUp:       make([]string, 0),
		DataFreshness:             time.Now().UTC().Format(time.RFC3339),
		CorrelationID:             req.CorrelationID,
	}

	switch req.EntityType {
	case EntityTypeRFQ:
		s.synthesizeRFQInsight(bCtx, insight)
	case EntityTypeShipment:
		s.synthesizeShipmentInsight(bCtx, insight)
	case EntityTypeInvoice:
		s.synthesizeInvoiceInsight(bCtx, insight)
	case EntityTypeCustomer:
		s.synthesizeCustomerInsight(bCtx, insight)
	default:
		insight.Title = fmt.Sprintf("%s Context Intelligence", req.EntityType)
		insight.Summary = fmt.Sprintf("Verified %s record #%d under organization #%d.", req.EntityType, req.EntityID, req.OrgID)
		insight.ConfidenceLevel = "HIGH"
		insight.DataCompletenessPercentage = 80
	}

	return insight, nil
}

func (s *defaultService) synthesizeRFQInsight(bCtx *BusinessContext, insight *IntelligenceInsight) {
	rec := bCtx.PrimaryRecord
	orig := fmt.Sprintf("%v", rec.KeyAttributes["origin"])
	dest := fmt.Sprintf("%v", rec.KeyAttributes["destination"])
	incoterms := fmt.Sprintf("%v", rec.KeyAttributes["incoterms"])
	status := rec.Status
	quotes := bCtx.RelatedRecords[EntityTypeQuotation]

	insight.Title = fmt.Sprintf("RFQ Intelligence: %s", rec.ReferenceNumber)
	insight.SupportingFieldReferences = append(insight.SupportingFieldReferences,
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "status", FieldValue: status},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "origin", FieldValue: orig},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "destination", FieldValue: dest},
	)

	// Incomplete data checks
	completeness := 100
	if orig == "" || dest == "" {
		completeness -= 30
		insight.Warnings = append(insight.Warnings, "Trade lane incomplete: Origin or Destination port is not specified.")
	}
	if incoterms == "" {
		completeness -= 15
		insight.Warnings = append(insight.Warnings, "Commercial terms: Incoterms missing from request specifications.")
	}
	if len(quotes) == 0 && (status == "ACTIVE" || status == "IN_REVIEW") {
		completeness -= 20
		insight.Warnings = append(insight.Warnings, "No commercial quotations generated yet for this active RFQ.")
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Run carrier spot rate search or generate quotation draft from tariffs.")
	} else if len(quotes) > 0 {
		insight.KeyHighlights = append(insight.KeyHighlights, fmt.Sprintf("%d commercial quotation(s) linked to this RFQ.", len(quotes)))
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Review quotation margins before issuing to customer.")
	}

	if status == "CONFIRMED" || status == "WON" {
		insight.KeyHighlights = append(insight.KeyHighlights, "RFQ is confirmed and ready for carrier booking handoff.")
	}

	insight.Summary = fmt.Sprintf("RFQ %s is currently in %s status covering lane %s to %s. Customer has %d linked quote(s).", rec.ReferenceNumber, status, orig, dest, len(quotes))
	insight.DataCompletenessPercentage = completeness
	if completeness >= 80 {
		insight.ConfidenceLevel = "HIGH"
	} else if completeness >= 50 {
		insight.ConfidenceLevel = "MEDIUM"
	} else {
		insight.ConfidenceLevel = "LOW"
	}
}

func (s *defaultService) synthesizeShipmentInsight(bCtx *BusinessContext, insight *IntelligenceInsight) {
	rec := bCtx.PrimaryRecord
	status := rec.Status
	carrier := fmt.Sprintf("%v", rec.KeyAttributes["carrier_scac"])
	exceptions := bCtx.OperationalExceptions
	milestones := bCtx.Milestones

	insight.Title = fmt.Sprintf("Shipment Operational Intelligence: %s", rec.ReferenceNumber)
	insight.SupportingFieldReferences = append(insight.SupportingFieldReferences,
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "status", FieldValue: status},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "carrier_scac", FieldValue: carrier},
	)

	activeExceptions := 0
	for _, e := range exceptions {
		if strings.ToUpper(e.Status) == "OPEN" || strings.ToUpper(e.Status) == "ACTIVE" {
			activeExceptions++
			insight.Warnings = append(insight.Warnings, fmt.Sprintf("Active [%s] Exception: %s - %s", e.Severity, e.Title, e.Description))
		}
	}

	completedMilestones := 0
	for _, m := range milestones {
		if m.Status == "COMPLETED" {
			completedMilestones++
		}
	}

	insight.KeyHighlights = append(insight.KeyHighlights, fmt.Sprintf("Progress: %d of %d milestones completed.", completedMilestones, len(milestones)))
	if activeExceptions > 0 {
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Resolve active operational exceptions with terminal operator or carrier.")
	}
	if len(milestones) == 0 {
		insight.Warnings = append(insight.Warnings, "No transit milestones recorded yet. Tracking poller awaiting carrier carrier EDI updates.")
	}

	insight.Summary = fmt.Sprintf("Shipment %s is %s with carrier %s. %d milestones logged with %d active exception(s).", rec.ReferenceNumber, status, carrier, len(milestones), activeExceptions)
	if activeExceptions > 0 {
		insight.ConfidenceLevel = "MEDIUM"
		insight.DataCompletenessPercentage = 85
	} else {
		insight.ConfidenceLevel = "HIGH"
		insight.DataCompletenessPercentage = 95
	}
}

func (s *defaultService) synthesizeInvoiceInsight(bCtx *BusinessContext, insight *IntelligenceInsight) {
	rec := bCtx.PrimaryRecord
	status := rec.Status
	total := rec.FinancialValues["total_amount"]
	balance := rec.FinancialValues["balance_due"]
	currency := rec.FinancialValues["currency"]
	dueDate := rec.KeyDates["due_date"]
	custName := fmt.Sprintf("%v", rec.KeyAttributes["customer_name"])

	insight.Title = fmt.Sprintf("Invoice Financial Intelligence: %s", rec.ReferenceNumber)
	insight.SupportingFieldReferences = append(insight.SupportingFieldReferences,
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "status", FieldValue: status},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "total_amount", FieldValue: fmt.Sprintf("%v %v", currency, total)},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "balance_due", FieldValue: fmt.Sprintf("%v %v", currency, balance)},
		FieldReference{SourceRecord: rec.ReferenceNumber, FieldName: "due_date", FieldValue: dueDate},
	)

	balFloat, _ := balance.(float64)
	if strings.ToUpper(status) == "OVERDUE" || (balFloat > 0 && dueDate != "" && dueDate < time.Now().Format("2006-01-02")) {
		insight.Warnings = append(insight.Warnings, fmt.Sprintf("Payment Overdue: %s %0.2f outstanding beyond due date (%s).", currency, balFloat, dueDate))
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Send payment reminder statement to customer billing contact.")
	} else if balFloat > 0 {
		insight.KeyHighlights = append(insight.KeyHighlights, fmt.Sprintf("Current outstanding balance: %s %0.2f.", currency, balFloat))
	} else {
		insight.KeyHighlights = append(insight.KeyHighlights, "Invoice fully settled. No remaining balance.")
	}

	insight.Summary = fmt.Sprintf("Invoice %s for %s (%s) has total value %s %0.2f with balance due of %s %0.2f.", rec.ReferenceNumber, custName, status, currency, total, currency, balFloat)
	insight.ConfidenceLevel = "HIGH"
	insight.DataCompletenessPercentage = 100
}

func (s *defaultService) synthesizeCustomerInsight(bCtx *BusinessContext, insight *IntelligenceInsight) {
	rec := bCtx.PrimaryRecord
	rfqs := bCtx.RelatedRecords[EntityTypeRFQ]
	invoices := bCtx.RelatedRecords[EntityTypeInvoice]
	quotes := bCtx.RelatedRecords[EntityTypeQuotation]
	creditStatus := fmt.Sprintf("%v", rec.FinancialValues["credit_status"])
	creditLimit := rec.FinancialValues["credit_limit"]
	healthScore := rec.KeyAttributes["health_score"]

	insight.Title = fmt.Sprintf("Customer 360° Intelligence: %s", rec.Title)
	insight.SupportingFieldReferences = append(insight.SupportingFieldReferences,
		FieldReference{SourceRecord: rec.Title, FieldName: "credit_status", FieldValue: creditStatus},
		FieldReference{SourceRecord: rec.Title, FieldName: "health_score", FieldValue: fmt.Sprintf("%v", healthScore)},
	)

	insight.KeyHighlights = append(insight.KeyHighlights,
		fmt.Sprintf("Portfolio activity: %d recent RFQ(s), %d commercial quote(s), %d invoice(s).", len(rfqs), len(quotes), len(invoices)),
		fmt.Sprintf("Credit profile: %s standing with limit %v.", creditStatus, creditLimit),
	)

	overdueCount := 0
	for _, inv := range invoices {
		if strings.ToUpper(inv.Status) == "OVERDUE" {
			overdueCount++
		}
	}
	if overdueCount > 0 {
		insight.Warnings = append(insight.Warnings, fmt.Sprintf("Customer has %d overdue invoice(s). Exercise caution before extending credit for new bookings.", overdueCount))
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Finance review required before confirming new rate quotes.")
	} else {
		insight.RecommendedFollowUp = append(insight.RecommendedFollowUp, "Customer account is in good operational and financial standing.")
	}

	insight.Summary = fmt.Sprintf("Customer %s (Status: %s) holds health score %v. Linked to %d RFQ(s) and %d invoice(s).", rec.Title, rec.Status, healthScore, len(rfqs), len(invoices))
	insight.ConfidenceLevel = "HIGH"
	insight.DataCompletenessPercentage = 95
}

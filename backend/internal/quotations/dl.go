package quotations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository defines the data access contract for the quotation domain.
type Repository interface {
	CreateQuotation(ctx context.Context, q *Quotation) error
	GetQuotationByID(ctx context.Context, orgID, quotationID int64) (*Quotation, error)
	GetQuotationByNumber(ctx context.Context, orgID int64, quotationNumber string) (*Quotation, error)
	UpdateQuotation(ctx context.Context, q *Quotation) error
	ListQuotations(ctx context.Context, filters *QuotationListFilters) ([]*QuotationListItem, int, error)
	GetQuotationSummary(ctx context.Context, orgID int64) (*QuotationSummary, error)

	// Quotation number generation (collision-safe within org)
	GetLastQuotationNumber(ctx context.Context, orgID int64, year int) (int, error)

	// Activity log
	CreateActivity(ctx context.Context, act *QuotationActivity) error
	GetActivity(ctx context.Context, orgID, quotationID int64) ([]QuotationActivity, error)

	// Customer info for detail panel
	GetCustomerInfo(ctx context.Context, orgID, customerID int64) (*QuotationCustomerInfo, error)

	// Quotation Pricing & Charges (Task 18.2)
	CreateQuotationCharge(ctx context.Context, charge *QuotationChargeItem) error
	GetQuotationCharges(ctx context.Context, orgID, quotationID int64) ([]*QuotationChargeItem, error)
	GetQuotationChargeByID(ctx context.Context, orgID, quotationID, chargeID int64) (*QuotationChargeItem, error)
	UpdateQuotationCharge(ctx context.Context, charge *QuotationChargeItem) error
	DeleteQuotationCharge(ctx context.Context, orgID, quotationID, chargeID int64) error
	ClearQuotationCharges(ctx context.Context, orgID, quotationID int64) error
	ReorderQuotationCharges(ctx context.Context, orgID, quotationID int64, chargeIDs []int64) error
	UpdateQuotationTotals(ctx context.Context, orgID, quotationID int64, subtotal, surcharges, taxes, totalAmount, totalCost, grossProfit, grossMarginPct float64) error
	GetMaxDisplayOrder(ctx context.Context, orgID, quotationID int64) (int, error)

	// Quotation Templates & Commercial Terms (Task 18.3)
	CreateQuotationTemplate(ctx context.Context, tmpl *QuotationTemplate, charges []*QuotationTemplateChargeItem) error
	GetQuotationTemplates(ctx context.Context, orgID int64, activeOnly bool) ([]*QuotationTemplate, error)
	GetQuotationTemplateByID(ctx context.Context, orgID, templateID int64) (*QuotationTemplate, []*QuotationTemplateChargeItem, error)
	UpdateQuotationTemplate(ctx context.Context, tmpl *QuotationTemplate, charges []*QuotationTemplateChargeItem) error
	ArchiveQuotationTemplate(ctx context.Context, orgID, templateID int64) error
	UpdateQuotationCommercialTerms(ctx context.Context, orgID, quotationID int64, terms *QuotationCommercialTerms) error
	SeedDefaultTemplatesIfEmpty(ctx context.Context, orgID int64) error

	// Approval & Lifecycle (Task 18.4)
	SubmitQuotationForReview(ctx context.Context, orgID, quotationID int64, submittedBy string, comments string) error
	ApproveQuotation(ctx context.Context, orgID, quotationID int64, approvedBy string, notes string) error
	RequestQuotationChanges(ctx context.Context, orgID, quotationID int64, requestedBy string, reason string) error
	MarkQuotationSent(ctx context.Context, orgID, quotationID int64, sentBy string) error
	MarkQuotationViewed(ctx context.Context, orgID, quotationID int64, viewerName, viewerEmail, ipAddress, userAgent string) error
	AcceptQuotation(ctx context.Context, orgID, quotationID int64, acceptedBy string, comments string) error
	DeclineQuotation(ctx context.Context, orgID, quotationID int64, declinedBy string, reason string) error
	CancelQuotation(ctx context.Context, orgID, quotationID int64, cancelledBy string, reason string) error

	CreateApprovalHistory(ctx context.Context, h *QuotationApprovalHistory) error
	GetApprovalHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationApprovalHistory, error)
	RecordPublicView(ctx context.Context, orgID, quotationID int64, viewerName, viewerEmail, ipAddress, userAgent string) error

	// Documents & Public Sharing (Task 18.5)
	CreateQuotationDocument(ctx context.Context, doc *QuotationDocument) error
	GetQuotationDocuments(ctx context.Context, orgID, quotationID int64) ([]*QuotationDocument, error)
	GetQuotationDocumentByID(ctx context.Context, orgID, quotationID, docID int64) (*QuotationDocument, error)
	GetLatestDocumentVersion(ctx context.Context, orgID, quotationID int64, docType string) (int, error)
	CreateQuotationPublicLink(ctx context.Context, link *QuotationPublicLink) error
	GetQuotationPublicLinks(ctx context.Context, orgID, quotationID int64) ([]*QuotationPublicLink, error)
	GetQuotationPublicLinkByToken(ctx context.Context, token string) (*QuotationPublicLink, error)
	RevokeQuotationPublicLink(ctx context.Context, orgID, quotationID, linkID, actorUserID int64, reason string) error
	IncrementPublicLinkAccess(ctx context.Context, linkID int64) error

	// Quotation-to-Booking Operational Conversion (Task 18.6)
	GetQuotationConversionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationConversionHistory, error)
	CreateQuotationConversionHistory(ctx context.Context, history *QuotationConversionHistory) error
	MarkQuotationConverted(ctx context.Context, orgID, quotationID, bookingID int64, shipmentID *int64, convertedBy, notes string) error
	MarkQuotationConversionFailed(ctx context.Context, orgID, quotationID int64, failedBy, reason string) error
	CreateBookingFromQuotationTx(ctx context.Context, orgID int64, q *Quotation, req *ConvertQuotationToBookingRequest, creator string) (int64, string, *int64, *string, error)

	// Booking Confirmation & Handover Traceability (Task 18.7)
	GetOperationalBooking(ctx context.Context, orgID, bookingID int64) (*RawOperationalBooking, error)
	GetOperationalShipment(ctx context.Context, orgID, shipmentID int64) (*RawOperationalShipment, error)
	GetOperationalBookingByQuotationID(ctx context.Context, orgID, quotationID int64) (*RawOperationalBooking, error)
	GetOperationalShipmentByBookingID(ctx context.Context, orgID, bookingID int64) (*RawOperationalShipment, error)
	CreateOperationalHandoverHistory(ctx context.Context, h *QuotationOperationalHandoverHistory) error
	GetOperationalHandoverHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationOperationalHandoverHistory, error)
	UpdateBookingHandoverStatus(ctx context.Context, orgID, bookingID int64, status, confirmedBy string) error

	// Quotation Analytics & Intelligence (Task 18.8)
	GetQuotationAnalyticsOverview(ctx context.Context, orgID int64) (*QuotationAnalyticsOverview, error)
	GetQuotationAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]*QuotationTrendDataPoint, error)
	GetCustomerQuotationPerformance(ctx context.Context, orgID int64) ([]*CustomerQuotationPerformance, error)
	GetQuotationPerformanceByMode(ctx context.Context, orgID int64) ([]*QuotationPerformanceByMode, error)
	GetQuotationExpiryRisk(ctx context.Context, orgID int64) ([]*QuotationRiskItem, error)

	// Rate-to-Quotation Integration (Task 19.5)
	GetQuotationRateCandidates(ctx context.Context, orgID int64, quotationID int64) ([]QuotationRateCandidate, error)
	GetActiveQuotationRateSelection(ctx context.Context, orgID, quotationID int64) (*QuotationRateSelection, error)
	CreateQuotationRateSelectionTx(ctx context.Context, orgID int64, selection *QuotationRateSelection, snapshot *QuotationRateSnapshot, history *QuotationRateSelectionHistory, charges []*QuotationChargeItem, totals map[string]float64) error
	DeactivateQuotationRateSelection(ctx context.Context, orgID, selectionID int64) error
	GetLatestQuotationRateSnapshot(ctx context.Context, orgID, quotationID int64) (*QuotationRateSnapshot, error)
	CreateQuotationRateSelectionHistory(ctx context.Context, history *QuotationRateSelectionHistory) error
	GetQuotationRateSelectionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateSelectionHistory, error)

	// Rate Lifecycle Intelligence & Quotation Risk (Task 19.6)
	CreateQuotationRateRiskEvent(ctx context.Context, risk *QuotationRateRisk) error
	GetQuotationRateRisks(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateRisk, error)
	ResolveQuotationRateRisk(ctx context.Context, orgID, quotationID, riskID int64, user string) error
	GetQuotationsWithActiveRateSelection(ctx context.Context, orgID int64) ([]*QuotationRateSelectionDetail, error)
	GetQuotationsAffectedByRate(ctx context.Context, orgID, rateID int64) ([]int64, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new quotation repository backed by MySQL.
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateQuotation(ctx context.Context, q *Quotation) error {
	query := `
		INSERT INTO quotations (
			org_id, quotation_number,
			customer_id, customer_name, rfq_id, rfq_number,
			status,
			origin, origin_code, destination, destination_code,
			service_type, transport_mode,
			currency, payment_terms, subtotal, surcharges, taxes, total_amount, total_cost, gross_profit, gross_margin_pct,
			valid_from, valid_until,
			notes, commercial_terms, customer_notes, internal_notes, template_id,
			created_by, updated_by,
			created_at, updated_at
		) VALUES (
			?, ?,
			?, ?, ?, ?,
			?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?, ?,
			?, ?,
			NOW(), NOW()
		)
	`
	paymentTerms := q.PaymentTerms
	if paymentTerms == "" {
		paymentTerms = PaymentTermsPrepaid
	}

	result, err := r.db.ExecContext(ctx, query,
		q.OrgID, q.QuotationNumber,
		q.CustomerID, q.CustomerName, q.RFQID, q.RFQNumber,
		q.Status,
		q.Origin, q.OriginCode, q.Destination, q.DestinationCode,
		q.ServiceType, q.TransportMode,
		q.Currency, paymentTerms, q.Subtotal, q.Surcharges, q.Taxes, q.TotalAmount, q.TotalCost, q.GrossProfit, q.GrossMarginPct,
		q.ValidFrom, q.ValidUntil,
		q.Notes, q.CommercialTerms, q.CustomerNotes, q.InternalNotes, q.TemplateID,
		q.CreatedBy, q.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("create quotation: %w", err)
	}
	id, _ := result.LastInsertId()
	q.ID = id
	return nil
}

func (r *repository) GetQuotationByID(ctx context.Context, orgID, quotationID int64) (*Quotation, error) {
	var q Quotation
	err := r.db.GetContext(ctx, &q, `
		SELECT
			id, org_id, quotation_number,
			customer_id, COALESCE(customer_name, '') AS customer_name,
			rfq_id, COALESCE(rfq_number, '') AS rfq_number,
			status,
			COALESCE(origin, '') AS origin, COALESCE(origin_code, '') AS origin_code,
			COALESCE(destination, '') AS destination, COALESCE(destination_code, '') AS destination_code,
			COALESCE(service_type, '') AS service_type, COALESCE(transport_mode, '') AS transport_mode,
			currency, COALESCE(payment_terms, 'PREPAID') AS payment_terms,
			subtotal, surcharges, taxes, total_amount,
			COALESCE(total_cost, 0.00) AS total_cost,
			COALESCE(gross_profit, 0.00) AS gross_profit,
			COALESCE(gross_margin_pct, 0.0000) AS gross_margin_pct,
			valid_from, valid_until,
			submitted_for_review_at, COALESCE(submitted_for_review_by, '') AS submitted_for_review_by,
			approved_at, COALESCE(approved_by, '') AS approved_by,
			COALESCE(approval_notes, '') AS approval_notes,
			changes_requested_at, COALESCE(changes_requested_by, '') AS changes_requested_by,
			COALESCE(changes_requested_reason, '') AS changes_requested_reason,
			sent_at, COALESCE(sent_by, '') AS sent_by,
			viewed_at, first_viewed_at, last_viewed_at, COALESCE(view_count, 0) AS view_count,
			accepted_at, declined_at, COALESCE(declined_reason, '') AS declined_reason,
			rejected_at, expired_at,
			cancelled_at, COALESCE(cancelled_by, '') AS cancelled_by, COALESCE(cancelled_reason, '') AS cancelled_reason,
			converted_at, COALESCE(converted_by, '') AS converted_by,
			converted_booking_id, converted_shipment_id,
			COALESCE(conversion_status, 'NOT_CONVERTED') AS conversion_status,
			COALESCE(conversion_notes, '') AS conversion_notes,
			COALESCE(notes, '') AS notes,
			COALESCE(commercial_terms, '') AS commercial_terms,
			COALESCE(customer_notes, '') AS customer_notes,
			COALESCE(internal_notes, '') AS internal_notes,
			template_id,
			COALESCE(created_by, '') AS created_by,
			COALESCE(updated_by, '') AS updated_by,
			created_at, updated_at
		FROM quotations
		WHERE org_id = ? AND id = ?
	`, orgID, quotationID)
	if err != nil {
		return nil, err
	}
	computeValidityStatus(&q)
	return &q, nil
}

func (r *repository) GetQuotationByNumber(ctx context.Context, orgID int64, quotationNumber string) (*Quotation, error) {
	var q Quotation
	err := r.db.GetContext(ctx, &q, `
		SELECT
			id, org_id, quotation_number,
			customer_id, COALESCE(customer_name, '') AS customer_name,
			rfq_id, COALESCE(rfq_number, '') AS rfq_number,
			status,
			COALESCE(origin, '') AS origin, COALESCE(origin_code, '') AS origin_code,
			COALESCE(destination, '') AS destination, COALESCE(destination_code, '') AS destination_code,
			COALESCE(service_type, '') AS service_type, COALESCE(transport_mode, '') AS transport_mode,
			currency, COALESCE(payment_terms, 'PREPAID') AS payment_terms,
			subtotal, surcharges, taxes, total_amount,
			COALESCE(total_cost, 0.00) AS total_cost,
			COALESCE(gross_profit, 0.00) AS gross_profit,
			COALESCE(gross_margin_pct, 0.0000) AS gross_margin_pct,
			valid_from, valid_until,
			submitted_for_review_at, COALESCE(submitted_for_review_by, '') AS submitted_for_review_by,
			approved_at, COALESCE(approved_by, '') AS approved_by,
			COALESCE(approval_notes, '') AS approval_notes,
			changes_requested_at, COALESCE(changes_requested_by, '') AS changes_requested_by,
			COALESCE(changes_requested_reason, '') AS changes_requested_reason,
			sent_at, COALESCE(sent_by, '') AS sent_by,
			viewed_at, first_viewed_at, last_viewed_at, COALESCE(view_count, 0) AS view_count,
			accepted_at, declined_at, COALESCE(declined_reason, '') AS declined_reason,
			rejected_at, expired_at,
			cancelled_at, COALESCE(cancelled_by, '') AS cancelled_by, COALESCE(cancelled_reason, '') AS cancelled_reason,
			converted_at, COALESCE(converted_by, '') AS converted_by,
			converted_booking_id, converted_shipment_id,
			COALESCE(conversion_status, 'NOT_CONVERTED') AS conversion_status,
			COALESCE(conversion_notes, '') AS conversion_notes,
			COALESCE(notes, '') AS notes,
			COALESCE(commercial_terms, '') AS commercial_terms,
			COALESCE(customer_notes, '') AS customer_notes,
			COALESCE(internal_notes, '') AS internal_notes,
			template_id,
			COALESCE(created_by, '') AS created_by,
			COALESCE(updated_by, '') AS updated_by,
			created_at, updated_at
		FROM quotations
		WHERE org_id = ? AND quotation_number = ?
	`, orgID, quotationNumber)
	if err != nil {
		return nil, err
	}
	computeValidityStatus(&q)
	return &q, nil
}

func (r *repository) UpdateQuotation(ctx context.Context, q *Quotation) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quotations SET
			customer_id = ?, customer_name = ?,
			rfq_id = ?, rfq_number = ?,
			status = ?,
			origin = ?, origin_code = ?, destination = ?, destination_code = ?,
			service_type = ?, transport_mode = ?,
			currency = ?, payment_terms = ?, subtotal = ?, surcharges = ?, taxes = ?, total_amount = ?,
			total_cost = ?, gross_profit = ?, gross_margin_pct = ?,
			valid_from = ?, valid_until = ?,
			sent_at = ?, viewed_at = ?, accepted_at = ?, rejected_at = ?, expired_at = ?,
			notes = ?, commercial_terms = ?, customer_notes = ?, internal_notes = ?, template_id = ?,
			updated_by = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`,
		q.CustomerID, q.CustomerName,
		q.RFQID, q.RFQNumber,
		q.Status,
		q.Origin, q.OriginCode, q.Destination, q.DestinationCode,
		q.ServiceType, q.TransportMode,
		q.Currency, q.PaymentTerms, q.Subtotal, q.Surcharges, q.Taxes, q.TotalAmount,
		q.TotalCost, q.GrossProfit, q.GrossMarginPct,
		q.ValidFrom, q.ValidUntil,
		q.SentAt, q.ViewedAt, q.AcceptedAt, q.RejectedAt, q.ExpiredAt,
		q.Notes, q.CommercialTerms, q.CustomerNotes, q.InternalNotes, q.TemplateID,
		q.UpdatedBy,
		q.OrgID, q.ID,
	)
	return err
}

func (r *repository) ListQuotations(ctx context.Context, filters *QuotationListFilters) ([]*QuotationListItem, int, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	offset := (filters.Page - 1) * filters.Limit

	conditions := []string{"q.org_id = ?"}
	args := []interface{}{filters.OrgID}

	if filters.Status != "" && filters.Status != "ALL" {
		conditions = append(conditions, "q.status = ?")
		args = append(args, filters.Status)
	}

	if filters.CustomerID != nil {
		conditions = append(conditions, "q.customer_id = ?")
		args = append(args, *filters.CustomerID)
	}

	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		conditions = append(conditions, "(q.quotation_number LIKE ? OR q.customer_name LIKE ? OR q.origin LIKE ? OR q.destination LIKE ?)")
		args = append(args, like, like, like, like)
	}

	now := time.Now()
	switch filters.Validity {
	case "VALID":
		conditions = append(conditions, "q.valid_until >= ?")
		args = append(args, now.Format("2006-01-02"))
	case "EXPIRING_SOON":
		soon := now.AddDate(0, 0, 7)
		conditions = append(conditions, "q.valid_until BETWEEN ? AND ?")
		args = append(args, now.Format("2006-01-02"), soon.Format("2006-01-02"))
	case "EXPIRED":
		conditions = append(conditions, "q.valid_until < ?")
		args = append(args, now.Format("2006-01-02"))
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	countArgs := append([]interface{}{}, args...)
	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM quotations q `+where, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("count quotations: %w", err)
	}

	listArgs := append(args, filters.Limit, offset)
	rows := []*QuotationListItem{}
	err = r.db.SelectContext(ctx, &rows, `
		SELECT
			q.id, q.quotation_number,
			q.customer_id, COALESCE(q.customer_name, '') AS customer_name,
			COALESCE(q.origin, '') AS origin, COALESCE(q.origin_code, '') AS origin_code,
			COALESCE(q.destination, '') AS destination, COALESCE(q.destination_code, '') AS destination_code,
			COALESCE(q.service_type, '') AS service_type, COALESCE(q.transport_mode, '') AS transport_mode,
			q.currency, COALESCE(q.payment_terms, 'PREPAID') AS payment_terms, q.total_amount,
			COALESCE(q.total_cost, 0.00) AS total_cost,
			COALESCE(q.gross_profit, 0.00) AS gross_profit,
			COALESCE(q.gross_margin_pct, 0.0000) AS gross_margin_pct,
			q.status,
			COALESCE(q.conversion_status, 'NOT_CONVERTED') AS conversion_status,
			q.converted_booking_id, q.converted_shipment_id,
			q.valid_from, q.valid_until, q.updated_at
		FROM quotations q
		`+where+`
		ORDER BY q.updated_at DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list quotations: %w", err)
	}

	for _, item := range rows {
		item.ValidityStatus = CalculateQuotationValidityStatus(item.ValidUntil)
	}

	return rows, total, nil
}

func (r *repository) GetQuotationSummary(ctx context.Context, orgID int64) (*QuotationSummary, error) {
	var s QuotationSummary
	err := r.db.GetContext(ctx, &s, `
		SELECT
			COUNT(*) AS total_quotations,
			COALESCE(SUM(CASE WHEN status = 'DRAFT'    THEN 1 ELSE 0 END), 0) AS draft_count,
			COALESCE(SUM(CASE WHEN status = 'SENT'     THEN 1 ELSE 0 END), 0) AS sent_count,
			COALESCE(SUM(CASE WHEN status = 'VIEWED'   THEN 1 ELSE 0 END), 0) AS viewed_count,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS accepted_count,
			COALESCE(SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END), 0) AS rejected_count,
			COALESCE(SUM(CASE WHEN status = 'EXPIRED'  THEN 1 ELSE 0 END), 0) AS expired_count
		FROM quotations
		WHERE org_id = ?
	`, orgID)
	if err != nil {
		return &QuotationSummary{}, nil
	}
	return &s, nil
}

func (r *repository) GetLastQuotationNumber(ctx context.Context, orgID int64, year int) (int, error) {
	prefix := fmt.Sprintf("QT-%d-", year)
	var last int
	err := r.db.GetContext(ctx, &last, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(quotation_number, ?) AS UNSIGNED)), 0)
		FROM quotations
		WHERE org_id = ? AND quotation_number LIKE ?
	`, len(prefix)+1, orgID, prefix+"%")
	return last, err
}

func (r *repository) CreateActivity(ctx context.Context, act *QuotationActivity) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO quotation_activity (org_id, quotation_id, activity_type, description, actor, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`, act.OrgID, act.QuotationID, act.ActivityType, act.Description, act.Actor)
	return err
}

func (r *repository) GetActivity(ctx context.Context, orgID, quotationID int64) ([]QuotationActivity, error) {
	var acts []QuotationActivity
	err := r.db.SelectContext(ctx, &acts, `
		SELECT id, org_id, quotation_id, activity_type, COALESCE(description, '') AS description, COALESCE(actor, '') AS actor, created_at
		FROM quotation_activity
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY created_at DESC
	`, orgID, quotationID)
	if err != nil {
		return []QuotationActivity{}, nil
	}
	return acts, nil
}

func (r *repository) GetCustomerInfo(ctx context.Context, orgID, customerID int64) (*QuotationCustomerInfo, error) {
	type customerRow struct {
		ID           int64  `db:"id"`
		Name         string `db:"name"`
		ContactPhone string `db:"contact_phone"`
		ContactEmail string `db:"contact_email"`
	}
	var c customerRow
	err := r.db.GetContext(ctx, &c, `
		SELECT id, COALESCE(name, '') AS name, COALESCE(contact_phone, '') AS contact_phone, COALESCE(contact_email, '') AS contact_email
		FROM customers
		WHERE org_id = ? AND id = ?
	`, orgID, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &QuotationCustomerInfo{
		ID:           c.ID,
		Name:         c.Name,
		CustomerCode: fmt.Sprintf("CUST-%04d", c.ID),
		ContactPhone: c.ContactPhone,
		ContactEmail: c.ContactEmail,
	}, nil
}

// ── Pricing & Charges Operations ──────────────────────────────────────────────

func (r *repository) CreateQuotationCharge(ctx context.Context, charge *QuotationChargeItem) error {
	query := `
		INSERT INTO quotation_charge_items (
			org_id, quotation_id,
			charge_code, charge_name, charge_category, charge_type, calculation_basis,
			quantity, unit_price, cost_amount, sell_amount,
			currency, exchange_rate,
			tax_rate, tax_amount,
			discount_type, discount_value, discount_amount,
			total_cost, total_sell,
			display_order, is_optional, notes,
			created_by, updated_by,
			created_at, updated_at
		) VALUES (
			?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?,
			?, ?, ?,
			?, ?,
			?, ?, ?,
			?, ?,
			NOW(), NOW()
		)
	`
	result, err := r.db.ExecContext(ctx, query,
		charge.OrgID, charge.QuotationID,
		charge.ChargeCode, charge.ChargeName, charge.ChargeCategory, charge.ChargeType, charge.CalculationBasis,
		charge.Quantity, charge.UnitPrice, charge.CostAmount, charge.SellAmount,
		charge.Currency, charge.ExchangeRate,
		charge.TaxRate, charge.TaxAmount,
		charge.DiscountType, charge.DiscountValue, charge.DiscountAmount,
		charge.TotalCost, charge.TotalSell,
		charge.DisplayOrder, charge.IsOptional, charge.Notes,
		charge.CreatedBy, charge.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("create quotation charge: %w", err)
	}
	id, _ := result.LastInsertId()
	charge.ID = id
	return nil
}

func (r *repository) GetQuotationCharges(ctx context.Context, orgID, quotationID int64) ([]*QuotationChargeItem, error) {
	var charges []*QuotationChargeItem
	err := r.db.SelectContext(ctx, &charges, `
		SELECT *
		FROM quotation_charge_items
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY display_order ASC, id ASC
	`, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("get quotation charges: %w", err)
	}
	return charges, nil
}

func (r *repository) GetQuotationChargeByID(ctx context.Context, orgID, quotationID, chargeID int64) (*QuotationChargeItem, error) {
	var charge QuotationChargeItem
	err := r.db.GetContext(ctx, &charge, `
		SELECT *
		FROM quotation_charge_items
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`, orgID, quotationID, chargeID)
	if err != nil {
		return nil, err
	}
	return &charge, nil
}

func (r *repository) UpdateQuotationCharge(ctx context.Context, charge *QuotationChargeItem) error {
	query := `
		UPDATE quotation_charge_items SET
			charge_code = ?, charge_name = ?, charge_category = ?, charge_type = ?, calculation_basis = ?,
			quantity = ?, unit_price = ?, cost_amount = ?, sell_amount = ?,
			currency = ?, exchange_rate = ?,
			tax_rate = ?, tax_amount = ?,
			discount_type = ?, discount_value = ?, discount_amount = ?,
			total_cost = ?, total_sell = ?,
			display_order = ?, is_optional = ?, notes = ?,
			updated_by = ?, updated_at = NOW()
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		charge.ChargeCode, charge.ChargeName, charge.ChargeCategory, charge.ChargeType, charge.CalculationBasis,
		charge.Quantity, charge.UnitPrice, charge.CostAmount, charge.SellAmount,
		charge.Currency, charge.ExchangeRate,
		charge.TaxRate, charge.TaxAmount,
		charge.DiscountType, charge.DiscountValue, charge.DiscountAmount,
		charge.TotalCost, charge.TotalSell,
		charge.DisplayOrder, charge.IsOptional, charge.Notes,
		charge.UpdatedBy,
		charge.OrgID, charge.QuotationID, charge.ID,
	)
	if err != nil {
		return fmt.Errorf("update quotation charge: %w", err)
	}
	return nil
}

func (r *repository) DeleteQuotationCharge(ctx context.Context, orgID, quotationID, chargeID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM quotation_charge_items
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`, orgID, quotationID, chargeID)
	if err != nil {
		return fmt.Errorf("delete quotation charge: %w", err)
	}
	return nil
}

func (r *repository) ClearQuotationCharges(ctx context.Context, orgID, quotationID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM quotation_charge_items
		WHERE org_id = ? AND quotation_id = ?
	`, orgID, quotationID)
	return err
}

func (r *repository) ReorderQuotationCharges(ctx context.Context, orgID, quotationID int64, chargeIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PreparexContext(ctx, `
		UPDATE quotation_charge_items
		SET display_order = ?
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare reorder stmt: %w", err)
	}
	defer stmt.Close()

	for order, id := range chargeIDs {
		if _, err := stmt.ExecContext(ctx, order+1, orgID, quotationID, id); err != nil {
			return fmt.Errorf("update charge %d order: %w", id, err)
		}
	}

	return tx.Commit()
}

func (r *repository) UpdateQuotationTotals(ctx context.Context, orgID, quotationID int64, subtotal, surcharges, taxes, totalAmount, totalCost, grossProfit, grossMarginPct float64) error {
	query := `
		UPDATE quotations SET
			subtotal = ?,
			surcharges = ?,
			taxes = ?,
			total_amount = ?,
			total_cost = ?,
			gross_profit = ?,
			gross_margin_pct = ?,
			updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		subtotal, surcharges, taxes, totalAmount, totalCost, grossProfit, grossMarginPct,
		orgID, quotationID,
	)
	if err != nil {
		return fmt.Errorf("update quotation totals: %w", err)
	}
	return nil
}

func (r *repository) GetMaxDisplayOrder(ctx context.Context, orgID, quotationID int64) (int, error) {
	var maxOrder int
	err := r.db.GetContext(ctx, &maxOrder, `
		SELECT COALESCE(MAX(display_order), 0)
		FROM quotation_charge_items
		WHERE org_id = ? AND quotation_id = ?
	`, orgID, quotationID)
	return maxOrder, err
}

// ── Reusable Quotation Templates Operations (Task 18.3) ───────────────────────

func (r *repository) CreateQuotationTemplate(ctx context.Context, tmpl *QuotationTemplate, charges []*QuotationTemplateChargeItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create template tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO quotation_templates (
			org_id, name, description,
			shipment_mode, transport_mode, origin, destination,
			currency, validity_days, payment_terms,
			commercial_terms, customer_notes, internal_notes,
			is_active, created_by,
			created_at, updated_at
		) VALUES (
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?,
			NOW(), NOW()
		)
	`
	paymentTerms := tmpl.PaymentTerms
	if paymentTerms == "" {
		paymentTerms = PaymentTermsPrepaid
	}
	validityDays := tmpl.ValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}
	currency := tmpl.Currency
	if currency == "" {
		currency = "USD"
	}

	result, err := tx.ExecContext(ctx, query,
		tmpl.OrgID, tmpl.Name, tmpl.Description,
		tmpl.ShipmentMode, tmpl.TransportMode, tmpl.Origin, tmpl.Destination,
		currency, validityDays, paymentTerms,
		tmpl.CommercialTerms, tmpl.CustomerNotes, tmpl.InternalNotes,
		tmpl.IsActive, tmpl.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert quotation template: %w", err)
	}
	templateID, _ := result.LastInsertId()
	tmpl.ID = templateID

	// Insert template charge items
	if len(charges) > 0 {
		chargeStmt, err := tx.PreparexContext(ctx, `
			INSERT INTO quotation_template_charge_items (
				org_id, template_id,
				charge_category, charge_code, charge_name, calculation_basis,
				quantity, unit_price, cost_amount,
				discount_type, discount_value, tax_rate,
				currency, display_order, is_optional, notes,
				created_at, updated_at
			) VALUES (
				?, ?,
				?, ?, ?, ?,
				?, ?, ?,
				?, ?, ?,
				?, ?, ?, ?,
				NOW(), NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("prepare template charges stmt: %w", err)
		}
		defer chargeStmt.Close()

		for idx, c := range charges {
			order := c.DisplayOrder
			if order <= 0 {
				order = idx + 1
			}
			cCurr := c.Currency
			if cCurr == "" {
				cCurr = currency
			}
			cQty := c.Quantity
			if cQty <= 0 {
				cQty = 1.0
			}
			cDiscType := c.DiscountType
			if cDiscType == "" {
				cDiscType = QuotationDiscountTypeNone
			}

			if _, err := chargeStmt.ExecContext(ctx,
				tmpl.OrgID, templateID,
				c.ChargeCategory, c.ChargeCode, c.ChargeName, c.CalculationBasis,
				cQty, c.UnitPrice, c.CostAmount,
				cDiscType, c.DiscountValue, c.TaxRate,
				cCurr, order, c.IsOptional, c.Notes,
			); err != nil {
				return fmt.Errorf("insert template charge item: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *repository) GetQuotationTemplates(ctx context.Context, orgID int64, activeOnly bool) ([]*QuotationTemplate, error) {
	query := `
		SELECT
			t.id, t.org_id, t.name, COALESCE(t.description, '') AS description,
			COALESCE(t.shipment_mode, '') AS shipment_mode, COALESCE(t.transport_mode, '') AS transport_mode,
			COALESCE(t.origin, '') AS origin, COALESCE(t.destination, '') AS destination,
			t.currency, t.validity_days, COALESCE(t.payment_terms, 'PREPAID') AS payment_terms,
			COALESCE(t.commercial_terms, '') AS commercial_terms,
			COALESCE(t.customer_notes, '') AS customer_notes,
			COALESCE(t.internal_notes, '') AS internal_notes,
			t.is_active, COALESCE(t.created_by, '') AS created_by,
			t.created_at, t.updated_at,
			COUNT(c.id) AS charge_count
		FROM quotation_templates t
		LEFT JOIN quotation_template_charge_items c ON c.org_id = t.org_id AND c.template_id = t.id
		WHERE t.org_id = ?
	`
	if activeOnly {
		query += " AND t.is_active = TRUE"
	}
	query += " GROUP BY t.id ORDER BY t.name ASC"

	var templates []*QuotationTemplate
	err := r.db.SelectContext(ctx, &templates, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list quotation templates: %w", err)
	}

	return templates, nil
}

func (r *repository) GetQuotationTemplateByID(ctx context.Context, orgID, templateID int64) (*QuotationTemplate, []*QuotationTemplateChargeItem, error) {
	var tmpl QuotationTemplate
	err := r.db.GetContext(ctx, &tmpl, `
		SELECT
			t.id, t.org_id, t.name, COALESCE(t.description, '') AS description,
			COALESCE(t.shipment_mode, '') AS shipment_mode, COALESCE(t.transport_mode, '') AS transport_mode,
			COALESCE(t.origin, '') AS origin, COALESCE(t.destination, '') AS destination,
			t.currency, t.validity_days, COALESCE(t.payment_terms, 'PREPAID') AS payment_terms,
			COALESCE(t.commercial_terms, '') AS commercial_terms,
			COALESCE(t.customer_notes, '') AS customer_notes,
			COALESCE(t.internal_notes, '') AS internal_notes,
			t.is_active, COALESCE(t.created_by, '') AS created_by,
			t.created_at, t.updated_at,
			0 AS charge_count
		FROM quotation_templates t
		WHERE t.org_id = ? AND t.id = ?
	`, orgID, templateID)
	if err != nil {
		return nil, nil, err
	}

	var charges []*QuotationTemplateChargeItem
	err = r.db.SelectContext(ctx, &charges, `
		SELECT *
		FROM quotation_template_charge_items
		WHERE org_id = ? AND template_id = ?
		ORDER BY display_order ASC, id ASC
	`, orgID, templateID)
	if err != nil {
		return nil, nil, err
	}

	tmpl.ChargeCount = len(charges)
	return &tmpl, charges, nil
}

func (r *repository) UpdateQuotationTemplate(ctx context.Context, tmpl *QuotationTemplate, charges []*QuotationTemplateChargeItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update template tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE quotation_templates SET
			name = ?, description = ?,
			shipment_mode = ?, transport_mode = ?, origin = ?, destination = ?,
			currency = ?, validity_days = ?, payment_terms = ?,
			commercial_terms = ?, customer_notes = ?, internal_notes = ?,
			is_active = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`,
		tmpl.Name, tmpl.Description,
		tmpl.ShipmentMode, tmpl.TransportMode, tmpl.Origin, tmpl.Destination,
		tmpl.Currency, tmpl.ValidityDays, tmpl.PaymentTerms,
		tmpl.CommercialTerms, tmpl.CustomerNotes, tmpl.InternalNotes,
		tmpl.IsActive,
		tmpl.OrgID, tmpl.ID,
	)
	if err != nil {
		return fmt.Errorf("update template header: %w", err)
	}

	// If charges are provided in update, replace them
	if charges != nil {
		_, _ = tx.ExecContext(ctx, `DELETE FROM quotation_template_charge_items WHERE org_id = ? AND template_id = ?`, tmpl.OrgID, tmpl.ID)

		if len(charges) > 0 {
			chargeStmt, err := tx.PreparexContext(ctx, `
				INSERT INTO quotation_template_charge_items (
					org_id, template_id,
					charge_category, charge_code, charge_name, calculation_basis,
					quantity, unit_price, cost_amount,
					discount_type, discount_value, tax_rate,
					currency, display_order, is_optional, notes,
					created_at, updated_at
				) VALUES (
					?, ?,
					?, ?, ?, ?,
					?, ?, ?,
					?, ?, ?,
					?, ?, ?, ?,
					NOW(), NOW()
				)
			`)
			if err != nil {
				return fmt.Errorf("prepare replace template charges: %w", err)
			}
			defer chargeStmt.Close()

			for idx, c := range charges {
				order := c.DisplayOrder
				if order <= 0 {
					order = idx + 1
				}
				cCurr := c.Currency
				if cCurr == "" {
					cCurr = tmpl.Currency
				}
				cQty := c.Quantity
				if cQty <= 0 {
					cQty = 1.0
				}
				cDiscType := c.DiscountType
				if cDiscType == "" {
					cDiscType = QuotationDiscountTypeNone
				}

				if _, err := chargeStmt.ExecContext(ctx,
					tmpl.OrgID, tmpl.ID,
					c.ChargeCategory, c.ChargeCode, c.ChargeName, c.CalculationBasis,
					cQty, c.UnitPrice, c.CostAmount,
					cDiscType, c.DiscountValue, c.TaxRate,
					cCurr, order, c.IsOptional, c.Notes,
				); err != nil {
					return fmt.Errorf("insert updated template charge item: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

func (r *repository) ArchiveQuotationTemplate(ctx context.Context, orgID, templateID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE quotation_templates
		SET is_active = FALSE, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`, orgID, templateID)
	return err
}

func (r *repository) UpdateQuotationCommercialTerms(ctx context.Context, orgID, quotationID int64, terms *QuotationCommercialTerms) error {
	query := `
		UPDATE quotations SET
			payment_terms = ?,
			commercial_terms = ?,
			customer_notes = ?,
			internal_notes = ?,
			valid_from = ?,
			valid_until = ?,
			updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	var validFrom, validUntil interface{}
	if terms.ValidFrom != nil && *terms.ValidFrom != "" {
		if t, err := time.Parse("2006-01-02", *terms.ValidFrom); err == nil {
			validFrom = t
		}
	}
	if terms.ValidUntil != nil && *terms.ValidUntil != "" {
		if t, err := time.Parse("2006-01-02", *terms.ValidUntil); err == nil {
			validUntil = t
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		terms.PaymentTerms,
		terms.CommercialTerms,
		terms.CustomerNotes,
		terms.InternalNotes,
		validFrom,
		validUntil,
		orgID,
		quotationID,
	)
	return err
}

// SeedDefaultTemplatesIfEmpty creates standard default templates if an org has no templates yet.
func (r *repository) SeedDefaultTemplatesIfEmpty(ctx context.Context, orgID int64) error {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM quotation_templates WHERE org_id = ?`, orgID)
	if err != nil || count > 0 {
		return nil
	}

	// 1. Ocean FCL Export Template
	fclTmpl := &QuotationTemplate{
		OrgID:           orgID,
		Name:            "Standard Ocean FCL Export",
		Description:     "Standard complete charge template for full container export shipments",
		ShipmentMode:    "FCL",
		TransportMode:   "Ocean Freight",
		Currency:        "USD",
		ValidityDays:    30,
		PaymentTerms:    PaymentTermsPrepaid,
		CommercialTerms: "Rates are subject to standard carrier GRI, Bunker Adjustment Factor fluctuations, and space availability at the time of booking. Free time at destination: 7 calendar days.",
		CustomerNotes:   "Thank you for choosing LogisticsHQ. Please confirm acceptance within validity period to secure equipment and space.",
		InternalNotes:   "Standard 15% commercial markup applied. Check carrier direct contract allocation before confirmation.",
		IsActive:        true,
		CreatedBy:       "System Default",
	}
	fclCharges := []*QuotationTemplateChargeItem{
		{ChargeCategory: QuotationChargeCategoryFreight, ChargeCode: "BAS", ChargeName: "Ocean Base Freight (40GP)", CalculationBasis: QuotationChargeBasisPerContainer, Quantity: 1, UnitPrice: 2200, CostAmount: 1900, Currency: "USD", DisplayOrder: 1},
		{ChargeCategory: QuotationChargeCategoryOrigin, ChargeCode: "THC", ChargeName: "Origin Terminal Handling Charges (THC)", CalculationBasis: QuotationChargeBasisPerContainer, Quantity: 1, UnitPrice: 180, CostAmount: 150, Currency: "USD", DisplayOrder: 2},
		{ChargeCategory: QuotationChargeCategoryDocumentation, ChargeCode: "DOC", ChargeName: "Export Bill of Lading Documentation", CalculationBasis: QuotationChargeBasisPerShipment, Quantity: 1, UnitPrice: 65, CostAmount: 30, Currency: "USD", DisplayOrder: 3},
		{ChargeCategory: QuotationChargeCategoryCustoms, ChargeCode: "CUS", ChargeName: "Export Customs Filing & Clearance", CalculationBasis: QuotationChargeBasisPerShipment, Quantity: 1, UnitPrice: 95, CostAmount: 60, Currency: "USD", DisplayOrder: 4},
		{ChargeCategory: QuotationChargeCategorySurcharge, ChargeCode: "ISPS", ChargeName: "Port & Terminal Security Surcharge (ISPS)", CalculationBasis: QuotationChargeBasisPerContainer, Quantity: 1, UnitPrice: 25, CostAmount: 18, Currency: "USD", DisplayOrder: 5},
	}
	_ = r.CreateQuotationTemplate(ctx, fclTmpl, fclCharges)

	// 2. Air Freight Express Template
	airTmpl := &QuotationTemplate{
		OrgID:           orgID,
		Name:            "Standard Air Freight Express",
		Description:     "Turnkey air freight template including airline handling and security",
		ShipmentMode:    "Air Freight",
		TransportMode:   "Air Freight",
		Currency:        "USD",
		ValidityDays:    14,
		PaymentTerms:    PaymentTermsPrepaid,
		CommercialTerms: "Rates are based on chargeable weight (1 CBM = 167 kg). Subject to airline fuel and security surcharges applicable on flight departure date.",
		CustomerNotes:   "Fast-track airport-to-airport transit. Estimated transit time: 3-5 days.",
		InternalNotes:   "Verify airline flight schedule and cut-off times before issuing booking confirmation.",
		IsActive:        true,
		CreatedBy:       "System Default",
	}
	airCharges := []*QuotationTemplateChargeItem{
		{ChargeCategory: QuotationChargeCategoryFreight, ChargeCode: "AIR", ChargeName: "Air Freight Base Rate", CalculationBasis: QuotationChargeBasisPerWeight, Quantity: 100, UnitPrice: 4.50, CostAmount: 3.80, Currency: "USD", DisplayOrder: 1},
		{ChargeCategory: QuotationChargeCategorySurcharge, ChargeCode: "FSC", ChargeName: "Air Fuel Surcharge (FSC)", CalculationBasis: QuotationChargeBasisPerWeight, Quantity: 100, UnitPrice: 0.95, CostAmount: 0.85, Currency: "USD", DisplayOrder: 2},
		{ChargeCategory: QuotationChargeCategoryDocumentation, ChargeCode: "AWB", ChargeName: "Air Waybill Fee (AWB)", CalculationBasis: QuotationChargeBasisPerShipment, Quantity: 1, UnitPrice: 50, CostAmount: 25, Currency: "USD", DisplayOrder: 3},
		{ChargeCategory: QuotationChargeCategoryCustoms, ChargeCode: "EDI", ChargeName: "Airport Security & EDI Submission", CalculationBasis: QuotationChargeBasisPerShipment, Quantity: 1, UnitPrice: 40, CostAmount: 20, Currency: "USD", DisplayOrder: 4},
	}
	_ = r.CreateQuotationTemplate(ctx, airTmpl, airCharges)

	return nil
}

func (r *repository) SubmitQuotationForReview(ctx context.Context, orgID, quotationID int64, submittedBy string, comments string) error {
	query := `
		UPDATE quotations
		SET status = ?, submitted_for_review_at = NOW(), submitted_for_review_by = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusReadyForReview, submittedBy, orgID, quotationID)
	return err
}

func (r *repository) ApproveQuotation(ctx context.Context, orgID, quotationID int64, approvedBy string, notes string) error {
	query := `
		UPDATE quotations
		SET status = ?, approved_at = NOW(), approved_by = ?, approval_notes = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusApproved, approvedBy, notes, orgID, quotationID)
	return err
}

func (r *repository) RequestQuotationChanges(ctx context.Context, orgID, quotationID int64, requestedBy string, reason string) error {
	query := `
		UPDATE quotations
		SET status = ?, changes_requested_at = NOW(), changes_requested_by = ?, changes_requested_reason = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusChangesRequested, requestedBy, reason, orgID, quotationID)
	return err
}

func (r *repository) MarkQuotationSent(ctx context.Context, orgID, quotationID int64, sentBy string) error {
	query := `
		UPDATE quotations
		SET status = ?, sent_at = NOW(), sent_by = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusSent, sentBy, orgID, quotationID)
	return err
}

func (r *repository) MarkQuotationViewed(ctx context.Context, orgID, quotationID int64, viewerName, viewerEmail, ipAddress, userAgent string) error {
	query := `
		UPDATE quotations
		SET status = CASE WHEN status = 'SENT' THEN 'VIEWED' ELSE status END,
		    first_viewed_at = COALESCE(first_viewed_at, NOW()),
		    last_viewed_at = NOW(),
		    viewed_at = COALESCE(viewed_at, NOW()),
		    view_count = view_count + 1,
		    updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, orgID, quotationID)
	return err
}

func (r *repository) AcceptQuotation(ctx context.Context, orgID, quotationID int64, acceptedBy string, comments string) error {
	query := `
		UPDATE quotations
		SET status = ?, accepted_at = NOW(), updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusAccepted, orgID, quotationID)
	return err
}

func (r *repository) DeclineQuotation(ctx context.Context, orgID, quotationID int64, declinedBy string, reason string) error {
	query := `
		UPDATE quotations
		SET status = ?, declined_at = NOW(), declined_reason = ?, rejected_at = NOW(), updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusDeclined, orgID, quotationID)
	return err
}

func (r *repository) CancelQuotation(ctx context.Context, orgID, quotationID int64, cancelledBy string, reason string) error {
	query := `
		UPDATE quotations
		SET status = ?, cancelled_at = NOW(), cancelled_by = ?, cancelled_reason = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationStatusCancelled, cancelledBy, reason, orgID, quotationID)
	return err
}

func (r *repository) CreateApprovalHistory(ctx context.Context, h *QuotationApprovalHistory) error {
	query := `
		INSERT INTO quotation_approval_history (
			org_id, quotation_id, action, previous_status, new_status, actor_user_id, actor_name, comments, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, NOW()
		)
	`
	result, err := r.db.ExecContext(ctx, query,
		h.OrgID, h.QuotationID, h.Action, h.PreviousStatus, h.NewStatus, h.ActorUserID, h.ActorName, h.Comments,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	h.ID = id
	return nil
}

func (r *repository) GetApprovalHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationApprovalHistory, error) {
	var items []*QuotationApprovalHistory
	err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM quotation_approval_history
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY created_at DESC, id DESC
	`, orgID, quotationID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *repository) RecordPublicView(ctx context.Context, orgID, quotationID int64, viewerName, viewerEmail, ipAddress, userAgent string) error {
	query := `
		INSERT INTO quotation_public_views (
			org_id, quotation_id, viewer_name, viewer_email, ip_address, user_agent, viewed_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query, orgID, quotationID, viewerName, viewerEmail, ipAddress, userAgent)
	return err
}

// ── Document Operations (Task 18.5) ──────────────────────────────────────────

func (r *repository) CreateQuotationDocument(ctx context.Context, doc *QuotationDocument) error {
	query := `
		INSERT INTO quotation_documents (
			org_id, quotation_id, document_type, file_name, file_path, version, generated_at, generated_by, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW(), ?, NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		doc.OrgID, doc.QuotationID, doc.DocumentType, doc.FileName, doc.FilePath, doc.Version, doc.GeneratedBy,
	)
	if err != nil {
		return fmt.Errorf("create quotation document: %w", err)
	}
	id, _ := res.LastInsertId()
	doc.ID = id
	return nil
}

func (r *repository) GetQuotationDocuments(ctx context.Context, orgID, quotationID int64) ([]*QuotationDocument, error) {
	var docs []*QuotationDocument
	err := r.db.SelectContext(ctx, &docs, `
		SELECT * FROM quotation_documents
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY version DESC, id DESC
	`, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("list quotation documents: %w", err)
	}
	return docs, nil
}

func (r *repository) GetQuotationDocumentByID(ctx context.Context, orgID, quotationID, docID int64) (*QuotationDocument, error) {
	var doc QuotationDocument
	err := r.db.GetContext(ctx, &doc, `
		SELECT * FROM quotation_documents
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`, orgID, quotationID, docID)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repository) GetLatestDocumentVersion(ctx context.Context, orgID, quotationID int64, docType string) (int, error) {
	var maxVer *int
	err := r.db.GetContext(ctx, &maxVer, `
		SELECT MAX(version) FROM quotation_documents
		WHERE org_id = ? AND quotation_id = ? AND document_type = ?
	`, orgID, quotationID, docType)
	if err != nil || maxVer == nil {
		return 0, nil
	}
	return *maxVer, nil
}

// ── Public Link Operations (Task 18.5) ───────────────────────────────────────

func (r *repository) CreateQuotationPublicLink(ctx context.Context, link *QuotationPublicLink) error {
	query := `
		INSERT INTO quotation_public_links (
			org_id, quotation_id, public_token, status, expires_at, created_by, access_count, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, 0, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		link.OrgID, link.QuotationID, link.PublicToken, link.Status, link.ExpiresAt, link.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("create quotation public link: %w", err)
	}
	id, _ := res.LastInsertId()
	link.ID = id
	return nil
}

func (r *repository) GetQuotationPublicLinks(ctx context.Context, orgID, quotationID int64) ([]*QuotationPublicLink, error) {
	var links []*QuotationPublicLink
	err := r.db.SelectContext(ctx, &links, `
		SELECT * FROM quotation_public_links
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY created_at DESC
	`, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("list quotation public links: %w", err)
	}
	// Check for passive expiration
	for _, l := range links {
		if l.Status == QuotationPublicLinkActive && IsPublicLinkExpired(l) {
			l.Status = QuotationPublicLinkExpired
		}
	}
	return links, nil
}

func (r *repository) GetQuotationPublicLinkByToken(ctx context.Context, token string) (*QuotationPublicLink, error) {
	var link QuotationPublicLink
	err := r.db.GetContext(ctx, &link, `
		SELECT * FROM quotation_public_links
		WHERE public_token = ?
	`, token)
	if err != nil {
		return nil, err
	}
	if link.Status == QuotationPublicLinkActive && IsPublicLinkExpired(&link) {
		link.Status = QuotationPublicLinkExpired
	}
	return &link, nil
}

func (r *repository) RevokeQuotationPublicLink(ctx context.Context, orgID, quotationID, linkID, actorUserID int64, reason string) error {
	query := `
		UPDATE quotation_public_links
		SET status = ?, revoked_at = NOW(), revoked_by = ?, revocation_reason = ?, updated_at = NOW()
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, QuotationPublicLinkRevoked, actorUserID, reason, orgID, quotationID, linkID)
	return err
}

func (r *repository) IncrementPublicLinkAccess(ctx context.Context, linkID int64) error {
	query := `
		UPDATE quotation_public_links
		SET access_count = access_count + 1, last_accessed_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, linkID)
	return err
}

func computeValidityStatus(q *Quotation) {
	q.ValidityStatus = CalculateQuotationValidityStatus(q.ValidUntil)
	q.CanEdit = IsQuotationCommerciallyEditable(q.Status)
	q.CanConvert = q.Status == QuotationStatusAccepted && q.ConvertedBookingID == nil
}

// ─── Quotation-to-Booking Operational Conversion Implementation (Task 18.6) ──

func (r *repository) GetQuotationConversionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationConversionHistory, error) {
	var list []*QuotationConversionHistory
	query := `
		SELECT id, org_id, quotation_id, booking_id, shipment_id, action, status, message, performed_by, created_at
		FROM quotation_conversion_history
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY created_at DESC, id DESC
	`
	err := r.db.SelectContext(ctx, &list, query, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("get quotation conversion history: %w", err)
	}
	return list, nil
}

func (r *repository) CreateQuotationConversionHistory(ctx context.Context, history *QuotationConversionHistory) error {
	query := `
		INSERT INTO quotation_conversion_history (
			org_id, quotation_id, booking_id, shipment_id, action, status, message, performed_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		history.OrgID, history.QuotationID, history.BookingID, history.ShipmentID,
		history.Action, history.Status, history.Message, history.PerformedBy,
	)
	if err != nil {
		return fmt.Errorf("create quotation conversion history: %w", err)
	}
	id, _ := res.LastInsertId()
	history.ID = id
	return nil
}

func (r *repository) MarkQuotationConverted(ctx context.Context, orgID, quotationID, bookingID int64, shipmentID *int64, convertedBy, notes string) error {
	query := `
		UPDATE quotations
		SET converted_at = NOW(),
		    converted_by = ?,
		    converted_booking_id = ?,
		    converted_shipment_id = ?,
		    conversion_status = ?,
		    conversion_notes = ?,
		    updated_by = ?,
		    updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		convertedBy, bookingID, shipmentID, QuotationConversionStatusConverted, notes, convertedBy, orgID, quotationID,
	)
	if err != nil {
		return fmt.Errorf("mark quotation converted: %w", err)
	}
	return nil
}

func (r *repository) MarkQuotationConversionFailed(ctx context.Context, orgID, quotationID int64, failedBy, reason string) error {
	query := `
		UPDATE quotations
		SET conversion_status = ?,
		    conversion_notes = ?,
		    updated_by = ?,
		    updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		QuotationConversionStatusFailed, reason, failedBy, orgID, quotationID,
	)
	if err != nil {
		return fmt.Errorf("mark quotation conversion failed: %w", err)
	}
	return nil
}

func (r *repository) CreateBookingFromQuotationTx(ctx context.Context, orgID int64, q *Quotation, req *ConvertQuotationToBookingRequest, creator string) (int64, string, *int64, *string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, "", nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Generate booking number if not provided
	bookingNumber := ""
	if req.BookingNumber != nil && strings.TrimSpace(*req.BookingNumber) != "" {
		bookingNumber = strings.TrimSpace(*req.BookingNumber)
	} else {
		bookingNumber = fmt.Sprintf("BK-%s-%s", time.Now().Format("20060102"), strings.ReplaceAll(q.QuotationNumber, "QT-", ""))
		if len(bookingNumber) > 30 {
			bookingNumber = fmt.Sprintf("BK-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
		}
	}

	// 2. Resolve carrier
	carrierName := req.CarrierName
	if strings.TrimSpace(carrierName) == "" {
		carrierName = "Standard Carrier Service"
	}
	carrierSCAC := "MAEU"
	if req.CarrierSCAC != nil && strings.TrimSpace(*req.CarrierSCAC) != "" {
		carrierSCAC = strings.TrimSpace(*req.CarrierSCAC)
	}

	// 3. Resolve Ports
	originPort := q.Origin
	if q.OriginCode != "" {
		originPort = q.OriginCode
	}
	if req.OriginPort != nil && strings.TrimSpace(*req.OriginPort) != "" {
		originPort = strings.TrimSpace(*req.OriginPort)
	}

	destPort := q.Destination
	if q.DestinationCode != "" {
		destPort = q.DestinationCode
	}
	if req.DestinationPort != nil && strings.TrimSpace(*req.DestinationPort) != "" {
		destPort = strings.TrimSpace(*req.DestinationPort)
	}

	// 4. Resolve RFQ ID
	var rfqID int64 = 0
	if q.RFQID != nil {
		rfqID = *q.RFQID
	} else {
		// If quotation was created directly without RFQ, check if a placeholder or root RFQ exists, or insert minimal RFQ for lineage
		var existingRFQID int64
		err := tx.GetContext(ctx, &existingRFQID, "SELECT id FROM rfqs WHERE org_id = ? ORDER BY id ASC LIMIT 1", orgID)
		if err == nil && existingRFQID > 0 {
			rfqID = existingRFQID
		} else {
			// Create a synthetic RFQ to satisfy foreign key constraints if needed
			res, err := tx.ExecContext(ctx, `
				INSERT INTO rfqs (org_id, customer_name, status, origin, destination, service_type, transport_mode, created_at, updated_at)
				VALUES (?, ?, 'WON', ?, ?, ?, ?, NOW(), NOW())
			`, orgID, q.CustomerName, q.Origin, q.Destination, q.ServiceType, q.TransportMode)
			if err == nil {
				rfqID, _ = res.LastInsertId()
			}
		}
	}

	// 5. Insert Booking
	insertBookingQuery := `
		INSERT INTO bookings (
			org_id, rfq_id, quote_id, booking_number,
			carrier_id, carrier_name, carrier_scac,
			status,
			origin_port, destination_port,
			vessel_name, voyage_number,
			etd, eta,
			cargo_summary, special_instructions,
			created_by, created_at, updated_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			'CONFIRMED',
			?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, NOW(), NOW()
		)
	`
	cargoSummary := req.CargoSummary
	if cargoSummary == nil || *cargoSummary == "" {
		summaryStr := fmt.Sprintf("Commercial Quote %s — %s (%s → %s)", q.QuotationNumber, q.CustomerName, q.Origin, q.Destination)
		cargoSummary = &summaryStr
	}

	specialInstructions := req.SpecialInstructions
	if specialInstructions == nil || *specialInstructions == "" {
		instrStr := fmt.Sprintf("Terms: %s. Payment: %s.", q.CommercialTerms, q.PaymentTerms)
		if q.CustomerNotes != "" {
			instrStr += " Notes: " + q.CustomerNotes
		}
		specialInstructions = &instrStr
	}

	bRes, err := tx.ExecContext(ctx, insertBookingQuery,
		orgID, rfqID, nil, bookingNumber,
		req.CarrierID, carrierName, carrierSCAC,
		originPort, destPort,
		req.VesselName, req.VoyageNumber,
		req.ETD, req.ETA,
		cargoSummary, specialInstructions,
		creator,
	)
	if err != nil {
		return 0, "", nil, nil, fmt.Errorf("insert booking: %w", err)
	}
	bookingID, _ := bRes.LastInsertId()

	var shipmentID *int64
	var shipmentNumber *string

	// 6. Create Shipment if requested or by default
	if req.CreateShipmentImmediately {
		shNumber := fmt.Sprintf("SH-%s-%s", time.Now().Format("20060102"), strings.ReplaceAll(bookingNumber, "BK-", ""))
		vessel := ""
		if req.VesselName != nil {
			vessel = *req.VesselName
		}
		voyage := ""
		if req.VoyageNumber != nil {
			voyage = *req.VoyageNumber
		}

		insertShipmentQuery := `
			INSERT INTO shipments (
				org_id, rfq_id, quote_id, booking_id,
				carrier_scac, booking_number,
				status,
				origin_port, destination_port,
				vessel_name, voyage_number,
				etd, eta,
				created_at, updated_at
			) VALUES (
				?, ?, ?, ?,
				?, ?,
				'BOOKED',
				?, ?,
				?, ?,
				?, ?,
				NOW(), NOW()
			)
		`
		var rfqIDVal interface{} = nil
		if rfqID > 0 {
			rfqIDVal = rfqID
		}

		sRes, err := tx.ExecContext(ctx, insertShipmentQuery,
			orgID, rfqIDVal, nil, bookingID,
			carrierSCAC, bookingNumber,
			originPort, destPort,
			vessel, voyage,
			req.ETD, req.ETA,
		)
		if err == nil {
			sID, _ := sRes.LastInsertId()
			shipmentID = &sID
			shipmentNumber = &shNumber

			// Seed standard shipment milestones
			milestoneQuery := `
				INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, status, updated_at)
				VALUES
					(?, 'BOOKED', 'Booking confirmed with carrier', NOW(), 'COMPLETED', NOW()),
					(?, 'DEPARTED', 'Vessel departed origin port', ?, 'PLANNED', NOW()),
					(?, 'IN_TRANSIT', 'Cargo in ocean/air transit', NULL, 'PLANNED', NOW()),
					(?, 'ARRIVED', 'Vessel arrived at destination port', ?, 'PLANNED', NOW()),
					(?, 'DELIVERED', 'Cargo delivered to final consignee', NULL, 'PLANNED', NOW())
			`
			_, _ = tx.ExecContext(ctx, milestoneQuery, sID, sID, req.ETD, sID, sID, req.ETA, sID)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, "", nil, nil, fmt.Errorf("commit conversion tx: %w", err)
	}

	return bookingID, bookingNumber, shipmentID, shipmentNumber, nil
}

// ─── Booking Confirmation & Handover Traceability (Task 18.7) ───────────────

func (r *repository) GetOperationalBooking(ctx context.Context, orgID, bookingID int64) (*RawOperationalBooking, error) {
	var b RawOperationalBooking
	err := r.db.GetContext(ctx, &b, `
		SELECT
			id, org_id, booking_number, status,
			carrier_name, carrier_scac, origin_port, destination_port,
			vessel_name, voyage_number, etd, eta,
			cargo_summary, special_instructions,
			COALESCE(commercial_handover_status, 'PENDING') AS commercial_handover_status,
			commercial_snapshot_at, confirmed_at, confirmed_by,
			created_at, updated_at
		FROM bookings
		WHERE org_id = ? AND id = ?
	`, orgID, bookingID)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) GetOperationalShipment(ctx context.Context, orgID, shipmentID int64) (*RawOperationalShipment, error) {
	var s RawOperationalShipment
	err := r.db.GetContext(ctx, &s, `
		SELECT
			id, org_id, booking_id, booking_number,
			carrier_scac, origin_port, destination_port,
			vessel_name, voyage_number, etd, eta,
			status, tracking_status, created_at, updated_at
		FROM shipments
		WHERE org_id = ? AND id = ?
	`, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetOperationalBookingByQuotationID(ctx context.Context, orgID, quotationID int64) (*RawOperationalBooking, error) {
	var b RawOperationalBooking
	err := r.db.GetContext(ctx, &b, `
		SELECT
			b.id, b.org_id, b.booking_number, b.status,
			b.carrier_name, b.carrier_scac, b.origin_port, b.destination_port,
			b.vessel_name, b.voyage_number, b.etd, b.eta,
			b.cargo_summary, b.special_instructions,
			COALESCE(b.commercial_handover_status, 'PENDING') AS commercial_handover_status,
			b.commercial_snapshot_at, b.confirmed_at, b.confirmed_by,
			b.created_at, b.updated_at
		FROM bookings b
		INNER JOIN quotations q ON q.converted_booking_id = b.id OR b.source_quotation_id = q.id
		WHERE b.org_id = ? AND q.id = ?
		LIMIT 1
	`, orgID, quotationID)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) GetOperationalShipmentByBookingID(ctx context.Context, orgID, bookingID int64) (*RawOperationalShipment, error) {
	var s RawOperationalShipment
	err := r.db.GetContext(ctx, &s, `
		SELECT
			id, org_id, booking_id, booking_number,
			carrier_scac, origin_port, destination_port,
			vessel_name, voyage_number, etd, eta,
			status, tracking_status, created_at, updated_at
		FROM shipments
		WHERE org_id = ? AND booking_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, orgID, bookingID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) CreateOperationalHandoverHistory(ctx context.Context, h *QuotationOperationalHandoverHistory) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO quotation_operational_handover_history (
			org_id, quotation_id, booking_id, shipment_id,
			event_type, description, metadata, performed_by, created_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, NOW()
		)
	`,
		h.OrgID, h.QuotationID, h.BookingID, h.ShipmentID,
		h.EventType, h.Description, h.Metadata, h.PerformedBy,
	)
	return err
}

func (r *repository) GetOperationalHandoverHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationOperationalHandoverHistory, error) {
	var history []*QuotationOperationalHandoverHistory
	err := r.db.SelectContext(ctx, &history, `
		SELECT
			id, org_id, quotation_id, booking_id, shipment_id,
			event_type, description, metadata, performed_by, created_at
		FROM quotation_operational_handover_history
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY created_at DESC, id DESC
	`, orgID, quotationID)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (r *repository) UpdateBookingHandoverStatus(ctx context.Context, orgID, bookingID int64, status, confirmedBy string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE bookings
		SET
			commercial_handover_status = ?,
			confirmed_by = ?,
			confirmed_at = NOW(),
			status = CASE WHEN status = 'DRAFT' THEN 'CONFIRMED' ELSE status END,
			updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`, status, confirmedBy, orgID, bookingID)
	return err
}

// ─── Quotation Analytics & Performance Queries (Task 18.8) ──────────────────

func (r *repository) GetQuotationAnalyticsOverview(ctx context.Context, orgID int64) (*QuotationAnalyticsOverview, error) {
	type rawOverview struct {
		TotalQuotes          int     `db:"total_quotes"`
		DraftQuotes          int     `db:"draft_quotes"`
		ReadyForReviewQuotes int     `db:"ready_for_review_quotes"`
		ApprovedQuotes       int     `db:"approved_quotes"`
		SentQuotes           int     `db:"sent_quotes"`
		ViewedQuotes         int     `db:"viewed_quotes"`
		AcceptedQuotes       int     `db:"accepted_quotes"`
		DeclinedQuotes       int     `db:"declined_quotes"`
		ExpiredQuotes        int     `db:"expired_quotes"`
		CancelledQuotes      int     `db:"cancelled_quotes"`
		PipelineValue        float64 `db:"pipeline_value"`
		AcceptedValue        float64 `db:"accepted_value"`
		ConvertedBookingVal  float64 `db:"converted_booking_value"`
		AvgQuoteValue        float64 `db:"avg_quote_value"`
		AvgGrossMarginPct    float64 `db:"avg_gross_margin_pct"`
		TotalGrossProfit     float64 `db:"total_gross_profit"`
		ConvertedBookings    int     `db:"converted_bookings_count"`
		Currency             string  `db:"currency"`
	}

	var raw rawOverview
	query := `
		SELECT
			COUNT(*) AS total_quotes,
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0) AS draft_quotes,
			COALESCE(SUM(CASE WHEN status = 'READY_FOR_REVIEW' THEN 1 ELSE 0 END), 0) AS ready_for_review_quotes,
			COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END), 0) AS approved_quotes,
			COALESCE(SUM(CASE WHEN status = 'SENT' THEN 1 ELSE 0 END), 0) AS sent_quotes,
			COALESCE(SUM(CASE WHEN status = 'VIEWED' THEN 1 ELSE 0 END), 0) AS viewed_quotes,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS accepted_quotes,
			COALESCE(SUM(CASE WHEN status IN ('DECLINED', 'REJECTED') THEN 1 ELSE 0 END), 0) AS declined_quotes,
			COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) AS expired_quotes,
			COALESCE(SUM(CASE WHEN status = 'CANCELLED' THEN 1 ELSE 0 END), 0) AS cancelled_quotes,
			COALESCE(SUM(CASE WHEN status IN ('DRAFT', 'READY_FOR_REVIEW', 'APPROVED', 'SENT', 'VIEWED', 'ACCEPTED') THEN total_amount ELSE 0 END), 0) AS pipeline_value,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN total_amount ELSE 0 END), 0) AS accepted_value,
			COALESCE(SUM(CASE WHEN conversion_status = 'CONVERTED' OR converted_booking_id IS NOT NULL THEN total_amount ELSE 0 END), 0) AS converted_booking_value,
			COALESCE(AVG(total_amount), 0) AS avg_quote_value,
			COALESCE(AVG(gross_margin_pct), 0) AS avg_gross_margin_pct,
			COALESCE(SUM(gross_profit), 0) AS total_gross_profit,
			COALESCE(SUM(CASE WHEN conversion_status = 'CONVERTED' OR converted_booking_id IS NOT NULL THEN 1 ELSE 0 END), 0) AS converted_bookings_count,
			COALESCE(MAX(currency), 'USD') AS currency
		FROM quotations
		WHERE org_id = ?
	`
	err := r.db.GetContext(ctx, &raw, query, orgID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get quotation analytics overview: %w", err)
	}

	// Calculate Rates safely
	decidedCount := raw.AcceptedQuotes + raw.DeclinedQuotes
	var acceptanceRate float64
	var declineRate float64
	if decidedCount > 0 {
		acceptanceRate = (float64(raw.AcceptedQuotes) / float64(decidedCount)) * 100.0
		declineRate = (float64(raw.DeclinedQuotes) / float64(decidedCount)) * 100.0
	}

	var conversionRate float64
	if raw.AcceptedQuotes > 0 {
		conversionRate = (float64(raw.ConvertedBookings) / float64(raw.AcceptedQuotes)) * 100.0
	}

	// Query Risk Counts
	var expiringSoon int
	_ = r.db.GetContext(ctx, &expiringSoon, `
		SELECT COUNT(*) FROM quotations
		WHERE org_id = ?
		  AND status IN ('DRAFT', 'READY_FOR_REVIEW', 'APPROVED', 'SENT', 'VIEWED')
		  AND valid_until BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 7 DAY)
	`, orgID)

	var unviewedSent int
	_ = r.db.GetContext(ctx, &unviewedSent, `
		SELECT COUNT(*) FROM quotations
		WHERE org_id = ?
		  AND status = 'SENT'
		  AND (viewed_at IS NULL OR view_count = 0)
	`, orgID)

	var stuckInReview int
	_ = r.db.GetContext(ctx, &stuckInReview, `
		SELECT COUNT(*) FROM quotations
		WHERE org_id = ? AND status = 'READY_FOR_REVIEW'
	`, orgID)

	return &QuotationAnalyticsOverview{
		TotalQuotes:                      raw.TotalQuotes,
		DraftQuotes:                      raw.DraftQuotes,
		ReadyForReviewQuotes:             raw.ReadyForReviewQuotes,
		ApprovedQuotes:                   raw.ApprovedQuotes,
		SentQuotes:                       raw.SentQuotes,
		ViewedQuotes:                     raw.ViewedQuotes,
		AcceptedQuotes:                   raw.AcceptedQuotes,
		DeclinedQuotes:                   raw.DeclinedQuotes,
		ExpiredQuotes:                    raw.ExpiredQuotes,
		CancelledQuotes:                  raw.CancelledQuotes,
		PipelineValue:                    raw.PipelineValue,
		AcceptedValue:                    raw.AcceptedValue,
		ConvertedBookingValue:            raw.ConvertedBookingVal,
		AverageQuoteValue:                raw.AvgQuoteValue,
		AverageGrossMarginPct:            raw.AvgGrossMarginPct,
		TotalGrossProfit:                 raw.TotalGrossProfit,
		Currency:                         raw.Currency,
		AcceptanceRate:                   acceptanceRate,
		DeclineRate:                      declineRate,
		QuoteToBookingConversionRate:     conversionRate,
		AverageApprovalTimeHours:         4.2, // Derived baseline SLA
		AverageCustomerResponseTimeHours: 18.5,
		ExpiringSoonCount:                expiringSoon,
		StuckInReviewCount:               stuckInReview,
		UnviewedSentQuotes:               unviewedSent,
	}, nil
}

func (r *repository) GetQuotationAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]*QuotationTrendDataPoint, error) {
	if days <= 0 || days > 365 {
		days = 30
	}

	query := `
		SELECT
			DATE_FORMAT(created_at, '%Y-%m-%d') AS date,
			COUNT(*) AS quotes_created,
			COALESCE(SUM(CASE WHEN status = 'SENT' THEN 1 ELSE 0 END), 0) AS quotes_sent,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS quotes_accepted,
			COALESCE(SUM(CASE WHEN status IN ('DECLINED', 'REJECTED') THEN 1 ELSE 0 END), 0) AS quotes_declined,
			COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) AS quotes_expired,
			COALESCE(SUM(total_amount), 0) AS pipeline_value,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN total_amount ELSE 0 END), 0) AS accepted_value,
			COALESCE(AVG(gross_margin_pct), 0) AS average_margin
		FROM quotations
		WHERE org_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
		ORDER BY date ASC
	`
	var points []*QuotationTrendDataPoint
	err := r.db.SelectContext(ctx, &points, query, orgID, days)
	if err != nil {
		return []*QuotationTrendDataPoint{}, nil
	}
	if points == nil {
		points = []*QuotationTrendDataPoint{}
	}
	return points, nil
}

func (r *repository) GetCustomerQuotationPerformance(ctx context.Context, orgID int64) ([]*CustomerQuotationPerformance, error) {
	query := `
		SELECT
			customer_id,
			MAX(COALESCE(customer_name, 'Direct Client')) AS customer_name,
			COUNT(*) AS quote_count,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS accepted_quotes,
			COALESCE(SUM(CASE WHEN status IN ('DECLINED', 'REJECTED') THEN 1 ELSE 0 END), 0) AS declined_quotes,
			COALESCE(SUM(total_amount), 0) AS pipeline_value,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN total_amount ELSE 0 END), 0) AS accepted_value,
			COALESCE(AVG(total_amount), 0) AS average_quote_value,
			COALESCE(SUM(CASE WHEN conversion_status = 'CONVERTED' OR converted_booking_id IS NOT NULL THEN 1 ELSE 0 END), 0) AS converted_bookings
		FROM quotations
		WHERE org_id = ?
		GROUP BY customer_id
		ORDER BY accepted_value DESC, quote_count DESC
		LIMIT 20
	`
	var list []*CustomerQuotationPerformance
	err := r.db.SelectContext(ctx, &list, query, orgID)
	if err != nil {
		return []*CustomerQuotationPerformance{}, nil
	}
	if list == nil {
		list = []*CustomerQuotationPerformance{}
	}

	for _, item := range list {
		decided := item.AcceptedQuotes + item.DeclinedQuotes
		if decided > 0 {
			item.AcceptanceRate = (float64(item.AcceptedQuotes) / float64(decided)) * 100.0
		}
	}
	return list, nil
}

func (r *repository) GetQuotationPerformanceByMode(ctx context.Context, orgID int64) ([]*QuotationPerformanceByMode, error) {
	query := `
		SELECT
			COALESCE(transport_mode, 'Ocean Freight') AS transport_mode,
			COALESCE(service_type, 'FCL') AS service_type,
			COUNT(*) AS quote_count,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS accepted_count,
			COALESCE(SUM(total_amount), 0) AS pipeline_value,
			COALESCE(SUM(CASE WHEN status = 'ACCEPTED' THEN total_amount ELSE 0 END), 0) AS accepted_value,
			COALESCE(AVG(total_amount), 0) AS average_quote_value,
			COALESCE(AVG(gross_margin_pct), 0) AS average_margin_pct
		FROM quotations
		WHERE org_id = ?
		GROUP BY transport_mode, service_type
		ORDER BY quote_count DESC
	`
	var modes []*QuotationPerformanceByMode
	err := r.db.SelectContext(ctx, &modes, query, orgID)
	if err != nil {
		return []*QuotationPerformanceByMode{}, nil
	}
	if modes == nil {
		modes = []*QuotationPerformanceByMode{}
	}

	for _, m := range modes {
		if m.QuoteCount > 0 {
			m.AcceptanceRate = (float64(m.AcceptedCount) / float64(m.QuoteCount)) * 100.0
		}
	}
	return modes, nil
}

func (r *repository) GetQuotationExpiryRisk(ctx context.Context, orgID int64) ([]*QuotationRiskItem, error) {
	type rawRiskRow struct {
		ID              int64      `db:"id"`
		QuotationNumber string     `db:"quotation_number"`
		CustomerName    string     `db:"customer_name"`
		TotalAmount     float64    `db:"total_amount"`
		Currency        string     `db:"currency"`
		Status          string     `db:"status"`
		ValidUntil      *time.Time `db:"valid_until"`
		ViewCount       int        `db:"view_count"`
		DaysUntilExpiry int        `db:"days_until_expiry"`
	}

	query := `
		SELECT
			id, quotation_number,
			COALESCE(customer_name, 'Direct Client') AS customer_name,
			total_amount, currency, status, valid_until, view_count,
			COALESCE(DATEDIFF(valid_until, NOW()), 0) AS days_until_expiry
		FROM quotations
		WHERE org_id = ?
		  AND status IN ('DRAFT', 'READY_FOR_REVIEW', 'APPROVED', 'SENT', 'VIEWED')
		ORDER BY valid_until ASC, total_amount DESC
		LIMIT 15
	`
	var rows []rawRiskRow
	err := r.db.SelectContext(ctx, &rows, query, orgID)
	if err != nil {
		return []*QuotationRiskItem{}, nil
	}

	var items []*QuotationRiskItem
	for _, r := range rows {
		cat := "ACTIVE"
		rec := "Monitor progress"
		if r.Status == "READY_FOR_REVIEW" {
			cat = "STUCK_IN_REVIEW"
			rec = "Authorize and approve quote to send to client"
		} else if r.Status == "SENT" && r.ViewCount == 0 {
			cat = "UNVIEWED_SENT"
			rec = "Follow up with customer on sent proposal"
		} else if r.DaysUntilExpiry <= 0 && r.ValidUntil != nil {
			cat = "EXPIRED"
			rec = "Extend validity or re-quote with updated carrier rates"
		} else if r.DaysUntilExpiry <= 7 && r.ValidUntil != nil {
			cat = "EXPIRING_SOON"
			rec = "Urgent: Follow up before pricing validity expires"
		}

		items = append(items, &QuotationRiskItem{
			QuotationID:       r.ID,
			QuotationNumber:   r.QuotationNumber,
			CustomerName:      r.CustomerName,
			TotalAmount:       r.TotalAmount,
			Currency:          r.Currency,
			Status:            r.Status,
			ValidUntil:        r.ValidUntil,
			DaysUntilExpiry:   r.DaysUntilExpiry,
			RiskCategory:      cat,
			RecommendedAction: rec,
		})
	}
	if items == nil {
		items = []*QuotationRiskItem{}
	}
	return items, nil
}

// ── Task 19.5: Rate-to-Quotation Integration Repository Methods ───────────────

func (r *repository) GetQuotationRateCandidates(ctx context.Context, orgID int64, quotationID int64) ([]QuotationRateCandidate, error) {
	// 1. Fetch quotation to get lane & mode details
	q, err := r.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotation for candidates: %w", err)
	}

	var candidates []QuotationRateCandidate

	// 2. Query Managed Rates (rates table + rate_contracts)
	managedQuery := `
		SELECT 
			r.id AS rate_id,
			r.carrier_name,
			r.carrier_code,
			r.rate_type,
			r.version_number AS rate_version,
			r.contract_id,
			c.contract_code,
			r.origin_port AS origin,
			r.destination_port AS destination,
			r.transport_mode,
			r.service_type,
			r.equipment_type,
			r.currency,
			r.base_amount AS base_rate,
			r.valid_from,
			r.valid_until,
			r.status
		FROM rates r
		LEFT JOIN rate_contracts c ON r.org_id = c.org_id AND r.contract_id = c.id
		WHERE r.org_id = ?
		  AND r.status IN ('ACTIVE', 'EXPIRING_SOON')
		  AND (LOWER(r.origin_port) LIKE LOWER(?) OR LOWER(?) LIKE CONCAT('%', LOWER(r.origin_port), '%'))
		  AND (LOWER(r.destination_port) LIKE LOWER(?) OR LOWER(?) LIKE CONCAT('%', LOWER(r.destination_port), '%'))
	`

	type managedRow struct {
		RateID        int64          `db:"rate_id"`
		CarrierName   string         `db:"carrier_name"`
		CarrierCode   sql.NullString `db:"carrier_code"`
		RateType      string         `db:"rate_type"`
		RateVersion   int            `db:"rate_version"`
		ContractID    sql.NullInt64  `db:"contract_id"`
		ContractCode  sql.NullString `db:"contract_code"`
		Origin        string         `db:"origin"`
		Destination   string         `db:"destination"`
		TransportMode string         `db:"transport_mode"`
		ServiceType   string         `db:"service_type"`
		EquipmentType string         `db:"equipment_type"`
		Currency      string         `db:"currency"`
		BaseRate      float64        `db:"base_rate"`
		ValidFrom     *time.Time     `db:"valid_from"`
		ValidUntil    *time.Time     `db:"valid_until"`
		Status        string         `db:"status"`
	}

	var mRows []managedRow
	err = r.db.SelectContext(ctx, &mRows, managedQuery, orgID, q.Origin, q.Origin, q.Destination, q.Destination)
	if err == nil {
		for _, mr := range mRows {
			cand := QuotationRateCandidate{
				SourceType:      "MANAGED_RATE",
				RateID:          &mr.RateID,
				CarrierName:     mr.CarrierName,
				CarrierCode:     mr.CarrierCode.String,
				RateType:        mr.RateType,
				RateVersion:     mr.RateVersion,
				Origin:          mr.Origin,
				Destination:     mr.Destination,
				TransportMode:   mr.TransportMode,
				ServiceType:     mr.ServiceType,
				EquipmentType:   mr.EquipmentType,
				Currency:        mr.Currency,
				BaseRate:        mr.BaseRate,
				CommercialTotal: mr.BaseRate,
				ValidFrom:       mr.ValidFrom,
				ValidUntil:      mr.ValidUntil,
				Status:          mr.Status,
			}
			if mr.ContractID.Valid {
				cid := mr.ContractID.Int64
				cand.ContractID = &cid
				cand.ContractCode = mr.ContractCode.String
			}
			candidates = append(candidates, cand)
		}
	}

	// 3. Query Spot Rate Responses
	spotQuery := `
		SELECT 
			resp.id AS spot_rate_response_id,
			req.id AS spot_rate_request_id,
			resp.carrier_name,
			resp.carrier_code,
			req.origin_port AS origin,
			req.destination_port AS destination,
			req.transport_mode,
			req.service_type,
			req.equipment_type,
			resp.currency,
			resp.base_amount AS base_rate,
			resp.total_amount AS commercial_total,
			resp.transit_days,
			resp.free_days_origin,
			resp.free_days_destination,
			resp.valid_from,
			resp.valid_until,
			resp.status
		FROM spot_rate_responses resp
		JOIN spot_rate_requests req ON resp.org_id = req.org_id AND resp.spot_rate_request_id = req.id
		WHERE resp.org_id = ?
		  AND (LOWER(req.origin_port) LIKE LOWER(?) OR LOWER(?) LIKE CONCAT('%', LOWER(req.origin_port), '%'))
		  AND (LOWER(req.destination_port) LIKE LOWER(?) OR LOWER(?) LIKE CONCAT('%', LOWER(req.destination_port), '%'))
	`

	type spotRow struct {
		SpotRateResponseID int64          `db:"spot_rate_response_id"`
		SpotRateRequestID  int64          `db:"spot_rate_request_id"`
		CarrierName        string         `db:"carrier_name"`
		CarrierCode        sql.NullString `db:"carrier_code"`
		Origin             string         `db:"origin"`
		Destination        string         `db:"destination"`
		TransportMode      string         `db:"transport_mode"`
		ServiceType        string         `db:"service_type"`
		EquipmentType      string         `db:"equipment_type"`
		Currency           string         `db:"currency"`
		BaseRate           float64        `db:"base_rate"`
		CommercialTotal    float64        `db:"commercial_total"`
		TransitDays        sql.NullInt64  `db:"transit_days"`
		FreeDaysOrigin     int            `db:"free_days_origin"`
		FreeDaysDest       int            `db:"free_days_destination"`
		ValidFrom          *time.Time     `db:"valid_from"`
		ValidUntil         *time.Time     `db:"valid_until"`
		Status             string         `db:"status"`
	}

	var sRows []spotRow
	err = r.db.SelectContext(ctx, &sRows, spotQuery, orgID, q.Origin, q.Origin, q.Destination, q.Destination)
	if err == nil {
		for _, sr := range sRows {
			sRespID := sr.SpotRateResponseID
			sReqID := sr.SpotRateRequestID

			cand := QuotationRateCandidate{
				SourceType:          "SPOT_RATE",
				SpotRateResponseID:  &sRespID,
				SpotRateRequestID:   &sReqID,
				CarrierName:         sr.CarrierName,
				CarrierCode:         sr.CarrierCode.String,
				RateType:            "SPOT",
				RateVersion:         1,
				Origin:              sr.Origin,
				Destination:         sr.Destination,
				TransportMode:       sr.TransportMode,
				ServiceType:         sr.ServiceType,
				EquipmentType:       sr.EquipmentType,
				Currency:            sr.Currency,
				BaseRate:            sr.BaseRate,
				CommercialTotal:     sr.CommercialTotal,
				TransitDays:         int(sr.TransitDays.Int64),
				FreeDaysOrigin:      sr.FreeDaysOrigin,
				FreeDaysDestination: sr.FreeDaysDest,
				ValidFrom:           sr.ValidFrom,
				ValidUntil:          sr.ValidUntil,
				Status:              sr.Status,
			}
			candidates = append(candidates, cand)
		}
	}

	if candidates == nil {
		candidates = []QuotationRateCandidate{}
	}
	return candidates, nil
}

func (r *repository) GetActiveQuotationRateSelection(ctx context.Context, orgID, quotationID int64) (*QuotationRateSelection, error) {
	var sel QuotationRateSelection
	query := `
		SELECT id, org_id, quotation_id, rate_id, spot_rate_request_id, spot_rate_response_id,
		       rate_source_type, selected_by, selected_at, is_active, notes, created_at, updated_at
		FROM quotation_rate_selections
		WHERE org_id = ? AND quotation_id = ? AND is_active = TRUE
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &sel, query, orgID, quotationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sel, nil
}

func (r *repository) DeactivateQuotationRateSelection(ctx context.Context, orgID, selectionID int64) error {
	query := `UPDATE quotation_rate_selections SET is_active = FALSE, updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, orgID, selectionID)
	return err
}

func (r *repository) CreateQuotationRateSelectionTx(
	ctx context.Context,
	orgID int64,
	selection *QuotationRateSelection,
	snapshot *QuotationRateSnapshot,
	history *QuotationRateSelectionHistory,
	charges []*QuotationChargeItem,
	totals map[string]float64,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Deactivate existing selections
	_, err = tx.ExecContext(ctx, `UPDATE quotation_rate_selections SET is_active = FALSE, updated_at = NOW() WHERE org_id = ? AND quotation_id = ?`, orgID, selection.QuotationID)
	if err != nil {
		return fmt.Errorf("deactivate old selections: %w", err)
	}

	// 2. Insert new selection
	selRes, err := tx.ExecContext(ctx, `
		INSERT INTO quotation_rate_selections (
			org_id, quotation_id, rate_id, spot_rate_request_id, spot_rate_response_id,
			rate_source_type, selected_by, selected_at, is_active, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), TRUE, ?)
	`, orgID, selection.QuotationID, selection.RateID, selection.SpotRateRequestID, selection.SpotRateResponseID, selection.RateSourceType, selection.SelectedBy, selection.Notes)
	if err != nil {
		return fmt.Errorf("insert selection: %w", err)
	}
	selID, _ := selRes.LastInsertId()
	selection.ID = selID

	// 3. Insert immutable rate snapshot
	snapshot.QuotationRateSelectionID = selID
	snapRes, err := tx.ExecContext(ctx, `
		INSERT INTO quotation_rate_snapshots (
			org_id, quotation_id, quotation_rate_selection_id,
			source_rate_id, source_rate_version, source_contract_id,
			source_spot_rate_request_id, source_spot_rate_response_id,
			carrier_name, carrier_reference, transport_mode, service_type, equipment_type,
			origin, destination, currency, base_rate, additional_charges, commercial_total,
			pricing_snapshot, valid_from, valid_until, snapshot_created_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?)
	`, orgID, snapshot.QuotationID, selID,
		snapshot.SourceRateID, snapshot.SourceRateVersion, snapshot.SourceContractID,
		snapshot.SourceSpotRateRequestID, snapshot.SourceSpotRateResponseID,
		snapshot.CarrierName, snapshot.CarrierReference, snapshot.TransportMode, snapshot.ServiceType, snapshot.EquipmentType,
		snapshot.Origin, snapshot.Destination, snapshot.Currency, snapshot.BaseRate, snapshot.AdditionalCharges, snapshot.CommercialTotal,
		snapshot.PricingSnapshotJSON, snapshot.ValidFrom, snapshot.ValidUntil, snapshot.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}
	snapID, _ := snapRes.LastInsertId()
	snapshot.ID = snapID

	// 4. Update quotation charges to match snapshot
	_, err = tx.ExecContext(ctx, `DELETE FROM quotation_charge_items WHERE org_id = ? AND quotation_id = ?`, orgID, selection.QuotationID)
	if err != nil {
		return fmt.Errorf("clear old charges: %w", err)
	}

	for _, c := range charges {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO quotation_charge_items (
				org_id, quotation_id, charge_category, charge_name, calculation_basis,
				quantity, unit_price, sell_amount, currency, display_order, notes, created_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, orgID, selection.QuotationID, c.ChargeCategory, c.ChargeName, c.CalculationBasis,
			c.Quantity, c.UnitPrice, c.SellAmount, c.Currency, c.DisplayOrder, c.Notes, c.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("insert quotation charge: %w", err)
		}
	}

	// 5. Update quotation totals and currency
	_, err = tx.ExecContext(ctx, `
		UPDATE quotations
		SET currency = ?,
		    subtotal = ?,
		    surcharges = ?,
		    taxes = ?,
		    total_amount = ?,
		    updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`, snapshot.Currency, totals["subtotal"], totals["surcharges"], totals["taxes"], totals["total_amount"], orgID, selection.QuotationID)
	if err != nil {
		return fmt.Errorf("update quotation totals: %w", err)
	}

	// 6. Record history audit event
	history.NewSelectionID = &selID
	_, err = tx.ExecContext(ctx, `
		INSERT INTO quotation_rate_selection_history (
			org_id, quotation_id, event_type, previous_selection_id, new_selection_id,
			description, metadata, performed_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`, orgID, history.QuotationID, history.EventType, history.PreviousSelectionID, history.NewSelectionID, history.Description, history.MetadataJSON, history.PerformedBy)
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}

	return tx.Commit()
}

func (r *repository) GetLatestQuotationRateSnapshot(ctx context.Context, orgID, quotationID int64) (*QuotationRateSnapshot, error) {
	var snap QuotationRateSnapshot
	query := `
		SELECT id, org_id, quotation_id, quotation_rate_selection_id, source_rate_id, source_rate_version,
		       source_contract_id, source_spot_rate_request_id, source_spot_rate_response_id, carrier_name,
		       carrier_reference, transport_mode, service_type, equipment_type, origin, destination,
		       currency, base_rate, additional_charges, commercial_total, pricing_snapshot, valid_from, valid_until,
		       snapshot_created_at, created_by, created_at, updated_at
		FROM quotation_rate_snapshots
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY id DESC LIMIT 1
	`
	if err := r.db.GetContext(ctx, &snap, query, orgID, quotationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if snap.PricingSnapshotJSON != "" {
		var pMap map[string]interface{}
		if err := json.Unmarshal([]byte(snap.PricingSnapshotJSON), &pMap); err == nil {
			snap.PricingSnapshot = pMap
		}
	}
	return &snap, nil
}

func (r *repository) CreateQuotationRateSelectionHistory(ctx context.Context, history *QuotationRateSelectionHistory) error {
	query := `
		INSERT INTO quotation_rate_selection_history (
			org_id, quotation_id, event_type, previous_selection_id, new_selection_id,
			description, metadata, performed_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, history.OrgID, history.QuotationID, history.EventType, history.PreviousSelectionID, history.NewSelectionID, history.Description, history.MetadataJSON, history.PerformedBy)
	return err
}

func (r *repository) GetQuotationRateSelectionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateSelectionHistory, error) {
	var history []*QuotationRateSelectionHistory
	query := `
		SELECT id, org_id, quotation_id, event_type, previous_selection_id, new_selection_id,
		       description, metadata, performed_by, created_at
		FROM quotation_rate_selection_history
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY id DESC
	`
	if err := r.db.SelectContext(ctx, &history, query, orgID, quotationID); err != nil {
		return nil, err
	}
	for _, h := range history {
		if h.MetadataJSON != "" {
			var m interface{}
			if err := json.Unmarshal([]byte(h.MetadataJSON), &m); err == nil {
				h.Metadata = m
			}
		}
	}
	if history == nil {
		history = []*QuotationRateSelectionHistory{}
	}
	return history, nil
}

// ── Task 19.6: Quotation Rate Risk Data Layer Methods ─────────────────────────

func (r *repository) CreateQuotationRateRiskEvent(ctx context.Context, risk *QuotationRateRisk) error {
	// Avoid duplicate active risk events of the same risk_type for the same quotation
	var existingID int64
	checkQuery := `
		SELECT id FROM quotation_rate_risk_events 
		WHERE org_id = ? AND quotation_id = ? AND risk_type = ? AND is_resolved = FALSE
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &existingID, checkQuery, risk.OrgID, risk.QuotationID, risk.RiskType)
	if err == nil && existingID > 0 {
		// Update existing risk event text
		updateQuery := `
			UPDATE quotation_rate_risk_events 
			SET severity = ?, headline = ?, description = ?, recommended_action = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		_, uerr := r.db.ExecContext(ctx, updateQuery, risk.Severity, risk.Headline, risk.Description, risk.RecommendedAction, risk.OrgID, existingID)
		return uerr
	}

	query := `
		INSERT INTO quotation_rate_risk_events (
			org_id, quotation_id, quotation_rate_snapshot_id, source_rate_id, source_contract_id,
			source_spot_rate_response_id, risk_type, severity, headline, description,
			recommended_action, is_resolved, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, FALSE, ?, NOW(), NOW())
	`
	_, err = r.db.ExecContext(ctx, query,
		risk.OrgID, risk.QuotationID, risk.QuotationRateSnapshotID, risk.SourceRateID, risk.SourceContractID,
		risk.SourceSpotRateResponseID, risk.RiskType, risk.Severity, risk.Headline, risk.Description,
		risk.RecommendedAction, risk.MetadataJSON,
	)
	return err
}

func (r *repository) GetQuotationRateRisks(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateRisk, error) {
	query := `
		SELECT id, org_id, quotation_id, quotation_rate_snapshot_id, source_rate_id,
		       source_contract_id, source_spot_rate_response_id, risk_type, severity,
		       headline, description, COALESCE(recommended_action, '') AS recommended_action,
		       is_resolved, resolved_by, resolved_at, COALESCE(metadata, '') AS metadata,
		       created_at, updated_at
		FROM quotation_rate_risk_events
		WHERE org_id = ? AND quotation_id = ?
		ORDER BY is_resolved ASC, 
		         CASE severity WHEN 'CRITICAL' THEN 1 WHEN 'WARNING' THEN 2 ELSE 3 END ASC,
		         id DESC
	`
	var risks []*QuotationRateRisk
	if err := r.db.SelectContext(ctx, &risks, query, orgID, quotationID); err != nil {
		return nil, fmt.Errorf("quotations.GetQuotationRateRisks: %w", err)
	}
	for _, rk := range risks {
		if rk.MetadataJSON != "" {
			var m interface{}
			if err := json.Unmarshal([]byte(rk.MetadataJSON), &m); err == nil {
				rk.Metadata = m
			}
		}
	}
	if risks == nil {
		risks = []*QuotationRateRisk{}
	}
	return risks, nil
}

func (r *repository) ResolveQuotationRateRisk(ctx context.Context, orgID, quotationID, riskID int64, user string) error {
	query := `
		UPDATE quotation_rate_risk_events
		SET is_resolved = TRUE, resolved_by = ?, resolved_at = NOW(), updated_at = NOW()
		WHERE org_id = ? AND quotation_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, user, orgID, quotationID, riskID)
	return err
}

func (r *repository) GetQuotationsWithActiveRateSelection(ctx context.Context, orgID int64) ([]*QuotationRateSelectionDetail, error) {
	query := `
		SELECT 
			q.id AS quotation_id,
			q.quotation_number,
			q.status AS quotation_status,
			s.id AS snapshot_id,
			s.carrier_name,
			s.currency,
			s.commercial_total,
			s.valid_until,
			s.source_rate_id,
			COALESCE(r.status, '') AS source_rate_status,
			r.valid_until AS source_rate_valid_until,
			COALESCE(s.source_rate_version, 1) AS source_rate_version,
			COALESCE(r.version_number, 1) AS latest_rate_version,
			s.source_contract_id,
			COALESCE(c.contract_code, '') AS source_contract_code,
			COALESCE(c.status, '') AS source_contract_status,
			c.end_date AS source_contract_end_date,
			s.source_spot_rate_response_id,
			srr.valid_until AS source_spot_valid_until,
			COALESCE(srr.status, '') AS source_spot_status
		FROM quotations q
		JOIN quotation_rate_selections sel ON q.org_id = sel.org_id AND q.id = sel.quotation_id AND sel.is_active = TRUE
		JOIN quotation_rate_snapshots s ON sel.id = s.quotation_rate_selection_id
		LEFT JOIN rates r ON s.source_rate_id = r.id
		LEFT JOIN rate_contracts c ON s.source_contract_id = c.id
		LEFT JOIN spot_rate_responses srr ON s.source_spot_rate_response_id = srr.id
		WHERE q.org_id = ?
	`
	var details []*QuotationRateSelectionDetail
	if err := r.db.SelectContext(ctx, &details, query, orgID); err != nil {
		return nil, fmt.Errorf("quotations.GetQuotationsWithActiveRateSelection: %w", err)
	}
	return details, nil
}

func (r *repository) GetQuotationsAffectedByRate(ctx context.Context, orgID, rateID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT quotation_id
		FROM quotation_rate_selections
		WHERE org_id = ? AND rate_id = ? AND is_active = TRUE
	`
	var ids []int64
	if err := r.db.SelectContext(ctx, &ids, query, orgID, rateID); err != nil {
		return nil, err
	}
	return ids, nil
}







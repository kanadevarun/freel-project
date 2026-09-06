package rfq

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	CreateRFQ(ctx context.Context, rfq *spec.RFQ) error
	GetRFQByID(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error)
	GetRFQByLeadID(ctx context.Context, orgID int32, leadID int64) (*spec.RFQ, error)
	GetRFQTimeline(ctx context.Context, orgID, rfqID int32, leadID *int64) ([]spec.TimelineEvent, error)
	ListRFQs(ctx context.Context, orgID int32, limit, offset int) ([]spec.RFQ, int, error)
	UpdateStage(ctx context.Context, orgID, rfqID int32, stage string) error
	UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error
	ConvertLead(ctx context.Context, orgID int32, leadID int64) error
	
	// Items & Quotes
	CreateRFQItem(ctx context.Context, item *spec.RFQItem) error
	CreateQuote(ctx context.Context, quote *spec.Quote) error
	GetQuotesByRFQ(ctx context.Context, rfqID int32) ([]spec.Quote, error)
	// ApproveQuote sets the chosen quote to APPROVED and rejects all others for this RFQ.
	ApproveQuote(ctx context.Context, rfqID, quoteID int32) error
	CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error

	// Quotes (Task 13)
	GetRFQQuotes(ctx context.Context, orgID, rfqID int32) ([]spec.RFQQuote, error)
	GetRFQQuoteByID(ctx context.Context, orgID, rfqID int32, quoteID int64) (*spec.RFQQuote, error)
	CreateRFQQuote(ctx context.Context, orgID int32, quote *spec.RFQQuote) error
	UpdateRFQQuote(ctx context.Context, orgID int32, quote *spec.RFQQuote) error
	UpdateRFQQuoteStatus(ctx context.Context, orgID, rfqID int32, quoteID int64, status string) error
	RecommendRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error
	ApproveRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, approver string) error
	SelectRFQQuoteForCustomer(ctx context.Context, orgID, rfqID int32, quoteID int64) error
	DeleteRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error

	// Documents (Task 12)
	GetRFQDocuments(ctx context.Context, orgID, rfqID int32) ([]spec.RFQDocument, error)
	GetRFQDocumentByID(ctx context.Context, orgID, rfqID int32, documentID int64) (*spec.RFQDocument, error)
	CreateRFQDocument(ctx context.Context, orgID int32, doc *spec.RFQDocument) error
	UpdateRFQDocument(ctx context.Context, orgID int32, doc *spec.RFQDocument) error
	UpdateRFQDocumentStatus(ctx context.Context, orgID, rfqID int32, documentID int64, status, reviewer string, rejectionReason *string) error
	DeleteRFQDocument(ctx context.Context, orgID, rfqID int32, documentID int64) error
	CreateActivity(ctx context.Context, orgID int32, entityType string, entityID int64, action, description, actor string) error

	// Bookings & Shipments (Task 14)
	GetRFQBookings(ctx context.Context, orgID, rfqID int32) ([]spec.RFQBooking, error)
	GetRFQBookingByID(ctx context.Context, orgID, rfqID int32, bookingID int64) (*spec.RFQBooking, error)
	CreateRFQBooking(ctx context.Context, orgID int32, booking *spec.RFQBooking) error
	UpdateRFQBookingStatus(ctx context.Context, orgID, rfqID int32, bookingID int64, status string) error
	GetRFQShipments(ctx context.Context, orgID, rfqID int32) ([]spec.RFQShipment, error)

	// Dedicated Booking Operations Workspace (Task 15)
	GetBookingsWorkspace(ctx context.Context, orgID int32, filter spec.BookingListFilter) ([]spec.BookingWorkspaceItem, spec.BookingKPIs, int, error)
	GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*spec.BookingDetailResponse, error)
	GetBookingByIDOnly(ctx context.Context, orgID int32, bookingID int64) (*spec.RFQBooking, error)
	UpdateBookingStatusDirect(ctx context.Context, orgID int32, bookingID int64, status string) error
	UpdateCarrierBookingResult(ctx context.Context, orgID int32, bookingID int64, carrierRef, carrierStatus string, confirmationRef, carrierError *string, bookedAt *time.Time, vesselName, voyageNum *string, etd, eta *time.Time, newStatus string) error
	GetEligibleRFQsForBooking(ctx context.Context, orgID int32) ([]spec.EligibleBookingRFQ, error)
	CreateShipmentFromBookingTx(ctx context.Context, orgID int32, bookingID int64, req spec.CreateShipmentFromBookingRequest, creator string) (*spec.RFQShipment, error)
	GetUniqueCarriersForBookings(ctx context.Context, orgID int32) ([]spec.BookingCarrierInfo, error)
}



type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) CreateRFQ(ctx context.Context, rfq *spec.RFQ) error {
	query := `
		INSERT INTO rfqs (org_id, rfq_number, customer_id, stage, origin, destination, incoterms, target_date, lead_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	rfq.CreatedAt = time.Now()
	rfq.UpdatedAt = time.Now()
	if rfq.Stage == "" {
		rfq.Stage = spec.StageRFQCreated
	}

	res, err := d.db.ExecContext(ctx, query, rfq.OrgID, rfq.RFQNumber, rfq.CustomerID, rfq.Stage, rfq.Origin, rfq.Destination, rfq.Incoterms, rfq.TargetDate, rfq.LeadID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rfq.ID = int32(id)
	return nil
}

func (d *dataLayer) ConvertLead(ctx context.Context, orgID int32, leadID int64) error {
	_, err := d.db.ExecContext(ctx, `UPDATE leads SET status = 'CONVERTED', updated_at = NOW() WHERE id = ? AND org_id = ?`, leadID, orgID)
	if err != nil {
		return err
	}

	var rfqNumber string
	_ = d.db.GetContext(ctx, &rfqNumber, `SELECT rfq_number FROM rfqs WHERE lead_id = ? AND org_id = ? ORDER BY created_at DESC LIMIT 1`, leadID, orgID)

	desc := "Lead was converted to RFQ."
	if rfqNumber != "" {
		desc = fmt.Sprintf("RFQ %s was created from this Lead.", rfqNumber)
	}

	_, _ = d.db.ExecContext(ctx, `
		INSERT INTO activities (org_id, entity_type, entity_id, action, description, created_at)
		VALUES (?, 'LEAD', ?, 'CONVERTED', ?, NOW())
	`, orgID, leadID, desc)

	return nil
}

func (d *dataLayer) GetRFQByID(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error) {
	var rfq spec.RFQ
	query := `
		SELECT
			r.id,
			r.org_id,
			r.rfq_number,
			r.customer_id,
			COALESCE(NULLIF(c.name, ''), l.company_name, '') AS customer_name,
			COALESCE(c.contact_email, l.email) AS customer_email,
			COALESCE(c.contact_phone, l.phone) AS customer_phone,
			COALESCE(c.contact_name, l.contact_name) AS customer_contact_name,
			r.stage,
			r.status,
			r.origin,
			r.destination,
			r.incoterms,
			r.target_date,
			r.sales_assignee_id,
			r.pricing_assignee_id,
			r.health_score,
			r.agent_status,
			r.lead_id,
			r.created_at,
			r.updated_at
		FROM rfqs r
		LEFT JOIN customers c ON c.id = r.customer_id
		LEFT JOIN leads l ON l.id = r.lead_id
		WHERE r.id = ? AND r.org_id = ?
	`
	err := d.db.GetContext(ctx, &rfq, query, rfqID, orgID)
	if err != nil {
		return nil, err
	}
	
	// Load items
	_ = d.db.SelectContext(ctx, &rfq.Items, `SELECT * FROM rfq_items WHERE rfq_id = ? ORDER BY id ASC`, rfq.ID)
	if rfq.Items == nil {
		rfq.Items = []spec.RFQItem{}
	}
	
	// Load quotes
	_ = d.db.SelectContext(ctx, &rfq.Quotes, `SELECT * FROM rfq_quotes WHERE rfq_id = ? ORDER BY created_at ASC`, rfq.ID)
	if rfq.Quotes == nil {
		rfq.Quotes = []spec.Quote{}
	}

	// Load documents
	docs, err := d.GetRFQDocuments(ctx, orgID, rfq.ID)
	if err == nil {
		rfq.Documents = docs
	} else {
		rfq.Documents = []spec.RFQDocument{}
	}

	return &rfq, nil
}


func (d *dataLayer) GetRFQByLeadID(ctx context.Context, orgID int32, leadID int64) (*spec.RFQ, error) {
	var rfq spec.RFQ
	query := `SELECT * FROM rfqs WHERE lead_id = ? AND org_id = ? LIMIT 1`
	err := d.db.GetContext(ctx, &rfq, query, leadID, orgID)
	if err != nil {
		return nil, err
	}
	return &rfq, nil
}

func (d *dataLayer) ListRFQs(ctx context.Context, orgID int32, limit, offset int) ([]spec.RFQ, int, error) {
	var rfqs []spec.RFQ

	// We LEFT JOIN customers so that RFQs without a customer record still show up.
	// COALESCE ensures customer_name is never null in the JSON output.
	query := `
		SELECT
			r.id,
			r.org_id,
			r.rfq_number,
			r.customer_id,
			COALESCE(c.name, '') AS customer_name,
			r.stage,
			r.origin,
			r.destination,
			r.incoterms,
			r.target_date,
			r.sales_assignee_id,
			r.pricing_assignee_id,
			r.health_score,
			r.agent_status,
			r.created_at,
			r.updated_at
		FROM rfqs r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.org_id = ?
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`
	err := d.db.SelectContext(ctx, &rfqs, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	d.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM rfqs WHERE org_id = ?`, orgID)

	return rfqs, total, nil
}

func (d *dataLayer) UpdateStage(ctx context.Context, orgID, rfqID int32, stage string) error {
	query := `UPDATE rfqs SET stage = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
	res, err := d.db.ExecContext(ctx, query, stage, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error {
	query := `UPDATE rfqs SET agent_status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
	res, err := d.db.ExecContext(ctx, query, status, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) CreateRFQItem(ctx context.Context, item *spec.RFQItem) error {
	query := `
		INSERT INTO rfq_items (rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	
	res, err := d.db.ExecContext(ctx, query, item.RFQID, item.Description, item.Quantity, item.WeightKG, item.VolumeCBM)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = int32(id)
	return nil
}

func (d *dataLayer) CreateQuote(ctx context.Context, quote *spec.Quote) error {
	query := `
		INSERT INTO rfq_quotes (rfq_id, carrier_name, transit_time_days, buy_price, sell_price, is_recommended, reliability_score, historical_success_rate, ai_reasoning, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	quote.CreatedAt = time.Now()
	quote.UpdatedAt = time.Now()
	if quote.Status == "" {
		quote.Status = "DRAFT"
	}
	
	res, err := d.db.ExecContext(ctx, query,
		quote.RFQID, quote.CarrierName, quote.TransitTimeDays, quote.BuyPrice, quote.SellPrice,
		quote.IsRecommended, quote.ReliabilityScore, quote.HistoricalSuccessRate, quote.AiReasoning, quote.Status,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	quote.ID = int32(id)
	return nil
}

func (d *dataLayer) GetQuotesByRFQ(ctx context.Context, rfqID int32) ([]spec.Quote, error) {
	var quotes []spec.Quote
	query := `SELECT * FROM rfq_quotes WHERE rfq_id = ? ORDER BY created_at ASC`
	err := d.db.SelectContext(ctx, &quotes, query, rfqID)
	return quotes, err
}

// ApproveQuote atomically marks the chosen quote as APPROVED and
// marks all other quotes for the same RFQ as REJECTED.
// We do this in a transaction so the DB is never left in a partially-approved state.
func (d *dataLayer) ApproveQuote(ctx context.Context, rfqID, quoteID int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reject all other quotes for this RFQ first
	_, err = tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET status = 'REJECTED', updated_at = NOW() WHERE rfq_id = ? AND id != ?`,
		rfqID, quoteID,
	)
	if err != nil {
		return err
	}

	// Approve the selected quote
	res, err := tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET status = 'APPROVED', is_recommended = TRUE, updated_at = NOW() WHERE id = ? AND rfq_id = ?`,
		quoteID, rfqID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d", quoteID, rfqID)
	}

	return tx.Commit()
}

func (d *dataLayer) GetRFQQuotes(ctx context.Context, orgID, rfqID int32) ([]spec.RFQQuote, error) {
	quotes := make([]spec.RFQQuote, 0)
	query := `
		SELECT q.id, q.rfq_id, r.org_id, q.carrier_id, q.carrier_name, q.quote_reference,
		       COALESCE(q.currency, 'USD') as currency, q.buy_price, q.sell_price,
		       COALESCE(q.ocean_freight, 0) as ocean_freight,
		       COALESCE(q.origin_charges, 0) as origin_charges,
		       COALESCE(q.destination_charges, 0) as destination_charges,
		       COALESCE(q.total_buy_price, q.buy_price) as total_buy_price,
		       q.transit_time_days, q.free_days, q.valid_from, q.valid_until,
		       q.etd, q.eta, q.notes, q.approved_by, q.approved_at,
		       COALESCE(q.charges, '{}') as charges, q.is_recommended,
		       COALESCE(q.reliability_score, 0) as reliability_score,
		       COALESCE(q.historical_success_rate, 0) as historical_success_rate,
		       q.ai_reasoning, q.status, q.created_at, q.updated_at
		FROM rfq_quotes q
		JOIN rfqs r ON q.rfq_id = r.id
		WHERE r.org_id = ? AND q.rfq_id = ?
		ORDER BY q.created_at ASC
	`
	err := d.db.SelectContext(ctx, &quotes, query, orgID, rfqID)
	if err != nil {
		return nil, err
	}
	for i := range quotes {
		quotes[i].UnmarshalCharges()
	}
	return quotes, nil
}

func (d *dataLayer) GetRFQQuoteByID(ctx context.Context, orgID, rfqID int32, quoteID int64) (*spec.RFQQuote, error) {
	var quote spec.RFQQuote
	query := `
		SELECT q.id, q.rfq_id, r.org_id, q.carrier_id, q.carrier_name, q.quote_reference,
		       COALESCE(q.currency, 'USD') as currency, q.buy_price, q.sell_price,
		       COALESCE(q.ocean_freight, 0) as ocean_freight,
		       COALESCE(q.origin_charges, 0) as origin_charges,
		       COALESCE(q.destination_charges, 0) as destination_charges,
		       COALESCE(q.total_buy_price, q.buy_price) as total_buy_price,
		       q.transit_time_days, q.free_days, q.valid_from, q.valid_until,
		       q.etd, q.eta, q.notes, q.approved_by, q.approved_at,
		       COALESCE(q.charges, '{}') as charges, q.is_recommended,
		       COALESCE(q.reliability_score, 0) as reliability_score,
		       COALESCE(q.historical_success_rate, 0) as historical_success_rate,
		       q.ai_reasoning, q.status, q.created_at, q.updated_at
		FROM rfq_quotes q
		JOIN rfqs r ON q.rfq_id = r.id
		WHERE r.org_id = ? AND q.rfq_id = ? AND q.id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &quote, query, orgID, rfqID, quoteID)
	if err != nil {
		return nil, err
	}
	quote.UnmarshalCharges()
	quote.MarginAmount, quote.MarginPercentage = CalculateQuoteMargin(quote.BuyPrice, quote.SellPrice)
	quote.ValidityStatus, quote.DaysUntilExpiry = EvaluateQuoteValidity(quote.ValidUntil, time.Now())
	return &quote, nil
}

func (d *dataLayer) CreateRFQQuote(ctx context.Context, orgID int32, quote *spec.RFQQuote) error {
	chargesJSON, _ := json.Marshal(quote.Charges)
	if len(chargesJSON) == 0 || string(chargesJSON) == "null" {
		chargesJSON = []byte("[]")
	}

	if quote.Currency == "" {
		quote.Currency = "USD"
	}
	if quote.Status == "" {
		quote.Status = spec.QuoteStatusDraft
	}

	query := `
		INSERT INTO rfq_quotes (
			rfq_id, carrier_id, carrier_name, quote_reference, currency,
			buy_price, sell_price, ocean_freight, origin_charges, destination_charges,
			total_buy_price, free_days, transit_time_days, valid_from, valid_until,
			etd, eta, notes, is_recommended, reliability_score, historical_success_rate,
			ai_reasoning, status, charges, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		quote.RFQID, quote.CarrierID, quote.CarrierName, quote.QuoteReference, quote.Currency,
		quote.BuyPrice, quote.SellPrice, quote.OceanFreight, quote.OriginCharges, quote.DestinationCharges,
		quote.TotalBuyPrice, quote.FreeDays, quote.TransitTimeDays, quote.ValidFrom, quote.ValidUntil,
		quote.ETD, quote.ETA, quote.Notes, quote.IsRecommended, quote.ReliabilityScore, quote.HistoricalSuccessRate,
		quote.AiReasoning, quote.Status, chargesJSON,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	quote.ID = id
	return nil
}

func (d *dataLayer) UpdateRFQQuote(ctx context.Context, orgID int32, quote *spec.RFQQuote) error {
	chargesJSON, _ := json.Marshal(quote.Charges)
	if len(chargesJSON) == 0 || string(chargesJSON) == "null" {
		chargesJSON = []byte("[]")
	}

	query := `
		UPDATE rfq_quotes
		SET carrier_id = ?, carrier_name = ?, quote_reference = ?, currency = ?,
		    buy_price = ?, sell_price = ?, ocean_freight = ?, origin_charges = ?,
		    destination_charges = ?, total_buy_price = ?, free_days = ?, transit_time_days = ?,
		    valid_from = ?, valid_until = ?, etd = ?, eta = ?, notes = ?,
		    status = ?, charges = ?, updated_at = NOW()
		WHERE id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)
	`
	_, err := d.db.ExecContext(ctx, query,
		quote.CarrierID, quote.CarrierName, quote.QuoteReference, quote.Currency,
		quote.BuyPrice, quote.SellPrice, quote.OceanFreight, quote.OriginCharges,
		quote.DestinationCharges, quote.TotalBuyPrice, quote.FreeDays, quote.TransitTimeDays,
		quote.ValidFrom, quote.ValidUntil, quote.ETD, quote.ETA, quote.Notes,
		quote.Status, chargesJSON, quote.ID, orgID,
	)
	return err
}

func (d *dataLayer) UpdateRFQQuoteStatus(ctx context.Context, orgID, rfqID int32, quoteID int64, status string) error {
	query := `
		UPDATE rfq_quotes
		SET status = ?, updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)
	`
	res, err := d.db.ExecContext(ctx, query, status, quoteID, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d in org %d", quoteID, rfqID, orgID)
	}
	return nil
}

func (d *dataLayer) RecommendRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear recommended flag from any existing recommended quote on this RFQ
	_, err = tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET is_recommended = FALSE, updated_at = NOW() WHERE rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)`,
		rfqID, orgID,
	)
	if err != nil {
		return err
	}

	// Mark target quote as recommended
	res, err := tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET is_recommended = TRUE, status = CASE WHEN status IN ('DRAFT', 'REQUESTED', 'RECEIVED', 'UNDER_REVIEW') THEN 'RECOMMENDED' ELSE status END, updated_at = NOW() WHERE id = ? AND rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)`,
		quoteID, rfqID, orgID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d in org %d", quoteID, rfqID, orgID)
	}

	return tx.Commit()
}

func (d *dataLayer) ApproveRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, approver string) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Mark target quote as APPROVED and set approver & timestamp
	res, err := tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET status = 'APPROVED', is_recommended = TRUE, approved_by = ?, approved_at = NOW(), updated_at = NOW() WHERE id = ? AND rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)`,
		approver, quoteID, rfqID, orgID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d in org %d", quoteID, rfqID, orgID)
	}

	// 2. Advance RFQ stage to STAGE_QUOTE_GENERATED if currently STAGE_RFQ_CREATED or STAGE_PRICING_ASSIGNED
	_, err = tx.ExecContext(ctx,
		`UPDATE rfqs SET stage = 'STAGE_QUOTE_GENERATED', updated_at = NOW() WHERE id = ? AND org_id = ? AND (stage = 'STAGE_RFQ_CREATED' OR stage = 'STAGE_PRICING_ASSIGNED' OR stage = 'DRAFT' OR stage IS NULL)`,
		rfqID, orgID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *dataLayer) SelectRFQQuoteForCustomer(ctx context.Context, orgID, rfqID int32, quoteID int64) error {
	query := `
		UPDATE rfq_quotes
		SET status = 'SELECTED_FOR_CUSTOMER', updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)
	`
	res, err := d.db.ExecContext(ctx, query, quoteID, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d in org %d", quoteID, rfqID, orgID)
	}
	return nil
}

func (d *dataLayer) DeleteRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error {
	// Soft delete / withdraw: marks status as WITHDRAWN and clears recommended flag
	query := `
		UPDATE rfq_quotes
		SET status = 'WITHDRAWN', is_recommended = FALSE, updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)
	`
	res, err := d.db.ExecContext(ctx, query, quoteID, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d in org %d", quoteID, rfqID, orgID)
	}
	return nil
}


func (d *dataLayer) CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}

	const query = `
INSERT INTO ai_processing_tasks (
    org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, 'QUEUED', NOW(), NOW()
)
`
	_, err = d.db.ExecContext(ctx, query, orgID, entityType, entityID, taskType, payloadJSON)
	return err
}

func (d *dataLayer) GetRFQTimeline(ctx context.Context, orgID, rfqID int32, leadID *int64) ([]spec.TimelineEvent, error) {
	events := make([]spec.TimelineEvent, 0)

	// 1. Fetch activities for RFQ and (if present) source Lead
	type dbActivity struct {
		ID          int64     `db:"id"`
		EntityType  string    `db:"entity_type"`
		EntityID    int64     `db:"entity_id"`
		Action      string    `db:"action"`
		Description string    `db:"description"`
		Actor       string    `db:"actor"`
		CreatedAt   time.Time `db:"created_at"`
	}

	var rawActivities []dbActivity
	var query string
	var args []interface{}

	if leadID != nil && *leadID > 0 {
		query = `
			SELECT 
				a.id,
				a.entity_type,
				a.entity_id,
				a.action,
				a.description,
				COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email, 'System') AS actor,
				a.created_at
			FROM activities a
			LEFT JOIN users u ON a.user_id = u.id
			WHERE a.org_id = ? AND (
				(a.entity_type = 'RFQ' AND a.entity_id = ?) OR
				(a.entity_type = 'LEAD' AND a.entity_id = ?) OR
				(a.entity_type = 'DOCUMENT' AND a.entity_id IN (SELECT id FROM rfq_documents WHERE rfq_id = ?)) OR
				(a.entity_type = 'QUOTE' AND a.entity_id IN (SELECT id FROM rfq_quotes WHERE rfq_id = ?)) OR
				(a.entity_type = 'BOOKING' AND a.entity_id IN (SELECT id FROM bookings WHERE rfq_id = ?)) OR
				(a.entity_type = 'SHIPMENT' AND a.entity_id IN (SELECT id FROM shipments WHERE rfq_id = ?))
			)
			ORDER BY a.created_at DESC
		`
		args = []interface{}{orgID, rfqID, *leadID, rfqID, rfqID, rfqID, rfqID}
	} else {
		query = `
			SELECT 
				a.id,
				a.entity_type,
				a.entity_id,
				a.action,
				a.description,
				COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email, 'System') AS actor,
				a.created_at
			FROM activities a
			LEFT JOIN users u ON a.user_id = u.id
			WHERE a.org_id = ? AND (
				(a.entity_type = 'RFQ' AND a.entity_id = ?) OR
				(a.entity_type = 'DOCUMENT' AND a.entity_id IN (SELECT id FROM rfq_documents WHERE rfq_id = ?)) OR
				(a.entity_type = 'QUOTE' AND a.entity_id IN (SELECT id FROM rfq_quotes WHERE rfq_id = ?)) OR
				(a.entity_type = 'BOOKING' AND a.entity_id IN (SELECT id FROM bookings WHERE rfq_id = ?)) OR
				(a.entity_type = 'SHIPMENT' AND a.entity_id IN (SELECT id FROM shipments WHERE rfq_id = ?))
			)
			ORDER BY a.created_at DESC
		`
		args = []interface{}{orgID, rfqID, rfqID, rfqID, rfqID, rfqID}
	}

	_ = d.db.SelectContext(ctx, &rawActivities, query, args...)

	for _, act := range rawActivities {
		cat := act.EntityType
		if act.Action == "EMAIL_INBOUND" || act.Action == "EMAIL_OUTBOUND" || act.Action == "EMAIL_SENT" {
			cat = "EMAIL"
		} else if act.Action == "AI_ENRICHED" || act.Action == "AI_PARSED" || act.Action == "AI_EXTRACTED" {
			cat = "AI"
		} else if act.Action == "QUOTE_GENERATED" || act.Action == "QUOTE_APPROVED" {
			cat = "QUOTE"
		}

		events = append(events, spec.TimelineEvent{
			ID:          fmt.Sprintf("act-%d", act.ID),
			EntityType:  act.EntityType,
			EntityID:    act.EntityID,
			Category:    cat,
			Action:      act.Action,
			Description: act.Description,
			Actor:       act.Actor,
			Timestamp:   act.CreatedAt,
		})
	}

	// 2. If Lead exists, also include interactions from lead_interactions
	if leadID != nil && *leadID > 0 {
		type dbInter struct {
			ID        int64     `db:"id"`
			Channel   string    `db:"channel"`
			Direction string    `db:"direction"`
			Subject   *string   `db:"subject"`
			Content   string    `db:"content"`
			Sender    *string   `db:"sender"`
			CreatedAt time.Time `db:"created_at"`
		}
		var rawInteractions []dbInter
		interQuery := `
			SELECT id, channel, direction, subject, content, sender, created_at
			FROM lead_interactions
			WHERE org_id = ? AND lead_id = ? AND status != 'IGNORED'
			ORDER BY created_at DESC
		`
		_ = d.db.SelectContext(ctx, &rawInteractions, interQuery, orgID, *leadID)

		for _, inter := range rawInteractions {
			subj := "Customer Inquiry"
			if inter.Subject != nil && *inter.Subject != "" {
				subj = *inter.Subject
			}
			actor := "Customer"
			if inter.Direction == "OUTBOUND" {
				actor = "Freight Forwarder"
			}
			if inter.Sender != nil && *inter.Sender != "" {
				actor = *inter.Sender
			}

			events = append(events, spec.TimelineEvent{
				ID:          fmt.Sprintf("inter-%d", inter.ID),
				EntityType:  "LEAD",
				EntityID:    *leadID,
				Category:    "EMAIL",
				Action:      fmt.Sprintf("EMAIL_%s", inter.Direction),
				Description: fmt.Sprintf("%s: %s", inter.Direction, subj),
				Actor:       actor,
				Timestamp:   inter.CreatedAt,
				Metadata: map[string]interface{}{
					"interaction_id": inter.ID,
					"direction":      inter.Direction,
					"channel":        inter.Channel,
				},
			})
		}
	}

	return events, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Document Management Datalayer Methods (Task 12)
// ──────────────────────────────────────────────────────────────────────────────

func (d *dataLayer) GetRFQDocuments(ctx context.Context, orgID, rfqID int32) ([]spec.RFQDocument, error) {
	query := `
		SELECT 
			id, org_id, rfq_id, document_type, document_name, description, status,
			file_name, file_url, file_size, mime_type, uploaded_by, uploaded_at,
			reviewed_by, reviewed_at, rejection_reason, expires_at,
			COALESCE(metadata, '{}') AS metadata, created_at, updated_at
		FROM rfq_documents
		WHERE rfq_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	var docs []spec.RFQDocument
	err := d.db.SelectContext(ctx, &docs, query, rfqID, orgID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []spec.RFQDocument{}
	}
	for i := range docs {
		docs[i].UnmarshalMetadata()
	}
	return docs, nil
}

func (d *dataLayer) GetRFQDocumentByID(ctx context.Context, orgID, rfqID int32, documentID int64) (*spec.RFQDocument, error) {
	query := `
		SELECT 
			id, org_id, rfq_id, document_type, document_name, description, status,
			file_name, file_url, file_size, mime_type, uploaded_by, uploaded_at,
			reviewed_by, reviewed_at, rejection_reason, expires_at,
			COALESCE(metadata, '{}') AS metadata, created_at, updated_at
		FROM rfq_documents
		WHERE id = ? AND rfq_id = ? AND org_id = ?
		LIMIT 1
	`
	var doc spec.RFQDocument
	err := d.db.GetContext(ctx, &doc, query, documentID, rfqID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	doc.UnmarshalMetadata()
	return &doc, nil
}

func (d *dataLayer) CreateRFQDocument(ctx context.Context, orgID int32, doc *spec.RFQDocument) error {
	query := `
		INSERT INTO rfq_documents (
			org_id, rfq_id, document_type, document_name, description, status,
			file_name, file_url, file_size, mime_type, uploaded_by, uploaded_at,
			reviewed_by, reviewed_at, rejection_reason, expires_at, metadata,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			NOW(), NOW()
		)
	`
	var metaJSON []byte
	if doc.Metadata != nil {
		metaJSON, _ = json.Marshal(doc.Metadata)
	}

	res, err := d.db.ExecContext(ctx, query,
		orgID, doc.RFQID, doc.DocumentType, doc.DocumentName, doc.Description, doc.Status,
		doc.FileName, doc.FileURL, doc.FileSize, doc.MimeType, doc.UploadedBy, doc.UploadedAt,
		doc.ReviewedBy, doc.ReviewedAt, doc.RejectionReason, doc.ExpiresAt, metaJSON,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	doc.ID = id
	doc.OrgID = orgID
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	return nil
}

func (d *dataLayer) UpdateRFQDocument(ctx context.Context, orgID int32, doc *spec.RFQDocument) error {
	query := `
		UPDATE rfq_documents SET
			document_type = ?,
			document_name = ?,
			description = ?,
			status = ?,
			file_name = ?,
			file_url = ?,
			file_size = ?,
			mime_type = ?,
			uploaded_by = ?,
			uploaded_at = ?,
			reviewed_by = ?,
			reviewed_at = ?,
			rejection_reason = ?,
			expires_at = ?,
			metadata = ?,
			updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND org_id = ?
	`
	var metaJSON []byte
	if doc.Metadata != nil {
		metaJSON, _ = json.Marshal(doc.Metadata)
	}

	_, err := d.db.ExecContext(ctx, query,
		doc.DocumentType, doc.DocumentName, doc.Description, doc.Status,
		doc.FileName, doc.FileURL, doc.FileSize, doc.MimeType, doc.UploadedBy, doc.UploadedAt,
		doc.ReviewedBy, doc.ReviewedAt, doc.RejectionReason, doc.ExpiresAt, metaJSON,
		doc.ID, doc.RFQID, orgID,
	)
	return err
}

func (d *dataLayer) UpdateRFQDocumentStatus(ctx context.Context, orgID int32, rfqID int32, documentID int64, status, reviewer string, rejectionReason *string) error {
	query := `
		UPDATE rfq_documents SET
			status = ?,
			reviewed_by = CASE WHEN ? != '' THEN ? ELSE reviewed_by END,
			reviewed_at = CASE WHEN ? IN ('APPROVED', 'REJECTED', 'UNDER_REVIEW') THEN NOW() ELSE reviewed_at END,
			rejection_reason = ?,
			updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND org_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, reviewer, reviewer, status, rejectionReason, documentID, rfqID, orgID)
	return err
}

func (d *dataLayer) DeleteRFQDocument(ctx context.Context, orgID int32, rfqID int32, documentID int64) error {
	query := `DELETE FROM rfq_documents WHERE id = ? AND rfq_id = ? AND org_id = ?`
	_, err := d.db.ExecContext(ctx, query, documentID, rfqID, orgID)
	return err
}

func (d *dataLayer) CreateActivity(ctx context.Context, orgID int32, entityType string, entityID int64, action, description, actor string) error {
	query := `
		INSERT INTO activities (org_id, entity_type, entity_id, action, description, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	_, err := d.db.ExecContext(ctx, query, orgID, entityType, entityID, action, description)
	return err
}

func (d *dataLayer) GetRFQBookings(ctx context.Context, orgID, rfqID int32) ([]spec.RFQBooking, error) {
	var bookings []spec.RFQBooking
	query := `
		SELECT id, org_id, rfq_id, quote_id, booking_number, carrier_id, carrier_name, carrier_scac,
		       carrier_booking_reference, carrier_booking_status, carrier_confirmation_reference,
		       carrier_booking_error, carrier_booked_at,
		       status, origin_port, destination_port, vessel_name, voyage_number, etd, eta,
		       cargo_summary, special_instructions, created_by, created_at, updated_at
		FROM bookings
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY created_at DESC
	`
	err := d.db.SelectContext(ctx, &bookings, query, orgID, rfqID)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (d *dataLayer) GetRFQBookingByID(ctx context.Context, orgID, rfqID int32, bookingID int64) (*spec.RFQBooking, error) {
	var booking spec.RFQBooking
	query := `
		SELECT id, org_id, rfq_id, quote_id, booking_number, carrier_id, carrier_name, carrier_scac,
		       carrier_booking_reference, carrier_booking_status, carrier_confirmation_reference,
		       carrier_booking_error, carrier_booked_at,
		       status, origin_port, destination_port, vessel_name, voyage_number, etd, eta,
		       cargo_summary, special_instructions, created_by, created_at, updated_at
		FROM bookings
		WHERE id = ? AND org_id = ? AND rfq_id = ?
	`
	err := d.db.GetContext(ctx, &booking, query, bookingID, orgID, rfqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}

func (d *dataLayer) CreateRFQBooking(ctx context.Context, orgID int32, booking *spec.RFQBooking) error {
	booking.OrgID = int64(orgID)
	if booking.Status == "" {
		booking.Status = spec.BookingStatusDraft
	}
	query := `
		INSERT INTO bookings (
			org_id, rfq_id, quote_id, booking_number, carrier_id, carrier_name, carrier_scac,
			status, origin_port, destination_port, vessel_name, voyage_number, etd, eta,
			cargo_summary, special_instructions, created_by, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		booking.OrgID, booking.RFQID, booking.QuoteID, booking.BookingNumber, booking.CarrierID, booking.CarrierName, booking.CarrierSCAC,
		booking.Status, booking.OriginPort, booking.DestinationPort, booking.VesselName, booking.VoyageNumber, booking.ETD, booking.ETA,
		booking.CargoSummary, booking.SpecialInstructions, booking.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	booking.ID = id
	return nil
}

func (d *dataLayer) UpdateRFQBookingStatus(ctx context.Context, orgID, rfqID int32, bookingID int64, status string) error {
	query := `
		UPDATE bookings
		SET status = ?, updated_at = NOW()
		WHERE id = ? AND rfq_id = ? AND org_id = ?
	`
	res, err := d.db.ExecContext(ctx, query, status, bookingID, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("booking %d not found for rfq %d in org %d", bookingID, rfqID, orgID)
	}
	return nil
}

func (d *dataLayer) GetRFQShipments(ctx context.Context, orgID, rfqID int32) ([]spec.RFQShipment, error) {
	type dbShipment struct {
		ID                    int64      `db:"id"`
		OrgID                 int64      `db:"org_id"`
		RFQID                 *int64     `db:"rfq_id"`
		QuoteID               *int64     `db:"quote_id"`
		BookingID             *int64     `db:"booking_id"`
		BookingNumber         *string    `db:"booking_number"`
		CarrierSCAC           string     `db:"carrier_scac"`
		Status                string     `db:"status"`
		OriginPort            string     `db:"origin_port"`
		DestinationPort       string     `db:"destination_port"`
		VesselName            *string    `db:"vessel_name"`
		VoyageNumber          *string    `db:"voyage_number"`
		ContainerNumbers      *string    `db:"container_numbers"`
		ETD                   *time.Time `db:"etd"`
		ETA                   *time.Time `db:"eta"`
		CreatedAt             time.Time  `db:"created_at"`
		UpdatedAt             time.Time  `db:"updated_at"`
		ActiveExceptionsCount int64      `db:"active_exceptions_count"`
	}

	var rawShipments []dbShipment
	// Look up by direct rfq_id OR via linked booking
	query := `
		SELECT s.id, s.org_id, s.rfq_id, s.quote_id, s.booking_id, s.booking_number, s.carrier_scac,
		       s.status, s.origin_port, s.destination_port, s.vessel_name, s.voyage_number,
		       CAST(s.container_numbers AS CHAR) AS container_numbers,
		       s.etd, s.eta, s.created_at, s.updated_at,
		       (SELECT COUNT(*) FROM shipment_exceptions WHERE shipment_id = s.id AND status NOT IN ('RESOLVED', 'DISMISSED')) AS active_exceptions_count
		FROM shipments s
		WHERE s.org_id = ? AND (
			s.rfq_id = ? OR 
			s.booking_id IN (SELECT id FROM bookings WHERE rfq_id = ? AND org_id = ?)
		)
		ORDER BY s.created_at DESC
	`
	err := d.db.SelectContext(ctx, &rawShipments, query, orgID, rfqID, rfqID, orgID)
	if err != nil {
		return nil, err
	}

	carrierMap := map[string]string{
		"MAEU": "Maersk Line",
		"HLCU": "Hapag-Lloyd",
		"MSCU": "MSC",
		"CMDU": "CMA CGM",
		"EGLV": "Evergreen Line",
		"COSU": "Cosco Shipping",
		"ONEY": "Ocean Network Express",
		"YMLU": "Yang Ming",
	}

	result := make([]spec.RFQShipment, 0, len(rawShipments))
	for _, s := range rawShipments {
		var containers []string
		if s.ContainerNumbers != nil && *s.ContainerNumbers != "" {
			_ = json.Unmarshal([]byte(*s.ContainerNumbers), &containers)
		}
		cName := carrierMap[s.CarrierSCAC]
		if cName == "" {
			cName = s.CarrierSCAC
		}

		// Milestone description
		var milestone *string
		switch s.Status {
		case "BOOKED":
			m := "Carrier Booking Confirmed"
			milestone = &m
		case "DEPARTED":
			m := "Vessel Departed Origin Port"
			milestone = &m
		case "IN_TRANSIT":
			m := "Vessel In Transit (Ocean Leg)"
			milestone = &m
		case "ARRIVED":
			m := "Vessel Arrived at Destination Port"
			milestone = &m
		case "DELIVERED":
			m := "Cargo Delivered to Consignee"
			milestone = &m
		default:
			m := "Operational Tracking Active"
			milestone = &m
		}

		result = append(result, spec.RFQShipment{
			ID:               s.ID,
			OrgID:            s.OrgID,
			RFQID:            s.RFQID,
			QuoteID:          s.QuoteID,
			BookingID:        s.BookingID,
			BookingNumber:    s.BookingNumber,
			CarrierSCAC:      s.CarrierSCAC,
			CarrierName:      cName,
			Status:           s.Status,
			OriginPort:       s.OriginPort,
			DestinationPort:  s.DestinationPort,
			VesselName:       s.VesselName,
			VoyageNumber:     s.VoyageNumber,
			ContainerNumbers: containers,
			ETD:              s.ETD,
			ETA:              s.ETA,
			CurrentMilestone: milestone,
			CreatedAt:        s.CreatedAt,
			UpdatedAt:        s.UpdatedAt,
			ActiveExceptionsCount: s.ActiveExceptionsCount,
		})
	}

	return result, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 15: Dedicated Booking Workspace Implementations
// ──────────────────────────────────────────────────────────────────────────────

func (d *dataLayer) GetBookingsWorkspace(ctx context.Context, orgID int32, filter spec.BookingListFilter) ([]spec.BookingWorkspaceItem, spec.BookingKPIs, int, error) {
	// 1. Calculate live KPIs for the organization
	var kpis spec.BookingKPIs
	kpiQuery := `
		SELECT 
			COUNT(*) AS total_bookings,
			COUNT(CASE WHEN status = 'DRAFT' THEN 1 END) AS draft,
			COUNT(CASE WHEN status = 'REQUESTED' THEN 1 END) AS requested,
			COUNT(CASE WHEN status = 'PENDING_CONFIRMATION' THEN 1 END) AS pending_confirmation,
			COUNT(CASE WHEN status = 'CONFIRMED' THEN 1 END) AS confirmed,
			COUNT(CASE WHEN status = 'CANCELLED' THEN 1 END) AS cancelled,
			COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) AS completed,
			COUNT(CASE WHEN status = 'CONFIRMED' AND etd IS NOT NULL AND etd >= NOW() AND etd <= DATE_ADD(NOW(), INTERVAL 7 DAY) THEN 1 END) AS departing_soon
		FROM bookings
		WHERE org_id = ?
	`
	_ = d.db.GetContext(ctx, &kpis, kpiQuery, orgID)

	// 2. Build filtered list query
	baseQuery := `
		FROM bookings b
		LEFT JOIN rfqs r ON b.rfq_id = r.id AND r.org_id = b.org_id
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN rfq_quotes q ON b.quote_id = q.id
		LEFT JOIN shipments s ON s.booking_id = b.id AND s.org_id = b.org_id AND s.status NOT IN ('CANCELLED')
		WHERE b.org_id = ?
	`
	args := []interface{}{orgID}

	if filter.Status != nil && *filter.Status != "" && *filter.Status != "ALL" {
		if *filter.Status == "PENDING_ACTION" {
			baseQuery += " AND b.status IN ('REQUESTED', 'PENDING_CONFIRMATION')"
		} else {
			baseQuery += " AND b.status = ?"
			args = append(args, *filter.Status)
		}
	}
	if filter.Carrier != nil && *filter.Carrier != "" {
		baseQuery += " AND (b.carrier_name LIKE ? OR b.carrier_scac LIKE ?)"
		carrierPattern := "%" + *filter.Carrier + "%"
		args = append(args, carrierPattern, carrierPattern)
	}
	if filter.OriginPort != nil && *filter.OriginPort != "" {
		baseQuery += " AND b.origin_port = ?"
		args = append(args, *filter.OriginPort)
	}
	if filter.DestinationPort != nil && *filter.DestinationPort != "" {
		baseQuery += " AND b.destination_port = ?"
		args = append(args, *filter.DestinationPort)
	}
	if filter.ETDFrom != nil {
		baseQuery += " AND b.etd >= ?"
		args = append(args, *filter.ETDFrom)
	}
	if filter.ETDTo != nil {
		baseQuery += " AND b.etd <= ?"
		args = append(args, *filter.ETDTo)
	}
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		baseQuery += ` AND (
			b.booking_number LIKE ? OR
			r.rfq_number LIKE ? OR
			b.carrier_name LIKE ? OR
			c.name LIKE ? OR
			b.origin_port LIKE ? OR
			b.destination_port LIKE ? OR
			b.vessel_name LIKE ? OR
			b.voyage_number LIKE ?
		)`
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Count total filtered items
	countQuery := "SELECT COUNT(DISTINCT b.id) " + baseQuery
	var totalItems int
	err := d.db.GetContext(ctx, &totalItems, countQuery, args...)
	if err != nil {
		return nil, kpis, 0, err
	}

	// Sorting & Pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	orderClause := " ORDER BY b.created_at DESC "
	if filter.SortBy != "" {
		dir := "DESC"
		if strings.ToUpper(filter.SortDir) == "ASC" {
			dir = "ASC"
		}
		switch filter.SortBy {
		case "booking_number":
			orderClause = fmt.Sprintf(" ORDER BY b.booking_number %s ", dir)
		case "carrier_name":
			orderClause = fmt.Sprintf(" ORDER BY b.carrier_name %s ", dir)
		case "status":
			orderClause = fmt.Sprintf(" ORDER BY b.status %s ", dir)
		case "etd":
			orderClause = fmt.Sprintf(" ORDER BY b.etd %s ", dir)
		case "created_at":
			orderClause = fmt.Sprintf(" ORDER BY b.created_at %s ", dir)
		}
	}

	selectQuery := `
		SELECT 
			b.id, b.org_id, b.rfq_id, COALESCE(r.rfq_number, CONCAT('RFQ-', b.rfq_id)) AS rfq_number,
			r.customer_id, COALESCE(c.name, 'Unknown Customer') AS customer_name,
			b.quote_id, q.quote_reference, q.sell_price AS quote_sell_price,
			COALESCE(q.currency, 'USD') AS currency,
			b.booking_number, b.carrier_id, b.carrier_name, b.carrier_scac,
			b.carrier_booking_reference, b.carrier_booking_status, b.carrier_confirmation_reference,
			b.carrier_booking_error, b.carrier_booked_at,
			b.status, b.origin_port, b.destination_port, b.vessel_name, b.voyage_number,
			b.etd, b.eta, b.cargo_summary, b.special_instructions, b.created_by,
			s.id AS shipment_id, s.status AS shipment_status,
			b.created_at, b.updated_at
	` + baseQuery + orderClause + " LIMIT ? OFFSET ?"

	args = append(args, limit, offset)

	var items []spec.BookingWorkspaceItem
	err = d.db.SelectContext(ctx, &items, selectQuery, args...)
	if err != nil {
		return nil, kpis, totalItems, err
	}
	if items == nil {
		items = []spec.BookingWorkspaceItem{}
	}

	return items, kpis, totalItems, nil
}

func (d *dataLayer) GetUniqueCarriersForBookings(ctx context.Context, orgID int32) ([]spec.BookingCarrierInfo, error) {
	query := `
		SELECT DISTINCT carrier_name, COALESCE(carrier_scac, '') AS carrier_scac
		FROM bookings
		WHERE org_id = ? AND carrier_name IS NOT NULL AND carrier_name != ''
		ORDER BY carrier_name ASC
	`
	var carriers []spec.BookingCarrierInfo
	err := d.db.SelectContext(ctx, &carriers, query, orgID)
	if err != nil {
		return nil, err
	}
	if carriers == nil {
		carriers = []spec.BookingCarrierInfo{}
	}
	return carriers, nil
}

func (d *dataLayer) GetBookingByIDOnly(ctx context.Context, orgID int32, bookingID int64) (*spec.RFQBooking, error) {
	var booking spec.RFQBooking
	query := `
		SELECT id, org_id, rfq_id, quote_id, booking_number, carrier_id, carrier_name, carrier_scac,
		       carrier_booking_reference, carrier_booking_status, carrier_confirmation_reference,
		       carrier_booking_error, carrier_booked_at,
		       status, origin_port, destination_port, vessel_name, voyage_number, etd, eta,
		       cargo_summary, special_instructions, created_by, created_at, updated_at
		FROM bookings
		WHERE id = ? AND org_id = ?
	`
	err := d.db.GetContext(ctx, &booking, query, bookingID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}

func (d *dataLayer) UpdateBookingStatusDirect(ctx context.Context, orgID int32, bookingID int64, status string) error {
	query := `
		UPDATE bookings
		SET status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	res, err := d.db.ExecContext(ctx, query, status, bookingID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("booking %d not found in org %d", bookingID, orgID)
	}
	return nil
}

func (d *dataLayer) UpdateCarrierBookingResult(ctx context.Context, orgID int32, bookingID int64, carrierRef, carrierStatus string, confirmationRef, carrierError *string, bookedAt *time.Time, vesselName, voyageNum *string, etd, eta *time.Time, newStatus string) error {
	query := `
		UPDATE bookings
		SET carrier_booking_reference = ?,
		    carrier_booking_status = ?,
		    carrier_confirmation_reference = ?,
		    carrier_booking_error = ?,
		    carrier_booked_at = ?,
		    vessel_name = COALESCE(?, vessel_name),
		    voyage_number = COALESCE(?, voyage_number),
		    etd = COALESCE(?, etd),
		    eta = COALESCE(?, eta),
		    status = ?,
		    updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	res, err := d.db.ExecContext(ctx, query,
		carrierRef, carrierStatus, confirmationRef, carrierError, bookedAt,
		vesselName, voyageNum, etd, eta,
		newStatus, bookingID, orgID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("booking %d not found in org %d", bookingID, orgID)
	}
	return nil
}

func (d *dataLayer) GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*spec.BookingDetailResponse, error) {
	// 1. Fetch Booking
	booking, err := d.GetBookingByIDOnly(ctx, orgID, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, nil
	}

	// 2. Fetch Source RFQ
	var sourceRFQ spec.BookingDetailSourceRFQ
	rfqQuery := `
		SELECT r.id, r.rfq_number, r.lead_id, r.customer_id, COALESCE(c.name, 'Unknown Customer') AS customer_name,
		       COALESCE(r.origin, '') AS origin_port, COALESCE(r.destination, '') AS destination_port, r.status, r.stage, r.created_at
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE r.id = ? AND r.org_id = ?
	`
	err = d.db.GetContext(ctx, &sourceRFQ, rfqQuery, booking.RFQID, orgID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 3. Fetch Commercial Quote if available
	var commQuote *spec.BookingDetailCommercialQuote
	if booking.QuoteID != nil && *booking.QuoteID > 0 {
		var q struct {
			ID             int64      `db:"id"`
			CarrierName    string     `db:"carrier_name"`
			CarrierID      *string    `db:"carrier_id"`
			QuoteReference *string    `db:"quote_reference"`
			Currency       string     `db:"currency"`
			BuyPrice       float64    `db:"buy_price"`
			SellPrice      float64    `db:"sell_price"`
			Status         string     `db:"status"`
			ValidUntil     *time.Time `db:"valid_until"`
			ApprovedAt     *time.Time `db:"approved_at"`
			ApprovedBy     *string    `db:"approved_by"`
		}
		quoteQuery := `
			SELECT id, carrier_name, carrier_id, quote_reference, COALESCE(currency, 'USD') AS currency,
			       buy_price, sell_price, status,
			       valid_until, approved_at, approved_by
			FROM rfq_quotes
			WHERE id = ?
		`
		if qErr := d.db.GetContext(ctx, &q, quoteQuery, *booking.QuoteID); qErr == nil {
			marginAmt := q.SellPrice - q.BuyPrice
			marginPct := 0.0
			if q.SellPrice > 0 {
				marginPct = (marginAmt / q.SellPrice) * 100
			}
			commQuote = &spec.BookingDetailCommercialQuote{
				ID:             q.ID,
				CarrierName:    q.CarrierName,
				CarrierSCAC:    q.CarrierID,
				QuoteReference: q.QuoteReference,
				Currency:       q.Currency,
				BuyPrice:       q.BuyPrice,
				SellPrice:      q.SellPrice,
				MarginAmount:   marginAmt,
				MarginPercent:  marginPct,
				Status:         q.Status,
				ValidUntil:     q.ValidUntil,
				ApprovedAt:     q.ApprovedAt,
				ApprovedBy:     q.ApprovedBy,
			}
		}
	}

	// 4. Fetch Cargo Summary from rfq_items
	var cargoSummary spec.BookingDetailCargoSummary
	cargoQuery := `
		SELECT 
			COALESCE(COUNT(*), 0) AS items_count,
			COALESCE(SUM(weight_kg), 0) AS total_weight_kg,
			COALESCE(SUM(volume_cbm), 0) AS total_volume_cbm,
			COALESCE(MAX(description), 'General Cargo') AS commodity
		FROM rfq_items
		WHERE rfq_id = ?
	`
	var cRow struct {
		ItemsCount     int     `db:"items_count"`
		TotalWeightKg  float64 `db:"total_weight_kg"`
		TotalVolumeCbm float64 `db:"total_volume_cbm"`
		Commodity      string  `db:"commodity"`
	}
	if cErr := d.db.GetContext(ctx, &cRow, cargoQuery, booking.RFQID); cErr == nil {
		cargoSummary = spec.BookingDetailCargoSummary{
			ItemsCount:     cRow.ItemsCount,
			TotalWeightKg:  cRow.TotalWeightKg,
			TotalVolumeCbm: cRow.TotalVolumeCbm,
			CargoType:      "FCL Ocean Freight",
			Commodity:      cRow.Commodity,
			PackagingType:  "Standard Export Packaging",
		}
	}

	// 5. Fetch Linked Shipment if any
	shipments, _ := d.GetRFQShipments(ctx, orgID, int32(booking.RFQID))
	var linkedShipment *spec.RFQShipment
	for _, s := range shipments {
		if s.BookingID != nil && *s.BookingID == booking.ID && s.Status != "CANCELLED" {
			sCopy := s
			linkedShipment = &sCopy
			break
		}
	}

	// 6. Fetch Activity Events for this booking + RFQ
	var rawActs []struct {
		ID          int64     `db:"id"`
		EntityType  string    `db:"entity_type"`
		EntityID    int64     `db:"entity_id"`
		Action      string    `db:"action"`
		Description string    `db:"description"`
		Actor       string    `db:"actor"`
		CreatedAt   time.Time `db:"created_at"`
	}
	actQuery := `
		SELECT id, entity_type, entity_id, action, description, actor, created_at
		FROM activities
		WHERE org_id = ? AND (
			(entity_type = 'BOOKING' AND entity_id = ?) OR
			(entity_type = 'RFQ' AND entity_id = ?) OR
			(entity_type = 'SHIPMENT' AND entity_id IN (SELECT id FROM shipments WHERE booking_id = ?))
		)
		ORDER BY created_at DESC
		LIMIT 20
	`
	_ = d.db.SelectContext(ctx, &rawActs, actQuery, orgID, booking.ID, booking.RFQID, booking.ID)
	actEvents := make([]spec.ActivityEvent, 0, len(rawActs))
	for _, a := range rawActs {
		actEvents = append(actEvents, spec.ActivityEvent{
			ID:          fmt.Sprintf("%d", a.ID),
			Type:        spec.ActivityEventType(a.Action),
			Category:    "OPERATIONS",
			Title:       strings.ReplaceAll(a.Action, "_", " "),
			Description: a.Description,
			Timestamp:   a.CreatedAt,
			ActorType:   "USER",
			ActorName:   a.Actor,
			SourceType:  "BOOKING",
			SourceID:    fmt.Sprintf("%d", booking.ID),
		})
	}

	// 7. Allowed contextual actions
	allowedActions := []string{}
	switch booking.Status {
	case spec.BookingStatusDraft:
		allowedActions = []string{"REQUEST_BOOKING", "CANCEL"}
	case spec.BookingStatusRequested:
		allowedActions = []string{"MARK_PENDING_CONFIRMATION", "CONFIRM_SPACE", "CANCEL"}
	case spec.BookingStatusPendingConfirmation:
		allowedActions = []string{"CONFIRM_SPACE", "CANCEL"}
	case spec.BookingStatusConfirmed:
		if linkedShipment == nil {
			allowedActions = []string{"CREATE_SHIPMENT"}
		} else {
			allowedActions = []string{"VIEW_SHIPMENT"}
		}
	}

	return &spec.BookingDetailResponse{
		Booking:         *booking,
		SourceRFQ:       sourceRFQ,
		CommercialQuote: commQuote,
		CargoSummary:    cargoSummary,
		LinkedShipment:  linkedShipment,
		ActivityEvents:  actEvents,
		AllowedActions:  allowedActions,
	}, nil
}

func (d *dataLayer) GetEligibleRFQsForBooking(ctx context.Context, orgID int32) ([]spec.EligibleBookingRFQ, error) {
	query := `
		SELECT 
			r.id AS rfq_id, COALESCE(r.rfq_number, CONCAT('RFQ-', r.id)) AS rfq_number,
			r.customer_id, COALESCE(c.name, 'Unknown Customer') AS customer_name,
			COALESCE(r.origin, '') AS origin_port, COALESCE(r.destination, '') AS destination_port, r.target_date,
			q.id AS approved_quote_id, q.quote_reference, q.carrier_name, q.carrier_id AS carrier_scac,
			COALESCE(q.currency, 'USD') AS currency, q.sell_price, q.buy_price, q.transit_time_days,
			COALESCE((SELECT description FROM rfq_items WHERE rfq_id = r.id LIMIT 1), 'General Cargo') AS cargo_description,
			COALESCE((SELECT SUM(quantity) FROM rfq_items WHERE rfq_id = r.id), 1) AS package_count,
			COALESCE((SELECT SUM(weight_kg) FROM rfq_items WHERE rfq_id = r.id), 0) AS total_weight_kg,
			COALESCE((SELECT SUM(volume_cbm) FROM rfq_items WHERE rfq_id = r.id), 0) AS total_volume_cbm
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		JOIN rfq_quotes q ON q.rfq_id = r.id AND q.status IN ('APPROVED', 'SELECTED_FOR_CUSTOMER')
		WHERE r.org_id = ?
		  AND r.id NOT IN (
			  SELECT rfq_id FROM bookings WHERE org_id = r.org_id AND status IN ('CONFIRMED', 'COMPLETED')
		  )
		ORDER BY r.created_at DESC
		LIMIT 50
	`
	var items []spec.EligibleBookingRFQ
	err := d.db.SelectContext(ctx, &items, query, orgID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []spec.EligibleBookingRFQ{}
	}
	return items, nil
}

func (d *dataLayer) CreateShipmentFromBookingTx(ctx context.Context, orgID int32, bookingID int64, req spec.CreateShipmentFromBookingRequest, creator string) (*spec.RFQShipment, error) {
	// 1. Fetch booking
	booking, err := d.GetBookingByIDOnly(ctx, orgID, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking %d not found in org %d", bookingID, orgID)
	}
	if booking.Status != spec.BookingStatusConfirmed {
		return nil, fmt.Errorf("booking must be in CONFIRMED status to create a shipment (current: %s)", booking.Status)
	}

	// 1.5. Deep Lineage Check: Verify that the associated RFQ exists and belongs to the caller's organization
	var rfqExists bool
	err = d.db.GetContext(ctx, &rfqExists, "SELECT EXISTS(SELECT 1 FROM rfqs WHERE id = ? AND org_id = ?)", booking.RFQID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify RFQ: %w", err)
	}
	if !rfqExists {
		return nil, fmt.Errorf("invalid lineage: RFQ %d not found or unauthorized", booking.RFQID)
	}

	// 2. Idempotency Check: if active shipment already exists for this booking, return it
	existingShipments, _ := d.GetRFQShipments(ctx, orgID, int32(booking.RFQID))
	for _, s := range existingShipments {
		if s.BookingID != nil && *s.BookingID == booking.ID && s.Status != "CANCELLED" {
			return &s, nil
		}
	}

	// 3. Resolve default vessel & voyage
	vessel := booking.VesselName
	if req.VesselName != nil && *req.VesselName != "" {
		vessel = req.VesselName
	}
	voyage := booking.VoyageNumber
	if req.VoyageNumber != nil && *req.VoyageNumber != "" {
		voyage = req.VoyageNumber
	}
	etd := booking.ETD
	if req.ETD != nil {
		etd = req.ETD
	}
	eta := booking.ETA
	if req.ETA != nil {
		eta = req.ETA
	}

	scac := "MAEU"
	if booking.CarrierSCAC != nil && *booking.CarrierSCAC != "" {
		scac = *booking.CarrierSCAC
	}

	containers := req.ContainerNumbers
	if len(containers) == 0 {
		containers = []string{"MSKU9012345"}
	}
	containersJSON, _ := json.Marshal(containers)
	containersStr := string(containersJSON)

	// 4. Insert shipment transactionally
	insertQuery := `
		INSERT INTO shipments (
			org_id, rfq_id, quote_id, booking_id, booking_number, carrier_scac,
			status, origin_port, destination_port, vessel_name, voyage_number,
			container_numbers, etd, eta, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			'BOOKED', ?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, insertQuery,
		orgID, booking.RFQID, booking.QuoteID, booking.ID, booking.BookingNumber, scac,
		booking.OriginPort, booking.DestinationPort, vessel, voyage,
		containersStr, etd, eta,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment record: %w", err)
	}

	shipmentID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Record activity
	_ = d.CreateActivity(ctx, orgID, "BOOKING", booking.ID, spec.ActionShipmentCreated,
		fmt.Sprintf("Shipment #%d created from Confirmed Booking %s", shipmentID, booking.BookingNumber),
		creator,
	)
	_ = d.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.ActionShipmentCreated,
		fmt.Sprintf("Shipment execution initiated for Booking %s (%s → %s)", booking.BookingNumber, booking.OriginPort, booking.DestinationPort),
		creator,
	)

	milestone := "Carrier Booking Confirmed"
	return &spec.RFQShipment{
		ID:               shipmentID,
		OrgID:            int64(orgID),
		RFQID:            &booking.RFQID,
		QuoteID:          booking.QuoteID,
		BookingID:        &booking.ID,
		BookingNumber:    &booking.BookingNumber,
		CarrierSCAC:      scac,
		CarrierName:      booking.CarrierName,
		Status:           "BOOKED",
		OriginPort:       booking.OriginPort,
		DestinationPort:  booking.DestinationPort,
		VesselName:       vessel,
		VoyageNumber:     voyage,
		ContainerNumbers: containers,
		ETD:              etd,
		ETA:              eta,
		CurrentMilestone: &milestone,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}




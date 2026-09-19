package recommendations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Generator evaluates deterministic business rules against real persisted data
type Generator interface {
	Generate(ctx context.Context, orgID int64, correlationID string, userID int64) (*GenerateResult, error)
	GenerateForAutomation(ctx context.Context, orgID int64, automationType string, correlationID string, userID int64, automationID int64, executionID int64) (*GenerateResult, error)
}

type generator struct {
	db   *sqlx.DB
	repo Repository
}

// NewGenerator creates a new deterministic recommendation generator
func NewGenerator(db *sqlx.DB, repo Repository) Generator {
	return &generator{db: db, repo: repo}
}

func (g *generator) Generate(ctx context.Context, orgID int64, correlationID string, userID int64) (*GenerateResult, error) {
	if correlationID == "" {
		correlationID = fmt.Sprintf("rec-gen-%d-%d", orgID, time.Now().UnixNano())
	}

	result := &GenerateResult{
		CorrelationID: correlationID,
		CalculatedAt:  time.Now(),
		RulesExecuted: []string{},
	}

	var candidates []*Recommendation

	// 1. Rule: OVERDUE_INVOICE_COLLECTIONS
	result.RulesExecuted = append(result.RulesExecuted, "OVERDUE_INVOICE_COLLECTIONS")
	c1, err := g.evaluateOverdueInvoices(ctx, orgID, correlationID)
	if err == nil && len(c1) > 0 {
		candidates = append(candidates, c1...)
	}

	// 2. Rule: UNRESOLVED_SHIPMENT_EXCEPTIONS
	result.RulesExecuted = append(result.RulesExecuted, "UNRESOLVED_SHIPMENT_EXCEPTIONS")
	c2, err := g.evaluateShipmentExceptions(ctx, orgID, correlationID)
	if err == nil && len(c2) > 0 {
		candidates = append(candidates, c2...)
	}

	// 3. Rule: RFQ_AWAITING_PRICING
	result.RulesExecuted = append(result.RulesExecuted, "RFQ_AWAITING_PRICING")
	c3, err := g.evaluateRFQs(ctx, orgID, correlationID)
	if err == nil && len(c3) > 0 {
		candidates = append(candidates, c3...)
	}

	// 4. Rule: QUOTATION_EXPIRING_SOON
	result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_EXPIRING_SOON")
	c4, err := g.evaluateQuotations(ctx, orgID, correlationID)
	if err == nil && len(c4) > 0 {
		candidates = append(candidates, c4...)
	}

	// 5. Rule: CONTRACT_EXPIRY_OR_RENEWAL
	result.RulesExecuted = append(result.RulesExecuted, "CONTRACT_EXPIRY_OR_RENEWAL")
	c5, err := g.evaluateContracts(ctx, orgID, correlationID)
	if err == nil && len(c5) > 0 {
		candidates = append(candidates, c5...)
	}

	// 6. Rule: HIGH_RISK_CUSTOMER_CREDIT
	result.RulesExecuted = append(result.RulesExecuted, "HIGH_RISK_CUSTOMER_CREDIT")
	c6, err := g.evaluateCustomerCredit(ctx, orgID, correlationID)
	if err == nil && len(c6) > 0 {
		candidates = append(candidates, c6...)
	}

	// 7. Rule: LEAD_UNATTENDED
	result.RulesExecuted = append(result.RulesExecuted, "LEAD_UNATTENDED")
	c7, err := g.evaluateLeads(ctx, orgID, correlationID)
	if err == nil && len(c7) > 0 {
		candidates = append(candidates, c7...)
	}

	// 8. Rule: CUSTOMER_INACTIVITY_CHECKIN (Phase 2 Task 2.2)
	result.RulesExecuted = append(result.RulesExecuted, "CUSTOMER_INACTIVITY_CHECKIN")
	c8, err := g.evaluateCustomerInactivity(ctx, orgID, correlationID)
	if err == nil && len(c8) > 0 {
		candidates = append(candidates, c8...)
	}

	// 9. Rule: RFQ_MISSING_INFO (Phase 2 Task 2.3)
	result.RulesExecuted = append(result.RulesExecuted, "RFQ_MISSING_INFO")
	c9, err := g.evaluateRFQMissingInfo(ctx, orgID, correlationID)
	if err == nil && len(c9) > 0 {
		candidates = append(candidates, c9...)
	}

	// 10. Rule: RFQ_RESPONSE_DEADLINE_RISK (Phase 2 Task 2.3)
	result.RulesExecuted = append(result.RulesExecuted, "RFQ_RESPONSE_DEADLINE_RISK")
	c10, err := g.evaluateRFQResponseDeadlines(ctx, orgID, correlationID)
	if err == nil && len(c10) > 0 {
		candidates = append(candidates, c10...)
	}

	// 11. Rule: QUOTATION_MARGIN_ALERT (Phase 2 Task 2.3)
	result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_MARGIN_ALERT")
	c11, err := g.evaluateQuotationMarginAndPricing(ctx, orgID, correlationID)
	if err == nil && len(c11) > 0 {
		candidates = append(candidates, c11...)
	}

	// 12. Rule: QUOTATION_PENDING_APPROVAL (Phase 2 Task 2.3)
	result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_PENDING_APPROVAL")
	c12, err := g.evaluateQuotationPendingApproval(ctx, orgID, correlationID)
	if err == nil && len(c12) > 0 {
		candidates = append(candidates, c12...)
	}

	// 13. Rule: QUOTATION_SENT_AWAITING_RESPONSE (Phase 2 Task 2.3)
	result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_SENT_AWAITING_RESPONSE")
	c13, err := g.evaluateQuotationSentAwaitingResponse(ctx, orgID, correlationID)
	if err == nil && len(c13) > 0 {
		candidates = append(candidates, c13...)
	}

	// 14. Rule: SHIPMENT_MILESTONE_DELAYED (Phase 2 Task 2.4)
	result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_MILESTONE_DELAYED")
	c14, err := g.evaluateShipmentMilestones(ctx, orgID, correlationID)
	if err == nil && len(c14) > 0 {
		candidates = append(candidates, c14...)
	}

	// 15. Rule: SHIPMENT_MISSING_OPERATIONAL_INFO (Phase 2 Task 2.4)
	result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_MISSING_OPERATIONAL_INFO")
	c15, err := g.evaluateShipmentMissingOperationalInfo(ctx, orgID, correlationID)
	if err == nil && len(c15) > 0 {
		candidates = append(candidates, c15...)
	}

	// 16. Rule: SHIPMENT_INACTIVE_TRACKING (Phase 2 Task 2.4)
	result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_INACTIVE_TRACKING")
	c16, err := g.evaluateShipmentInactiveTracking(ctx, orgID, correlationID)
	if err == nil && len(c16) > 0 {
		candidates = append(candidates, c16...)
	}

	// 17. Rule: SHIPMENT_OVERDUE_DELIVERY (Phase 2 Task 2.4)
	result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_OVERDUE_DELIVERY")
	c17, err := g.evaluateShipmentOverdueDelivery(ctx, orgID, correlationID)
	if err == nil && len(c17) > 0 {
		candidates = append(candidates, c17...)
	}

	// 18. Rule: UPCOMING_COLLECTION_PRIORITY (Phase 2 Task 2.5)
	result.RulesExecuted = append(result.RulesExecuted, "UPCOMING_COLLECTION_PRIORITY")
	c18, err := g.evaluateUpcomingCollectionPriorities(ctx, orgID, correlationID)
	if err == nil && len(c18) > 0 {
		candidates = append(candidates, c18...)
	}

	// 19. Rule: PAYMENT_RISK_SIGNALS (Phase 2 Task 2.5)
	result.RulesExecuted = append(result.RulesExecuted, "PAYMENT_RISK_SIGNALS")
	c19, err := g.evaluatePaymentRiskSignals(ctx, orgID, correlationID)
	if err == nil && len(c19) > 0 {
		candidates = append(candidates, c19...)
	}

	// 20. Rule: INVOICE_DATA_QUALITY_ISSUE (Phase 2 Task 2.5)
	result.RulesExecuted = append(result.RulesExecuted, "INVOICE_DATA_QUALITY_ISSUE")
	c20, err := g.evaluateInvoiceDataQualityIssues(ctx, orgID, correlationID)
	if err == nil && len(c20) > 0 {
		candidates = append(candidates, c20...)
	}

	// 21. Rule: DOCUMENT_EXPIRY_AND_MISSING_INFO (Phase 2 Task 2.6)
	result.RulesExecuted = append(result.RulesExecuted, "DOCUMENT_EXPIRY_AND_MISSING_INFO")
	c21, err := g.evaluateDocumentExpiryAndMissingInfo(ctx, orgID, correlationID)
	if err == nil && len(c21) > 0 {
		candidates = append(candidates, c21...)
	}

	// 22. Rule: COMPLIANCE_AND_OBLIGATION_RISKS (Phase 2 Task 2.6)
	result.RulesExecuted = append(result.RulesExecuted, "COMPLIANCE_AND_OBLIGATION_RISKS")
	c22, err := g.evaluateComplianceRequirementsAndObligations(ctx, orgID, correlationID)
	if err == nil && len(c22) > 0 {
		candidates = append(candidates, c22...)
	}

	result.TotalEvaluated = len(candidates)

	// Persist / Deduplicate candidates
	for _, cand := range candidates {
		if cand.DraftStatus == "" {
			cand.DraftStatus = "NONE"
		}
		existing, err := g.repo.GetByDedupHash(ctx, orgID, cand.DedupHash)
		if err != nil {
			continue
		}

		if existing != nil {
			// Update freshness and evidence if currently active
			if existing.Status == StatusNew || existing.Status == StatusReviewed || existing.Status == StatusAssigned || existing.Status == StatusApproved {
				_ = g.repo.UpdateFreshnessAndEvidence(ctx, existing.ID, orgID, time.Now(), cand.EvidenceJSON, cand.Confidence, cand.ConfidenceScore)
				result.UpdatedCount++
			} else {
				result.UnchangedCount++
			}
		} else {
			// Insert new recommendation
			created, err := g.repo.Create(ctx, cand)
			if err == nil && created != nil {
				result.CreatedCount++
				if created.Priority == PriorityCritical || created.Priority == PriorityHigh {
					result.TopPriorityItems = append(result.TopPriorityItems, created)
				}
			}
		}
	}

	return result, nil
}

func (g *generator) GenerateForAutomation(ctx context.Context, orgID int64, automationType string, correlationID string, userID int64, automationID int64, executionID int64) (*GenerateResult, error) {
	if correlationID == "" {
		correlationID = fmt.Sprintf("auto-exec-%d-%d", automationID, time.Now().UnixNano())
	}

	result := &GenerateResult{
		CorrelationID: correlationID,
		CalculatedAt:  time.Now(),
		RulesExecuted: []string{},
	}

	var candidates []*Recommendation

	switch automationType {
	case "DAILY_OVERDUE_INVOICE_REVIEW":
		result.RulesExecuted = append(result.RulesExecuted, "OVERDUE_INVOICE_COLLECTIONS")
		if c, err := g.evaluateOverdueInvoices(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "UPCOMING_COLLECTION_PRIORITY")
		if c, err := g.evaluateUpcomingCollectionPriorities(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "PAYMENT_RISK_SIGNALS")
		if c, err := g.evaluatePaymentRiskSignals(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "INVOICE_DATA_QUALITY_ISSUE")
		if c, err := g.evaluateInvoiceDataQualityIssues(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "HIGH_RISK_CUSTOMER_CREDIT")
		if c, err := g.evaluateCustomerCredit(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}

	case "DAILY_CONTRACT_DOCUMENT_EXPIRY_REVIEW":
		result.RulesExecuted = append(result.RulesExecuted, "CONTRACT_EXPIRY_OR_RENEWAL")
		if c, err := g.evaluateContracts(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "DOCUMENT_EXPIRY_AND_MISSING_INFO")
		if c, err := g.evaluateDocumentExpiryAndMissingInfo(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "COMPLIANCE_AND_OBLIGATION_RISKS")
		if c, err := g.evaluateComplianceRequirementsAndObligations(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}

	case "SHIPMENT_EXCEPTION_REVIEW":
		result.RulesExecuted = append(result.RulesExecuted, "UNRESOLVED_SHIPMENT_EXCEPTIONS")
		if c, err := g.evaluateShipmentExceptions(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_MILESTONE_DELAYED")
		if c, err := g.evaluateShipmentMilestones(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_MISSING_OPERATIONAL_INFO")
		if c, err := g.evaluateShipmentMissingOperationalInfo(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_INACTIVE_TRACKING")
		if c, err := g.evaluateShipmentInactiveTracking(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "SHIPMENT_OVERDUE_DELIVERY")
		if c, err := g.evaluateShipmentOverdueDelivery(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}

	case "RFQ_QUOTATION_REVIEW":
		result.RulesExecuted = append(result.RulesExecuted, "RFQ_AWAITING_PRICING")
		if c, err := g.evaluateRFQs(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_EXPIRING_SOON")
		if c, err := g.evaluateQuotations(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "RFQ_MISSING_INFO")
		if c, err := g.evaluateRFQMissingInfo(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "RFQ_RESPONSE_DEADLINE_RISK")
		if c, err := g.evaluateRFQResponseDeadlines(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_MARGIN_ALERT")
		if c, err := g.evaluateQuotationMarginAndPricing(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_PENDING_APPROVAL")
		if c, err := g.evaluateQuotationPendingApproval(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "QUOTATION_SENT_AWAITING_RESPONSE")
		if c, err := g.evaluateQuotationSentAwaitingResponse(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}

	case "CUSTOMER_FOLLOWUP_REVIEW":
		result.RulesExecuted = append(result.RulesExecuted, "CUSTOMER_INACTIVITY_CHECKIN")
		if c, err := g.evaluateCustomerInactivity(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "HIGH_RISK_CUSTOMER_CREDIT")
		if c, err := g.evaluateCustomerCredit(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}
		result.RulesExecuted = append(result.RulesExecuted, "LEAD_UNATTENDED")
		if c, err := g.evaluateLeads(ctx, orgID, correlationID); err == nil && len(c) > 0 {
			candidates = append(candidates, c...)
		}

	case "DAILY_OPERATIONAL_SUMMARY":
		fallthrough
	default:
		// Consolidates all rules across operational assistants
		resAll, err := g.Generate(ctx, orgID, correlationID, userID)
		if err != nil {
			return nil, err
		}
		// Link top priority recommendations to this automation execution
		for _, rec := range resAll.TopPriorityItems {
			_ = g.repo.UpdateExecutionLink(ctx, rec.ID, orgID, automationID, executionID)
		}
		return resAll, nil
	}

	result.TotalEvaluated = len(candidates)

	// Persist / Deduplicate candidates with automation_id and execution_id linkage
	for _, cand := range candidates {
		cand.AutomationID = &automationID
		cand.ExecutionID = &executionID
		if cand.DraftStatus == "" {
			cand.DraftStatus = "NONE"
		}
		existing, err := g.repo.GetByDedupHash(ctx, orgID, cand.DedupHash)
		if err != nil {
			continue
		}

		if existing != nil {
			if existing.Status == StatusNew || existing.Status == StatusReviewed || existing.Status == StatusAssigned || existing.Status == StatusApproved {
				_ = g.repo.UpdateFreshnessAndEvidence(ctx, existing.ID, orgID, time.Now(), cand.EvidenceJSON, cand.Confidence, cand.ConfidenceScore)
				_ = g.repo.UpdateExecutionLink(ctx, existing.ID, orgID, automationID, executionID)
				result.UpdatedCount++
			} else {
				result.UnchangedCount++
			}
		} else {
			created, err := g.repo.Create(ctx, cand)
			if err == nil && created != nil {
				result.CreatedCount++
				if created.Priority == PriorityCritical || created.Priority == PriorityHigh {
					result.TopPriorityItems = append(result.TopPriorityItems, created)
				}
			}
		}
	}

	return result, nil
}

func computeDedupHash(orgID int64, sourceType string, sourceID int64, category string, ruleApplied string) string {
	raw := fmt.Sprintf("%d:%s:%d:%s:%s", orgID, sourceType, sourceID, category, ruleApplied)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

func (g *generator) evaluateOverdueInvoices(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type invRow struct {
		ID            int64     `db:"id"`
		InvoiceNumber string    `db:"invoice_number"`
		VendorName    string    `db:"vendor_name"`
		Currency      string    `db:"currency"`
		TotalAmount   float64   `db:"total_amount"`
		Status        string    `db:"status"`
		ShipmentID    int64     `db:"shipment_id"`
		CreatedAt     time.Time `db:"created_at"`
	}

	query := `
		SELECT id, invoice_number, vendor_name, currency, total_amount, status, shipment_id, created_at
		FROM shipment_invoices
		WHERE org_id = ? AND status = 'ISSUED' AND created_at < DATE_SUB(NOW(), INTERVAL 14 DAY)
		ORDER BY total_amount DESC
		LIMIT 10
	`
	var rows []invRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation

	// 1a. Customer Receivables Overdue (Customer Follow-Up Focus)
	type custInvRow struct {
		ID                  int64          `db:"id"`
		InvoiceNumber       string         `db:"invoice_number"`
		CustomerID          int64          `db:"customer_id"`
		CustomerName        string         `db:"customer_name"`
		Currency            string         `db:"currency"`
		TotalAmount         float64        `db:"total_amount"`
		BalanceDue          float64        `db:"balance_due"`
		Status              string         `db:"status"`
		DueDate             *time.Time     `db:"due_date"`
		CreatedAt           time.Time      `db:"created_at"`
		AccountOwnerID      sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName    sql.NullString `db:"account_owner_name"`
		CustomerOverdueCount int           `db:"customer_overdue_count"`
	}

	custInvQuery := `
		SELECT ci.id, ci.invoice_number, ci.customer_id, ci.customer_name, ci.currency,
		       ci.total_amount, ci.balance_due, ci.status, ci.due_date, ci.created_at,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name,
		       (SELECT COUNT(*) FROM customer_invoices ci2 WHERE ci2.org_id = ci.org_id AND ci2.customer_id = ci.customer_id AND (ci2.status = 'Overdue' OR (ci2.due_date < CURDATE() AND ci2.balance_due > 0))) AS customer_overdue_count
		FROM customer_invoices ci
		JOIN customers c ON ci.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE ci.org_id = ? AND (ci.status = 'Overdue' OR (ci.due_date < CURDATE() AND ci.balance_due > 0))
		ORDER BY ci.balance_due DESC
		LIMIT 20
	`
	var cRows []custInvRow
	if err := g.db.SelectContext(ctx, &cRows, custInvQuery, orgID); err == nil {
		for _, row := range cRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			daysOverdue := 1
			dueStr := "Past due"
			if row.DueDate != nil {
				dueStr = row.DueDate.Format("2006-01-02")
				calcDays := int(time.Since(*row.DueDate).Hours() / 24)
				if calcDays > 0 {
					daysOverdue = calcDays
				}
			}
			var agingBand string
			switch {
			case daysOverdue <= 15:
				agingBand = "1-15 days"
			case daysOverdue <= 30:
				agingBand = "16-30 days"
			case daysOverdue <= 60:
				agingBand = "31-60 days"
			default:
				agingBand = "60+ days"
			}

			paidAmount := row.TotalAmount - row.BalanceDue
			if paidAmount < 0 {
				paidAmount = 0
			}

			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "status",
					ObservedValue:  row.Status,
					Description:    "Receivable invoice in Overdue settlement state",
				},
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "days_overdue",
					ObservedValue:  fmt.Sprintf("%d days (%s)", daysOverdue, agingBand),
					Description:    "Aging duration past contractual credit maturity window",
				},
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "balance_due",
					ObservedValue:  fmt.Sprintf("%s %.2f (Total: %.2f, Paid: %.2f)", row.Currency, row.BalanceDue, row.TotalAmount, paidAmount),
					Description:    "Outstanding balance awaiting accounts receivable settlement",
				},
				{
					SourceModule:   "customers",
					SourceEntityID: row.CustomerID,
					SourceRef:      row.CustomerName,
					FieldName:      "customer_id",
					ObservedValue:  row.CustomerID,
					Description:    "Debtor account associated with invoice",
				},
			}

			if row.CustomerOverdueCount > 1 {
				evidence = append(evidence, EvidenceItem{
					SourceModule:   "customers",
					SourceEntityID: row.CustomerID,
					SourceRef:      row.CustomerName,
					FieldName:      "customer_overdue_count",
					ObservedValue:  row.CustomerOverdueCount,
					Description:    fmt.Sprintf("Customer has %d total overdue invoices compounding credit risk", row.CustomerOverdueCount),
				})
			}

			evidenceBytes, _ := json.Marshal(evidence)

			priority := PriorityHigh
			risk := RiskHigh
			if row.BalanceDue > 10000 || daysOverdue > 30 || row.CustomerOverdueCount >= 3 {
				priority = PriorityCritical
				risk = RiskCritical
			}

			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID
			ft := DraftTypeOverduePaymentNotice
			var ownerID *int64
			var ownerName *string
			if row.AccountOwnerID.Valid {
				id := row.AccountOwnerID.Int64
				ownerID = &id
			}
			if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
				name := row.AccountOwnerName.String
				ownerName = &name
			}

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Overdue Invoice %s: %s (%d Days Overdue, %s %.2f)", sourceRef, row.CustomerName, daysOverdue, row.Currency, row.BalanceDue),
				Description:        fmt.Sprintf("Receivable invoice %s for %s (%s %.2f, due %s, %d days overdue [%s]) remains unpaid with outstanding balance %s %.2f. Recommend accounts receivable payment reminder.", sourceRef, row.CustomerName, row.Currency, row.TotalAmount, dueStr, daysOverdue, agingBand, row.Currency, row.BalanceDue),
				Category:           CategoryFinance,
				Priority:           priority,
				RiskLevel:          risk,
				Confidence:         "HIGH",
				ConfidenceScore:    0.95,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Transmit payment reminder statement to customer finance desk and verify expected remittance date.",
				ActionType:         ActionTypeReviewOverdueInvoice,
				Status:             StatusNew,
				RequiresApproval:   false,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "OVERDUE_INVOICE_COLLECTIONS",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "OVERDUE_CUSTOMER_INVOICE"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				FollowupType:       &ft,
				SuggestedOwnerID:   ownerID,
				SuggestedOwnerName: ownerName,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 1b. Shipment Vendor Invoices
	for _, row := range rows {
		sourceRef := row.InvoiceNumber
		if sourceRef == "" {
			sourceRef = fmt.Sprintf("INV-%d", row.ID)
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "finance",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "status",
				ObservedValue:  row.Status,
				Description:    "Invoice remains in ISSUED status past standard 14-day payment window",
			},
			{
				SourceModule:   "finance",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "total_amount",
				ObservedValue:  fmt.Sprintf("%s %.2f", row.Currency, row.TotalAmount),
				Description:    "Outstanding invoice balance awaiting settlement",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      fmt.Sprintf("SH-%d", row.ShipmentID),
				FieldName:      "shipment_id",
				ObservedValue:  row.ShipmentID,
				Description:    "Associated operational shipment file",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		priority := PriorityHigh
		risk := RiskHigh
		if row.TotalAmount > 15000 {
			priority = PriorityCritical
			risk = RiskCritical
		}

		ft := FollowupTypeInvoiceReminder
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceInvoice,
			SourceID:          row.ID,
			SourceReference:   sourceRef,
			Title:             fmt.Sprintf("Overdue Invoice %s Awaiting Settlement", sourceRef),
			Description:       fmt.Sprintf("Invoice %s for %s (%s %.2f) was issued on %s and is past due. Follow up with vendor/customer to reconcile settlement.", sourceRef, row.VendorName, row.Currency, row.TotalAmount, row.CreatedAt.Format("2006-01-02")),
			Category:          CategoryFinance,
			Priority:          priority,
			RiskLevel:         risk,
			Confidence:        "HIGH",
			ConfidenceScore:   0.95,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Review invoice line items, verify shipment delivery proof, and transmit statement reminder.",
			ActionType:        "REVIEW_INVOICE",
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "OVERDUE_INVOICE_COLLECTIONS",
			DedupHash:         computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "OVERDUE_INVOICE_COLLECTIONS"),
			FollowupType:      &ft,
			DraftStatus:       "NONE",
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateShipmentExceptions(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type excRow struct {
		ID            int64          `db:"id"`
		ShipmentID    int64          `db:"shipment_id"`
		ExceptionType string         `db:"exception_type"`
		Severity      string         `db:"severity"`
		Title         string         `db:"title"`
		Description   sql.NullString `db:"description"`
		BookingNumber sql.NullString `db:"booking_number"`
		BookingID     sql.NullInt64  `db:"booking_id"`
		CarrierSCAC   string         `db:"carrier_scac"`
		OriginPort    string         `db:"origin_port"`
		DestPort      string         `db:"destination_port"`
		CustomerID    int64          `db:"customer_id"`
		CustomerName  string         `db:"customer_name"`
		CreatedAt     time.Time      `db:"created_at"`
		MultiCount    int            `db:"multi_count"`
	}

	query := `
		SELECT se.id, se.shipment_id, se.exception_type, se.severity, se.title, se.description, se.created_at,
		       s.booking_number, s.booking_id, s.carrier_scac, s.origin_port, s.destination_port,
		       COALESCE(q.customer_id, r.customer_id, 0) AS customer_id,
		       COALESCE(q.customer_name, c.name, '') AS customer_name,
		       (SELECT COUNT(*) FROM shipment_exceptions se2 WHERE se2.shipment_id = se.shipment_id AND se2.org_id = se.org_id AND se2.resolved = 0 AND se2.status NOT IN ('RESOLVED', 'CLOSED')) AS multi_count
		FROM shipment_exceptions se
		JOIN shipments s ON s.id = se.shipment_id AND s.org_id = se.org_id
		LEFT JOIN quotations q ON s.quote_id = q.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE se.org_id = ? AND se.resolved = 0 AND se.status NOT IN ('RESOLVED', 'CLOSED')
		ORDER BY se.created_at DESC
		LIMIT 20
	`
	var rows []excRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		shipmentRef := fmt.Sprintf("SH-%d", row.ShipmentID)
		if row.BookingNumber.Valid && row.BookingNumber.String != "" {
			shipmentRef = row.BookingNumber.String
		}

		desc := "Exception logged on transit milestone."
		if row.Description.Valid && row.Description.String != "" {
			desc = row.Description.String
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "exception_type",
				ObservedValue:  row.ExceptionType,
				Description:    fmt.Sprintf("Active unresolved exception: %s", row.Title),
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "severity",
				ObservedValue:  row.Severity,
				Description:    fmt.Sprintf("Severity assigned by operations monitor: %s", row.Severity),
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "transit_route",
				ObservedValue:  fmt.Sprintf("%s -> %s via %s", row.OriginPort, row.DestPort, row.CarrierSCAC),
				Description:    "Operational shipping lane affected by event",
			},
		}

		if row.MultiCount > 1 {
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "multiple_exceptions",
				ObservedValue:  fmt.Sprintf("%d unresolved exceptions", row.MultiCount),
				Description:    "Multiple concurrent active exceptions detected on single shipment requiring urgent operational triage",
			})
		}

		evidenceBytes, _ := json.Marshal(evidence)

		priority := PriorityHigh
		risk := RiskHigh
		requiresApproval := false

		if row.Severity == "CRITICAL" || row.MultiCount > 1 {
			priority = PriorityCritical
			risk = RiskCritical
			requiresApproval = true
		} else if row.Severity == "LOW" {
			priority = PriorityLow
			risk = RiskLow
		}

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
			if row.CustomerName != "" {
				cname := row.CustomerName
				custName = &cname
			}
		}

		draftType := DraftTypeCustomerShipmentUpdate
		actionType := ActionTypeInvestigateException
		if priority == PriorityCritical {
			draftType = DraftTypeExceptionEscalation
			actionType = ActionTypeEscalateOperationalRisk
		}

		var bkgID *int64
		if row.BookingID.Valid && row.BookingID.Int64 > 0 {
			bid := row.BookingID.Int64
			bkgID = &bid
		}
		shID := row.ShipmentID
		excID := row.ID

		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceShipment,
			SourceID:          row.ShipmentID,
			SourceReference:   shipmentRef,
			Title:             fmt.Sprintf("Shipment %s Exception: %s", shipmentRef, row.Title),
			Description:       fmt.Sprintf("Unresolved %s exception on shipment %s (%s to %s via %s): %s", row.Severity, shipmentRef, row.OriginPort, row.DestPort, row.CarrierSCAC, desc),
			Category:          CategoryOperations,
			Priority:          priority,
			RiskLevel:         risk,
			Confidence:        "HIGH",
			ConfidenceScore:   0.94,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Coordinate with ocean carrier / freight terminal to clear exception and notify consignee with verified updates.",
			ActionType:        actionType,
			Status:            StatusNew,
			RequiresApproval:  requiresApproval,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "UNRESOLVED_SHIPMENT_EXCEPTIONS",
			DedupHash:         computeDedupHash(orgID, SourceShipment, row.ShipmentID, CategoryOperations, fmt.Sprintf("EXC_%d", row.ID)),
			CustomerID:        custID,
			CustomerName:      custName,
			FollowupType:      &draftType,
			DraftStatus:       "NONE",
			ShipmentID:        &shID,
			ExceptionID:       &excID,
			BookingID:         bkgID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateRFQs(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type rfqRow struct {
		ID                int64          `db:"id"`
		RFQNumber         sql.NullString `db:"rfq_number"`
		CustomerID        int64          `db:"customer_id"`
		CustomerName      sql.NullString `db:"customer_name"`
		Origin            sql.NullString `db:"origin"`
		Destination       sql.NullString `db:"destination"`
		Stage             sql.NullString `db:"stage"`
		Status            sql.NullString `db:"status"`
		HealthScore       int            `db:"health_score"`
		SalesAssigneeID   sql.NullInt64  `db:"sales_assignee_id"`
		SalesAssigneeName sql.NullString `db:"sales_assignee_name"`
		CreatedAt         time.Time      `db:"created_at"`
	}

	query := `
		SELECT r.id, r.rfq_number, r.customer_id, c.name AS customer_name,
		       r.origin, r.destination, r.stage, r.status, r.health_score, r.created_at,
		       r.sales_assignee_id, CONCAT(u.first_name, ' ', u.last_name) AS sales_assignee_name
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN users u ON r.sales_assignee_id = u.id
		WHERE r.org_id = ? AND r.status IN ('DRAFT', 'SUBMITTED', 'PENDING')
		ORDER BY r.created_at ASC
		LIMIT 10
	`
	var rows []rfqRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		rfqRef := fmt.Sprintf("RFQ-%d", row.ID)
		if row.RFQNumber.Valid && row.RFQNumber.String != "" {
			rfqRef = row.RFQNumber.String
		}

		origin := "Origin Pending"
		if row.Origin.Valid && row.Origin.String != "" {
			origin = row.Origin.String
		}
		dest := "Destination Pending"
		if row.Destination.Valid && row.Destination.String != "" {
			dest = row.Destination.String
		}
		st := "DRAFT"
		if row.Status.Valid && row.Status.String != "" {
			st = row.Status.String
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "status",
				ObservedValue:  st,
				Description:    "RFQ is awaiting quote generation or carrier rate sourcing",
			},
			{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "lane",
				ObservedValue:  fmt.Sprintf("%s -> %s", origin, dest),
				Description:    "Lane routing specifications requested by shipper",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		cid := row.CustomerID
		var cname *string
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			n := row.CustomerName.String
			cname = &n
		}
		ft := FollowupTypeRFQFollowup
		var ownerID *int64
		var ownerName *string
		if row.SalesAssigneeID.Valid {
			id := row.SalesAssigneeID.Int64
			ownerID = &id
		}
		if row.SalesAssigneeName.Valid && row.SalesAssigneeName.String != "" {
			name := row.SalesAssigneeName.String
			ownerName = &name
		}

		rfqID := row.ID
		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceRFQ,
			SourceID:           row.ID,
			SourceReference:    rfqRef,
			Title:              fmt.Sprintf("RFQ %s Awaiting Commercial Pricing", rfqRef),
			Description:        fmt.Sprintf("Inquiry %s (%s to %s) is in %s state. Rate cards or spot rates should be sourced to issue quotation.", rfqRef, origin, dest, st),
			Category:           CategoryRFQ,
			Priority:           PriorityMedium,
			RiskLevel:          RiskMedium,
			Confidence:         "HIGH",
			ConfidenceScore:    0.88,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Review carrier spot rates for lane and construct competitive commercial quotation.",
			ActionType:         "SOURCE_RATES",
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "RFQ_AWAITING_PRICING",
			DedupHash:          computeDedupHash(orgID, SourceRFQ, row.ID, CategoryRFQ, "RFQ_AWAITING_PRICING"),
			CustomerID:         &cid,
			CustomerName:       cname,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        "NONE",
			RFQID:              &rfqID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateQuotations(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type qRow struct {
		ID               int64          `db:"id"`
		QuotationNumber  string         `db:"quotation_number"`
		CustomerID       int64          `db:"customer_id"`
		CustomerName     sql.NullString `db:"customer_name"`
		Status           string         `db:"status"`
		Currency         string         `db:"currency"`
		TotalAmount      float64        `db:"total_amount"`
		GrossMarginPct   float64        `db:"gross_margin_pct"`
		ValidUntil       *time.Time     `db:"valid_until"`
		CreatedAt        time.Time      `db:"created_at"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT q.id, q.quotation_number, COALESCE(q.customer_id, 0) AS customer_id,
		       q.customer_name, q.status, q.currency, q.total_amount, q.gross_margin_pct,
		       q.valid_until, q.created_at,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM quotations q
		LEFT JOIN customers c ON q.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE q.org_id = ? AND q.status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'SENT')
		  AND q.valid_until IS NOT NULL AND q.valid_until BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 14 DAY)
		ORDER BY q.valid_until ASC
		LIMIT 10
	`
	var rows []qRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		cust := "Customer"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			cust = row.CustomerName.String
		}

		validStr := "soon"
		if row.ValidUntil != nil {
			validStr = row.ValidUntil.Format("2006-01-02")
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      row.QuotationNumber,
				FieldName:      "valid_until",
				ObservedValue:  validStr,
				Description:    "Offer expires within 14 calendar days",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      row.QuotationNumber,
				FieldName:      "gross_margin_pct",
				ObservedValue:  fmt.Sprintf("%.1f%%", row.GrossMarginPct),
				Description:    "Estimated commercial gross margin percentage",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
		}
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			name := row.CustomerName.String
			custName = &name
		}
		ft := FollowupTypeQuotationFollowup
		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		quoteID := row.ID
		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceQuotation,
			SourceID:           row.ID,
			SourceReference:    row.QuotationNumber,
			Title:              fmt.Sprintf("Quotation %s Approaching Expiry", row.QuotationNumber),
			Description:        fmt.Sprintf("Quotation %s for %s (%s %.2f, %.1f%% margin) is valid until %s. Confirm commercial acceptance or extend validity.", row.QuotationNumber, cust, row.Currency, row.TotalAmount, row.GrossMarginPct, validStr),
			Category:           CategoryQuotation,
			Priority:           PriorityMedium,
			RiskLevel:          RiskMedium,
			Confidence:         "HIGH",
			ConfidenceScore:    0.90,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Follow up with customer buyer and assess validity extension or conversion to booking.",
			ActionType:         "FOLLOW_UP_QUOTE",
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "QUOTATION_EXPIRING_SOON",
			DedupHash:          computeDedupHash(orgID, SourceQuotation, row.ID, CategoryQuotation, "QUOTATION_EXPIRING_SOON"),
			CustomerID:         custID,
			CustomerName:       custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        "NONE",
			QuotationID:        &quoteID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateContracts(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type cRow struct {
		ID                int64           `db:"id"`
		ContractReference string          `db:"contract_reference"`
		ContractName      string          `db:"contract_name"`
		ContractType      string          `db:"contract_type"`
		PartyID           int64           `db:"party_id"`
		PartyName         string          `db:"party_name"`
		Status            string          `db:"status"`
		ExpiryDate        *time.Time      `db:"expiry_date"`
		ContractValue     sql.NullFloat64 `db:"contract_value"`
		Currency          sql.NullString  `db:"currency"`
		Owner             sql.NullString  `db:"owner"`
	}

	query := `
		SELECT c.id, c.contract_reference, c.contract_name, c.contract_type, c.party_id, c.party_name, c.status,
		       c.expiry_date, c.contract_value, c.currency, c.owner
		FROM contracts c
		WHERE c.org_id = ? AND c.status NOT IN ('TERMINATED', 'CANCELLED', 'ARCHIVED')
		ORDER BY c.expiry_date ASC
	`
	var rows []cRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	now := time.Now()

	for _, row := range rows {
		contractID := row.ID
		val := 0.0
		if row.ContractValue.Valid {
			val = row.ContractValue.Float64
		}
		curr := "USD"
		if row.Currency.Valid && row.Currency.String != "" {
			curr = row.Currency.String
		}
		ownerStr := ""
		if row.Owner.Valid {
			ownerStr = row.Owner.String
		}

		// 1. Expired contracts
		if row.ExpiryDate != nil && row.ExpiryDate.Before(now) {
			daysSince := int(now.Sub(*row.ExpiryDate).Hours() / 24)
			expStr := row.ExpiryDate.Format("2006-01-02")
			evidence := []EvidenceItem{
				{
					SourceModule:   "contracts",
					SourceEntityID: row.ID,
					SourceRef:      row.ContractReference,
					FieldName:      "expiry_date",
					ObservedValue:  expStr,
					Description:    fmt.Sprintf("Contract concluded on %s (%d days past expiry)", expStr, daysSince),
				},
				{
					SourceModule:   "contracts",
					SourceEntityID: row.ID,
					SourceRef:      row.ContractReference,
					FieldName:      "status",
					ObservedValue:  row.Status,
					Description:    fmt.Sprintf("Current operational status is '%s'", row.Status),
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "CONTRACT_EXPIRED"
			prio := PriorityCritical
			if row.Status == "DRAFT" {
				prio = PriorityHigh
			}
			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceContract,
				SourceID:           row.ID,
				SourceReference:    row.ContractReference,
				Title:              fmt.Sprintf("Expired Contract: %s (%s)", row.ContractReference, row.PartyName),
				Description:        fmt.Sprintf("Agreement '%s' with %s (%s %.2f) expired on %s. Continued operations require immediate term extension or formal contract closure.", row.ContractName, row.PartyName, curr, val, expStr),
				Category:           CategoryContract,
				Priority:           prio,
				RiskLevel:          RiskCritical,
				Confidence:         "HIGH",
				ConfidenceScore:    0.98,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Review expired terms, confirm operational commitments, and verify if formal renewal or closeout is required.",
				ActionType:         ActionTypeReviewExpiringContract,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        rule,
				DedupHash:          computeDedupHash(orgID, SourceContract, row.ID, CategoryContract, rule),
				ContractID:         &contractID,
				SuggestedOwnerName: &ownerStr,
				DraftStatus:        DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		} else if row.ExpiryDate != nil && row.ExpiryDate.Before(now.AddDate(0, 0, 60)) {
			// 2. Expiring soon (within 60 days)
			daysUntil := int(row.ExpiryDate.Sub(now).Hours() / 24)
			expStr := row.ExpiryDate.Format("2006-01-02")
			evidence := []EvidenceItem{
				{
					SourceModule:   "contracts",
					SourceEntityID: row.ID,
					SourceRef:      row.ContractReference,
					FieldName:      "expiry_date",
					ObservedValue:  expStr,
					Description:    fmt.Sprintf("Contract scheduled to conclude in %d days (%s)", daysUntil, expStr),
				},
				{
					SourceModule:   "contracts",
					SourceEntityID: row.ID,
					SourceRef:      row.ContractReference,
					FieldName:      "contract_value",
					ObservedValue:  fmt.Sprintf("%s %.2f", curr, val),
					Description:    "Commercial agreement monetary baseline",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "CONTRACT_EXPIRING_SOON"
			prio := PriorityHigh
			risk := RiskHigh
			if daysUntil <= 14 {
				prio = PriorityCritical
				risk = RiskCritical
			}
			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceContract,
				SourceID:           row.ID,
				SourceReference:    row.ContractReference,
				Title:              fmt.Sprintf("Contract Expiry Notice: %s (%s)", row.ContractReference, row.PartyName),
				Description:        fmt.Sprintf("Agreement '%s' with %s (%s %.2f) expires on %s (%d days remaining). Timely notice period verification and renewal preparation advised.", row.ContractName, row.PartyName, curr, val, expStr, daysUntil),
				Category:           CategoryContract,
				Priority:           prio,
				RiskLevel:          risk,
				Confidence:         "HIGH",
				ConfidenceScore:    0.95,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Prepare contract renewal reminder and verify counterparty renewal notice obligations.",
				ActionType:         ActionTypePrepareRenewalReminder,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        rule,
				DedupHash:          computeDedupHash(orgID, SourceContract, row.ID, CategoryContract, rule),
				ContractID:         &contractID,
				SuggestedOwnerName: &ownerStr,
				DraftStatus:        DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		}

		// 3. Missing owner
		if strings.TrimSpace(ownerStr) == "" && (row.Status == "ACTIVE" || row.Status == "DRAFT") {
			evidence := []EvidenceItem{
				{
					SourceModule:   "contracts",
					SourceEntityID: row.ID,
					SourceRef:      row.ContractReference,
					FieldName:      "owner",
					ObservedValue:  "UNASSIGNED",
					Description:    "No legal or account representative designated as contract owner",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "CONTRACT_MISSING_OWNER"
			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceContract,
				SourceID:           row.ID,
				SourceReference:    row.ContractReference,
				Title:              fmt.Sprintf("Unassigned Contract Owner: %s", row.ContractReference),
				Description:        fmt.Sprintf("Agreement '%s' with %s has no assigned owner. Accountability for renewal and compliance requires designating an owner.", row.ContractName, row.PartyName),
				Category:           CategoryContract,
				Priority:           PriorityMedium,
				RiskLevel:          RiskMedium,
				Confidence:         "HIGH",
				ConfidenceScore:    0.90,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Assign an internal contract owner from the legal or account management team.",
				ActionType:         ActionTypeAssignContractOwner,
				Status:             StatusNew,
				RequiresApproval:   false,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        rule,
				DedupHash:          computeDedupHash(orgID, SourceContract, row.ID, CategoryContract, rule),
				ContractID:         &contractID,
				DraftStatus:        DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		}
	}

	return candidates, nil
}

func (g *generator) evaluateDocumentExpiryAndMissingInfo(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	var candidates []*Recommendation
	now := time.Now()

	// A. Shipment Documents: expired, missing, or discrepancy
	type docRow struct {
		ID           int64          `db:"id"`
		ShipmentID   *int64         `db:"shipment_id"`
		DocType      string         `db:"doc_type"`
		FileName     string         `db:"file_name"`
		Status       string         `db:"status"`
		ExpiresAt    *time.Time     `db:"expires_at"`
		DocumentDate *time.Time     `db:"document_date"`
		AISummary    sql.NullString `db:"ai_summary"`
	}

	var sDocs []docRow
	q := `
		SELECT id, shipment_id, doc_type, file_name, status, expires_at, document_date, ai_summary
		FROM shipment_documents
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT 50
	`
	_ = g.db.SelectContext(ctx, &sDocs, q, orgID)

	for _, d := range sDocs {
		docIDStr := fmt.Sprintf("%d", d.ID)
		if d.ExpiresAt != nil && d.ExpiresAt.Before(now) {
			expStr := d.ExpiresAt.Format("2006-01-02")
			evidence := []EvidenceItem{
				{
					SourceModule:   "shipment_documents",
					SourceEntityID: d.ID,
					SourceRef:      d.FileName,
					FieldName:      "expires_at",
					ObservedValue:  expStr,
					Description:    fmt.Sprintf("Document validity expired on %s", expStr),
				},
				{
					SourceModule:   "shipment_documents",
					SourceEntityID: d.ID,
					SourceRef:      d.FileName,
					FieldName:      "doc_type",
					ObservedValue:  d.DocType,
					Description:    "Configured shipment regulatory document type",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "DOCUMENT_EXPIRED"
			cand := &Recommendation{
				OrgID:             orgID,
				SourceType:        SourceDocument,
				SourceID:          d.ID,
				SourceReference:   d.FileName,
				Title:             fmt.Sprintf("Expired Document: %s (%s)", d.DocType, d.FileName),
				Description:       fmt.Sprintf("Required document '%s' (type %s) expired on %s. Updated certificate or filing must be requested.", d.FileName, d.DocType, expStr),
				Category:          CategoryCompliance,
				Priority:          PriorityHigh,
				RiskLevel:         RiskHigh,
				Confidence:        "HIGH",
				ConfidenceScore:   0.95,
				EvidenceJSON:      string(evidenceBytes),
				RecommendedAction: "Request replacement or renewed document from counterparty.",
				ActionType:        ActionTypeRequestMissingDocument,
				Status:            StatusNew,
				RequiresApproval:  false,
				CorrelationID:     correlationID,
				CreatedBy:         "SYSTEM",
				GeneratedBy:       "DETERMINISTIC_RULES",
				RuleApplied:       rule,
				DedupHash:         computeDedupHash(orgID, SourceDocument, d.ID, CategoryCompliance, rule),
				DocumentID:        &docIDStr,
				ShipmentID:        d.ShipmentID,
				DraftStatus:       DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		}

		if strings.EqualFold(d.Status, "MISSING") {
			evidence := []EvidenceItem{
				{
					SourceModule:   "shipment_documents",
					SourceEntityID: d.ID,
					SourceRef:      d.FileName,
					FieldName:      "status",
					ObservedValue:  "MISSING",
					Description:    "Checklist requirement has no verified uploaded file",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "DOCUMENT_MISSING_REQUIRED"
			cand := &Recommendation{
				OrgID:             orgID,
				SourceType:        SourceDocument,
				SourceID:          d.ID,
				SourceReference:   d.FileName,
				Title:             fmt.Sprintf("Missing Required Document: %s", d.DocType),
				Description:       fmt.Sprintf("Checklist requirement '%s' for shipment is currently missing. Follow-up is required to avoid clearance delays.", d.DocType),
				Category:          CategoryCompliance,
				Priority:          PriorityHigh,
				RiskLevel:         RiskHigh,
				Confidence:        "HIGH",
				ConfidenceScore:   0.93,
				EvidenceJSON:      string(evidenceBytes),
				RecommendedAction: "Request missing compliance document from shipper or carrier.",
				ActionType:        ActionTypeRequestMissingDocument,
				Status:            StatusNew,
				RequiresApproval:  false,
				CorrelationID:     correlationID,
				CreatedBy:         "SYSTEM",
				GeneratedBy:       "DETERMINISTIC_RULES",
				RuleApplied:       rule,
				DedupHash:         computeDedupHash(orgID, SourceDocument, d.ID, CategoryCompliance, rule),
				DocumentID:        &docIDStr,
				ShipmentID:        d.ShipmentID,
				DraftStatus:       DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		}

		if strings.EqualFold(d.Status, "DISCREPANCY") {
			discDesc := "Extracted document fields conflict with shipment master data"
			if d.AISummary.Valid && d.AISummary.String != "" {
				discDesc = d.AISummary.String
			}
			evidence := []EvidenceItem{
				{
					SourceModule:   "shipment_documents",
					SourceEntityID: d.ID,
					SourceRef:      d.FileName,
					FieldName:      "status",
					ObservedValue:  "DISCREPANCY",
					Description:    discDesc,
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			rule := "DOCUMENT_DISCREPANCY_DETECTED"
			cand := &Recommendation{
				OrgID:             orgID,
				SourceType:        SourceDocument,
				SourceID:          d.ID,
				SourceReference:   d.FileName,
				Title:             fmt.Sprintf("Document Discrepancy: %s", d.FileName),
				Description:       fmt.Sprintf("Discrepancy detected on document '%s' (%s). Operational review required before customs submission.", d.FileName, d.DocType),
				Category:          CategoryCompliance,
				Priority:          PriorityHigh,
				RiskLevel:         RiskHigh,
				Confidence:        "HIGH",
				ConfidenceScore:   0.91,
				EvidenceJSON:      string(evidenceBytes),
				RecommendedAction: "Review document discrepancies against shipment manifest and reconcile values.",
				ActionType:        ActionTypeVerifyDocumentMetadata,
				Status:            StatusNew,
				RequiresApproval:  false,
				CorrelationID:     correlationID,
				CreatedBy:         "SYSTEM",
				GeneratedBy:       "DETERMINISTIC_RULES",
				RuleApplied:       rule,
				DedupHash:         computeDedupHash(orgID, SourceDocument, d.ID, CategoryCompliance, rule),
				DocumentID:        &docIDStr,
				ShipmentID:        d.ShipmentID,
				DraftStatus:       DraftStatusNotGenerated,
			}
			candidates = append(candidates, cand)
		}
	}

	// B. Contract Documents: Pending extraction review
	type cDocRow struct {
		ID                 string         `db:"id"`
		FileName           string         `db:"file_name"`
		CarrierName        sql.NullString `db:"carrier_name"`
		Status             string         `db:"status"`
		PendingReviewCount int            `db:"pending_review_count"`
		FailedRateCount    int            `db:"failed_rate_count"`
	}
	var cDocs []cDocRow
	cq := `
		SELECT id, file_name, carrier_name, status, pending_review_count, failed_rate_count
		FROM contract_documents
		WHERE org_id = ? AND (pending_review_count > 0 OR status IN ('PENDING_REVIEW', 'PENDING_EXTRACTION'))
		LIMIT 10
	`
	_ = g.db.SelectContext(ctx, &cDocs, cq, orgID)

	for _, cd := range cDocs {
		carrier := "Carrier"
		if cd.CarrierName.Valid && cd.CarrierName.String != "" {
			carrier = cd.CarrierName.String
		}
		evidence := []EvidenceItem{
			{
				SourceModule:   "contract_documents",
				SourceEntityID: 0,
				SourceRef:      cd.FileName,
				FieldName:      "status",
				ObservedValue:  cd.Status,
				Description:    fmt.Sprintf("%d extracted rate item(s) awaiting human validation", cd.PendingReviewCount),
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)
		rule := "CONTRACT_DOC_PENDING_REVIEW"
		docIDCopy := cd.ID
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceDocument,
			SourceID:          0,
			SourceReference:   cd.FileName,
			Title:             fmt.Sprintf("Contract Document Review: %s", cd.FileName),
			Description:       fmt.Sprintf("Contract document '%s' from %s has %d rate item(s) pending human verification before rate card activation.", cd.FileName, carrier, cd.PendingReviewCount),
			Category:          CategoryCompliance,
			Priority:          PriorityMedium,
			RiskLevel:         RiskMedium,
			Confidence:        "MEDIUM",
			ConfidenceScore:   0.82,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Verify extracted rates in Document Workspace and confirm OCR validity.",
			ActionType:        ActionTypeVerifyDocumentMetadata,
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       rule,
			DedupHash:         computeDedupHash(orgID, SourceDocument, int64(len(cd.ID)), CategoryCompliance, rule+cd.ID),
			DocumentID:        &docIDCopy,
			DraftStatus:       DraftStatusNotGenerated,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateComplianceRequirementsAndObligations(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	var candidates []*Recommendation
	now := time.Now()

	// A. Compliance Requirements
	type reqRow struct {
		ID                int64          `db:"id"`
		ContractID        int64          `db:"contract_id"`
		ContractReference string         `db:"contract_reference"`
		PartyName         string         `db:"party_name"`
		RequirementType   string         `db:"requirement_type"`
		Title             string         `db:"title"`
		Description       sql.NullString `db:"description"`
		ResponsibleParty  string         `db:"responsible_party"`
		ValidUntil        *time.Time     `db:"valid_until"`
		Status            string         `db:"status"`
		RiskSeverity      string         `db:"risk_severity"`
	}

	var reqs []reqRow
	rq := `
		SELECT r.id, r.contract_id, c.contract_reference, c.party_name, r.requirement_type,
		       r.title, r.description, r.responsible_party, r.valid_until, r.status, r.risk_severity
		FROM contract_compliance_requirements r
		JOIN contracts c ON c.id = r.contract_id
		WHERE r.org_id = ? AND r.status NOT IN ('COMPLIANT', 'VERIFIED')
		ORDER BY r.id ASC
		LIMIT 20
	`
	_ = g.db.SelectContext(ctx, &reqs, rq, orgID)

	for _, r := range reqs {
		prio := PriorityMedium
		risk := RiskMedium
		switch strings.ToUpper(r.RiskSeverity) {
		case "CRITICAL":
			prio = PriorityCritical
			risk = RiskCritical
		case "HIGH":
			prio = PriorityHigh
			risk = RiskHigh
		}

		validStr := "Unspecified"
		isOverdue := false
		if r.ValidUntil != nil {
			validStr = r.ValidUntil.Format("2006-01-02")
			if r.ValidUntil.Before(now) {
				isOverdue = true
				prio = PriorityCritical
				risk = RiskCritical
			}
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "contract_compliance_requirements",
				SourceEntityID: r.ID,
				SourceRef:      r.ContractReference,
				FieldName:      "status",
				ObservedValue:  r.Status,
				Description:    fmt.Sprintf("Compliance requirement '%s' status is %s", r.Title, r.Status),
			},
			{
				SourceModule:   "contract_compliance_requirements",
				SourceEntityID: r.ID,
				SourceRef:      r.ContractReference,
				FieldName:      "valid_until",
				ObservedValue:  validStr,
				Description:    "Regulatory or contractual validity date",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)
		rule := "COMPLIANCE_REQUIREMENT_PENDING"
		if isOverdue {
			rule = "COMPLIANCE_REQUIREMENT_OVERDUE"
		}

		candID := r.ID
		candContractID := r.ContractID
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceCompliance,
			SourceID:          r.ID,
			SourceReference:   r.ContractReference,
			Title:             fmt.Sprintf("Compliance Review Required: %s (%s)", r.Title, r.PartyName),
			Description:       fmt.Sprintf("Contract %s requirement '%s' (%s) assigned to %s is %s (valid until %s). Verification is required.", r.ContractReference, r.Title, r.RequirementType, r.ResponsibleParty, r.Status, validStr),
			Category:          CategoryCompliance,
			Priority:          prio,
			RiskLevel:         risk,
			Confidence:        "HIGH",
			ConfidenceScore:   0.94,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Review compliance checklist, verify submitted credentials, and update compliance state.",
			ActionType:        ActionTypeReviewComplianceChecklist,
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       rule,
			DedupHash:         computeDedupHash(orgID, SourceCompliance, r.ID, CategoryCompliance, rule),
			ComplianceID:      &candID,
			ContractID:        &candContractID,
			DraftStatus:       DraftStatusNotGenerated,
		}
		candidates = append(candidates, cand)
	}

	// B. Contract Obligations: Overdue or unassigned
	type obRow struct {
		ID                  int64          `db:"id"`
		ContractID          int64          `db:"contract_id"`
		ContractReference   string         `db:"contract_reference"`
		PartyName           string         `db:"party_name"`
		ObligationReference string         `db:"obligation_reference"`
		Title               string         `db:"title"`
		ResponsibleParty    string         `db:"responsible_party"`
		Owner               sql.NullString `db:"owner"`
		Priority            string         `db:"priority"`
		DueDate             *time.Time     `db:"due_date"`
		Status              string         `db:"status"`
	}
	var obs []obRow
	oq := `
		SELECT o.id, o.contract_id, c.contract_reference, c.party_name, o.obligation_reference,
		       o.title, o.responsible_party, o.owner, o.priority, o.due_date, o.status
		FROM contract_obligations o
		JOIN contracts c ON c.id = o.contract_id
		WHERE o.org_id = ? AND o.status = 'ACTIVE' AND (o.due_date < NOW() OR o.owner IS NULL OR TRIM(o.owner) = '')
		ORDER BY o.due_date ASC
		LIMIT 10
	`
	_ = g.db.SelectContext(ctx, &obs, oq, orgID)

	for _, o := range obs {
		dueStr := "No deadline set"
		isOverdue := false
		if o.DueDate != nil {
			dueStr = o.DueDate.Format("2006-01-02")
			if o.DueDate.Before(now) {
				isOverdue = true
			}
		}

		rule := "CONTRACT_OBLIGATION_UNASSIGNED"
		desc := fmt.Sprintf("Operational obligation '%s' (%s) under contract %s has no assigned owner.", o.Title, o.ObligationReference, o.ContractReference)
		if isOverdue {
			rule = "CONTRACT_OBLIGATION_OVERDUE"
			desc = fmt.Sprintf("Operational obligation '%s' under contract %s was due on %s and remains unfulfilled.", o.Title, o.ContractReference, dueStr)
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "contract_obligations",
				SourceEntityID: o.ID,
				SourceRef:      o.ObligationReference,
				FieldName:      "status",
				ObservedValue:  o.Status,
				Description:    desc,
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		candContractID := o.ContractID
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceContract,
			SourceID:          o.ContractID,
			SourceReference:   o.ObligationReference,
			Title:             fmt.Sprintf("Contract Obligation Follow-up: %s", o.Title),
			Description:       desc,
			Category:          CategoryContract,
			Priority:          PriorityHigh,
			RiskLevel:         RiskHigh,
			Confidence:        "HIGH",
			ConfidenceScore:   0.92,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Review operational obligation terms and designate responsible team member.",
			ActionType:        ActionTypeReviewContractObligation,
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       rule,
			DedupHash:         computeDedupHash(orgID, SourceContract, o.ID, CategoryContract, rule),
			ContractID:        &candContractID,
			DraftStatus:       DraftStatusNotGenerated,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateCustomerCredit(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type custRow struct {
		ID               int64          `db:"id"`
		Name             string         `db:"name"`
		CustomerCode     sql.NullString `db:"customer_code"`
		CreditStatus     string         `db:"credit_status"`
		CreditLimit      float64        `db:"credit_limit"`
		HealthScore      int            `db:"health_score"`
		PaymentTerms     string         `db:"payment_terms"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT c.id, c.name, c.customer_code, c.credit_status, c.credit_limit, c.health_score, c.payment_terms,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM customers c
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE c.org_id = ? AND (c.credit_status IN ('OVERDUE', 'WATCHLIST', 'SUSPENDED') OR c.health_score < 60)
		ORDER BY c.health_score ASC
		LIMIT 10
	`
	var rows []custRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		code := fmt.Sprintf("CUST-%d", row.ID)
		if row.CustomerCode.Valid && row.CustomerCode.String != "" {
			code = row.CustomerCode.String
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "customers",
				SourceEntityID: row.ID,
				SourceRef:      code,
				FieldName:      "credit_status",
				ObservedValue:  row.CreditStatus,
				Description:    "Customer credit health categorized outside standard parameters",
			},
			{
				SourceModule:   "customers",
				SourceEntityID: row.ID,
				SourceRef:      code,
				FieldName:      "health_score",
				ObservedValue:  row.HealthScore,
				Description:    "Composite customer health score below minimum 60 threshold",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		priority := PriorityHigh
		risk := RiskHigh
		if row.CreditStatus == "SUSPENDED" {
			priority = PriorityCritical
			risk = RiskCritical
		}

		cid := row.ID
		cname := row.Name
		ft := FollowupTypeAccountReview
		if row.HealthScore < 60 {
			ft = FollowupTypeServiceRecovery
		}
		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceCustomer,
			SourceID:           row.ID,
			SourceReference:    code,
			Title:              fmt.Sprintf("Customer %s Credit Watch Alert", row.Name),
			Description:        fmt.Sprintf("Shipper account %s (%s) has credit status %s and health score %d/100. Review outstanding exposure before releasing new bookings.", row.Name, code, row.CreditStatus, row.HealthScore),
			Category:           CategoryCustomer,
			Priority:           priority,
			RiskLevel:          risk,
			Confidence:         "HIGH",
			ConfidenceScore:    0.91,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Verify ledger balances with finance and confirm prepayment requirement for pending freight.",
			ActionType:         "CREDIT_REVIEW",
			Status:             StatusNew,
			RequiresApproval:   true,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "HIGH_RISK_CUSTOMER_CREDIT",
			DedupHash:          computeDedupHash(orgID, SourceCustomer, row.ID, CategoryCustomer, "HIGH_RISK_CUSTOMER_CREDIT"),
			CustomerID:         &cid,
			CustomerName:       &cname,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        "NONE",
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateLeads(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type leadRow struct {
		ID          int64          `db:"id"`
		CompanyName string         `db:"company_name"`
		ContactName sql.NullString `db:"contact_name"`
		Status      string         `db:"status"`
		AIScore     int            `db:"ai_score"`
		CreatedAt   time.Time      `db:"created_at"`
	}

	query := `
		SELECT id, company_name, contact_name, status, ai_score, created_at
		FROM leads
		WHERE org_id = ? AND status = 'NEW' AND assigned_to IS NULL AND created_at < DATE_SUB(NOW(), INTERVAL 2 DAY)
		ORDER BY ai_score DESC
		LIMIT 10
	`
	var rows []leadRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		leadRef := fmt.Sprintf("LEAD-%d", row.ID)
		contact := "Commercial Contact"
		if row.ContactName.Valid && row.ContactName.String != "" {
			contact = row.ContactName.String
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "leads",
				SourceEntityID: row.ID,
				SourceRef:      leadRef,
				FieldName:      "status",
				ObservedValue:  row.Status,
				Description:    "Inbound prospect unassigned past 48 hours",
			},
			{
				SourceModule:   "leads",
				SourceEntityID: row.ID,
				SourceRef:      leadRef,
				FieldName:      "ai_score",
				ObservedValue:  row.AIScore,
				Description:    "Lead qualification priority index",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		ft := FollowupTypeGeneralCheckin
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceLead,
			SourceID:          row.ID,
			SourceReference:   leadRef,
			Title:             fmt.Sprintf("Unassigned Lead %s Requires Attention", row.CompanyName),
			Description:       fmt.Sprintf("New sales inquiry for %s (Contact: %s, Score: %d) has not been assigned to a business development representative.", row.CompanyName, contact, row.AIScore),
			Category:          CategorySales,
			Priority:          PriorityMedium,
			RiskLevel:         RiskLow,
			Confidence:        "HIGH",
			ConfidenceScore:   0.85,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Assign sales owner and initiate introductory freight requirements discovery.",
			ActionType:        "ASSIGN_LEAD",
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "LEAD_UNATTENDED",
			DedupHash:         computeDedupHash(orgID, SourceLead, row.ID, CategorySales, "LEAD_UNATTENDED"),
			FollowupType:      &ft,
			DraftStatus:       "NONE",
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateCustomerInactivity(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type inactiveCust struct {
		ID               int64          `db:"id"`
		Name             string         `db:"name"`
		HealthScore      int            `db:"health_score"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT c.id, c.name, c.health_score, c.account_owner_id,
		       CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM customers c
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE c.org_id = ? AND c.status = 'ACTIVE'
		  AND c.id NOT IN (
		      SELECT DISTINCT customer_id FROM customer_invoices WHERE org_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
		  )
		  AND c.id NOT IN (
		      SELECT DISTINCT customer_id FROM rfqs WHERE org_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
		  )
		ORDER BY c.health_score ASC
		LIMIT 5
	`
	var rows []inactiveCust
	if err := g.db.SelectContext(ctx, &rows, query, orgID, orgID, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		sourceRef := fmt.Sprintf("CUST-%d", row.ID)
		evidence := []EvidenceItem{
			{
				SourceModule:   "customers",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "inactivity_period",
				ObservedValue:  ">30 days without new RFQs or invoices",
				Description:    "No commercial bookings, RFQs, or billing entries recorded within the last 30 days",
			},
			{
				SourceModule:   "customers",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "health_score",
				ObservedValue:  row.HealthScore,
				Description:    "Current customer health rating",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		custID := row.ID
		custName := row.Name
		ft := FollowupTypeGeneralCheckin

		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceCustomer,
			SourceID:           row.ID,
			SourceReference:    sourceRef,
			Title:              fmt.Sprintf("Customer %s: Relationship Check-In Recommended", row.Name),
			Description:        fmt.Sprintf("Customer %s has no active freight bookings, RFQs, or invoice activity in the last 30 days. Recommend a relationship check-in to confirm ongoing freight pipeline.", row.Name),
			Category:           CategoryCustomer,
			Priority:           PriorityMedium,
			RiskLevel:          RiskLow,
			Confidence:         "HIGH",
			ConfidenceScore:    0.88,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Schedule check-in call with client shipping coordinator to review upcoming trade lane requirements.",
			ActionType:         "SCHEDULE_CHECKIN",
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "CUSTOMER_INACTIVITY_CHECKIN",
			DedupHash:          computeDedupHash(orgID, SourceCustomer, row.ID, CategoryCustomer, "CUSTOMER_INACTIVITY_CHECKIN"),
			CustomerID:         &custID,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        "NONE",
		}
		candidates = append(candidates, cand)
	}
	return candidates, nil
}

// ── Phase 2 Task 2.3: RFQ & Quotation Assistant Deterministic Rules ──────────

func (g *generator) evaluateRFQMissingInfo(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type rfqMissingRow struct {
		ID                int64          `db:"id"`
		RFQNumber         sql.NullString `db:"rfq_number"`
		CustomerID        int64          `db:"customer_id"`
		CustomerName      sql.NullString `db:"customer_name"`
		Origin            sql.NullString `db:"origin"`
		Destination       sql.NullString `db:"destination"`
		Incoterms         sql.NullString `db:"incoterms"`
		TargetDate        *time.Time     `db:"target_date"`
		Status            sql.NullString `db:"status"`
		Stage             sql.NullString `db:"stage"`
		SalesAssigneeID   sql.NullInt64  `db:"sales_assignee_id"`
		SalesAssigneeName sql.NullString `db:"sales_assignee_name"`
		CreatedAt         time.Time      `db:"created_at"`
	}

	query := `
		SELECT r.id, r.rfq_number, r.customer_id, c.name AS customer_name,
		       r.origin, r.destination, r.incoterms, r.target_date, r.status, r.stage,
		       r.sales_assignee_id, CONCAT(u.first_name, ' ', u.last_name) AS sales_assignee_name,
		       r.created_at
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN users u ON r.sales_assignee_id = u.id
		WHERE r.org_id = ? AND r.status IN ('DRAFT', 'SUBMITTED', 'PENDING', 'NEW', 'IN_REVIEW')
		ORDER BY r.created_at ASC
		LIMIT 20
	`
	var rows []rfqMissingRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		var missingFields []string
		var evidence []EvidenceItem

		rfqRef := fmt.Sprintf("RFQ-%d", row.ID)
		if row.RFQNumber.Valid && row.RFQNumber.String != "" {
			rfqRef = row.RFQNumber.String
		}

		if !row.Origin.Valid || strings.TrimSpace(row.Origin.String) == "" {
			missingFields = append(missingFields, "origin")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "origin",
				ObservedValue:  "NULL / Missing",
				Description:    "Origin location/port is undefined; carrier lane pricing cannot be resolved",
			})
		}
		if !row.Destination.Valid || strings.TrimSpace(row.Destination.String) == "" {
			missingFields = append(missingFields, "destination")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "destination",
				ObservedValue:  "NULL / Missing",
				Description:    "Destination location/port is undefined; delivery route and terminal tariffs cannot be determined",
			})
		}
		if !row.Incoterms.Valid || strings.TrimSpace(row.Incoterms.String) == "" {
			missingFields = append(missingFields, "incoterms")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "incoterms",
				ObservedValue:  "NULL / Missing",
				Description:    "Commercial incoterms (e.g. FOB, CIF, EXW) not specified; freight liability and cost boundary unresolved",
			})
		}
		if row.TargetDate == nil {
			missingFields = append(missingFields, "target_date")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "target_date",
				ObservedValue:  "NULL / Missing",
				Description:    "Customer target delivery/sailing date is unspecified",
			})
		}

		// Also check cargo items in rfq_items
		type itemSummary struct {
			ItemCount   int     `db:"item_count"`
			TotalWeight float64 `db:"total_weight"`
			TotalVolume float64 `db:"total_volume"`
		}
		var is itemSummary
		_ = g.db.GetContext(ctx, &is, `
			SELECT COUNT(*) AS item_count,
			       COALESCE(SUM(weight_kg), 0.0) AS total_weight,
			       COALESCE(SUM(volume_cbm), 0.0) AS total_volume
			FROM rfq_items
			WHERE rfq_id = ?
		`, row.ID)

		if is.ItemCount == 0 {
			missingFields = append(missingFields, "cargo_items")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq_items",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "cargo_items",
				ObservedValue:  "0 items",
				Description:    "No cargo line items or package descriptions recorded on inquiry",
			})
		} else if is.TotalWeight <= 0 && is.TotalVolume <= 0 {
			missingFields = append(missingFields, "cargo_weight_or_volume")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "rfq_items",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "weight_and_volume",
				ObservedValue:  "0 kg / 0 cbm",
				Description:    "Cargo items lack physical weight and dimensional volume data required for ocean/air rating",
			})
		}

		if len(missingFields) == 0 {
			continue // Fully complete RFQ; do not flag as missing info!
		}

		evidenceBytes, _ := json.Marshal(evidence)

		custName := "Shipper"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			custName = row.CustomerName.String
		}

		var ownerID *int64
		var ownerName *string
		if row.SalesAssigneeID.Valid {
			id := row.SalesAssigneeID.Int64
			ownerID = &id
		}
		if row.SalesAssigneeName.Valid && row.SalesAssigneeName.String != "" {
			name := row.SalesAssigneeName.String
			ownerName = &name
		}

		priority := PriorityHigh
		risk := RiskMedium
		if len(missingFields) >= 2 {
			risk = RiskHigh
		}

		rfqID := row.ID
		cid := row.CustomerID
		ft := DraftTypeRFQClarification

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceRFQ,
			SourceID:           row.ID,
			SourceReference:    rfqRef,
			Title:              fmt.Sprintf("RFQ %s: Missing Required Commercial Information (%s)", rfqRef, strings.Join(missingFields, ", ")),
			Description:        fmt.Sprintf("Inquiry %s for %s lacks required commercial specifications: %s. Rate calculation and carrier tender cannot proceed without complete information.", rfqRef, custName, strings.Join(missingFields, ", ")),
			Category:           CategoryRFQ,
			Priority:           priority,
			RiskLevel:          risk,
			Confidence:         "HIGH",
			ConfidenceScore:    0.95,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  fmt.Sprintf("Request clarification from %s regarding %s before proceeding with quotation.", custName, strings.Join(missingFields, ", ")),
			ActionType:         ActionTypeRequestRFQClarification,
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "RFQ_MISSING_INFO",
			DedupHash:          computeDedupHash(orgID, SourceRFQ, row.ID, CategoryRFQ, "RFQ_MISSING_INFO"),
			CustomerID:         &cid,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        DraftStatusNotGenerated,
			RFQID:              &rfqID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateRFQResponseDeadlines(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type rfqDeadlineRow struct {
		ID                int64          `db:"id"`
		RFQNumber         sql.NullString `db:"rfq_number"`
		CustomerID        int64          `db:"customer_id"`
		CustomerName      sql.NullString `db:"customer_name"`
		Origin            sql.NullString `db:"origin"`
		Destination       sql.NullString `db:"destination"`
		TargetDate        *time.Time     `db:"target_date"`
		Status            sql.NullString `db:"status"`
		SalesAssigneeID   sql.NullInt64  `db:"sales_assignee_id"`
		SalesAssigneeName sql.NullString `db:"sales_assignee_name"`
		CreatedAt         time.Time      `db:"created_at"`
		QuotesCount       int            `db:"quotes_count"`
	}

	query := `
		SELECT r.id, r.rfq_number, r.customer_id, c.name AS customer_name,
		       r.origin, r.destination, r.target_date, r.status,
		       r.sales_assignee_id, CONCAT(u.first_name, ' ', u.last_name) AS sales_assignee_name,
		       r.created_at,
		       (SELECT COUNT(*) FROM quotations q WHERE q.rfq_id = r.id AND q.org_id = r.org_id) AS quotes_count
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN users u ON r.sales_assignee_id = u.id
		WHERE r.org_id = ? AND r.status IN ('SUBMITTED', 'PENDING')
		  AND (
		      (r.target_date IS NOT NULL AND r.target_date <= DATE_ADD(NOW(), INTERVAL 3 DAY))
		      OR (r.created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR))
		  )
		ORDER BY r.created_at ASC
		LIMIT 15
	`
	var rows []rfqDeadlineRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		rfqRef := fmt.Sprintf("RFQ-%d", row.ID)
		if row.RFQNumber.Valid && row.RFQNumber.String != "" {
			rfqRef = row.RFQNumber.String
		}

		targetStr := "Unspecified"
		isOverdue := false
		if row.TargetDate != nil {
			targetStr = row.TargetDate.Format("2006-01-02")
			if row.TargetDate.Before(time.Now()) {
				isOverdue = true
			}
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "target_date",
				ObservedValue:  targetStr,
				Description:    "Customer requested target fulfillment date",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "commercial_quotes_count",
				ObservedValue:  row.QuotesCount,
				Description:    "Active commercial proposals generated for inquiry",
			},
			{
				SourceModule:   "rfq",
				SourceEntityID: row.ID,
				SourceRef:      rfqRef,
				FieldName:      "submitted_at",
				ObservedValue:  row.CreatedAt.Format("2006-01-02 15:04"),
				Description:    "Elapsed time since RFQ submission awaiting commercial response",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		custName := "Shipper"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			custName = row.CustomerName.String
		}

		var ownerID *int64
		var ownerName *string
		if row.SalesAssigneeID.Valid {
			id := row.SalesAssigneeID.Int64
			ownerID = &id
		}
		if row.SalesAssigneeName.Valid && row.SalesAssigneeName.String != "" {
			name := row.SalesAssigneeName.String
			ownerName = &name
		}

		priority := PriorityHigh
		risk := RiskHigh
		title := fmt.Sprintf("RFQ %s Approaching Response Deadline (%s)", rfqRef, targetStr)
		if isOverdue {
			priority = PriorityCritical
			risk = RiskCritical
			title = fmt.Sprintf("RFQ %s Response Overdue (Target: %s)", rfqRef, targetStr)
		} else if row.QuotesCount == 0 {
			title = fmt.Sprintf("RFQ %s: No Quotation Prepared After Submission", rfqRef)
		}

		rfqID := row.ID
		cid := row.CustomerID
		ft := DraftTypeRFQClarification

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceRFQ,
			SourceID:           row.ID,
			SourceReference:    rfqRef,
			Title:              title,
			Description:        fmt.Sprintf("Inquiry %s for %s was submitted on %s (Target: %s) with %d quotations issued. Prioritize commercial rate sourcing to avoid response delay risk.", rfqRef, custName, row.CreatedAt.Format("2006-01-02"), targetStr, row.QuotesCount),
			Category:           CategoryRFQ,
			Priority:           priority,
			RiskLevel:          risk,
			Confidence:         "HIGH",
			ConfidenceScore:    0.92,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Source carrier rates, prepare commercial quotation, and assign sales representative.",
			ActionType:         ActionTypeSourceRates,
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "RFQ_RESPONSE_DEADLINE_RISK",
			DedupHash:          computeDedupHash(orgID, SourceRFQ, row.ID, CategoryRFQ, "RFQ_RESPONSE_DEADLINE_RISK"),
			CustomerID:         &cid,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        DraftStatusNotGenerated,
			RFQID:              &rfqID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateQuotationMarginAndPricing(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type quoteMarginRow struct {
		ID               int64          `db:"id"`
		QuotationNumber  string         `db:"quotation_number"`
		RFQID            sql.NullInt64  `db:"rfq_id"`
		CustomerID       sql.NullInt64  `db:"customer_id"`
		CustomerName     sql.NullString `db:"customer_name"`
		Status           string         `db:"status"`
		Currency         string         `db:"currency"`
		TotalAmount      float64        `db:"total_amount"`
		TotalCost        float64        `db:"total_cost"`
		GrossProfit      float64        `db:"gross_profit"`
		GrossMarginPct   float64        `db:"gross_margin_pct"`
		ValidUntil       *time.Time     `db:"valid_until"`
		CreatedAt        time.Time      `db:"created_at"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT q.id, q.quotation_number, q.rfq_id, q.customer_id, q.customer_name,
		       q.status, q.currency, q.total_amount, q.total_cost, q.gross_profit,
		       q.gross_margin_pct, q.valid_until, q.created_at,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM quotations q
		LEFT JOIN customers c ON q.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE q.org_id = ? AND q.status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'SENT')
		  AND (
		      q.gross_margin_pct < 10.0
		      OR q.gross_margin_pct < 0.0
		      OR (q.total_cost = 0.0 AND q.total_amount > 0.0)
		      OR q.total_amount = 0.0
		  )
		ORDER BY q.gross_margin_pct ASC
		LIMIT 15
	`
	var rows []quoteMarginRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		quoteRef := row.QuotationNumber
		if quoteRef == "" {
			quoteRef = fmt.Sprintf("QT-%d", row.ID)
		}

		custName := "Client Account"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			custName = row.CustomerName.String
		}

		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		priority := PriorityHigh
		risk := RiskHigh
		var concernDesc string
		if row.GrossMarginPct < 0 {
			priority = PriorityCritical
			risk = RiskCritical
			concernDesc = fmt.Sprintf("Negative commercial gross margin: %.2f%% (Loss: %s %.2f)", row.GrossMarginPct, row.Currency, -row.GrossProfit)
		} else if row.TotalCost == 0 && row.TotalAmount > 0 {
			priority = PriorityCritical
			risk = RiskCritical
			concernDesc = "Missing cost components (Total buy cost is $0.00 while sell price is positive)"
		} else if row.TotalAmount == 0 {
			priority = PriorityHigh
			risk = RiskMedium
			concernDesc = "Missing commercial price (Total sell amount is $0.00)"
		} else {
			priority = PriorityHigh
			risk = RiskHigh
			concernDesc = fmt.Sprintf("Low gross margin below 10.0%% policy threshold: %.2f%% (%s %.2f profit)", row.GrossMarginPct, row.Currency, row.GrossProfit)
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "gross_margin_pct",
				ObservedValue:  fmt.Sprintf("%.2f%%", row.GrossMarginPct),
				Description:    "Backend-calculated commercial gross margin percentage",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "total_amount_vs_cost",
				ObservedValue:  fmt.Sprintf("Sell: %s %.2f | Buy: %s %.2f", row.Currency, row.TotalAmount, row.Currency, row.TotalCost),
				Description:    "Persisted sell price versus carrier buy cost structure",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "status",
				ObservedValue:  row.Status,
				Description:    "Current commercial proposal lifecycle status",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		quoteID := row.ID
		var rfqID *int64
		if row.RFQID.Valid && row.RFQID.Int64 > 0 {
			id := row.RFQID.Int64
			rfqID = &id
		}
		var cid *int64
		if row.CustomerID.Valid && row.CustomerID.Int64 > 0 {
			id := row.CustomerID.Int64
			cid = &id
		}
		ft := DraftTypeInternalPricingReview

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceQuotation,
			SourceID:           row.ID,
			SourceReference:    quoteRef,
			Title:              fmt.Sprintf("Quotation %s: Pricing & Margin Concern (%s)", quoteRef, concernDesc),
			Description:        fmt.Sprintf("Quotation %s for %s (%s %.2f) exhibits a commercial pricing risk: %s. Human pricing review required prior to customer transmission.", quoteRef, custName, row.Currency, row.TotalAmount, concernDesc),
			Category:           CategoryPricing,
			Priority:           priority,
			RiskLevel:          risk,
			Confidence:         "HIGH",
			ConfidenceScore:    0.98,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Conduct internal pricing review. Verify carrier buy costs, accessorials, and minimum markup before release.",
			ActionType:         ActionTypeRequestPricingApproval,
			Status:             StatusNew,
			RequiresApproval:   true,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "QUOTATION_MARGIN_ALERT",
			DedupHash:          computeDedupHash(orgID, SourceQuotation, row.ID, CategoryPricing, "QUOTATION_MARGIN_ALERT"),
			CustomerID:         cid,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        DraftStatusNotGenerated,
			RFQID:              rfqID,
			QuotationID:        &quoteID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateQuotationPendingApproval(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type quotePendingRow struct {
		ID               int64          `db:"id"`
		QuotationNumber  string         `db:"quotation_number"`
		RFQID            sql.NullInt64  `db:"rfq_id"`
		CustomerID       sql.NullInt64  `db:"customer_id"`
		CustomerName     sql.NullString `db:"customer_name"`
		Status           string         `db:"status"`
		Currency         string         `db:"currency"`
		TotalAmount      float64        `db:"total_amount"`
		TotalCost        float64        `db:"total_cost"`
		GrossMarginPct   float64        `db:"gross_margin_pct"`
		ValidUntil       *time.Time     `db:"valid_until"`
		CreatedAt        time.Time      `db:"created_at"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT q.id, q.quotation_number, q.rfq_id, q.customer_id, q.customer_name,
		       q.status, q.currency, q.total_amount, q.total_cost,
		       q.gross_margin_pct, q.valid_until, q.created_at,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM quotations q
		LEFT JOIN customers c ON q.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE q.org_id = ? AND q.status = 'PENDING_APPROVAL'
		ORDER BY q.created_at ASC
		LIMIT 15
	`
	var rows []quotePendingRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		quoteRef := row.QuotationNumber
		if quoteRef == "" {
			quoteRef = fmt.Sprintf("QT-%d", row.ID)
		}

		custName := "Client"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			custName = row.CustomerName.String
		}

		validStr := "Unspecified"
		if row.ValidUntil != nil {
			validStr = row.ValidUntil.Format("2006-01-02")
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "status",
				ObservedValue:  row.Status,
				Description:    "Quotation submitted and waiting for commercial management approval",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "commercial_terms",
				ObservedValue:  fmt.Sprintf("%s %.2f (%.1f%% margin)", row.Currency, row.TotalAmount, row.GrossMarginPct),
				Description:    "Persisted proposal valuation and margin awaiting review",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "valid_until",
				ObservedValue:  validStr,
				Description:    "Commercial proposal expiry timeline",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		quoteID := row.ID
		var rfqID *int64
		if row.RFQID.Valid && row.RFQID.Int64 > 0 {
			id := row.RFQID.Int64
			rfqID = &id
		}
		var cid *int64
		if row.CustomerID.Valid && row.CustomerID.Int64 > 0 {
			id := row.CustomerID.Int64
			cid = &id
		}
		ft := DraftTypeQuotationApproval

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceQuotation,
			SourceID:           row.ID,
			SourceReference:    quoteRef,
			Title:              fmt.Sprintf("Quotation %s Awaiting Commercial Approval", quoteRef),
			Description:        fmt.Sprintf("Quotation %s for %s (%s %.2f, %.1f%% margin) is pending commercial management approval prior to customer dispatch. Valid until %s.", quoteRef, custName, row.Currency, row.TotalAmount, row.GrossMarginPct, validStr),
			Category:           CategoryQuotation,
			Priority:           PriorityHigh,
			RiskLevel:          RiskHigh,
			Confidence:         "HIGH",
			ConfidenceScore:    0.95,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Review commercial margin and route through HITL approval workflow to release quotation.",
			ActionType:         ActionTypeRequestPricingApproval,
			Status:             StatusNew,
			RequiresApproval:   true,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "QUOTATION_PENDING_APPROVAL",
			DedupHash:          computeDedupHash(orgID, SourceQuotation, row.ID, CategoryQuotation, "QUOTATION_PENDING_APPROVAL"),
			CustomerID:         cid,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        DraftStatusNotGenerated,
			RFQID:              rfqID,
			QuotationID:        &quoteID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateQuotationSentAwaitingResponse(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type quoteSentRow struct {
		ID               int64          `db:"id"`
		QuotationNumber  string         `db:"quotation_number"`
		RFQID            sql.NullInt64  `db:"rfq_id"`
		CustomerID       sql.NullInt64  `db:"customer_id"`
		CustomerName     sql.NullString `db:"customer_name"`
		Status           string         `db:"status"`
		Currency         string         `db:"currency"`
		TotalAmount      float64        `db:"total_amount"`
		GrossMarginPct   float64        `db:"gross_margin_pct"`
		SentAt           *time.Time     `db:"sent_at"`
		ValidUntil       *time.Time     `db:"valid_until"`
		AccountOwnerID   sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName sql.NullString `db:"account_owner_name"`
	}

	query := `
		SELECT q.id, q.quotation_number, q.rfq_id, q.customer_id, q.customer_name,
		       q.status, q.currency, q.total_amount, q.gross_margin_pct,
		       q.sent_at, q.valid_until,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name
		FROM quotations q
		LEFT JOIN customers c ON q.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE q.org_id = ? AND q.status = 'SENT'
		  AND q.sent_at IS NOT NULL AND q.sent_at < DATE_SUB(NOW(), INTERVAL 3 DAY)
		ORDER BY q.sent_at ASC
		LIMIT 15
	`
	var rows []quoteSentRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		quoteRef := row.QuotationNumber
		if quoteRef == "" {
			quoteRef = fmt.Sprintf("QT-%d", row.ID)
		}

		custName := "Client"
		if row.CustomerName.Valid && row.CustomerName.String != "" {
			custName = row.CustomerName.String
		}

		sentStr := "3+ days ago"
		if row.SentAt != nil {
			sentStr = row.SentAt.Format("2006-01-02")
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "sent_at",
				ObservedValue:  sentStr,
				Description:    "Proposal sent to customer over 3 days ago without recorded response",
			},
			{
				SourceModule:   "quotations",
				SourceEntityID: row.ID,
				SourceRef:      quoteRef,
				FieldName:      "total_amount",
				ObservedValue:  fmt.Sprintf("%s %.2f", row.Currency, row.TotalAmount),
				Description:    "Outstanding quoted value awaiting commercial confirmation",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		quoteID := row.ID
		var rfqID *int64
		if row.RFQID.Valid && row.RFQID.Int64 > 0 {
			id := row.RFQID.Int64
			rfqID = &id
		}
		var cid *int64
		if row.CustomerID.Valid && row.CustomerID.Int64 > 0 {
			id := row.CustomerID.Int64
			cid = &id
		}
		ft := DraftTypeQuotationFollowup

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceQuotation,
			SourceID:           row.ID,
			SourceReference:    quoteRef,
			Title:              fmt.Sprintf("Quotation %s Sent But Awaiting Customer Response", quoteRef),
			Description:        fmt.Sprintf("Quotation %s for %s (%s %.2f) was transmitted on %s and has had no customer response. Proactive commercial follow-up recommended.", quoteRef, custName, row.Currency, row.TotalAmount, sentStr),
			Category:           CategoryQuotation,
			Priority:           PriorityMedium,
			RiskLevel:          RiskMedium,
			Confidence:         "HIGH",
			ConfidenceScore:    0.90,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Transmit polite follow-up message to customer procurement contact to verify receipt and assist with booking questions.",
			ActionType:         ActionTypeFollowupQuote,
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "QUOTATION_SENT_AWAITING_RESPONSE",
			DedupHash:          computeDedupHash(orgID, SourceQuotation, row.ID, CategoryQuotation, "QUOTATION_SENT_AWAITING_RESPONSE"),
			CustomerID:         cid,
			CustomerName:       &custName,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        DraftStatusNotGenerated,
			RFQID:              rfqID,
			QuotationID:        &quoteID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateShipmentMilestones(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type msRow struct {
		MilestoneID    int64          `db:"milestone_id"`
		ShipmentID     int64          `db:"shipment_id"`
		MilestoneCode  string         `db:"milestone_code"`
		Description    sql.NullString `db:"description"`
		PlannedDate    time.Time      `db:"planned_date"`
		Status         string         `db:"status"`
		BookingNumber  sql.NullString `db:"booking_number"`
		BookingID      sql.NullInt64  `db:"booking_id"`
		CarrierSCAC    string         `db:"carrier_scac"`
		OriginPort     string         `db:"origin_port"`
		DestPort       string         `db:"destination_port"`
		ShipmentStatus string         `db:"shipment_status"`
		CustomerID     int64          `db:"customer_id"`
		CustomerName   string         `db:"customer_name"`
	}

	query := `
		SELECT sm.id AS milestone_id, sm.shipment_id, sm.milestone_code, sm.description, sm.planned_date, sm.status,
		       s.booking_number, s.booking_id, s.carrier_scac, s.origin_port, s.destination_port, s.status AS shipment_status,
		       COALESCE(q.customer_id, r.customer_id, 0) AS customer_id,
		       COALESCE(q.customer_name, c.name, '') AS customer_name
		FROM shipment_milestones sm
		JOIN shipments s ON s.id = sm.shipment_id AND s.org_id = ?
		LEFT JOIN quotations q ON s.quote_id = q.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE sm.status = 'PLANNED'
		  AND sm.planned_date IS NOT NULL
		  AND sm.planned_date < NOW()
		  AND s.status NOT IN ('DELIVERED', 'COMPLETED', 'CANCELLED')
		ORDER BY sm.planned_date ASC
		LIMIT 25
	`
	var rows []msRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		shipmentRef := fmt.Sprintf("SH-%d", row.ShipmentID)
		if row.BookingNumber.Valid && row.BookingNumber.String != "" {
			shipmentRef = row.BookingNumber.String
		}

		delayDuration := time.Since(row.PlannedDate)
		delayHours := int(delayDuration.Hours())
		delayDays := int(delayDuration.Hours() / 24)

		priority := PriorityHigh
		risk := RiskHigh
		if delayDays >= 3 {
			priority = PriorityCritical
			risk = RiskCritical
		} else if delayDays < 1 {
			priority = PriorityMedium
			risk = RiskMedium
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "milestone_code",
				ObservedValue:  row.MilestoneCode,
				Description:    "Operational transit milestone flagged as delayed",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "planned_date",
				ObservedValue:  row.PlannedDate.Format(time.RFC3339),
				Description:    "Planned schedule date recorded in shipment execution plan",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "delay_duration",
				ObservedValue:  fmt.Sprintf("%d hours (%d days overdue)", delayHours, delayDays),
				Description:    "Milestone is past its planned date without confirmation of completion",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ShipmentID,
				SourceRef:      shipmentRef,
				FieldName:      "carrier_scac",
				ObservedValue:  row.CarrierSCAC,
				Description:    "Operating ocean/air freight carrier responsible for milestone leg",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
			if row.CustomerName != "" {
				cname := row.CustomerName
				custName = &cname
			}
		}

		ft := DraftTypeCarrierClarification
		var bkgID *int64
		if row.BookingID.Valid && row.BookingID.Int64 > 0 {
			bid := row.BookingID.Int64
			bkgID = &bid
		}
		mID := row.MilestoneID
		sID := row.ShipmentID

		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceShipment,
			SourceID:          row.ShipmentID,
			SourceReference:   shipmentRef,
			Title:             fmt.Sprintf("Delayed Milestone: %s on Shipment %s", row.MilestoneCode, shipmentRef),
			Description:       fmt.Sprintf("Milestone %s was planned for %s (%d days ago) on lane %s to %s via carrier %s, but actual completion has not been confirmed. Current shipment status: %s.", row.MilestoneCode, row.PlannedDate.Format("2006-01-02 15:04"), delayDays, row.OriginPort, row.DestPort, row.CarrierSCAC, row.ShipmentStatus),
			Category:          CategoryOperations,
			Priority:          priority,
			RiskLevel:         risk,
			Confidence:        "HIGH",
			ConfidenceScore:   0.92,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: fmt.Sprintf("Query carrier %s operations desk to verify %s milestone status and request revised schedule.", row.CarrierSCAC, row.MilestoneCode),
			ActionType:        ActionTypeRequestCarrierClarification,
			Status:            StatusNew,
			RequiresApproval:  priority == PriorityCritical,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "SHIPMENT_MILESTONE_DELAYED",
			DedupHash:         computeDedupHash(orgID, SourceShipment, row.ShipmentID, CategoryOperations, fmt.Sprintf("DELAYED_MILESTONE_%d", row.MilestoneID)),
			CustomerID:        custID,
			CustomerName:      custName,
			FollowupType:      &ft,
			DraftStatus:       "NONE",
			ShipmentID:        &sID,
			MilestoneID:       &mID,
			BookingID:         bkgID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateShipmentMissingOperationalInfo(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type shipRow struct {
		ID               int64          `db:"id"`
		BookingNumber    sql.NullString `db:"booking_number"`
		BookingID        sql.NullInt64  `db:"booking_id"`
		CarrierSCAC      string         `db:"carrier_scac"`
		Status           string         `db:"status"`
		OriginPort       string         `db:"origin_port"`
		DestPort         string         `db:"destination_port"`
		VesselName       sql.NullString `db:"vessel_name"`
		VoyageNumber     sql.NullString `db:"voyage_number"`
		ContainerNumbers sql.NullString `db:"container_numbers"`
		MBLNumber        sql.NullString `db:"mbl_number"`
		ETD              *time.Time     `db:"etd"`
		ETA              *time.Time     `db:"eta"`
		MilestoneCount   int            `db:"milestone_count"`
		CustomerID       int64          `db:"customer_id"`
		CustomerName     string         `db:"customer_name"`
	}

	query := `
		SELECT s.id, s.booking_number, s.booking_id, s.carrier_scac, s.status, s.origin_port, s.destination_port,
		       s.vessel_name, s.voyage_number, s.container_numbers, s.mbl_number, s.etd, s.eta,
		       (SELECT COUNT(*) FROM shipment_milestones sm WHERE sm.shipment_id = s.id) AS milestone_count,
		       COALESCE(q.customer_id, r.customer_id, 0) AS customer_id,
		       COALESCE(q.customer_name, c.name, '') AS customer_name
		FROM shipments s
		LEFT JOIN quotations q ON s.quote_id = q.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE s.org_id = ?
		  AND s.status NOT IN ('DELIVERED', 'COMPLETED', 'CANCELLED')
		ORDER BY s.id DESC
		LIMIT 50
	`
	var rows []shipRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		shipmentRef := fmt.Sprintf("SH-%d", row.ID)
		if row.BookingNumber.Valid && row.BookingNumber.String != "" {
			shipmentRef = row.BookingNumber.String
		}

		var missingFields []string
		var evidence []EvidenceItem

		if row.CarrierSCAC == "" || row.CarrierSCAC == "UNKNOWN" {
			missingFields = append(missingFields, "Operating Carrier SCAC")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "carrier_scac",
				ObservedValue:  "MISSING",
				Description:    "Carrier code is not assigned",
			})
		}

		if (!row.VesselName.Valid || row.VesselName.String == "") && (!row.MBLNumber.Valid || row.MBLNumber.String == "") {
			missingFields = append(missingFields, "Transport Reference (Vessel / MBL)")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "transport_reference",
				ObservedValue:  "MISSING",
				Description:    "Neither vessel name nor Master B/L number is recorded",
			})
		}

		if row.Status == "IN_TRANSIT" || row.Status == "DEPARTED" {
			cnums := ""
			if row.ContainerNumbers.Valid {
				cnums = strings.TrimSpace(row.ContainerNumbers.String)
			}
			if cnums == "" || cnums == "{}" || cnums == "[]" {
				missingFields = append(missingFields, "Container Identification Numbers")
				evidence = append(evidence, EvidenceItem{
					SourceModule:   "shipments",
					SourceEntityID: row.ID,
					SourceRef:      shipmentRef,
					FieldName:      "container_numbers",
					ObservedValue:  "EMPTY",
					Description:    "Shipment is departed/in-transit but has no container numbers assigned",
				})
			}
		}

		if row.ETA == nil {
			missingFields = append(missingFields, "Estimated Time of Arrival (ETA)")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "eta",
				ObservedValue:  "MISSING",
				Description:    "Official estimated time of arrival date is not recorded",
			})
		}

		if row.MilestoneCount == 0 {
			missingFields = append(missingFields, "Execution Milestones")
			evidence = append(evidence, EvidenceItem{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "milestones",
				ObservedValue:  "0 milestones",
				Description:    "No tracking or execution milestones have been generated for this active shipment",
			})
		}

		if len(missingFields) == 0 {
			continue
		}

		priority := PriorityMedium
		risk := RiskMedium
		if len(missingFields) >= 3 || (row.Status == "IN_TRANSIT" && (!row.ContainerNumbers.Valid || row.ContainerNumbers.String == "")) {
			priority = PriorityHigh
			risk = RiskHigh
		}

		evidenceBytes, _ := json.Marshal(evidence)

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
			if row.CustomerName != "" {
				cname := row.CustomerName
				custName = &cname
			}
		}

		ft := DraftTypeMissingDocumentRequest
		var bkgID *int64
		if row.BookingID.Valid && row.BookingID.Int64 > 0 {
			bid := row.BookingID.Int64
			bkgID = &bid
		}
		sID := row.ID

		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceShipment,
			SourceID:          row.ID,
			SourceReference:   shipmentRef,
			Title:             fmt.Sprintf("Missing Operational Information on Shipment %s", shipmentRef),
			Description:       fmt.Sprintf("Shipment %s has incomplete operational records. Missing: %s. Route: %s to %s. Status: %s.", shipmentRef, strings.Join(missingFields, ", "), row.OriginPort, row.DestPort, row.Status),
			Category:          CategoryOperations,
			Priority:          priority,
			RiskLevel:         risk,
			Confidence:        "HIGH",
			ConfidenceScore:   0.90,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Request missing transport identifiers and complete carrier booking documentation before downstream execution.",
			ActionType:        ActionTypeRequestMissingDocuments,
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "SHIPMENT_MISSING_OPERATIONAL_INFO",
			DedupHash:         computeDedupHash(orgID, SourceShipment, row.ID, CategoryOperations, "MISSING_OPS_INFO"),
			CustomerID:        custID,
			CustomerName:      custName,
			FollowupType:      &ft,
			DraftStatus:       "NONE",
			ShipmentID:        &sID,
			BookingID:         bkgID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateShipmentInactiveTracking(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type trackRow struct {
		ID            int64          `db:"id"`
		BookingNumber sql.NullString `db:"booking_number"`
		BookingID     sql.NullInt64  `db:"booking_id"`
		CarrierSCAC   string         `db:"carrier_scac"`
		Status        string         `db:"status"`
		OriginPort    string         `db:"origin_port"`
		DestPort      string         `db:"destination_port"`
		UpdatedAt     time.Time      `db:"updated_at"`
		CustomerID    int64          `db:"customer_id"`
		CustomerName  string         `db:"customer_name"`
	}

	// Shipments in transit with no update in the last 5 days
	query := `
		SELECT s.id, s.booking_number, s.booking_id, s.carrier_scac, s.status, s.origin_port, s.destination_port, s.updated_at,
		       COALESCE(q.customer_id, r.customer_id, 0) AS customer_id,
		       COALESCE(q.customer_name, c.name, '') AS customer_name
		FROM shipments s
		LEFT JOIN quotations q ON s.quote_id = q.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE s.org_id = ?
		  AND s.status IN ('DEPARTED', 'IN_TRANSIT')
		  AND s.updated_at < DATE_SUB(NOW(), INTERVAL 5 DAY)
		ORDER BY s.updated_at ASC
		LIMIT 20
	`
	var rows []trackRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		shipmentRef := fmt.Sprintf("SH-%d", row.ID)
		if row.BookingNumber.Valid && row.BookingNumber.String != "" {
			shipmentRef = row.BookingNumber.String
		}

		daysSinceUpdate := int(time.Since(row.UpdatedAt).Hours() / 24)

		evidence := []EvidenceItem{
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "last_updated_at",
				ObservedValue:  row.UpdatedAt.Format(time.RFC3339),
				Description:    fmt.Sprintf("Last tracking / status update recorded %d days ago", daysSinceUpdate),
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "tracking_status",
				ObservedValue:  "No recent tracking update was found.",
				Description:    "The available data does not confirm the current physical location.",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "operating_carrier",
				ObservedValue:  row.CarrierSCAC,
				Description:    fmt.Sprintf("Carrier %s has not submitted automatic or manual milestone events", row.CarrierSCAC),
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
			if row.CustomerName != "" {
				cname := row.CustomerName
				custName = &cname
			}
		}

		ft := DraftTypeCarrierClarification
		var bkgID *int64
		if row.BookingID.Valid && row.BookingID.Int64 > 0 {
			bid := row.BookingID.Int64
			bkgID = &bid
		}
		sID := row.ID

		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceShipment,
			SourceID:          row.ID,
			SourceReference:   shipmentRef,
			Title:             fmt.Sprintf("Inactive Tracking on Shipment %s (%d Days)", shipmentRef, daysSinceUpdate),
			Description:       fmt.Sprintf("No recent tracking update was found for shipment %s (%s to %s via %s). Last recorded update was on %s (%d days ago). The available data does not confirm the current physical location.", shipmentRef, row.OriginPort, row.DestPort, row.CarrierSCAC, row.UpdatedAt.Format("2006-01-02"), daysSinceUpdate),
			Category:          CategoryOperations,
			Priority:          PriorityMedium,
			RiskLevel:         RiskMedium,
			Confidence:        "MEDIUM",
			ConfidenceScore:   0.82,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: fmt.Sprintf("Contact carrier %s EDI / tracking desk or local agent to verify current vessel location and container status.", row.CarrierSCAC),
			ActionType:        ActionTypeRequestCarrierClarification,
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "SHIPMENT_INACTIVE_TRACKING",
			DedupHash:         computeDedupHash(orgID, SourceShipment, row.ID, CategoryOperations, "INACTIVE_TRACKING"),
			CustomerID:        custID,
			CustomerName:      custName,
			FollowupType:      &ft,
			DraftStatus:       "NONE",
			ShipmentID:        &sID,
			BookingID:         bkgID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

func (g *generator) evaluateShipmentOverdueDelivery(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type delRow struct {
		ID            int64          `db:"id"`
		BookingNumber sql.NullString `db:"booking_number"`
		BookingID     sql.NullInt64  `db:"booking_id"`
		CarrierSCAC   string         `db:"carrier_scac"`
		Status        string         `db:"status"`
		OriginPort    string         `db:"origin_port"`
		DestPort      string         `db:"destination_port"`
		ETA           time.Time      `db:"eta"`
		CustomerID    int64          `db:"customer_id"`
		CustomerName  string         `db:"customer_name"`
	}

	query := `
		SELECT s.id, s.booking_number, s.booking_id, s.carrier_scac, s.status, s.origin_port, s.destination_port, s.eta,
		       COALESCE(q.customer_id, r.customer_id, 0) AS customer_id,
		       COALESCE(q.customer_name, c.name, '') AS customer_name
		FROM shipments s
		LEFT JOIN quotations q ON s.quote_id = q.id
		LEFT JOIN rfqs r ON s.rfq_id = r.id
		LEFT JOIN customers c ON r.customer_id = c.id
		WHERE s.org_id = ?
		  AND s.eta IS NOT NULL
		  AND s.eta < NOW()
		  AND s.status NOT IN ('DELIVERED', 'COMPLETED', 'CANCELLED')
		ORDER BY s.eta ASC
		LIMIT 20
	`
	var rows []delRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		shipmentRef := fmt.Sprintf("SH-%d", row.ID)
		if row.BookingNumber.Valid && row.BookingNumber.String != "" {
			shipmentRef = row.BookingNumber.String
		}

		daysPastETA := int(time.Since(row.ETA).Hours() / 24)

		evidence := []EvidenceItem{
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "official_eta",
				ObservedValue:  row.ETA.Format(time.RFC3339),
				Description:    "Recorded estimated time of arrival date in shipment contract",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "overdue_duration",
				ObservedValue:  fmt.Sprintf("%d days past ETA", daysPastETA),
				Description:    "Destination arrival date has passed without delivery confirmation",
			},
			{
				SourceModule:   "shipments",
				SourceEntityID: row.ID,
				SourceRef:      shipmentRef,
				FieldName:      "current_status",
				ObservedValue:  row.Status,
				Description:    "Shipment remains in active transit status instead of delivered",
			},
		}
		evidenceBytes, _ := json.Marshal(evidence)

		var custID *int64
		var custName *string
		if row.CustomerID > 0 {
			cid := row.CustomerID
			custID = &cid
			if row.CustomerName != "" {
				cname := row.CustomerName
				custName = &cname
			}
		}

		ft := DraftTypeDeliveryClarification
		var bkgID *int64
		if row.BookingID.Valid && row.BookingID.Int64 > 0 {
			bid := row.BookingID.Int64
			bkgID = &bid
		}
		sID := row.ID

		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceShipment,
			SourceID:          row.ID,
			SourceReference:   shipmentRef,
			Title:             fmt.Sprintf("Delivery Date Passed: Shipment %s (%d Days Overdue)", shipmentRef, daysPastETA),
			Description:       fmt.Sprintf("Shipment %s had an official ETA of %s, which passed %d days ago without confirmed delivery. Current status: %s (Carrier %s, %s to %s).", shipmentRef, row.ETA.Format("2006-01-02"), daysPastETA, row.Status, row.CarrierSCAC, row.OriginPort, row.DestPort),
			Category:          CategoryOperations,
			Priority:          PriorityCritical,
			RiskLevel:         RiskCritical,
			Confidence:        "HIGH",
			ConfidenceScore:   0.95,
			EvidenceJSON:      string(evidenceBytes),
			RecommendedAction: "Clarify destination discharge and customs release status with destination port agent and advise consignee with verified update.",
			ActionType:        ActionTypeInvestigateException,
			Status:            StatusNew,
			RequiresApproval:  true,
			CorrelationID:     correlationID,
			CreatedBy:         "SYSTEM",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "SHIPMENT_OVERDUE_DELIVERY",
			DedupHash:         computeDedupHash(orgID, SourceShipment, row.ID, CategoryOperations, "OVERDUE_DELIVERY"),
			CustomerID:        custID,
			CustomerName:      custName,
			FollowupType:      &ft,
			DraftStatus:       "NONE",
			ShipmentID:        &sID,
			BookingID:         bkgID,
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

// 18. Rule: UPCOMING_COLLECTION_PRIORITY (Phase 2 Task 2.5)
func (g *generator) evaluateUpcomingCollectionPriorities(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	type upcomingRow struct {
		ID                  int64          `db:"id"`
		InvoiceNumber       string         `db:"invoice_number"`
		CustomerID          int64          `db:"customer_id"`
		CustomerName        string         `db:"customer_name"`
		Currency            string         `db:"currency"`
		TotalAmount         float64        `db:"total_amount"`
		BalanceDue          float64        `db:"balance_due"`
		Status              string         `db:"status"`
		DueDate             *time.Time     `db:"due_date"`
		AccountOwnerID      sql.NullInt64  `db:"account_owner_id"`
		AccountOwnerName    sql.NullString `db:"account_owner_name"`
		PriorLateCount      int            `db:"prior_late_count"`
	}

	query := `
		SELECT ci.id, ci.invoice_number, ci.customer_id, ci.customer_name, ci.currency,
		       ci.total_amount, ci.balance_due, ci.status, ci.due_date,
		       c.account_owner_id, CONCAT(u.first_name, ' ', u.last_name) AS account_owner_name,
		       (SELECT COUNT(*) FROM customer_invoices ci2 WHERE ci2.org_id = ci.org_id AND ci2.customer_id = ci.customer_id AND (ci2.status = 'Overdue' OR (ci2.due_date < CURDATE() AND ci2.balance_due > 0))) AS prior_late_count
		FROM customer_invoices ci
		JOIN customers c ON ci.customer_id = c.id
		LEFT JOIN users u ON c.account_owner_id = u.id
		WHERE ci.org_id = ?
		  AND ci.due_date >= CURDATE()
		  AND ci.due_date <= DATE_ADD(CURDATE(), INTERVAL 7 DAY)
		  AND ci.balance_due > 0
		  AND ci.status NOT IN ('Paid', 'Void', 'Cancelled', 'Disputed')
		ORDER BY ci.due_date ASC, ci.balance_due DESC
		LIMIT 15
	`

	var rows []upcomingRow
	if err := g.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return nil, err
	}

	var candidates []*Recommendation
	for _, row := range rows {
		sourceRef := row.InvoiceNumber
		if sourceRef == "" {
			sourceRef = fmt.Sprintf("INV-%d", row.ID)
		}

		daysUntil := 0
		dueStr := "Due soon"
		if row.DueDate != nil {
			dueStr = row.DueDate.Format("2006-01-02")
			calcDays := int(time.Until(*row.DueDate).Hours() / 24)
			if calcDays >= 0 {
				daysUntil = calcDays + 1
			}
		}

		evidence := []EvidenceItem{
			{
				SourceModule:   "invoices",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "due_date",
				ObservedValue:  fmt.Sprintf("%s (%d days remaining)", dueStr, daysUntil),
				Description:    "Invoice maturity date approaching within the 7-day collection window",
			},
			{
				SourceModule:   "invoices",
				SourceEntityID: row.ID,
				SourceRef:      sourceRef,
				FieldName:      "balance_due",
				ObservedValue:  fmt.Sprintf("%s %.2f", row.Currency, row.BalanceDue),
				Description:    "Payable balance awaiting settlement remittance",
			},
			{
				SourceModule:   "customers",
				SourceEntityID: row.CustomerID,
				SourceRef:      row.CustomerName,
				FieldName:      "customer_id",
				ObservedValue:  row.CustomerID,
				Description:    "Debtor account associated with invoice",
			},
		}

		priority := PriorityMedium
		risk := RiskLow
		if row.BalanceDue > 15000 || row.PriorLateCount > 0 {
			priority = PriorityHigh
			risk = RiskMedium
			if row.PriorLateCount > 0 {
				evidence = append(evidence, EvidenceItem{
					SourceModule:   "customers",
					SourceEntityID: row.CustomerID,
					SourceRef:      row.CustomerName,
					FieldName:      "prior_late_count",
					ObservedValue:  row.PriorLateCount,
					Description:    fmt.Sprintf("Customer has %d overdue items in payment history", row.PriorLateCount),
				})
			}
		}

		custID := row.CustomerID
		custName := row.CustomerName
		invID := row.ID
		ft := DraftTypeFirstPaymentReminder
		var ownerID *int64
		var ownerName *string
		if row.AccountOwnerID.Valid {
			id := row.AccountOwnerID.Int64
			ownerID = &id
		}
		if row.AccountOwnerName.Valid && row.AccountOwnerName.String != "" {
			name := row.AccountOwnerName.String
			ownerName = &name
		}

		evidenceBytes, _ := json.Marshal(evidence)

		cand := &Recommendation{
			OrgID:              orgID,
			SourceType:         SourceInvoice,
			SourceID:           row.ID,
			SourceReference:    sourceRef,
			Title:              fmt.Sprintf("Upcoming Due (%d Days): %s - Invoice %s (%s %.2f)", daysUntil, row.CustomerName, sourceRef, row.Currency, row.BalanceDue),
			Description:        fmt.Sprintf("Invoice %s for %s (%s %.2f, due %s, %d days remaining) has no recorded settlement. Recommend proactive courtesy reminder prior to due date.", sourceRef, row.CustomerName, row.Currency, row.TotalAmount, dueStr, daysUntil),
			Category:           CategoryFinance,
			Priority:           priority,
			RiskLevel:          risk,
			Confidence:         "HIGH",
			ConfidenceScore:    0.92,
			EvidenceJSON:       string(evidenceBytes),
			RecommendedAction:  "Transmit pre-due courtesy payment schedule reminder to customer accounts payable desk.",
			ActionType:         ActionTypePreparePaymentReminder,
			Status:             StatusNew,
			RequiresApproval:   false,
			CorrelationID:      correlationID,
			CreatedBy:          "SYSTEM",
			GeneratedBy:        "DETERMINISTIC_RULES",
			RuleApplied:        "UPCOMING_COLLECTION_PRIORITY",
			DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "UPCOMING_COLLECTION_PRIORITY"),
			CustomerID:         &custID,
			CustomerName:       &custName,
			InvoiceID:          &invID,
			FollowupType:       &ft,
			SuggestedOwnerID:   ownerID,
			SuggestedOwnerName: ownerName,
			DraftStatus:        "NONE",
		}
		candidates = append(candidates, cand)
	}

	return candidates, nil
}

// 19. Rule: PAYMENT_RISK_SIGNALS (Phase 2 Task 2.5)
func (g *generator) evaluatePaymentRiskSignals(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	var candidates []*Recommendation

	// 19a. Stalled Partial Payment
	type partialRow struct {
		ID            int64      `db:"id"`
		InvoiceNumber string     `db:"invoice_number"`
		CustomerID    int64      `db:"customer_id"`
		CustomerName  string     `db:"customer_name"`
		Currency      string     `db:"currency"`
		TotalAmount   float64    `db:"total_amount"`
		BalanceDue    float64    `db:"balance_due"`
		Status        string     `db:"status"`
		DueDate       *time.Time `db:"due_date"`
	}

	partialQuery := `
		SELECT id, invoice_number, customer_id, customer_name, currency, total_amount, balance_due, status, due_date
		FROM customer_invoices
		WHERE org_id = ?
		  AND balance_due > 0
		  AND balance_due < total_amount
		  AND (status = 'Overdue' OR (due_date < CURDATE() AND due_date IS NOT NULL))
		ORDER BY balance_due DESC
		LIMIT 10
	`
	var pRows []partialRow
	if err := g.db.SelectContext(ctx, &pRows, partialQuery, orgID); err == nil {
		for _, row := range pRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			paidAmount := row.TotalAmount - row.BalanceDue
			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "partial_settlement",
					ObservedValue:  fmt.Sprintf("Paid: %s %.2f, Remaining: %s %.2f of %s %.2f", row.Currency, paidAmount, row.Currency, row.BalanceDue, row.Currency, row.TotalAmount),
					Description:    "Partial payment was received but remainder is overdue without further remittance",
				},
				{
					SourceModule:   "customers",
					SourceEntityID: row.CustomerID,
					SourceRef:      row.CustomerName,
					FieldName:      "customer_id",
					ObservedValue:  row.CustomerID,
					Description:    "Debtor account with stalled partial payment",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID
			ft := DraftTypePaymentReconciliationQuery

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Stalled Partial Payment: Invoice %s (%s %.2f Remaining Past Due)", sourceRef, row.Currency, row.BalanceDue),
				Description:        fmt.Sprintf("Invoice %s for %s received a partial remittance of %s %.2f, but the remaining balance of %s %.2f is overdue. Recommend payment verification and reconciliation inquiry.", sourceRef, row.CustomerName, row.Currency, paidAmount, row.Currency, row.BalanceDue),
				Category:           CategoryFinance,
				Priority:           PriorityHigh,
				RiskLevel:          RiskHigh,
				Confidence:         "HIGH",
				ConfidenceScore:    0.94,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Contact customer accounts payable to verify remaining balance payment schedule and check bank reconciliation.",
				ActionType:         ActionTypeVerifyPaymentStatus,
				Status:             StatusNew,
				RequiresApproval:   false,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "PAYMENT_RISK_SIGNALS",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "STALLED_PARTIAL_PAYMENT"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				FollowupType:       &ft,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 19b. Unresolved Invoice Disputes
	type disputeRow struct {
		ID            int64      `db:"id"`
		InvoiceNumber string     `db:"invoice_number"`
		CustomerID    int64      `db:"customer_id"`
		CustomerName  string     `db:"customer_name"`
		Currency      string     `db:"currency"`
		TotalAmount   float64    `db:"total_amount"`
		BalanceDue    float64    `db:"balance_due"`
		Status        string     `db:"status"`
	}

	disputeQuery := `
		SELECT id, invoice_number, customer_id, customer_name, currency, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ?
		  AND (status = 'Disputed' OR status = 'Dispute')
		ORDER BY balance_due DESC
		LIMIT 10
	`
	var dRows []disputeRow
	if err := g.db.SelectContext(ctx, &dRows, disputeQuery, orgID); err == nil {
		for _, row := range dRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "status",
					ObservedValue:  row.Status,
					Description:    "Invoice is under formal commercial dispute blocking regular payment",
				},
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "disputed_balance",
					ObservedValue:  fmt.Sprintf("%s %.2f", row.Currency, row.BalanceDue),
					Description:    "Total disputed receivables value at risk",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID
			ft := DraftTypeFinanceInternalEscalation

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Unresolved Commercial Dispute: Invoice %s (%s %.2f)", sourceRef, row.Currency, row.BalanceDue),
				Description:        fmt.Sprintf("Invoice %s for %s (%s %.2f) is marked Disputed. Requires commercial review and internal finance escalation before collection actions can proceed.", sourceRef, row.CustomerName, row.Currency, row.BalanceDue),
				Category:           CategoryFinance,
				Priority:           PriorityCritical,
				RiskLevel:          RiskCritical,
				Confidence:         "HIGH",
				ConfidenceScore:    0.98,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Initiate commercial finance dispute review and coordinate with sales representative.",
				ActionType:         ActionTypeReviewInvoiceDispute,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "PAYMENT_RISK_SIGNALS",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "UNRESOLVED_DISPUTE"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				FollowupType:       &ft,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 19c. Status & Balance Discrepancies
	type mismatchRow struct {
		ID            int64   `db:"id"`
		InvoiceNumber string  `db:"invoice_number"`
		CustomerID    int64   `db:"customer_id"`
		CustomerName  string  `db:"customer_name"`
		Currency      string  `db:"currency"`
		TotalAmount   float64 `db:"total_amount"`
		BalanceDue    float64 `db:"balance_due"`
		Status        string  `db:"status"`
	}

	mismatchQuery := `
		SELECT id, invoice_number, customer_id, customer_name, currency, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ?
		  AND (
		      (status = 'Paid' AND balance_due > 0.01)
		   OR (status IN ('Overdue', 'Unpaid') AND balance_due <= 0.001)
		  )
		ORDER BY id DESC
		LIMIT 10
	`
	var mRows []mismatchRow
	if err := g.db.SelectContext(ctx, &mRows, mismatchQuery, orgID); err == nil {
		for _, row := range mRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			mismatchDesc := ""
			if row.Status == "Paid" && row.BalanceDue > 0.01 {
				mismatchDesc = fmt.Sprintf("Invoice marked 'Paid' but carries unpaid balance %s %.2f", row.Currency, row.BalanceDue)
			} else {
				mismatchDesc = fmt.Sprintf("Invoice marked '%s' but carries zero balance (%.2f)", row.Status, row.BalanceDue)
			}

			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "status_balance_mismatch",
					ObservedValue:  fmt.Sprintf("Status: %s, BalanceDue: %.2f", row.Status, row.BalanceDue),
					Description:    mismatchDesc,
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID
			ft := DraftTypeFinanceInternalEscalation

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Status/Balance Mismatch: Invoice %s (%s)", sourceRef, row.Status),
				Description:        fmt.Sprintf("Invoice %s for %s has contradictory ledger state: %s. Requires finance ledger audit.", sourceRef, row.CustomerName, mismatchDesc),
				Category:           CategoryFinance,
				Priority:           PriorityCritical,
				RiskLevel:          RiskHigh,
				Confidence:         "HIGH",
				ConfidenceScore:    0.99,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Reconcile payment records and correct settlement ledger status.",
				ActionType:         ActionTypeReconcilePaymentInfo,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "PAYMENT_RISK_SIGNALS",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryFinance, "STATUS_BALANCE_MISMATCH"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				FollowupType:       &ft,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 19d. Compounding Customer Overdue Exposure (Multi-invoice customer default risk)
	type multiOverdueRow struct {
		CustomerID   int64   `db:"customer_id"`
		CustomerName string  `db:"customer_name"`
		OverdueCount int     `db:"overdue_count"`
		TotalOverdue float64 `db:"total_overdue"`
		Currency     string  `db:"currency"`
	}

	multiOverdueQuery := `
		SELECT c.id as customer_id, c.name as customer_name,
		       COUNT(ci.id) as overdue_count,
		       SUM(ci.balance_due) as total_overdue,
		       COALESCE(MAX(ci.currency), 'USD') as currency
		FROM customer_invoices ci
		JOIN customers c ON ci.customer_id = c.id
		WHERE ci.org_id = ?
		  AND (ci.status = 'Overdue' OR (ci.due_date < CURDATE() AND ci.balance_due > 0))
		GROUP BY c.id, c.name
		HAVING overdue_count >= 2
		ORDER BY total_overdue DESC
		LIMIT 10
	`
	var moRows []multiOverdueRow
	if err := g.db.SelectContext(ctx, &moRows, multiOverdueQuery, orgID); err == nil {
		for _, row := range moRows {
			evidence := []EvidenceItem{
				{
					SourceModule:   "customers",
					SourceEntityID: row.CustomerID,
					SourceRef:      row.CustomerName,
					FieldName:      "compounding_overdue",
					ObservedValue:  fmt.Sprintf("%d overdue invoices totaling %s %.2f", row.OverdueCount, row.Currency, row.TotalOverdue),
					Description:    "Debtor account has multiple concurrent delinquent invoices accumulating significant AR exposure",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			ft := DraftTypeCustomerStatementSummary

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceCustomer,
				SourceID:           row.CustomerID,
				SourceReference:    row.CustomerName,
				Title:              fmt.Sprintf("Compounding Credit Risk: Customer %s (%d Overdue Invoices, %s %.2f)", row.CustomerName, row.OverdueCount, row.Currency, row.TotalOverdue),
				Description:        fmt.Sprintf("Customer %s has %d separate overdue invoices with combined delinquent exposure of %s %.2f. Suggest credit hold evaluation and consolidated statement review.", row.CustomerName, row.OverdueCount, row.Currency, row.TotalOverdue),
				Category:           CategoryFinance,
				Priority:           PriorityCritical,
				RiskLevel:          RiskCritical,
				Confidence:         "HIGH",
				ConfidenceScore:    0.96,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Escalate to senior credit controller, evaluate operational credit limits, and issue consolidated statement.",
				ActionType:         ActionTypeEscalateCollectionRisk,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "PAYMENT_RISK_SIGNALS",
				DedupHash:          computeDedupHash(orgID, SourceCustomer, row.CustomerID, CategoryFinance, "COMPOUNDING_CUSTOMER_RISK"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				FollowupType:       &ft,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	return candidates, nil
}

// 20. Rule: INVOICE_DATA_QUALITY_ISSUE (Phase 2 Task 2.5)
func (g *generator) evaluateInvoiceDataQualityIssues(ctx context.Context, orgID int64, correlationID string) ([]*Recommendation, error) {
	var candidates []*Recommendation

	// 20a. Missing Due Date on open invoices
	type missingDateRow struct {
		ID            int64   `db:"id"`
		InvoiceNumber string  `db:"invoice_number"`
		CustomerID    int64   `db:"customer_id"`
		CustomerName  string  `db:"customer_name"`
		Currency      string  `db:"currency"`
		TotalAmount   float64 `db:"total_amount"`
		BalanceDue    float64 `db:"balance_due"`
		Status        string  `db:"status"`
	}

	missingDateQuery := `
		SELECT id, invoice_number, customer_id, customer_name, currency, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ?
		  AND due_date IS NULL
		  AND status NOT IN ('Paid', 'Void', 'Cancelled')
		ORDER BY id DESC
		LIMIT 10
	`
	var mdRows []missingDateRow
	if err := g.db.SelectContext(ctx, &mdRows, missingDateQuery, orgID); err == nil {
		for _, row := range mdRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "due_date",
					ObservedValue:  "NULL",
					Description:    "Invoice has no contractual payment due date defined",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Data Quality: Invoice %s Missing Maturity Due Date", sourceRef),
				Description:        fmt.Sprintf("Active invoice %s (%s %.2f, status: %s) has no due date recorded. Without a due date, aging and payment reminders cannot function reliably.", sourceRef, row.Currency, row.TotalAmount, row.Status),
				Category:           CategoryDataQuality,
				Priority:           PriorityHigh,
				RiskLevel:          RiskMedium,
				Confidence:         "HIGH",
				ConfidenceScore:    0.99,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Update invoice record with customer contractual payment term due date.",
				ActionType:         ActionTypeRequestFinanceReview,
				Status:             StatusNew,
				RequiresApproval:   false,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "INVOICE_DATA_QUALITY_ISSUE",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryDataQuality, "MISSING_DUE_DATE"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 20b. Negative or impossible balances
	type invalidBalanceRow struct {
		ID            int64   `db:"id"`
		InvoiceNumber string  `db:"invoice_number"`
		CustomerID    int64   `db:"customer_id"`
		CustomerName  string  `db:"customer_name"`
		Currency      string  `db:"currency"`
		TotalAmount   float64 `db:"total_amount"`
		BalanceDue    float64 `db:"balance_due"`
		Status        string  `db:"status"`
	}

	invalidBalanceQuery := `
		SELECT id, invoice_number, customer_id, customer_name, currency, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ?
		  AND (balance_due < 0 OR balance_due > total_amount OR total_amount <= 0)
		ORDER BY id DESC
		LIMIT 10
	`
	var ibRows []invalidBalanceRow
	if err := g.db.SelectContext(ctx, &ibRows, invalidBalanceQuery, orgID); err == nil {
		for _, row := range ibRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			reason := "Invalid balance configuration"
			if row.BalanceDue < 0 {
				reason = fmt.Sprintf("Negative balance due (%.2f)", row.BalanceDue)
			} else if row.BalanceDue > row.TotalAmount {
				reason = fmt.Sprintf("Balance due (%.2f) exceeds total invoice amount (%.2f)", row.BalanceDue, row.TotalAmount)
			} else if row.TotalAmount <= 0 {
				reason = fmt.Sprintf("Non-positive invoice total amount (%.2f)", row.TotalAmount)
			}

			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "financial_totals",
					ObservedValue:  fmt.Sprintf("Total: %.2f, BalanceDue: %.2f", row.TotalAmount, row.BalanceDue),
					Description:    reason,
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			custID := row.CustomerID
			custName := row.CustomerName
			invID := row.ID

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Data Quality: Invoice %s Impossible Financial Totals", sourceRef),
				Description:        fmt.Sprintf("Invoice %s for %s contains an impossible financial balance: %s. Requires ledger correction.", sourceRef, row.CustomerName, reason),
				Category:           CategoryDataQuality,
				Priority:           PriorityCritical,
				RiskLevel:          RiskCritical,
				Confidence:         "HIGH",
				ConfidenceScore:    0.99,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Audit invoice ledger balance and adjust line item total or payment allocation.",
				ActionType:         ActionTypeRequestFinanceReview,
				Status:             StatusNew,
				RequiresApproval:   true,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "INVOICE_DATA_QUALITY_ISSUE",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryDataQuality, "IMPOSSIBLE_TOTALS"),
				CustomerID:         &custID,
				CustomerName:       &custName,
				InvoiceID:          &invID,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	// 20c. Missing Customer Association
	type missingCustRow struct {
		ID            int64   `db:"id"`
		InvoiceNumber string  `db:"invoice_number"`
		Currency      string  `db:"currency"`
		TotalAmount   float64 `db:"total_amount"`
		BalanceDue    float64 `db:"balance_due"`
		Status        string  `db:"status"`
	}

	missingCustQuery := `
		SELECT id, invoice_number, currency, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ?
		  AND (customer_id IS NULL OR customer_id = 0)
		ORDER BY id DESC
		LIMIT 10
	`
	var mcRows []missingCustRow
	if err := g.db.SelectContext(ctx, &mcRows, missingCustQuery, orgID); err == nil {
		for _, row := range mcRows {
			sourceRef := row.InvoiceNumber
			if sourceRef == "" {
				sourceRef = fmt.Sprintf("INV-%d", row.ID)
			}
			evidence := []EvidenceItem{
				{
					SourceModule:   "invoices",
					SourceEntityID: row.ID,
					SourceRef:      sourceRef,
					FieldName:      "customer_id",
					ObservedValue:  "Unassigned / 0",
					Description:    "Invoice has no associated customer account entity",
				},
			}
			evidenceBytes, _ := json.Marshal(evidence)
			invID := row.ID

			cand := &Recommendation{
				OrgID:              orgID,
				SourceType:         SourceInvoice,
				SourceID:           row.ID,
				SourceReference:    sourceRef,
				Title:              fmt.Sprintf("Data Quality: Invoice %s Has No Customer Association", sourceRef),
				Description:        fmt.Sprintf("Invoice %s (%s %.2f) lacks an associated customer account. It cannot be billed or collected without assigning a debtor account.", sourceRef, row.Currency, row.TotalAmount),
				Category:           CategoryDataQuality,
				Priority:           PriorityHigh,
				RiskLevel:          RiskHigh,
				Confidence:         "HIGH",
				ConfidenceScore:    0.99,
				EvidenceJSON:       string(evidenceBytes),
				RecommendedAction:  "Link invoice to valid customer account in the CRM directory.",
				ActionType:         ActionTypeRequestFinanceReview,
				Status:             StatusNew,
				RequiresApproval:   false,
				CorrelationID:      correlationID,
				CreatedBy:          "SYSTEM",
				GeneratedBy:        "DETERMINISTIC_RULES",
				RuleApplied:        "INVOICE_DATA_QUALITY_ISSUE",
				DedupHash:          computeDedupHash(orgID, SourceInvoice, row.ID, CategoryDataQuality, "MISSING_CUSTOMER_LINK"),
				InvoiceID:          &invID,
				DraftStatus:        "NONE",
			}
			candidates = append(candidates, cand)
		}
	}

	return candidates, nil
}


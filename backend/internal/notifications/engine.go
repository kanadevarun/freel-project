package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Engine interface {
	Evaluate(ctx context.Context, orgID int64) (int, error)
}

type engineImpl struct {
	db   *sqlx.DB
	repo Repository
}

func NewEngine(db *sqlx.DB, repo Repository) Engine {
	return &engineImpl{
		db:   db,
		repo: repo,
	}
}

func makeDedupHash(orgID int64, module, recordType, recordID, notifType string) string {
	raw := fmt.Sprintf("%d:%s:%s:%s:%s", orgID, module, recordType, recordID, notifType)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (e *engineImpl) Evaluate(ctx context.Context, orgID int64) (int, error) {
	createdCount := 0

	// 1. Evaluate Pending Approvals
	c, err := e.evaluateApprovals(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating approvals: %v", err)
	}
	createdCount += c

	// 2. Evaluate High-Priority & Assigned Recommendations
	c, err = e.evaluateRecommendations(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating recommendations: %v", err)
	}
	createdCount += c

	// 3. Evaluate Failed Automation Executions
	c, err = e.evaluateAutomations(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating automations: %v", err)
	}
	createdCount += c

	// 4. Evaluate Shipment Tracking Alerts
	c, err = e.evaluateShipmentAlerts(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating shipment alerts: %v", err)
	}
	createdCount += c

	// 5. Evaluate Overdue Customer Invoices
	c, err = e.evaluateInvoices(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating invoices: %v", err)
	}
	createdCount += c

	// 6. Evaluate Rate Contracts Expiry & Document Reviews
	c, err = e.evaluateContracts(ctx, orgID)
	if err != nil {
		log.Printf("[Notifications Engine] Error evaluating contracts: %v", err)
	}
	createdCount += c

	return createdCount, nil
}

func (e *engineImpl) evaluateApprovals(ctx context.Context, orgID int64) (int, error) {
	type approvalRow struct {
		ID           int64     `db:"id"`
		Title        string    `db:"title"`
		Type         string    `db:"type"`
		Priority     string    `db:"priority"`
		RiskLevel    string    `db:"risk_level"`
		Status       string    `db:"status"`
		CreatedAt    time.Time `db:"created_at"`
		Department   *string   `db:"department"`
		RequestedBy  string    `db:"requested_by_name"`
		AssignedTo   *string   `db:"assigned_to"`
	}

	query := `
		SELECT id, title, type, priority, risk_level, status, created_at, department, requested_by_name, assigned_to
		FROM approval_requests
		WHERE org_id = ? AND status = 'Pending'
		ORDER BY created_at ASC
	`
	var rows []approvalRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

	for _, app := range rows {
		dedup := makeDedupHash(orgID, ModuleApprovals, "APPROVAL", fmt.Sprintf("%d", app.ID), TypePendingApproval)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		hoursPending := int(now.Sub(app.CreatedAt).Hours())
		actionURL := fmt.Sprintf("/dashboard/approvals?id=%d", app.ID)

		if existing == nil {
			sev := SeverityMedium
			if app.Priority == "HIGH" || app.RiskLevel == "HIGH" || app.RiskLevel == "CRITICAL" {
				sev = SeverityHigh
			}
			if hoursPending >= 48 {
				sev = SeverityCritical
			}

			notif := &Notification{
				OrgID:            orgID,
				SourceModule:     ModuleApprovals,
				SourceRecordType: "APPROVAL",
				SourceRecordID:   fmt.Sprintf("%d", app.ID),
				ApprovalID:       &app.ID,
				NotificationType: TypePendingApproval,
				Title:            fmt.Sprintf("Approval Required: %s", app.Title),
				Message:          fmt.Sprintf("A %s request submitted by %s requires review and human-in-the-loop sign-off.", app.Type, app.RequestedBy),
				Severity:         sev,
				Priority:         app.Priority,
				Status:           StatusActive,
				ActionRequired:   true,
				ActionURL:        &actionURL,
				DedupHash:        dedup,
				CorrelationID:    fmt.Sprintf("corr-appr-%d", app.ID),
				IsEscalated:      hoursPending >= 48,
				EscalationLevel:  0,
			}
			if hoursPending >= 48 {
				notif.EscalationLevel = 1
				notif.EscalatedAt = &now
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
				if hoursPending >= 48 {
					_ = e.repo.Escalate(ctx, orgID, notif.ID, SeverityCritical, "Approval pending beyond 48-hour SLA window", "TIME_THRESHOLD", 48, notif.CorrelationID)
				}
			}
		} else {
			// Check escalation if pending > 48h and not yet escalated
			if hoursPending >= 48 && !existing.IsEscalated {
				_ = e.repo.Escalate(ctx, orgID, existing.ID, SeverityCritical, "Approval pending beyond 48-hour SLA window", "TIME_THRESHOLD", 48, existing.CorrelationID)
			}
		}
	}

	return created, nil
}

func (e *engineImpl) evaluateRecommendations(ctx context.Context, orgID int64) (int, error) {
	type recRow struct {
		ID           int64     `db:"id"`
		Title        string    `db:"title"`
		Category     string    `db:"category"`
		Priority     string    `db:"priority"`
		RiskLevel    string    `db:"risk_level"`
		Status       string    `db:"status"`
		AssigneeID   *int64    `db:"assignee_id"`
		AssigneeName *string   `db:"assignee_name"`
		CreatedAt    time.Time `db:"created_at"`
	}

	query := `
		SELECT id, title, category, priority, risk_level, status, assignee_id, assignee_name, created_at
		FROM ai_recommendations
		WHERE org_id = ? AND status IN ('new', 'active') 
		  AND (priority IN ('HIGH', 'URGENT', 'CRITICAL') OR risk_level IN ('HIGH', 'CRITICAL') OR assignee_id IS NOT NULL)
		ORDER BY created_at ASC
	`
	var rows []recRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

	for _, rec := range rows {
		notifType := TypeHighPriorityRecommendation
		if rec.AssigneeID != nil && *rec.AssigneeID > 0 {
			notifType = TypeAssignedRecommendation
		}

		dedup := makeDedupHash(orgID, ModuleRecommendations, "RECOMMENDATION", fmt.Sprintf("%d", rec.ID), notifType)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		hoursActive := int(now.Sub(rec.CreatedAt).Hours())
		actionURL := fmt.Sprintf("/dashboard/recommendations?id=%d", rec.ID)

		if existing == nil {
			sev := SeverityHigh
			if rec.Priority == "URGENT" || rec.RiskLevel == "CRITICAL" {
				sev = SeverityCritical
			}
			if notifType == TypeAssignedRecommendation && sev != SeverityCritical {
				sev = SeverityMedium
			}

			notif := &Notification{
				OrgID:            orgID,
				UserID:           rec.AssigneeID,
				SourceModule:     ModuleRecommendations,
				SourceRecordType: "RECOMMENDATION",
				SourceRecordID:   fmt.Sprintf("%d", rec.ID),
				RecommendationID: &rec.ID,
				NotificationType: notifType,
				Title:            fmt.Sprintf("AI Insight: %s", rec.Title),
				Message:          fmt.Sprintf("A %s priority recommendation in %s requires review.", rec.Priority, rec.Category),
				Severity:         sev,
				Priority:         rec.Priority,
				Status:           StatusActive,
				ActionRequired:   true,
				ActionURL:        &actionURL,
				DedupHash:        dedup,
				CorrelationID:    fmt.Sprintf("corr-rec-%d", rec.ID),
			}

			if hoursActive >= 24 && sev == SeverityHigh {
				notif.IsEscalated = true
				notif.EscalationLevel = 1
				notif.EscalatedAt = &now
				notif.Severity = SeverityCritical
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
				if notif.IsEscalated {
					_ = e.repo.Escalate(ctx, orgID, notif.ID, SeverityCritical, "High priority AI recommendation unresolved after 24 hours", "TIME_THRESHOLD", 24, notif.CorrelationID)
				}
			}
		} else {
			// Escalate if > 24 hours unresolved and not escalated
			if hoursActive >= 24 && !existing.IsEscalated && existing.Severity == SeverityHigh {
				_ = e.repo.Escalate(ctx, orgID, existing.ID, SeverityCritical, "High priority AI recommendation unresolved after 24 hours", "TIME_THRESHOLD", 24, existing.CorrelationID)
			}
		}
	}

	return created, nil
}

func (e *engineImpl) evaluateAutomations(ctx context.Context, orgID int64) (int, error) {
	type autoRow struct {
		ID           int64     `db:"id"`
		AutomationID int64     `db:"automation_id"`
		Status       string    `db:"status"`
		ErrorMessage *string   `db:"error_message"`
		RetryCount   int       `db:"retry_count"`
		CreatedAt    time.Time `db:"created_at"`
	}

	query := `
		SELECT id, automation_id, status, error_message, retry_count, created_at
		FROM ai_automation_executions
		WHERE org_id = ? AND status = 'FAILED'
		ORDER BY created_at DESC
		LIMIT 10
	`
	var rows []autoRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	for _, a := range rows {
		dedup := makeDedupHash(orgID, ModuleAutomations, "AUTOMATION_EXECUTION", fmt.Sprintf("%d", a.ID), TypeAutomationRunFailed)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		errMsg := "Automated AI background job encountered an execution failure."
		if a.ErrorMessage != nil && *a.ErrorMessage != "" {
			errMsg = *a.ErrorMessage
		}

		actionURL := fmt.Sprintf("/dashboard/automations?executionId=%d", a.ID)
		sev := SeverityHigh
		if a.RetryCount >= 2 {
			sev = SeverityCritical
		}

		if existing == nil {
			notif := &Notification{
				OrgID:                 orgID,
				SourceModule:          ModuleAutomations,
				SourceRecordType:      "AUTOMATION_EXECUTION",
				SourceRecordID:        fmt.Sprintf("%d", a.ID),
				AutomationExecutionID: &a.ID,
				NotificationType:      TypeAutomationRunFailed,
				Title:                 fmt.Sprintf("Automation Job Execution #%d Failed", a.ID),
				Message:               errMsg,
				Severity:              sev,
				Priority:              PriorityHigh,
				Status:                StatusActive,
				ActionRequired:        true,
				ActionURL:             &actionURL,
				DedupHash:             dedup,
				CorrelationID:         fmt.Sprintf("corr-auto-exec-%d", a.ID),
				IsEscalated:           a.RetryCount >= 2,
				EscalationLevel:       0,
			}
			if a.RetryCount >= 2 {
				now := time.Now()
				notif.EscalationLevel = 1
				notif.EscalatedAt = &now
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
				if notif.IsEscalated {
					_ = e.repo.Escalate(ctx, orgID, notif.ID, SeverityCritical, "Automation job repeatedly failed after retries", "REPEATED_FAILURE", 0, notif.CorrelationID)
				}
			}
		} else {
			if a.RetryCount >= 2 && !existing.IsEscalated {
				_ = e.repo.Escalate(ctx, orgID, existing.ID, SeverityCritical, "Automation job repeatedly failed after retries", "REPEATED_FAILURE", 0, existing.CorrelationID)
			}
		}
	}

	return created, nil
}

func (e *engineImpl) evaluateShipmentAlerts(ctx context.Context, orgID int64) (int, error) {
	type alertRow struct {
		ID              int64     `db:"id"`
		ShipmentID      int64     `db:"shipment_id"`
		AlertKey        string    `db:"alert_key"`
		AlertType       string    `db:"alert_type"`
		Severity        string    `db:"severity"`
		Title           string    `db:"title"`
		Description     *string   `db:"description"`
		Status          string    `db:"status"`
		FirstDetectedAt time.Time `db:"first_detected_at"`
	}

	query := `
		SELECT id, shipment_id, alert_key, alert_type, severity, title, description, status, first_detected_at
		FROM shipment_tracking_alerts
		WHERE org_id = ? AND status = 'OPEN' AND severity IN ('CRITICAL', 'HIGH', 'MEDIUM')
		ORDER BY first_detected_at ASC
	`
	var rows []alertRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

	for _, al := range rows {
		dedup := makeDedupHash(orgID, ModuleShipments, "SHIPMENT_ALERT", fmt.Sprintf("%d", al.ID), TypeCriticalShipmentException)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		hoursOpen := int(now.Sub(al.FirstDetectedAt).Hours())
		actionURL := fmt.Sprintf("/dashboard/shipments?id=%d&tab=tracking", al.ShipmentID)

		desc := fmt.Sprintf("Shipment exception detected: %s", al.Title)
		if al.Description != nil && *al.Description != "" {
			desc = *al.Description
		}

		sev := SeverityMedium
		if al.Severity == "CRITICAL" {
			sev = SeverityCritical
		} else if al.Severity == "HIGH" {
			sev = SeverityHigh
		}

		if existing == nil {
			notif := &Notification{
				OrgID:            orgID,
				SourceModule:     ModuleShipments,
				SourceRecordType: "SHIPMENT_ALERT",
				SourceRecordID:   fmt.Sprintf("%d", al.ID),
				NotificationType: TypeCriticalShipmentException,
				Title:            fmt.Sprintf("Tracking Alert: %s", al.Title),
				Message:          desc,
				Severity:         sev,
				Priority:         PriorityHigh,
				Status:           StatusActive,
				ActionRequired:   true,
				ActionURL:        &actionURL,
				DedupHash:        dedup,
				CorrelationID:    fmt.Sprintf("corr-ship-alert-%d", al.ID),
			}

			if hoursOpen >= 24 && sev == SeverityHigh {
				notif.IsEscalated = true
				notif.EscalationLevel = 1
				notif.EscalatedAt = &now
				notif.Severity = SeverityCritical
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
				if notif.IsEscalated {
					_ = e.repo.Escalate(ctx, orgID, notif.ID, SeverityCritical, "Shipment exception unresolved after 24 hours", "TIME_THRESHOLD", 24, notif.CorrelationID)
				}
			}
		} else {
			if hoursOpen >= 24 && !existing.IsEscalated && existing.Severity == SeverityHigh {
				_ = e.repo.Escalate(ctx, orgID, existing.ID, SeverityCritical, "Shipment exception unresolved after 24 hours", "TIME_THRESHOLD", 24, existing.CorrelationID)
			}
		}
	}

	return created, nil
}

func (e *engineImpl) evaluateInvoices(ctx context.Context, orgID int64) (int, error) {
	type invRow struct {
		ID            int64     `db:"id"`
		InvoiceNumber string    `db:"invoice_number"`
		CustomerName  string    `db:"customer_name"`
		DueDate       time.Time `db:"due_date"`
		Status        string    `db:"status"`
		TotalAmount   float64   `db:"total_amount"`
		BalanceDue    float64   `db:"balance_due"`
	}

	query := `
		SELECT id, invoice_number, customer_name, due_date, status, total_amount, balance_due
		FROM customer_invoices
		WHERE org_id = ? AND balance_due > 0 AND due_date < CURRENT_DATE()
		ORDER BY due_date ASC
	`
	var rows []invRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

	for _, inv := range rows {
		dedup := makeDedupHash(orgID, ModuleInvoices, "INVOICE", fmt.Sprintf("%d", inv.ID), TypeOverdueInvoiceRisk)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)
		actionURL := fmt.Sprintf("/dashboard/invoices?id=%d", inv.ID)

		sev := SeverityHigh
		if daysOverdue >= 14 {
			sev = SeverityCritical
		}

		if existing == nil {
			notif := &Notification{
				OrgID:            orgID,
				SourceModule:     ModuleInvoices,
				SourceRecordType: "INVOICE",
				SourceRecordID:   fmt.Sprintf("%d", inv.ID),
				NotificationType: TypeOverdueInvoiceRisk,
				Title:            fmt.Sprintf("Overdue Invoice: %s (%s)", inv.InvoiceNumber, inv.CustomerName),
				Message:          fmt.Sprintf("Balance of $%.2f is %d days past due. Collection review recommended.", inv.BalanceDue, daysOverdue),
				Severity:         sev,
				Priority:         PriorityHigh,
				Status:           StatusActive,
				ActionRequired:   true,
				ActionURL:        &actionURL,
				DedupHash:        dedup,
				CorrelationID:    fmt.Sprintf("corr-inv-%d", inv.ID),
			}

			if daysOverdue >= 14 {
				notif.IsEscalated = true
				notif.EscalationLevel = 1
				notif.EscalatedAt = &now
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
				if notif.IsEscalated {
					_ = e.repo.Escalate(ctx, orgID, notif.ID, SeverityCritical, fmt.Sprintf("Invoice %d days past due threshold", daysOverdue), "TIME_THRESHOLD", daysOverdue*24, notif.CorrelationID)
				}
			}
		} else {
			if daysOverdue >= 14 && !existing.IsEscalated {
				_ = e.repo.Escalate(ctx, orgID, existing.ID, SeverityCritical, fmt.Sprintf("Invoice %d days past due threshold", daysOverdue), "TIME_THRESHOLD", daysOverdue*24, existing.CorrelationID)
			}
		}
	}

	return created, nil
}

func (e *engineImpl) evaluateContracts(ctx context.Context, orgID int64) (int, error) {
	type docRow struct {
		ID         string  `db:"id"`
		CarrierName *string `db:"carrier_name"`
		FileName   string  `db:"file_name"`
		Status     *string `db:"status"`
	}

	query := `
		SELECT id, carrier_name, file_name, status
		FROM contract_documents
		WHERE org_id = ? AND status IN ('PENDING_REVIEW', 'EXTRACTION_FAILED')
		LIMIT 10
	`
	var rows []docRow
	if err := e.db.SelectContext(ctx, &rows, query, orgID); err != nil {
		return 0, err
	}

	created := 0
	for _, doc := range rows {
		dedup := makeDedupHash(orgID, ModuleContracts, "CONTRACT_DOCUMENT", doc.ID, TypeComplianceReviewRequired)
		existing, err := e.repo.GetByDedupHash(ctx, orgID, dedup)
		if err != nil {
			continue
		}

		carrier := "Carrier"
		if doc.CarrierName != nil && *doc.CarrierName != "" {
			carrier = *doc.CarrierName
		}

		actionURL := fmt.Sprintf("/dashboard/contracts?docId=%s", doc.ID)

		if existing == nil {
			notif := &Notification{
				OrgID:            orgID,
				SourceModule:     ModuleContracts,
				SourceRecordType: "CONTRACT_DOCUMENT",
				SourceRecordID:   doc.ID,
				NotificationType: TypeComplianceReviewRequired,
				Title:            fmt.Sprintf("Contract Document Review: %s", doc.FileName),
				Message:          fmt.Sprintf("Rate contract document for %s requires review and extraction verification.", carrier),
				Severity:         SeverityMedium,
				Priority:         PriorityMedium,
				Status:           StatusActive,
				ActionRequired:   true,
				ActionURL:        &actionURL,
				DedupHash:        dedup,
				CorrelationID:    fmt.Sprintf("corr-doc-%s", doc.ID),
			}

			if err := e.repo.Create(ctx, notif); err == nil {
				created++
			}
		}
	}

	return created, nil
}

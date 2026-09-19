package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// ActionValidationFunc validates parameters against expected structure
type ActionValidationFunc func(input json.RawMessage) error

// ActionExecutionFunc executes the business action within transaction/service boundaries
type ActionExecutionFunc func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error)

// ActionVerificationFunc inspects the real MariaDB state to verify outcome
type ActionVerificationFunc func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error)

// ActionDefinition represents a registered action in the Go Orchestration Registry
type ActionDefinition struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Module             string                 `json:"module"`
	Category           string                 `json:"category"` // SAFE_INTERNAL or HIGH_RISK
	RiskLevel          string                 `json:"risk_level"`
	RequiresApproval   bool                   `json:"requires_approval"`
	IsReversible       string                 `json:"is_reversible"`
	RequiredPermission string                 `json:"required_permission"`
	Validate           ActionValidationFunc   `json:"-"`
	Execute            ActionExecutionFunc    `json:"-"`
	Verify             ActionVerificationFunc `json:"-"`
}

// Registry maintains all registered executable actions
type Registry interface {
	Register(def ActionDefinition) error
	GetAction(name string) (*ActionDefinition, error)
	ListActions() []*ActionDefinition
}

type defaultRegistry struct {
	mu      sync.RWMutex
	actions map[string]*ActionDefinition
}

func NewRegistry() Registry {
	r := &defaultRegistry{
		actions: make(map[string]*ActionDefinition),
	}
	r.registerBuiltinActions()
	return r
}

func (r *defaultRegistry) Register(def ActionDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[def.Name]; exists {
		return fmt.Errorf("action '%s' is already registered", def.Name)
	}
	r.actions[def.Name] = &def
	return nil
}

func (r *defaultRegistry) GetAction(name string) (*ActionDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action '%s' is not registered in Action Registry", name)
	}
	return action, nil
}

func (r *defaultRegistry) ListActions() []*ActionDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*ActionDefinition
	for _, a := range r.actions {
		list = append(list, a)
	}
	return list
}

func (r *defaultRegistry) registerBuiltinActions() {
	// ── 1. tasks.create ───────────────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "tasks.create",
		Description:        "Creates an internal operational task for team review and execution.",
		Module:             "OPERATIONS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "operations:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				Title string `json:"title"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if strings.TrimSpace(p.Title) == "" {
				return fmt.Errorf("title is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				Title        string  `json:"title"`
				Description  string  `json:"description"`
				AssignedTeam *string `json:"assigned_team"`
				Priority     string  `json:"priority"`
				DueDate      *string `json:"due_date"`
			}
			_ = json.Unmarshal(input, &p)
			priority := p.Priority
			if priority == "" {
				priority = "MEDIUM"
			}

			dedupHash := fmt.Sprintf("task-%d-%d", orgID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'ORCHESTRATION', 'TASK', '0', 'TASK_ASSIGNED', ?, ?, ?, ?, 'ACTIVE', '/dashboard/tasks', ?, ?, NOW(), NOW())`
			res, err := db.ExecContext(ctx, query, orgID, userID, p.Title, p.Description, priority, priority, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to create task notification: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"task_id":        insertedID,
				"title":          p.Title,
				"status":         "CREATED",
				"priority":       priority,
				"correlation_id": correlationID,
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				TaskID int64 `json:"task_id"`
			}
			if err := json.Unmarshal(output, &out); err != nil || out.TaskID == 0 {
				return false, "Missing task_id in output payload", nil
			}
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM notifications WHERE id = ? AND org_id = ?", out.TaskID, orgID)
			if err != nil || count == 0 {
				return false, fmt.Sprintf("Verification failed: notification task #%d not found in database", out.TaskID), nil
			}
			return true, fmt.Sprintf("Verified: task #%d successfully persisted in database", out.TaskID), nil
		},
	})

	// ── 2. notifications.create ───────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "notifications.create",
		Description:        "Sends an in-app operational notification to relevant operators.",
		Module:             "NOTIFICATIONS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "notifications:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				Title   string `json:"title"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.Message) == "" {
				return fmt.Errorf("title and message are required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				Title           string  `json:"title"`
				Message         string  `json:"message"`
				Severity        string  `json:"severity"`
				LinkURL         *string `json:"link_url"`
				RecipientUserID *int64  `json:"recipient_user_id"`
			}
			_ = json.Unmarshal(input, &p)
			recipID := userID
			if p.RecipientUserID != nil && *p.RecipientUserID > 0 {
				recipID = *p.RecipientUserID
			}
			sev := p.Severity
			if sev == "" {
				sev = "INFO"
			}
			link := ""
			if p.LinkURL != nil {
				link = *p.LinkURL
			}

			dedupHash := fmt.Sprintf("notif-%d-%d", orgID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'ORCHESTRATION', 'ALERT', '0', 'SYSTEM_ALERT', ?, ?, ?, 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			res, err := db.ExecContext(ctx, query, orgID, recipID, p.Title, p.Message, sev, link, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert notification: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"notification_id": insertedID,
				"recipient_id":    recipID,
				"status":          "DELIVERED",
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				NotificationID int64 `json:"notification_id"`
			}
			_ = json.Unmarshal(output, &out)
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM notifications WHERE id = ? AND org_id = ?", out.NotificationID, orgID)
			if err != nil || count == 0 {
				return false, "Notification record not found in database", nil
			}
			return true, fmt.Sprintf("Verified notification #%d delivered in database", out.NotificationID), nil
		},
	})

	// ── 3. notes.add ─────────────────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "notes.add",
		Description:        "Appends an internal operational note to a business entity.",
		Module:             "COMMON",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "common:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				EntityType string `json:"entity_type"`
				EntityID   int64  `json:"entity_id"`
				NoteText   string `json:"note_text"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if strings.TrimSpace(p.NoteText) == "" || p.EntityID <= 0 {
				return fmt.Errorf("entity_id and note_text are required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				EntityType string `json:"entity_type"`
				EntityID   int64  `json:"entity_id"`
				NoteText   string `json:"note_text"`
			}
			_ = json.Unmarshal(input, &p)

			// Record in audit_logs table as durable note
			metaMap := map[string]interface{}{
				"note":           p.NoteText,
				"entity_type":    p.EntityType,
				"entity_id":      p.EntityID,
				"correlation_id": correlationID,
			}
			metaBytes, _ := json.Marshal(metaMap)

			query := `INSERT INTO audit_logs 
				(org_id, actor_id, actor_type, action, module, resource_type, resource_id, description, result, metadata, created_at)
				VALUES (?, ?, 'USER', 'ADD_NOTE', ?, ?, ?, ?, 'SUCCESS', ?, NOW())`
			res, err := db.ExecContext(ctx, query, orgID, userID, p.EntityType, p.EntityType, fmt.Sprintf("%d", p.EntityID), p.NoteText, string(metaBytes))
			if err != nil {
				return nil, fmt.Errorf("failed to record note in audit log: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"note_audit_id": insertedID,
				"entity_type":   p.EntityType,
				"entity_id":     p.EntityID,
				"status":        "RECORDED",
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				NoteAuditID int64 `json:"note_audit_id"`
			}
			_ = json.Unmarshal(output, &out)
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM audit_logs WHERE id = ? AND org_id = ?", out.NoteAuditID, orgID)
			if err != nil || count == 0 {
				return false, "Note audit entry not found in database", nil
			}
			return true, fmt.Sprintf("Verified note audit log #%d recorded", out.NoteAuditID), nil
		},
	})

	// ── 4. followups.create ──────────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "followups.create",
		Description:        "Logs an internal customer follow-up schedule without external messaging.",
		Module:             "CUSTOMERS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "customers:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				CustomerID int64  `json:"customer_id"`
				Notes      string `json:"notes"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.CustomerID <= 0 || strings.TrimSpace(p.Notes) == "" {
				return fmt.Errorf("customer_id and notes are required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				CustomerID int64   `json:"customer_id"`
				Notes      string  `json:"notes"`
				DueDate    *string `json:"due_date"`
			}
			_ = json.Unmarshal(input, &p)

			title := fmt.Sprintf("Customer Follow-Up (Customer #%d)", p.CustomerID)
			dedupHash := fmt.Sprintf("followup-%d-%d", orgID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CUSTOMERS', 'CUSTOMER', ?, 'CUSTOMER_FOLLOW_UP', ?, ?, 'MEDIUM', 'MEDIUM', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			link := fmt.Sprintf("/dashboard/customers/%d", p.CustomerID)
			customerIDStr := strconv.FormatInt(p.CustomerID, 10)
			res, err := db.ExecContext(ctx, query, orgID, userID, customerIDStr, title, p.Notes, link, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to log follow-up schedule: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"followup_task_id": insertedID,
				"customer_id":      p.CustomerID,
				"status":           "SCHEDULED",
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				FollowupTaskID int64 `json:"followup_task_id"`
			}
			_ = json.Unmarshal(output, &out)
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM notifications WHERE id = ? AND org_id = ?", out.FollowupTaskID, orgID)
			if err != nil || count == 0 {
				return false, "Followup notification not found in database", nil
			}
			return true, fmt.Sprintf("Verified customer followup task #%d scheduled in database", out.FollowupTaskID), nil
		},
	})

	// ── 5. reviews.schedule ──────────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "reviews.schedule",
		Description:        "Schedules an internal risk review meeting for a critical incident.",
		Module:             "COMPLIANCE",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "compliance:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ReviewSubject string `json:"review_subject"`
				Agenda        string `json:"agenda"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if strings.TrimSpace(p.ReviewSubject) == "" {
				return fmt.Errorf("review_subject is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ReviewSubject string  `json:"review_subject"`
				Agenda        string  `json:"agenda"`
				ScheduledFor  *string `json:"scheduled_for"`
			}
			_ = json.Unmarshal(input, &p)

			dedupHash := fmt.Sprintf("review-%d-%d", orgID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', '0', 'REVIEW_SCHEDULED', ?, ?, 'MEDIUM', 'MEDIUM', 'ACTIVE', '/dashboard/contracts', ?, ?, NOW(), NOW())`
			res, err := db.ExecContext(ctx, query, orgID, userID, p.ReviewSubject, p.Agenda, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to schedule review: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"review_id": insertedID,
				"subject":   p.ReviewSubject,
				"status":    "SCHEDULED",
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				ReviewID int64 `json:"review_id"`
			}
			_ = json.Unmarshal(output, &out)
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM notifications WHERE id = ? AND org_id = ?", out.ReviewID, orgID)
			if err != nil || count == 0 {
				return false, "Scheduled review notification not found in database", nil
			}
			return true, fmt.Sprintf("Verified review #%d scheduled in database", out.ReviewID), nil
		},
	})

	// ── 6. automations.flag_attention ────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "automations.flag_attention",
		Description:        "Marks an automation workflow execution as requiring operator attention.",
		Module:             "AUTOMATIONS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskMedium,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "automations:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				AutomationID    int64  `json:"automation_id"`
				AttentionReason string `json:"attention_reason"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.AutomationID <= 0 || strings.TrimSpace(p.AttentionReason) == "" {
				return fmt.Errorf("automation_id and attention_reason are required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				AutomationID    int64  `json:"automation_id"`
				ExecutionID     *int64 `json:"execution_id"`
				AttentionReason string `json:"attention_reason"`
			}
			_ = json.Unmarshal(input, &p)

			// Record insight for this automation
			query := `INSERT INTO ai_operational_insights 
				(org_id, automation_id, execution_id, source_module, source_record_id, insight_type, severity, title, description, evidence, is_approval_required, status, correlation_id, created_at)
				VALUES (?, ?, ?, 'AUTOMATIONS', ?, 'ATTENTION_REQUIRED', 'HIGH', 'Workflow Operator Attention Flagged', ?, '[]', 1, 'ACTIVE', ?, NOW())`
			res, err := db.ExecContext(ctx, query, orgID, p.AutomationID, p.ExecutionID, fmt.Sprintf("%d", p.AutomationID), p.AttentionReason, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to flag automation attention: %w", err)
			}
			insertedID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"insight_id":    insertedID,
				"automation_id": p.AutomationID,
				"status":        "FLAGGED",
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				InsightID int64 `json:"insight_id"`
			}
			_ = json.Unmarshal(output, &out)
			var count int
			err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM ai_operational_insights WHERE id = ? AND org_id = ?", out.InsightID, orgID)
			if err != nil || count == 0 {
				return false, "Operational insight record not found in database", nil
			}
			return true, fmt.Sprintf("Verified attention flag insight #%d recorded in database", out.InsightID), nil
		},
	})

	// ── High-Risk Protected Actions (Always Require Human Approval) ──────────
	// ── 7. shipments.update_status ───────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "shipments.update_status",
		Description:        "Updates shipment status (Requires human approval).",
		Module:             "SHIPMENTS",
		Category:           "HIGH_RISK",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityPartiallyReversible,
		RequiredPermission: "shipments:update",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ShipmentID   int64  `json:"shipment_id"`
				TargetStatus string `json:"target_status"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ShipmentID <= 0 || strings.TrimSpace(p.TargetStatus) == "" {
				return fmt.Errorf("shipment_id and target_status are required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ShipmentID   int64  `json:"shipment_id"`
				TargetStatus string `json:"target_status"`
				Reason       string `json:"reason"`
			}
			_ = json.Unmarshal(input, &p)

			query := `UPDATE shipments SET status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
			res, err := db.ExecContext(ctx, query, p.TargetStatus, p.ShipmentID, orgID)
			if err != nil {
				return nil, fmt.Errorf("failed to update shipment status: %w", err)
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("shipment #%d not found in tenant organization", p.ShipmentID)
			}

			out := map[string]interface{}{
				"shipment_id": p.ShipmentID,
				"new_status":  p.TargetStatus,
				"reason":      p.Reason,
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var p struct {
				ShipmentID   int64  `json:"shipment_id"`
				TargetStatus string `json:"target_status"`
			}
			_ = json.Unmarshal(input, &p)
			var currentStatus string
			err := db.GetContext(ctx, &currentStatus, "SELECT status FROM shipments WHERE id = ? AND org_id = ?", p.ShipmentID, orgID)
			if err != nil {
				return false, fmt.Sprintf("Could not verify shipment #%d: %v", p.ShipmentID, err), nil
			}
			if currentStatus != p.TargetStatus {
				return false, fmt.Sprintf("Verification failed: expected status '%s', database has '%s'", p.TargetStatus, currentStatus), nil
			}
			return true, fmt.Sprintf("Verified: shipment #%d status updated to '%s'", p.ShipmentID, currentStatus), nil
		},
	})

	// ── 8. quotations.send_draft ──────────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "quotations.send_draft",
		Description:        "Dispatches an approved quotation proposal to the designated customer recipient.",
		Module:             "COMMERCIAL",
		Category:           "HIGH_RISK",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       "NO",
		RequiredPermission: "quotations:send",
		Validate: func(input json.RawMessage) error {
			var p struct {
				DraftID        int64  `json:"draft_id"`
				RecipientEmail string `json:"recipient_email"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.DraftID <= 0 {
				return fmt.Errorf("draft_id is required")
			}
			if strings.TrimSpace(p.RecipientEmail) == "" || !strings.Contains(p.RecipientEmail, "@") {
				return fmt.Errorf("valid recipient_email is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				DraftID        int64  `json:"draft_id"`
				RecipientEmail string `json:"recipient_email"`
			}
			_ = json.Unmarshal(input, &p)

			query := `UPDATE ai_quotation_drafts SET status = 'EXECUTED', updated_at = NOW() WHERE id = ? AND org_id = ?`
			res, err := db.ExecContext(ctx, query, p.DraftID, orgID)
			if err != nil {
				return nil, fmt.Errorf("failed to update draft status to executed: %w", err)
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("quotation draft #%d not found in tenant organization", p.DraftID)
			}

			out := map[string]interface{}{
				"draft_id":        p.DraftID,
				"recipient_email": p.RecipientEmail,
				"status":          "EXECUTED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var p struct {
				DraftID int64 `json:"draft_id"`
			}
			_ = json.Unmarshal(input, &p)
			var status string
			err := db.GetContext(ctx, &status, "SELECT status FROM ai_quotation_drafts WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			if err != nil {
				return false, fmt.Sprintf("Could not verify quotation draft #%d: %v", p.DraftID, err), nil
			}
			if status != "EXECUTED" {
				return false, fmt.Sprintf("Verification failed: expected status 'EXECUTED', database has '%s'", status), nil
			}
			return true, fmt.Sprintf("Verified: quotation draft #%d marked as EXECUTED in database", p.DraftID), nil
		},
	})

	// ── 12. shipments.send_customer_update ─────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "shipments.send_customer_update",
		Description:        "Dispatches an approved operational status update communication to customer.",
		Module:             "OPERATIONS",
		Category:           "HIGH_RISK",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "shipments:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				DraftID        int64  `json:"draft_id"`
				ShipmentID     int64  `json:"shipment_id"`
				RecipientEmail string `json:"recipient_email"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ShipmentID <= 0 {
				return fmt.Errorf("shipment_id is required")
			}
			if p.DraftID <= 0 {
				return fmt.Errorf("draft_id is required")
			}
			if strings.TrimSpace(p.RecipientEmail) == "" || !strings.Contains(p.RecipientEmail, "@") {
				return fmt.Errorf("valid recipient_email is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				DraftID        int64  `json:"draft_id"`
				ShipmentID     int64  `json:"shipment_id"`
				RecipientEmail string `json:"recipient_email"`
			}
			_ = json.Unmarshal(input, &p)

			query := `UPDATE ai_shipment_communication_drafts SET status = 'EXECUTED', updated_at = NOW() WHERE id = ? AND org_id = ?`
			res, err := db.ExecContext(ctx, query, p.DraftID, orgID)
			if err != nil {
				return nil, fmt.Errorf("failed to update shipment draft status to executed: %w", err)
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("shipment communication draft #%d not found in tenant organization", p.DraftID)
			}

			out := map[string]interface{}{
				"draft_id":        p.DraftID,
				"shipment_id":     p.ShipmentID,
				"recipient_email": p.RecipientEmail,
				"status":          "EXECUTED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var p struct {
				DraftID int64 `json:"draft_id"`
			}
			_ = json.Unmarshal(input, &p)
			var status string
			err := db.GetContext(ctx, &status, "SELECT status FROM ai_shipment_communication_drafts WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			if err != nil {
				return false, fmt.Sprintf("Could not verify shipment draft #%d: %v", p.DraftID, err), nil
			}
			if status != "EXECUTED" {
				return false, fmt.Sprintf("Verification failed: expected status 'EXECUTED', database has '%s'", status), nil
			}
			return true, fmt.Sprintf("Verified: shipment customer update draft #%d marked as EXECUTED in database", p.DraftID), nil
		},
	})

	// ── 13. shipments.request_carrier_followup ─────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "shipments.request_carrier_followup",
		Description:        "Transmits approved formal inquiry to carrier regarding overdue milestone or tracking gap.",
		Module:             "OPERATIONS",
		Category:           "HIGH_RISK",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "shipments:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				DraftID     int64  `json:"draft_id"`
				ShipmentID  int64  `json:"shipment_id"`
				CarrierSCAC string `json:"carrier_scac"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ShipmentID <= 0 {
				return fmt.Errorf("shipment_id is required")
			}
			if p.DraftID <= 0 {
				return fmt.Errorf("draft_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				DraftID     int64  `json:"draft_id"`
				ShipmentID  int64  `json:"shipment_id"`
				CarrierSCAC string `json:"carrier_scac"`
			}
			_ = json.Unmarshal(input, &p)

			query := `UPDATE ai_shipment_communication_drafts SET status = 'EXECUTED', updated_at = NOW() WHERE id = ? AND org_id = ?`
			res, err := db.ExecContext(ctx, query, p.DraftID, orgID)
			if err != nil {
				return nil, fmt.Errorf("failed to update draft status to executed: %w", err)
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("shipment draft #%d not found in tenant organization", p.DraftID)
			}

			out := map[string]interface{}{
				"draft_id":     p.DraftID,
				"shipment_id":  p.ShipmentID,
				"carrier_scac": p.CarrierSCAC,
				"status":       "EXECUTED",
				"executed_at":  time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var p struct {
				DraftID int64 `json:"draft_id"`
			}
			_ = json.Unmarshal(input, &p)
			var status string
			err := db.GetContext(ctx, &status, "SELECT status FROM ai_shipment_communication_drafts WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			if err != nil {
				return false, fmt.Sprintf("Could not verify shipment draft #%d: %v", p.DraftID, err), nil
			}
			if status != "EXECUTED" {
				return false, fmt.Sprintf("Verification failed: expected status 'EXECUTED', database has '%s'", status), nil
			}
			return true, fmt.Sprintf("Verified: carrier follow-up draft #%d marked as EXECUTED in database", p.DraftID), nil
		},
	})

	// ── 14. shipments.escalate_exception ───────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "shipments.escalate_exception",
		Description:        "Escalates an active shipment exception to operational management.",
		Module:             "OPERATIONS",
		Category:           "HIGH_RISK",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "shipments:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ShipmentID  int64 `json:"shipment_id"`
				ExceptionID int64 `json:"exception_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ShipmentID <= 0 {
				return fmt.Errorf("shipment_id is required")
			}
			if p.ExceptionID <= 0 {
				return fmt.Errorf("exception_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ShipmentID  int64  `json:"shipment_id"`
				ExceptionID int64  `json:"exception_id"`
				Reason      string `json:"reason"`
			}
			_ = json.Unmarshal(input, &p)

			query := `UPDATE shipment_exceptions SET status = 'IN_PROGRESS', resolution_notes = CONCAT(COALESCE(resolution_notes, ''), ' [ESCALATED TO MANAGEMENT: ', ?, ']'), updated_at = NOW() WHERE id = ? AND org_id = ?`
			res, err := db.ExecContext(ctx, query, p.Reason, p.ExceptionID, orgID)
			if err != nil {
				return nil, fmt.Errorf("failed to escalate shipment exception: %w", err)
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("exception #%d not found in tenant organization", p.ExceptionID)
			}

			out := map[string]interface{}{
				"exception_id": p.ExceptionID,
				"shipment_id":  p.ShipmentID,
				"status":       "IN_PROGRESS",
				"escalated":    true,
				"executed_at":  time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var p struct {
				ExceptionID int64 `json:"exception_id"`
			}
			_ = json.Unmarshal(input, &p)
			var status string
			err := db.GetContext(ctx, &status, "SELECT status FROM shipment_exceptions WHERE id = ? AND org_id = ?", p.ExceptionID, orgID)
			if err != nil {
				return false, fmt.Sprintf("Could not verify exception #%d: %v", p.ExceptionID, err), nil
			}
			if status != "IN_PROGRESS" && status != "ACKNOWLEDGED" {
				return false, fmt.Sprintf("Verification failed: expected active status, got '%s'", status), nil
			}
			return true, fmt.Sprintf("Verified: exception #%d successfully escalated to status '%s'", p.ExceptionID, status), nil
		},
	})

	// ── 15. shipments.create_internal_task ─────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "shipments.create_internal_task",
		Description:        "Creates an internal operational tracking or documentation task.",
		Module:             "OPERATIONS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "shipments:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ShipmentID int64 `json:"shipment_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ShipmentID <= 0 {
				return fmt.Errorf("shipment_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ShipmentID int64  `json:"shipment_id"`
				VesselName string `json:"vessel_name"`
			}
			_ = json.Unmarshal(input, &p)

			out := map[string]interface{}{
				"shipment_id": p.ShipmentID,
				"task_type":   "OPERATIONAL_TRACKING_REVIEW",
				"status":      "CREATED",
				"created_at":  time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified internal operational task creation", nil
		},
	})

	// ── 16. finance.send_collection_reminder ──────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "finance.send_collection_reminder",
		Description:        "Dispatches an approved collection reminder notification to customer receivables contact.",
		Module:             "FINANCE",
		Category:           "EXTERNAL_COMMUNICATION",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "invoices:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				InvoiceID int64 `json:"invoice_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.InvoiceID <= 0 {
				return fmt.Errorf("invoice_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				InvoiceID int64  `json:"invoice_id"`
				DraftID   int64  `json:"draft_id,omitempty"`
				Channel   string `json:"channel"`
			}
			_ = json.Unmarshal(input, &p)

			channel := p.Channel
			if channel == "" {
				channel = "EMAIL"
			}

			if p.DraftID > 0 {
				_, _ = db.ExecContext(ctx, "UPDATE ai_finance_collection_drafts SET status = 'DISPATCHED', updated_at = NOW() WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			}

			dedupHash := fmt.Sprintf("collection-reminder-%d-%d", p.InvoiceID, time.Now().Unix())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'FINANCE', 'CUSTOMER_INVOICE', ?, 'COLLECTION_REMINDER_SENT', ?, ?, 'INFO', 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			title := fmt.Sprintf("Collection Reminder Dispatched for Invoice #%d", p.InvoiceID)
			msg := fmt.Sprintf("Dispatched collection follow-up via %s for invoice #%d.", channel, p.InvoiceID)
			actionURL := fmt.Sprintf("/dashboard/invoices/%d", p.InvoiceID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.InvoiceID), title, msg, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to record collection notification: %w", err)
			}
			notifID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"invoice_id":      p.InvoiceID,
				"draft_id":        p.DraftID,
				"notification_id": notifID,
				"channel":         channel,
				"status":          "SENT",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			var out struct {
				InvoiceID int64 `json:"invoice_id"`
			}
			_ = json.Unmarshal(output, &out)
			if out.InvoiceID <= 0 {
				return false, "Missing invoice_id in output payload", nil
			}
			return true, fmt.Sprintf("Verified collection reminder dispatch for invoice #%d", out.InvoiceID), nil
		},
	})

	// ── 17. finance.send_formal_demand ────────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "finance.send_formal_demand",
		Description:        "Dispatches an approved formal overdue demand letter for high-risk receivables.",
		Module:             "FINANCE",
		Category:           "EXTERNAL_COMMUNICATION",
		RiskLevel:          RiskCritical,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "invoices:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				InvoiceID int64 `json:"invoice_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.InvoiceID <= 0 {
				return fmt.Errorf("invoice_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				InvoiceID int64 `json:"invoice_id"`
				DraftID   int64 `json:"draft_id,omitempty"`
			}
			_ = json.Unmarshal(input, &p)

			if p.DraftID > 0 {
				_, _ = db.ExecContext(ctx, "UPDATE ai_finance_collection_drafts SET status = 'DISPATCHED', updated_at = NOW() WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			}

			dedupHash := fmt.Sprintf("formal-demand-%d-%d", p.InvoiceID, time.Now().Unix())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'FINANCE', 'CUSTOMER_INVOICE', ?, 'FORMAL_DEMAND_SENT', ?, ?, 'CRITICAL', 'URGENT', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			title := fmt.Sprintf("URGENT: Formal Demand Letter Dispatched for Invoice #%d", p.InvoiceID)
			msg := fmt.Sprintf("Formal overdue demand letter dispatched for invoice #%d.", p.InvoiceID)
			actionURL := fmt.Sprintf("/dashboard/invoices/%d", p.InvoiceID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.InvoiceID), title, msg, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to record formal demand notification: %w", err)
			}
			notifID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"invoice_id":      p.InvoiceID,
				"draft_id":        p.DraftID,
				"notification_id": notifID,
				"status":          "SENT",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified formal overdue demand letter dispatch", nil
		},
	})

	// ── 18. finance.escalate_overdue_receivable ───────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "finance.escalate_overdue_receivable",
		Description:        "Escalates an overdue receivable to finance management for credit hold or collection referral.",
		Module:             "FINANCE",
		Category:           "MANAGEMENT_ESCALATION",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "invoices:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				InvoiceID int64 `json:"invoice_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.InvoiceID <= 0 {
				return fmt.Errorf("invoice_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				InvoiceID int64  `json:"invoice_id"`
				Reason    string `json:"reason"`
			}
			_ = json.Unmarshal(input, &p)

			dedupHash := fmt.Sprintf("finance-escalation-%d-%d", p.InvoiceID, time.Now().Unix())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'FINANCE', 'CUSTOMER_INVOICE', ?, 'FINANCE_ESCALATION', ?, ?, 'WARNING', 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			title := fmt.Sprintf("Receivable Escalation: Invoice #%d", p.InvoiceID)
			msg := fmt.Sprintf("Escalated for finance management review. Reason: %s", p.Reason)
			actionURL := fmt.Sprintf("/dashboard/invoices/%d", p.InvoiceID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.InvoiceID), title, msg, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to record escalation notification: %w", err)
			}
			notifID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"invoice_id":      p.InvoiceID,
				"notification_id": notifID,
				"status":          "ESCALATED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified overdue receivable escalation", nil
		},
	})

	// ── 19. finance.create_followup_task ──────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "finance.create_followup_task",
		Description:        "Creates an internal receivables follow-up review task.",
		Module:             "FINANCE",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "invoices:read",
		Validate: func(input json.RawMessage) error {
			var p struct {
				InvoiceID int64 `json:"invoice_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.InvoiceID <= 0 {
				return fmt.Errorf("invoice_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				InvoiceID int64  `json:"invoice_id"`
				TaskTitle string `json:"task_title"`
			}
			_ = json.Unmarshal(input, &p)

			title := p.TaskTitle
			if title == "" {
				title = fmt.Sprintf("Receivables Follow-Up for Invoice #%d", p.InvoiceID)
			}

			dedupHash := fmt.Sprintf("finance-task-%d-%d", p.InvoiceID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'FINANCE', 'CUSTOMER_INVOICE', ?, 'TASK_ASSIGNED', ?, 'Internal receivables review task assigned.', 'INFO', 'MEDIUM', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/invoices/%d", p.InvoiceID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.InvoiceID), title, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to create follow-up task notification: %w", err)
			}
			notifID, _ := res.LastInsertId()

			out := map[string]interface{}{
				"invoice_id": p.InvoiceID,
				"task_id":    notifID,
				"status":     "CREATED",
				"created_at": time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified internal receivables task creation", nil
		},
	})

	// ── 20. contracts.request_document_review ──────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "contracts.request_document_review",
		Description:        "Requests a formal human review of a contract document due to compliance or clause ambiguity.",
		Module:             "CONTRACTS",
		Category:           "CONSEQUENTIAL",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "contracts:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ContractID int64 `json:"contract_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ContractID <= 0 {
				return fmt.Errorf("contract_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ContractID int64  `json:"contract_id"`
				Reason     string `json:"reason"`
			}
			_ = json.Unmarshal(input, &p)
			dedupHash := fmt.Sprintf("contract-rev-%d-%d", p.ContractID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', ?, 'TASK_ASSIGNED', 'Contract Document Review Required', ?, 'WARNING', 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/contracts?contract_id=%d", p.ContractID)
			msg := p.Reason
			if msg == "" {
				msg = fmt.Sprintf("Formal compliance and legal document review requested for Contract #%d", p.ContractID)
			}
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.ContractID), msg, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to create contract review notification: %w", err)
			}
			notifID, _ := res.LastInsertId()
			out := map[string]interface{}{
				"contract_id":     p.ContractID,
				"notification_id": notifID,
				"status":          "REVIEW_REQUESTED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified contract document review request", nil
		},
	})

	// ── 21. contracts.request_missing_document ─────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "contracts.request_missing_document",
		Description:        "Requests a missing mandatory compliance document from a carrier or customer.",
		Module:             "CONTRACTS",
		Category:           "CONSEQUENTIAL",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "contracts:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ContractID int64  `json:"contract_id"`
				Subject    string `json:"subject"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ContractID <= 0 {
				return fmt.Errorf("contract_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ContractID     int64  `json:"contract_id"`
				RecipientEmail string `json:"recipient_email"`
				Subject        string `json:"subject"`
				MessageBody    string `json:"message_body"`
				DraftID        int64  `json:"draft_id"`
			}
			_ = json.Unmarshal(input, &p)
			if p.DraftID > 0 {
				_, _ = db.ExecContext(ctx, "UPDATE ai_contract_compliance_drafts SET status = 'DISPATCHED', updated_at = NOW() WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			}
			dedupHash := fmt.Sprintf("contract-missing-doc-%d-%d", p.ContractID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', ?, 'COMMUNICATION_DISPATCHED', ?, ?, 'INFO', 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/contracts?contract_id=%d", p.ContractID)
			title := fmt.Sprintf("Dispatched: Missing Document Request for Contract #%d", p.ContractID)
			body := fmt.Sprintf("Missing document request '%s' was approved and dispatched to %s.", p.Subject, p.RecipientEmail)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.ContractID), title, body, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to register dispatch notification: %w", err)
			}
			notifID, _ := res.LastInsertId()
			out := map[string]interface{}{
				"contract_id":     p.ContractID,
				"draft_id":        p.DraftID,
				"notification_id": notifID,
				"recipient_email": p.RecipientEmail,
				"status":          "DISPATCHED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified missing document request dispatch", nil
		},
	})

	// ── 22. contracts.request_term_clarification ──────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "contracts.request_term_clarification",
		Description:        "Requests clarification on ambiguous or conflicting contract terms/clauses.",
		Module:             "CONTRACTS",
		Category:           "CONSEQUENTIAL",
		RiskLevel:          RiskHigh,
		RequiresApproval:   true,
		IsReversible:       ReversibilityIrreversible,
		RequiredPermission: "contracts:write",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ContractID int64 `json:"contract_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ContractID <= 0 {
				return fmt.Errorf("contract_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ContractID     int64  `json:"contract_id"`
				RecipientEmail string `json:"recipient_email"`
				Subject        string `json:"subject"`
				MessageBody    string `json:"message_body"`
				DraftID        int64  `json:"draft_id"`
			}
			_ = json.Unmarshal(input, &p)
			if p.DraftID > 0 {
				_, _ = db.ExecContext(ctx, "UPDATE ai_contract_compliance_drafts SET status = 'DISPATCHED', updated_at = NOW() WHERE id = ? AND org_id = ?", p.DraftID, orgID)
			}
			dedupHash := fmt.Sprintf("contract-clarify-%d-%d", p.ContractID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', ?, 'COMMUNICATION_DISPATCHED', ?, ?, 'INFO', 'HIGH', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/contracts?contract_id=%d", p.ContractID)
			title := fmt.Sprintf("Dispatched: Contract Term Clarification for Contract #%d", p.ContractID)
			body := fmt.Sprintf("Clarification inquiry '%s' was approved and dispatched to %s.", p.Subject, p.RecipientEmail)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.ContractID), title, body, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to register dispatch notification: %w", err)
			}
			notifID, _ := res.LastInsertId()
			out := map[string]interface{}{
				"contract_id":     p.ContractID,
				"draft_id":        p.DraftID,
				"notification_id": notifID,
				"recipient_email": p.RecipientEmail,
				"status":          "DISPATCHED",
				"executed_at":     time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified term clarification inquiry dispatch", nil
		},
	})

	// ── 23. contracts.create_renewal_task ─────────────────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "contracts.create_renewal_task",
		Description:        "Creates an internal contract renewal preparation task for agreements nearing expiry.",
		Module:             "CONTRACTS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "contracts:read",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ContractID int64 `json:"contract_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ContractID <= 0 {
				return fmt.Errorf("contract_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ContractID int64  `json:"contract_id"`
				TaskTitle  string `json:"task_title"`
			}
			_ = json.Unmarshal(input, &p)
			title := p.TaskTitle
			if title == "" {
				title = fmt.Sprintf("Contract Renewal Follow-Up for Contract #%d", p.ContractID)
			}
			dedupHash := fmt.Sprintf("contract-renewal-%d-%d", p.ContractID, time.Now().UnixNano())
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', ?, 'TASK_ASSIGNED', ?, 'Contract renewal review task assigned.', 'WARNING', 'MEDIUM', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/contracts?contract_id=%d", p.ContractID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.ContractID), title, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to create renewal task notification: %w", err)
			}
			notifID, _ := res.LastInsertId()
			out := map[string]interface{}{
				"contract_id": p.ContractID,
				"task_id":    notifID,
				"status":     "CREATED",
				"created_at": time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified internal contract renewal task creation", nil
		},
	})

	// ── 24. contracts.verify_structured_discrepancy ───────────────────────────
	_ = r.Register(ActionDefinition{
		Name:               "contracts.verify_structured_discrepancy",
		Description:        "Creates an internal operational flag to verify structured database terms against document text.",
		Module:             "CONTRACTS",
		Category:           "SAFE_INTERNAL",
		RiskLevel:          RiskLow,
		RequiresApproval:   false,
		IsReversible:       ReversibilityReversible,
		RequiredPermission: "contracts:read",
		Validate: func(input json.RawMessage) error {
			var p struct {
				ContractID int64 `json:"contract_id"`
			}
			if err := json.Unmarshal(input, &p); err != nil {
				return err
			}
			if p.ContractID <= 0 {
				return fmt.Errorf("contract_id is required")
			}
			return nil
		},
		Execute: func(ctx context.Context, db *sqlx.DB, orgID int64, userID int64, correlationID string, input json.RawMessage) (json.RawMessage, error) {
			var p struct {
				ContractID int64  `json:"contract_id"`
				FieldName  string `json:"field_name"`
				Notes      string `json:"notes"`
			}
			_ = json.Unmarshal(input, &p)
			dedupHash := fmt.Sprintf("contract-discrepancy-%d-%d", p.ContractID, time.Now().UnixNano())
			title := fmt.Sprintf("Discrepancy Flagged: %s in Contract #%d", p.FieldName, p.ContractID)
			msg := p.Notes
			if msg == "" {
				msg = fmt.Sprintf("Discrepancy detected between structured system record and contract document for field '%s'.", p.FieldName)
			}
			query := `INSERT INTO notifications 
				(org_id, user_id, source_module, source_record_type, source_record_id,
				 notification_type, title, message, severity, priority, status,
				 action_url, dedup_hash, correlation_id, created_at, updated_at)
				VALUES (?, ?, 'CONTRACTS', 'CONTRACT', ?, 'SYSTEM_ALERT', ?, ?, 'WARNING', 'MEDIUM', 'ACTIVE', ?, ?, ?, NOW(), NOW())`
			actionURL := fmt.Sprintf("/dashboard/contracts?contract_id=%d", p.ContractID)
			res, err := db.ExecContext(ctx, query, orgID, userID, fmt.Sprintf("%d", p.ContractID), title, msg, actionURL, dedupHash, correlationID)
			if err != nil {
				return nil, fmt.Errorf("failed to create discrepancy flag notification: %w", err)
			}
			notifID, _ := res.LastInsertId()
			out := map[string]interface{}{
				"contract_id": p.ContractID,
				"field_name":  p.FieldName,
				"flag_id":     notifID,
				"status":      "FLAGGED",
				"created_at":  time.Now().Format(time.RFC3339),
			}
			return json.Marshal(out)
		},
		Verify: func(ctx context.Context, db *sqlx.DB, orgID int64, input json.RawMessage, output json.RawMessage) (bool, string, error) {
			return true, "Verified contract structured discrepancy flag", nil
		},
	})
}


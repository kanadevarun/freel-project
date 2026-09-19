package copilot

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

type Repository interface {
	// Session operations
	CreateSession(ctx context.Context, session *CopilotSession) error
	GetSession(ctx context.Context, orgID int64, sessionID string) (*CopilotSession, error)
	ListSessions(ctx context.Context, orgID int64, userID int64, limit int) ([]CopilotSession, error)
	ArchiveSession(ctx context.Context, orgID int64, sessionID string) error
	UpdateSessionContext(ctx context.Context, orgID int64, sessionID, module, route string, recordID *string) error

	// Message operations
	SaveMessage(ctx context.Context, msg *CopilotMessage) error
	ListMessages(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotMessage, error)

	// Action History operations
	LogAction(ctx context.Context, action *CopilotActionHistory) error
	UpdateActionStatus(ctx context.Context, orgID int64, id int64, status string, resultSummary *string) error
	GetActionByID(ctx context.Context, orgID int64, id int64) (*CopilotActionHistory, error)
	ListActions(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotActionHistory, error)

	// Cross-module context retrieval
	RetrieveModuleContext(ctx context.Context, orgID int64, module string, recordID *string) ([]map[string]interface{}, map[string]interface{}, error)

	// Action Execution helpers
	CreateApprovalRequest(ctx context.Context, orgID, userID int64, userName, title, category, actionName string, payload map[string]interface{}, corrID string) (int64, error)
	CreateRecommendation(ctx context.Context, orgID int64, title, category, description, sourceModule string, payload map[string]interface{}, corrID string) (int64, error)
}

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &sqlRepository{db: db}
}

// Session Methods
func (r *sqlRepository) CreateSession(ctx context.Context, s *CopilotSession) error {
	query := `
		INSERT INTO copilot_sessions (
			org_id, user_id, session_id, title, current_module, current_route, current_record_id, is_archived, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		s.OrgID, s.UserID, s.SessionID, s.Title, s.CurrentModule, s.CurrentRoute, s.CurrentRecordID, s.IsArchived,
	)
	if err != nil {
		return fmt.Errorf("failed to insert copilot session: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.ID = id
	}
	return nil
}

func (r *sqlRepository) GetSession(ctx context.Context, orgID int64, sessionID string) (*CopilotSession, error) {
	query := `
		SELECT id, org_id, user_id, session_id, title, current_module, current_route, current_record_id, is_archived, created_at, updated_at
		FROM copilot_sessions
		WHERE org_id = ? AND session_id = ?
		LIMIT 1
	`
	var s CopilotSession
	err := r.db.GetContext(ctx, &s, query, orgID, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get copilot session: %w", err)
	}
	return &s, nil
}

func (r *sqlRepository) ListSessions(ctx context.Context, orgID int64, userID int64, limit int) ([]CopilotSession, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `
		SELECT id, org_id, user_id, session_id, title, current_module, current_route, current_record_id, is_archived, created_at, updated_at
		FROM copilot_sessions
		WHERE org_id = ? AND user_id = ? AND is_archived = 0
		ORDER BY updated_at DESC
		LIMIT ?
	`
	var sessions []CopilotSession
	err := r.db.SelectContext(ctx, &sessions, query, orgID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list copilot sessions: %w", err)
	}
	return sessions, nil
}

func (r *sqlRepository) ArchiveSession(ctx context.Context, orgID int64, sessionID string) error {
	query := `
		UPDATE copilot_sessions
		SET is_archived = 1, updated_at = NOW()
		WHERE org_id = ? AND session_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, orgID, sessionID)
	return err
}

func (r *sqlRepository) UpdateSessionContext(ctx context.Context, orgID int64, sessionID, module, route string, recordID *string) error {
	query := `
		UPDATE copilot_sessions
		SET current_module = ?, current_route = ?, current_record_id = ?, updated_at = NOW()
		WHERE org_id = ? AND session_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, module, route, recordID, orgID, sessionID)
	return err
}

// Message Methods
func (r *sqlRepository) SaveMessage(ctx context.Context, m *CopilotMessage) error {
	query := `
		INSERT INTO copilot_messages (
			session_id, org_id, user_id, role, content, confirmed_facts, source_references,
			action_proposals, draft_content, draft_type, confidence_score, correlation_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		m.SessionID, m.OrgID, m.UserID, m.Role, m.Content, m.ConfirmedFacts, m.SourceReferences,
		m.ActionProposals, m.DraftContent, m.DraftType, m.ConfidenceScore, m.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert copilot message: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		m.ID = id
	}
	// touch session updated_at
	_, _ = r.db.ExecContext(ctx, `UPDATE copilot_sessions SET updated_at = NOW() WHERE org_id = ? AND session_id = ?`, m.OrgID, m.SessionID)
	return nil
}

func (r *sqlRepository) ListMessages(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT id, session_id, org_id, user_id, role, content, confirmed_facts, source_references,
		       action_proposals, draft_content, draft_type, confidence_score, correlation_id, created_at
		FROM copilot_messages
		WHERE org_id = ? AND session_id = ?
		ORDER BY created_at ASC
		LIMIT ?
	`
	var messages []CopilotMessage
	err := r.db.SelectContext(ctx, &messages, query, orgID, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list copilot messages: %w", err)
	}
	return messages, nil
}

// Action History Methods
func (r *sqlRepository) LogAction(ctx context.Context, a *CopilotActionHistory) error {
	query := `
		INSERT INTO copilot_action_history (
			org_id, user_id, session_id, action_type, action_title, action_payload,
			approval_id, recommendation_id, status, result_summary, correlation_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		a.OrgID, a.UserID, a.SessionID, a.ActionType, a.ActionTitle, a.ActionPayload,
		a.ApprovalID, a.RecommendationID, a.Status, a.ResultSummary, a.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to log copilot action: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		a.ID = id
	}
	return nil
}

func (r *sqlRepository) UpdateActionStatus(ctx context.Context, orgID int64, id int64, status string, resultSummary *string) error {
	query := `
		UPDATE copilot_action_history
		SET status = ?, result_summary = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, resultSummary, orgID, id)
	return err
}

func (r *sqlRepository) GetActionByID(ctx context.Context, orgID int64, id int64) (*CopilotActionHistory, error) {
	query := `
		SELECT id, org_id, user_id, session_id, action_type, action_title, action_payload,
		       approval_id, recommendation_id, status, result_summary, correlation_id, created_at, updated_at
		FROM copilot_action_history
		WHERE org_id = ? AND id = ?
		LIMIT 1
	`
	var a CopilotActionHistory
	err := r.db.GetContext(ctx, &a, query, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	return &a, nil
}

func (r *sqlRepository) ListActions(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotActionHistory, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var query string
	var args []interface{}
	if sessionID != "" {
		query = `
			SELECT id, org_id, user_id, session_id, action_type, action_title, action_payload,
			       approval_id, recommendation_id, status, result_summary, correlation_id, created_at, updated_at
			FROM copilot_action_history
			WHERE org_id = ? AND session_id = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{orgID, sessionID, limit}
	} else {
		query = `
			SELECT id, org_id, user_id, session_id, action_type, action_title, action_payload,
			       approval_id, recommendation_id, status, result_summary, correlation_id, created_at, updated_at
			FROM copilot_action_history
			WHERE org_id = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{orgID, limit}
	}
	var actions []CopilotActionHistory
	err := r.db.SelectContext(ctx, &actions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}
	return actions, nil
}

// RetrieveModuleContext queries authorized records and metrics strictly filtered by orgID.
func (r *sqlRepository) RetrieveModuleContext(ctx context.Context, orgID int64, module string, recordID *string) ([]map[string]interface{}, map[string]interface{}, error) {
	records := make([]map[string]interface{}, 0)
	metrics := make(map[string]interface{})

	modUpper := strings.ToUpper(strings.TrimSpace(module))

	switch modUpper {
	case "DASHBOARD":
		// Compute aggregate operational metrics
		var totalShipments, openInvoices, activeRFQs, pendingApprovals int
		_ = r.db.GetContext(ctx, &totalShipments, `SELECT COUNT(*) FROM shipments WHERE org_id = ?`, orgID)
		_ = r.db.GetContext(ctx, &openInvoices, `SELECT COUNT(*) FROM invoices WHERE org_id = ? AND status != 'PAID'`, orgID)
		_ = r.db.GetContext(ctx, &activeRFQs, `SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND status NOT IN ('CLOSED', 'REJECTED')`, orgID)
		_ = r.db.GetContext(ctx, &pendingApprovals, `SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND status IN ('Pending', 'PENDING_APPROVAL', 'In Review')`, orgID)

		metrics["total_shipments"] = totalShipments
		metrics["open_invoices"] = openInvoices
		metrics["active_rfqs"] = activeRFQs
		metrics["pending_approvals"] = pendingApprovals

		// Sample 5 active shipments
		rows, err := r.db.QueryxContext(ctx, `SELECT id, booking_number, carrier_scac, status, etd, eta FROM shipments WHERE org_id = ? ORDER BY id DESC LIMIT 5`, orgID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				m := make(map[string]interface{})
				if err := rows.MapScan(m); err == nil {
					records = append(records, sanitizeDBMap(m))
				}
			}
		}

	case "SHIPMENTS", "TRACKING", "EXCEPTIONS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, booking_number, mbl_number, hbl_number, carrier_scac, vessel_name, voyage_number,
				       origin_port, destination_port, status, etd, eta, created_at
				FROM shipments
				WHERE org_id = ? AND (id = ? OR booking_number = ? OR mbl_number = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, booking_number, mbl_number, hbl_number, carrier_scac, vessel_name, voyage_number,
				       origin_port, destination_port, status, etd, eta
				FROM shipments
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "INVOICES":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, number, amount_due, amount_paid, status, issued_at, paid_at
				FROM invoices
				WHERE org_id = ? AND (id = ? OR number = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, number, amount_due, amount_paid, status, issued_at, paid_at
				FROM invoices
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "CUSTOMERS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, name, domain, industry, contact_name, contact_email, contact_phone, customer_code
				FROM customers
				WHERE org_id = ? AND (id = ? OR customer_code = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, name, domain, industry, contact_name, contact_email, contact_phone, customer_code
				FROM customers
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "LEADS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, company_name, contact_name, email, phone, source, status, ai_score
				FROM leads
				WHERE org_id = ? AND (id = ? OR company_name = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, company_name, contact_name, email, phone, source, status, ai_score
				FROM leads
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "RFQS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, rfq_number, customer_id, stage, status, origin, destination, incoterms
				FROM rfqs
				WHERE org_id = ? AND (id = ? OR rfq_number = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, rfq_number, customer_id, stage, status, origin, destination, incoterms
				FROM rfqs
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "QUOTATIONS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, quotation_number, customer_name, rfq_number, status, origin
				FROM quotations
				WHERE org_id = ? AND (id = ? OR quotation_number = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, quotation_number, customer_name, rfq_number, status, origin
				FROM quotations
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "BOOKINGS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, booking_number, carrier_name, carrier_scac, carrier_booking_reference, carrier_booking_status
				FROM bookings
				WHERE org_id = ? AND (id = ? OR booking_number = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, booking_number, carrier_name, carrier_scac, carrier_booking_reference, carrier_booking_status
				FROM bookings
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "CONTRACTS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, contract_reference, contract_name, contract_type, party_name, transport_mode,
				       status, currency, contract_value, effective_date, expiry_date, owner
				FROM contracts
				WHERE org_id = ? AND (id = ? OR contract_reference = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, contract_reference, contract_name, contract_type, party_name, transport_mode,
				       status, currency, contract_value, effective_date, expiry_date, owner
				FROM contracts
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "DOCUMENTS":
		query := `
			SELECT id, contract_id, file_name, document_type, upload_status, created_at
			FROM contract_documents
			WHERE org_id = ?
			ORDER BY id DESC
			LIMIT 10
		`
		rows, err := r.db.QueryxContext(ctx, query, orgID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				m := make(map[string]interface{})
				if err := rows.MapScan(m); err == nil {
					records = append(records, sanitizeDBMap(m))
				}
			}
		}

	case "COMPLIANCE":
		query := `
			SELECT id, contract_id, overall_status, overall_score, risk_level, review_summary, created_at
			FROM ai_contract_compliance_reviews
			WHERE org_id = ?
			ORDER BY id DESC
			LIMIT 10
		`
		rows, err := r.db.QueryxContext(ctx, query, orgID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				m := make(map[string]interface{})
				if err := rows.MapScan(m); err == nil {
					records = append(records, sanitizeDBMap(m))
				}
			}
		}

	case "APPROVALS":
		if recordID != nil && *recordID != "" {
			query := `
				SELECT id, request_code, title, category, type, status, priority, customer_name, due_text, action_name
				FROM approval_requests
				WHERE org_id = ? AND (id = ? OR request_code = ?)
				LIMIT 1
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID, *recordID, *recordID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		} else {
			query := `
				SELECT id, request_code, title, category, type, status, priority, customer_name, due_text, action_name
				FROM approval_requests
				WHERE org_id = ?
				ORDER BY id DESC
				LIMIT 10
			`
			rows, err := r.db.QueryxContext(ctx, query, orgID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					m := make(map[string]interface{})
					if err := rows.MapScan(m); err == nil {
						records = append(records, sanitizeDBMap(m))
					}
				}
			}
		}

	case "NOTIFICATIONS":
		query := `
			SELECT id, title, message, severity, priority, status, is_read, is_escalated, created_at
			FROM notifications
			WHERE org_id = ?
			ORDER BY id DESC
			LIMIT 10
		`
		rows, err := r.db.QueryxContext(ctx, query, orgID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				m := make(map[string]interface{})
				if err := rows.MapScan(m); err == nil {
					records = append(records, sanitizeDBMap(m))
				}
			}
		}

	case "AUDIT_LOGS", "AI_WORKFORCE":
		query := `
			SELECT id, user_id, action, entity_type, entity_id, created_at
			FROM audit_logs
			WHERE org_id = ?
			ORDER BY id DESC
			LIMIT 10
		`
		rows, err := r.db.QueryxContext(ctx, query, orgID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				m := make(map[string]interface{})
				if err := rows.MapScan(m); err == nil {
					records = append(records, sanitizeDBMap(m))
				}
			}
		}
	}

	metrics["record_count"] = len(records)
	return records, metrics, nil
}

// CreateApprovalRequest inserts a pending approval request into approval_requests for HITL gating.
func (r *sqlRepository) CreateApprovalRequest(ctx context.Context, orgID, userID int64, userName, title, category, actionName string, payload map[string]interface{}, corrID string) (int64, error) {
	requestCode := fmt.Sprintf("APR-COPILOT-%d", time.Now().UnixNano()%1000000)
	payloadJSON, _ := json.Marshal(payload)
	payloadStr := string(payloadJSON)

	query := `
		INSERT INTO approval_requests (
			org_id, request_code, title, category, type, status, priority,
			requested_by_id, requested_by_name, description,
			actor_type, source, action_name, risk_level, proposed_payload,
			correlation_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, 'AI Copilot Action', 'Pending', 'MEDIUM',
			?, ?, ?,
			'AI_COPILOT', 'COPILOT', ?, 'MEDIUM', ?,
			?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		orgID, requestCode, title, category,
		userID, userName, fmt.Sprintf("Copilot requested action: %s", actionName),
		actionName, payloadStr, corrID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create approval request: %w", err)
	}
	return res.LastInsertId()
}

// CreateRecommendation inserts an actionable recommendation into ai_recommendations table.
func (r *sqlRepository) CreateRecommendation(ctx context.Context, orgID int64, title, category, description, sourceModule string, payload map[string]interface{}, corrID string) (int64, error) {
	recCode := fmt.Sprintf("REC-COPILOT-%d", time.Now().UnixNano()%1000000)
	payloadJSON, _ := json.Marshal(payload)
	payloadStr := string(payloadJSON)

	query := `
		INSERT INTO ai_recommendations (
			org_id, source_type, source_id, source_reference, title, description,
			category, priority, risk_level, confidence, confidence_score, evidence,
			recommended_action, action_type, status, freshness, requires_approval,
			correlation_id, created_by, generated_by, rule_applied, dedup_hash, draft_status,
			created_at, updated_at
		) VALUES (
			?, ?, 0, ?, ?, ?,
			?, 'medium', 'medium', 'HIGH', 0.95, ?,
			?, 'COPILOT_ACTION', 'new', NOW(), 0,
			?, 'AI_COPILOT', 'COPILOT', 'COPILOT_USER_REQUEST', ?, 'NONE',
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		orgID, sourceModule, recCode, title, description,
		category, payloadStr,
		description, corrID, recCode,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create recommendation: %w", err)
	}
	return res.LastInsertId()
}

// Helper to convert byte slices (often from MySQL strings/dates) into clean strings/numbers
func sanitizeDBMap(m map[string]interface{}) map[string]interface{} {
	clean := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case []byte:
			clean[k] = string(val)
		case time.Time:
			clean[k] = val.Format(time.RFC3339)
		default:
			clean[k] = val
		}
	}
	return clean
}

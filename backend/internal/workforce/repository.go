package workforce

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrAgentNotFound   = errors.New("workforce agent not found")
	ErrTaskNotFound    = errors.New("workforce task not found")
	ErrHandoffNotFound = errors.New("workforce handoff not found")
)

type Repository interface {
	SeedBaselineAgents(ctx context.Context) error
	ListAgents(ctx context.Context, orgID int64, agentType string, isEnabled *bool) ([]*WorkforceAgent, error)
	GetAgent(ctx context.Context, orgID int64, agentID string) (*WorkforceAgent, error)
	UpsertAgent(ctx context.Context, agent *WorkforceAgent) error

	CreateTask(ctx context.Context, task *WorkforceTask) error
	GetTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*WorkforceTask, int, error)
	UpdateTaskStatus(ctx context.Context, orgID int64, taskID string, status TaskStatus, result json.RawMessage, confidence float64, errCode, errMsg *string) error
	GetTaskHierarchy(ctx context.Context, orgID int64, rootOrTaskID string) (*TaskHierarchyNode, error)

	CreateMessage(ctx context.Context, msg *WorkforceMessage) error
	ListMessages(ctx context.Context, orgID int64, taskID string) ([]*WorkforceMessage, error)

	CreateContextItem(ctx context.Context, ctxItem *WorkforceContextItem) error
	ListContextItems(ctx context.Context, orgID int64, taskID string) ([]*WorkforceContextItem, error)

	CreateHandoff(ctx context.Context, handoff *WorkforceHandoff) error
	ListHandoffs(ctx context.Context, orgID int64, taskID string) ([]*WorkforceHandoff, error)
	UpdateHandoffStatus(ctx context.Context, orgID int64, handoffID string, status HandoffStatus) error

	// Phase 6.5: Shipment & Exception context extraction
	GetShipmentContext(ctx context.Context, orgID int64, shipmentID string) (map[string]interface{}, error)
	GetExceptionContext(ctx context.Context, orgID int64, shipmentID, exceptionID string) (map[string]interface{}, error)

	// Phase 6.6: Commercial context extraction (Customer, Lead, RFQ, Invoice)
	GetCustomerContext(ctx context.Context, orgID int64, customerID string) (map[string]interface{}, error)
	GetLeadContext(ctx context.Context, orgID int64, leadID string) (map[string]interface{}, error)
	GetRFQContext(ctx context.Context, orgID int64, rfqID string) (map[string]interface{}, error)
	GetInvoiceContext(ctx context.Context, orgID int64, customerID, invoiceID string) (map[string]interface{}, error)
	GetContractContext(ctx context.Context, orgID int64, contractID string) (map[string]interface{}, error)
	GetComplianceContext(ctx context.Context, orgID int64, shipmentID string) (map[string]interface{}, error)

	// Phase 6.8: Memory & Learning from Outcomes
	GetMemoryContext(ctx context.Context, orgID int64, domain, entityType, entityID string, limit int) ([]MemoryItemDTO, error)
	RecordOutcome(ctx context.Context, outcome *AgentOutcome) error
	ListOutcomes(ctx context.Context, orgID int64, entityType, entityID string, limit int) ([]AgentOutcome, error)
	GetOutcomesSummary(ctx context.Context, orgID int64) (map[string]interface{}, error)

	// Phase 6.9: Governed Autonomy & Command Center
	GetWorkforceHealth(ctx context.Context, orgID int64) (*WorkforceHealthSummary, error)
	GetAgentWorkload(ctx context.Context, orgID int64) ([]AgentWorkloadMetrics, error)
	UpdateAgentControl(ctx context.Context, orgID int64, agentID string, isEnabled bool, healthStatus string, autonomyLevel string) error
}

type MySQLRepository struct {
	db *sqlx.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: sqlx.NewDb(db, "mysql")}
}

func NewMySQLRepositorySqlx(db *sqlx.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// SeedBaselineAgents seeds the 10 specialized workforce agents if not already in DB
func (r *MySQLRepository) SeedBaselineAgents(ctx context.Context) error {
	baseline := []struct {
		AgentID         string
		AgentType       AgentType
		Name            string
		Description     string
		Capabilities    []string
		AllowedTasks    []string
		AllowedEntities []string
		AutonomyLevel   string
	}{
		{
			AgentID:         "planning_agent",
			AgentType:       AgentTypeCoordinator,
			Name:            "Operational Planning Coordinator",
			Description:     "Coordinates complex multi-step logistics objectives, constructs execution plans, and delegates to specialist agents.",
			Capabilities:    []string{CapPlanningCreate, CapTaskDelegate, CapPlanningEvaluate, CapMonitoringObserve},
			AllowedTasks:    []string{"PLAN_CREATION", "WORKFORCE_DELEGATION", "INCIDENT_COORDINATION", "MULTI_AGENT_EXECUTION"},
			AllowedEntities: []string{"SHIPMENT", "INVOICE", "CUSTOMER", "RFQ", "CONTRACT"},
			AutonomyLevel:   "LEVEL_2_PREPARE",
		},
		{
			AgentID:         "shipment_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Shipment Operations Specialist",
			Description:     "Analyzes shipment operational status, ETA movements, container tracking, and carrier milestones.",
			Capabilities:    []string{CapShipmentRead, CapShipmentAnalyze, CapShipmentPredict, CapMonitoringObserve},
			AllowedTasks:    []string{"SHIPMENT_INSPECTION", "TRACKING_UPDATE", "MILESTONE_VERIFICATION", "ETA_FORECAST"},
			AllowedEntities: []string{"SHIPMENT", "CONTAINER", "CARRIER", "VESSEL"},
			AutonomyLevel:   "LEVEL_2_PREPARE",
		},
		{
			AgentID:         "exception_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Exception Resolution Specialist",
			Description:     "Assesses disruptions, port congestion, weather anomalies, and drafts operational mitigation options.",
			Capabilities:    []string{CapExceptionRead, CapExceptionAnalyze, CapExceptionRecommend, CapShipmentRead},
			AllowedTasks:    []string{"EXCEPTION_TRIAGE", "REROUTE_ASSESSMENT", "INCIDENT_ANALYSIS", "DISRUPTION_MITIGATION"},
			AllowedEntities: []string{"SHIPMENT", "EXCEPTION", "CARRIER", "DISRUPTION"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "customer_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Customer Intelligence Specialist",
			Description:     "Handles customer communication analysis, response drafts, sentiment tracking, and relationship health.",
			Capabilities:    []string{CapCustomerRead, CapCustomerAnalyze, CapCustomerFollowupRecommend},
			AllowedTasks:    []string{"CUSTOMER_FOLLOWUP", "SENTIMENT_ANALYSIS", "INQUIRY_TRIAGE", "OUTREACH_RECOMMENDATION"},
			AllowedEntities: []string{"CUSTOMER", "LEAD", "COMMUNICATION", "ACCOUNT"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "pricing_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Pricing & Margin Specialist",
			Description:     "Evaluates freight rate market trends, spot quotes, margin thresholds, and surcharge reconciliation.",
			Capabilities:    []string{CapRFQRead, CapPricingAnalyze, CapPricingRecommend, "rate.read"},
			AllowedTasks:    []string{"RATE_BENCHMARK", "QUOTE_OPTIMIZATION", "MARGIN_ANALYSIS", "RFQ_PRICING"},
			AllowedEntities: []string{"RFQ", "QUOTATION", "RATE", "SURCHARGE"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "finance_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Finance & Collections Specialist",
			Description:     "Inspects freight billing discrepancies, aging receivables, collections escalation, and credit limits.",
			Capabilities:    []string{CapInvoiceRead, CapFinanceAnalyze, CapFinanceRecommend},
			AllowedTasks:    []string{"INVOICE_AUDIT", "COLLECTIONS_TRIAGE", "DISCREPANCY_RECONCILE", "CASH_FLOW_RISK"},
			AllowedEntities: []string{"INVOICE", "PAYMENT", "CUSTOMER", "LEDGER"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "contract_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Contract Agreement Specialist",
			Description:     "Extracts contractual conditions, free days, rate validity, and cross-references operational events with contract terms.",
			Capabilities:    []string{CapContractRead, CapContractAnalyze},
			AllowedTasks:    []string{"CONTRACT_ANALYSIS", "CLAUSE_EXTRACTION", "TERMS_VERIFICATION", "FREE_DAYS_CHECK"},
			AllowedEntities: []string{"CONTRACT", "AGREEMENT", "RATE_VERSION", "CARRIER"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "compliance_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Contract Compliance Specialist",
			Description:     "Audits contract terms, commercial obligations, SLA adherence, and regulatory documentation.",
			Capabilities:    []string{CapComplianceRead, CapComplianceAnalyze, "document.read"},
			AllowedTasks:    []string{"COMPLIANCE_AUDIT", "DOCUMENT_VERIFICATION", "CUSTOMS_SCREENING", "HAZMAT_CHECK"},
			AllowedEntities: []string{"CONTRACT", "DOCUMENT", "CUSTOMS", "COMPLIANCE_FILING"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
		{
			AgentID:         "monitoring_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Workforce & Systems Observer",
			Description:     "Continuously monitors workforce tasks, detects stalled tasks, tracks agent SLA, and flags anomalies.",
			Capabilities:    []string{CapWorkforceObserve, CapTaskMonitor, CapMonitoringObserve, "anomaly.detect"},
			AllowedTasks:    []string{"STATE_OBSERVATION", "ANOMALY_DETECTION", "WORKFORCE_HEALTH_CHECK", "TASK_MONITORING"},
			AllowedEntities: []string{"WORKFORCE_TASK", "AGENT", "SYSTEM_HEALTH"},
			AutonomyLevel:   "LEVEL_0_OBSERVE",
		},
		{
			AgentID:         "memory_agent",
			AgentType:       AgentTypeSpecialist,
			Name:            "Operational Memory & Learning Specialist",
			Description:     "Retrieves prior outcomes, identifies recurring operational patterns, and provides learned lessons to workforce agents.",
			Capabilities:    []string{CapMemoryRetrieve, CapMemoryAnalyze, CapOutcomeRecord},
			AllowedTasks:    []string{"MEMORY_RETRIEVAL", "PATTERN_ANALYSIS", "OUTCOME_SYNTHESIS", "LEARNING_CLASSIFICATION"},
			AllowedEntities: []string{"OUTCOME", "MEMORY", "PLAN", "LESSON_LEARNED"},
			AutonomyLevel:   "LEVEL_1_RECOMMEND",
		},
	}


	for _, a := range baseline {
		capsJSON, _ := json.Marshal(a.Capabilities)
		tasksJSON, _ := json.Marshal(a.AllowedTasks)
		entitiesJSON, _ := json.Marshal(a.AllowedEntities)

		query := `
			INSERT INTO workforce_agents 
				(org_id, agent_id, agent_type, name, description, capabilities, allowed_tasks, allowed_entities, autonomy_level, is_enabled, version, prompt_version, health_status)
			VALUES 
				(0, ?, ?, ?, ?, ?, ?, ?, ?, 1, '1.0.0', '1.0.0', 'HEALTHY')
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				description = VALUES(description),
				capabilities = VALUES(capabilities),
				allowed_tasks = VALUES(allowed_tasks),
				allowed_entities = VALUES(allowed_entities),
				autonomy_level = VALUES(autonomy_level)
		`
		_, err := r.db.ExecContext(ctx, query,
			a.AgentID, string(a.AgentType), a.Name, a.Description,
			capsJSON, tasksJSON, entitiesJSON, a.AutonomyLevel,
		)
		if err != nil {
			return fmt.Errorf("failed seeding baseline agent %s: %w", a.AgentID, err)
		}
	}
	return nil
}

func (r *MySQLRepository) ListAgents(ctx context.Context, orgID int64, agentType string, isEnabled *bool) ([]*WorkforceAgent, error) {
	query := `
		SELECT id, org_id, agent_id, agent_type, name, description, capabilities, allowed_tasks, 
		       allowed_entities, autonomy_level, is_enabled, version, prompt_version, health_status, 
		       last_execution_at, COALESCE(last_execution_info, '{}') AS last_execution_info, created_at, updated_at
		FROM workforce_agents
		WHERE (org_id = ? OR org_id = 0)
	`
	args := []interface{}{orgID}

	if agentType != "" {
		query += " AND agent_type = ?"
		args = append(args, agentType)
	}
	if isEnabled != nil {
		query += " AND is_enabled = ?"
		args = append(args, *isEnabled)
	}
	query += " ORDER BY org_id DESC, name ASC"

	var agents []*WorkforceAgent
	err := r.db.SelectContext(ctx, &agents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed listing workforce agents: %w", err)
	}
	return agents, nil
}

func (r *MySQLRepository) GetAgent(ctx context.Context, orgID int64, agentID string) (*WorkforceAgent, error) {
	query := `
		SELECT id, org_id, agent_id, agent_type, name, description, capabilities, allowed_tasks, 
		       allowed_entities, autonomy_level, is_enabled, version, prompt_version, health_status, 
		       last_execution_at, COALESCE(last_execution_info, '{}') AS last_execution_info, created_at, updated_at
		FROM workforce_agents
		WHERE (org_id = ? OR org_id = 0) AND agent_id = ?
		ORDER BY org_id DESC
		LIMIT 1
	`
	var agent WorkforceAgent
	err := r.db.GetContext(ctx, &agent, query, orgID, agentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAgentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed getting workforce agent %s: %w", agentID, err)
	}
	return &agent, nil
}

func (r *MySQLRepository) UpsertAgent(ctx context.Context, agent *WorkforceAgent) error {
	query := `
		INSERT INTO workforce_agents 
			(org_id, agent_id, agent_type, name, description, capabilities, allowed_tasks, allowed_entities, autonomy_level, is_enabled, version, prompt_version, health_status)
		VALUES 
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			description = VALUES(description),
			capabilities = VALUES(capabilities),
			allowed_tasks = VALUES(allowed_tasks),
			allowed_entities = VALUES(allowed_entities),
			autonomy_level = VALUES(autonomy_level),
			is_enabled = VALUES(is_enabled),
			health_status = VALUES(health_status),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query,
		agent.OrgID, agent.AgentID, string(agent.AgentType), agent.Name, agent.Description,
		agent.Capabilities, agent.AllowedTasks, agent.AllowedEntities, agent.AutonomyLevel,
		agent.IsEnabled, agent.Version, agent.PromptVersion, agent.HealthStatus,
	)
	if err != nil {
		return fmt.Errorf("failed upserting workforce agent: %w", err)
	}
	return nil
}

func (r *MySQLRepository) CreateTask(ctx context.Context, task *WorkforceTask) error {
	query := `
		INSERT INTO workforce_tasks 
			(task_id, org_id, objective, initiating_event, parent_task_id, root_task_id, assigned_agent_id, status, priority, context_reference, required_capabilities, dependencies, result, confidence, error_code, error_message, correlation_id, started_at, completed_at)
		VALUES 
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		task.TaskID, task.OrgID, task.Objective, task.InitiatingEvent, task.ParentTaskID, task.RootTaskID,
		task.AssignedAgentID, string(task.Status), string(task.Priority), task.ContextReference,
		task.RequiredCapabilities, task.Dependencies, task.Result, task.Confidence, task.ErrorCode,
		task.ErrorMessage, task.CorrelationID, task.StartedAt, task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed inserting workforce task: %w", err)
	}
	return nil
}

func (r *MySQLRepository) GetTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, error) {
	query := `
		SELECT id, task_id, org_id, objective, initiating_event, parent_task_id, root_task_id, 
		       assigned_agent_id, status, priority, 
		       COALESCE(context_reference, '{}') AS context_reference, 
		       required_capabilities, 
		       COALESCE(dependencies, '[]') AS dependencies, 
		       COALESCE(result, '{}') AS result, 
		       confidence, error_code, error_message, correlation_id, 
		       started_at, completed_at, created_at, updated_at
		FROM workforce_tasks
		WHERE org_id = ? AND task_id = ?
	`
	var task WorkforceTask
	err := r.db.GetContext(ctx, &task, query, orgID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed getting workforce task: %w", err)
	}
	return &task, nil
}

func (r *MySQLRepository) ListTasks(ctx context.Context, filter TaskFilter) ([]*WorkforceTask, int, error) {
	where := "WHERE org_id = ?"
	args := []interface{}{filter.OrgID}

	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.AssignedAgentID != "" {
		where += " AND assigned_agent_id = ?"
		args = append(args, filter.AssignedAgentID)
	}
	if filter.ParentTaskID != nil {
		where += " AND parent_task_id = ?"
		args = append(args, *filter.ParentTaskID)
	}
	if filter.RootTaskID != nil {
		where += " AND root_task_id = ?"
		args = append(args, *filter.RootTaskID)
	}
	if filter.CorrelationID != "" {
		where += " AND correlation_id = ?"
		args = append(args, filter.CorrelationID)
	}

	countQuery := "SELECT COUNT(*) FROM workforce_tasks " + where
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed counting workforce tasks: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, task_id, org_id, objective, initiating_event, parent_task_id, root_task_id, 
		       assigned_agent_id, status, priority, 
		       COALESCE(context_reference, '{}') AS context_reference, 
		       required_capabilities, 
		       COALESCE(dependencies, '[]') AS dependencies, 
		       COALESCE(result, '{}') AS result, 
		       confidence, error_code, error_message, correlation_id, 
		       started_at, completed_at, created_at, updated_at
		FROM workforce_tasks ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var tasks []*WorkforceTask
	err = r.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed querying workforce tasks: %w", err)
	}
	return tasks, total, nil
}

func (r *MySQLRepository) UpdateTaskStatus(ctx context.Context, orgID int64, taskID string, status TaskStatus, result json.RawMessage, confidence float64, errCode, errMsg *string) error {
	var completedAt *time.Time
	if status == TaskStatusCompleted || status == TaskStatusFailed || status == TaskStatusCancelled {
		now := time.Now().UTC()
		completedAt = &now
	}

	query := `
		UPDATE workforce_tasks
		SET status = ?,
		    result = COALESCE(?, result),
		    confidence = CASE WHEN ? > 0 THEN ? ELSE confidence END,
		    error_code = ?,
		    error_message = ?,
		    completed_at = COALESCE(?, completed_at),
		    updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND task_id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		string(status), result, confidence, confidence, errCode, errMsg, completedAt, orgID, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed updating workforce task status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *MySQLRepository) GetTaskHierarchy(ctx context.Context, orgID int64, rootOrTaskID string) (*TaskHierarchyNode, error) {
	// First fetch the target task
	rootTask, err := r.GetTask(ctx, orgID, rootOrTaskID)
	if err != nil {
		return nil, err
	}

	rootNode := &TaskHierarchyNode{
		Task:     rootTask,
		Children: make([]*TaskHierarchyNode, 0),
	}

	// Fetch all descendant tasks sharing root_task_id or parent_task_id
	var descendants []*WorkforceTask
	query := `
		SELECT id, task_id, org_id, objective, initiating_event, parent_task_id, root_task_id, 
		       assigned_agent_id, status, priority, 
		       COALESCE(context_reference, '{}') AS context_reference, 
		       required_capabilities, 
		       COALESCE(dependencies, '[]') AS dependencies, 
		       COALESCE(result, '{}') AS result, 
		       confidence, error_code, error_message, correlation_id, 
		       started_at, completed_at, created_at, updated_at
		FROM workforce_tasks
		WHERE org_id = ? AND (root_task_id = ? OR parent_task_id = ?)
		ORDER BY created_at ASC
	`
	err = r.db.SelectContext(ctx, &descendants, query, orgID, rootOrTaskID, rootOrTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed querying child tasks: %w", err)
	}

	// Index nodes by task_id
	nodeMap := make(map[string]*TaskHierarchyNode)
	nodeMap[rootTask.TaskID] = rootNode

	for _, d := range descendants {
		if d.TaskID == rootTask.TaskID {
			continue
		}
		nodeMap[d.TaskID] = &TaskHierarchyNode{
			Task:     d,
			Children: make([]*TaskHierarchyNode, 0),
		}
	}

	// Link children to parents
	for _, d := range descendants {
		if d.TaskID == rootTask.TaskID {
			continue
		}
		childNode := nodeMap[d.TaskID]
		if d.ParentTaskID != nil && *d.ParentTaskID != "" {
			if parentNode, exists := nodeMap[*d.ParentTaskID]; exists {
				parentNode.Children = append(parentNode.Children, childNode)
				continue
			}
		}
		// If direct parent not in map, link to root
		rootNode.Children = append(rootNode.Children, childNode)
	}

	return rootNode, nil
}

func (r *MySQLRepository) CreateMessage(ctx context.Context, msg *WorkforceMessage) error {
	query := `
		INSERT INTO workforce_messages 
			(message_id, org_id, task_id, parent_task_id, sender_agent_id, recipient_agent_id, message_type, objective, requested_capability, context_reference, payload, priority, correlation_id)
		VALUES 
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.MessageID, msg.OrgID, msg.TaskID, msg.ParentTaskID, msg.SenderAgentID,
		msg.RecipientAgentID, string(msg.MessageType), msg.Objective, msg.RequestedCapability,
		msg.ContextReference, msg.Payload, string(msg.Priority), msg.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed inserting workforce message: %w", err)
	}
	return nil
}

func (r *MySQLRepository) ListMessages(ctx context.Context, orgID int64, taskID string) ([]*WorkforceMessage, error) {
	query := `
		SELECT id, message_id, org_id, task_id, parent_task_id, sender_agent_id, recipient_agent_id, 
		       message_type, objective, requested_capability, 
		       COALESCE(context_reference, '{}') AS context_reference, 
		       COALESCE(payload, '{}') AS payload, 
		       priority, correlation_id, created_at
		FROM workforce_messages
		WHERE org_id = ? AND task_id = ?
		ORDER BY created_at ASC
	`
	var messages []*WorkforceMessage
	err := r.db.SelectContext(ctx, &messages, query, orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed querying workforce messages: %w", err)
	}
	return messages, nil
}

func (r *MySQLRepository) CreateContextItem(ctx context.Context, ctxItem *WorkforceContextItem) error {
	query := `
		INSERT INTO workforce_contexts 
			(context_id, org_id, task_id, item_type, entity_type, entity_id, source_agent_id, provenance_id, content, confidence, is_authoritative)
		VALUES 
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		ctxItem.ContextID, ctxItem.OrgID, ctxItem.TaskID, string(ctxItem.ItemType),
		ctxItem.EntityType, ctxItem.EntityID, ctxItem.SourceAgentID, ctxItem.ProvenanceID,
		ctxItem.Content, ctxItem.Confidence, ctxItem.IsAuthoritative,
	)
	if err != nil {
		return fmt.Errorf("failed inserting workforce context item: %w", err)
	}
	return nil
}

func (r *MySQLRepository) ListContextItems(ctx context.Context, orgID int64, taskID string) ([]*WorkforceContextItem, error) {
	query := `
		SELECT id, context_id, org_id, task_id, item_type, entity_type, entity_id, source_agent_id, 
		       provenance_id, content, confidence, is_authoritative, created_at
		FROM workforce_contexts
		WHERE org_id = ? AND task_id = ?
		ORDER BY is_authoritative DESC, created_at ASC
	`
	var items []*WorkforceContextItem
	err := r.db.SelectContext(ctx, &items, query, orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed querying workforce contexts: %w", err)
	}
	return items, nil
}

func (r *MySQLRepository) CreateHandoff(ctx context.Context, handoff *WorkforceHandoff) error {
	query := `
		INSERT INTO workforce_handoffs 
			(handoff_id, org_id, task_id, originating_task_id, source_agent_id, destination_agent_id, reason, objective, required_capability, context_references, current_findings, expected_output, confidence, status)
		VALUES 
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		handoff.HandoffID, handoff.OrgID, handoff.TaskID, handoff.OriginatingTaskID,
		handoff.SourceAgentID, handoff.DestinationAgentID, handoff.Reason, handoff.Objective,
		handoff.RequiredCapability, handoff.ContextReferences, handoff.CurrentFindings,
		handoff.ExpectedOutput, handoff.Confidence, string(handoff.Status),
	)
	if err != nil {
		return fmt.Errorf("failed inserting workforce handoff: %w", err)
	}
	return nil
}

func (r *MySQLRepository) ListHandoffs(ctx context.Context, orgID int64, taskID string) ([]*WorkforceHandoff, error) {
	query := `
		SELECT id, handoff_id, org_id, task_id, originating_task_id, source_agent_id, 
		       destination_agent_id, reason, objective, required_capability, 
		       COALESCE(context_references, '[]') AS context_references, 
		       COALESCE(current_findings, '{}') AS current_findings, 
		       COALESCE(expected_output, '{}') AS expected_output, 
		       confidence, status, created_at
		FROM workforce_handoffs
		WHERE org_id = ? AND task_id = ?
		ORDER BY created_at ASC
	`
	var handoffs []*WorkforceHandoff
	err := r.db.SelectContext(ctx, &handoffs, query, orgID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed querying workforce handoffs: %w", err)
	}
	return handoffs, nil
}

func (r *MySQLRepository) UpdateHandoffStatus(ctx context.Context, orgID int64, handoffID string, status HandoffStatus) error {
	query := `
		UPDATE workforce_handoffs
		SET status = ?
		WHERE org_id = ? AND handoff_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, string(status), orgID, handoffID)
	if err != nil {
		return fmt.Errorf("failed updating handoff status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrHandoffNotFound
	}
	return nil
}

func (r *MySQLRepository) GetShipmentContext(ctx context.Context, orgID int64, shipmentID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var shRow struct {
		ID              int64   `db:"id"`
		OrgID           int64   `db:"org_id"`
		BookingNumber   *string `db:"booking_number"`
		CarrierSCAC     string  `db:"carrier_scac"`
		OriginPort      string  `db:"origin_port"`
		DestinationPort string  `db:"destination_port"`
		Status          *string `db:"status"`
		CurrentRisk     string  `db:"current_risk_level"`
	}

	query := `
		SELECT id, org_id, booking_number, carrier_scac, origin_port, destination_port, status, current_risk_level
		FROM shipments
		WHERE org_id = ? AND (id = ? OR booking_number = ?)
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &shRow, query, orgID, shipmentID, shipmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying shipment for context: %w", err)
	}

	result["id"] = shRow.ID
	result["shipment_id"] = fmt.Sprintf("%d", shRow.ID)
	result["org_id"] = shRow.OrgID
	if shRow.BookingNumber != nil {
		result["booking_number"] = *shRow.BookingNumber
	}
	result["carrier_scac"] = shRow.CarrierSCAC
	result["carrier"] = shRow.CarrierSCAC
	result["origin_port"] = shRow.OriginPort
	result["origin"] = shRow.OriginPort
	result["destination_port"] = shRow.DestinationPort
	result["destination"] = shRow.DestinationPort
	if shRow.Status != nil {
		result["status"] = *shRow.Status
	}
	result["current_risk_level"] = shRow.CurrentRisk

	// Query Milestones
	var milestones []struct {
		MilestoneCode string  `db:"milestone_code"`
		Status        string  `db:"status"`
		PlannedDate   *string `db:"planned_date"`
		ActualDate    *string `db:"actual_date"`
	}
	msQuery := `
		SELECT milestone_code, status, DATE_FORMAT(planned_date, '%Y-%m-%dT%H:%i:%sZ') as planned_date,
		       DATE_FORMAT(actual_date, '%Y-%m-%dT%H:%i:%sZ') as actual_date
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY id ASC
	`
	_ = r.db.SelectContext(ctx, &milestones, msQuery, shRow.ID)
	if len(milestones) > 0 {
		msList := make([]map[string]interface{}, 0, len(milestones))
		for _, m := range milestones {
			item := map[string]interface{}{
				"milestone_code": m.MilestoneCode,
				"status":         m.Status,
			}
			if m.PlannedDate != nil {
				item["planned_date"] = *m.PlannedDate
			}
			if m.ActualDate != nil {
				item["actual_date"] = *m.ActualDate
			}
			msList = append(msList, item)
		}
		result["milestones"] = msList
	}

	return result, nil
}

func (r *MySQLRepository) GetExceptionContext(ctx context.Context, orgID int64, shipmentID, exceptionID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var exRow struct {
		ID            int64   `db:"id"`
		OrgID         int64   `db:"org_id"`
		ShipmentID    int64   `db:"shipment_id"`
		ExceptionType string  `db:"exception_type"`
		Severity      string  `db:"severity"`
		Status        string  `db:"status"`
		Description   *string `db:"description"`
	}

	query := `
		SELECT id, org_id, shipment_id, exception_type, severity, status, description
		FROM shipment_exceptions
		WHERE org_id = ? AND (id = ? OR shipment_id = ?)
		ORDER BY id DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &exRow, query, orgID, exceptionID, shipmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying exception for context: %w", err)
	}

	result["id"] = exRow.ID
	result["exception_id"] = fmt.Sprintf("%d", exRow.ID)
	result["shipment_id"] = fmt.Sprintf("%d", exRow.ShipmentID)
	result["org_id"] = exRow.OrgID
	result["exception_type"] = exRow.ExceptionType
	result["severity"] = exRow.Severity
	result["status"] = exRow.Status
	if exRow.Description != nil {
		result["description"] = *exRow.Description
	}

	return result, nil
}

// Phase 6.6: Commercial Context Extractors

func (r *MySQLRepository) GetCustomerContext(ctx context.Context, orgID int64, customerID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var custRow struct {
		ID           int64   `db:"id"`
		OrgID        int64   `db:"org_id"`
		Name         string  `db:"name"`
		CustomerCode *string `db:"customer_code"`
		Status       string  `db:"status"`
		CreditStatus *string `db:"credit_status"`
	}

	query := `
		SELECT id, org_id, name, customer_code, status, credit_status
		FROM customers
		WHERE org_id = ? AND (id = ? OR customer_code = ?)
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &custRow, query, orgID, customerID, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying customer for context: %w", err)
	}

	result["id"] = custRow.ID
	result["customer_id"] = custRow.ID
	result["org_id"] = custRow.OrgID
	result["name"] = custRow.Name
	result["status"] = custRow.Status
	if custRow.CustomerCode != nil {
		result["customer_code"] = *custRow.CustomerCode
	}
	if custRow.CreditStatus != nil {
		result["credit_status"] = *custRow.CreditStatus
	}

	// Fetch recent overdue or outstanding invoices for this customer
	var invRows []struct {
		ID            int64   `db:"id"`
		InvoiceNumber string  `db:"invoice_number"`
		TotalAmount   float64 `db:"total_amount"`
		BalanceDue    float64 `db:"balance_due"`
		Status        string  `db:"status"`
	}
	invQuery := `
		SELECT id, invoice_number, total_amount, balance_due, status
		FROM customer_invoices
		WHERE org_id = ? AND customer_id = ?
		ORDER BY id DESC LIMIT 5
	`
	_ = r.db.SelectContext(ctx, &invRows, invQuery, orgID, custRow.ID)
	if len(invRows) > 0 {
		invList := make([]map[string]interface{}, 0, len(invRows))
		for _, inv := range invRows {
			invList = append(invList, map[string]interface{}{
				"id":             inv.ID,
				"invoice_number": inv.InvoiceNumber,
				"total_amount":   inv.TotalAmount,
				"balance_due":    inv.BalanceDue,
				"status":         inv.Status,
			})
		}
		result["recent_invoices"] = invList
	}

	return result, nil
}

func (r *MySQLRepository) GetLeadContext(ctx context.Context, orgID int64, leadID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var leadRow struct {
		ID          int64    `db:"id"`
		OrgID       int64    `db:"org_id"`
		CompanyName string   `db:"company_name"`
		ContactName *string  `db:"contact_name"`
		Email       *string  `db:"email"`
		Phone       *string  `db:"phone"`
		Source      *string  `db:"source"`
		Status      string   `db:"status"`
		AIScore     *float64 `db:"ai_score"`
	}

	query := `
		SELECT id, org_id, company_name, contact_name, email, phone, source, status, ai_score
		FROM leads
		WHERE org_id = ? AND id = ?
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &leadRow, query, orgID, leadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying lead for context: %w", err)
	}

	result["id"] = leadRow.ID
	result["lead_id"] = leadRow.ID
	result["org_id"] = leadRow.OrgID
	result["company_name"] = leadRow.CompanyName
	result["status"] = leadRow.Status
	if leadRow.ContactName != nil {
		result["contact_name"] = *leadRow.ContactName
	}
	if leadRow.Email != nil {
		result["email"] = *leadRow.Email
	}
	if leadRow.Phone != nil {
		result["phone"] = *leadRow.Phone
	}
	if leadRow.Source != nil {
		result["source"] = *leadRow.Source
	}
	if leadRow.AIScore != nil {
		result["ai_score"] = *leadRow.AIScore
	}

	return result, nil
}

func (r *MySQLRepository) GetRFQContext(ctx context.Context, orgID int64, rfqID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var rfqRow struct {
		ID          int64   `db:"id"`
		OrgID       int64   `db:"org_id"`
		RFQNumber   string  `db:"rfq_number"`
		CustomerID  *int64  `db:"customer_id"`
		Stage       string  `db:"stage"`
		Status      string  `db:"status"`
		Origin      *string `db:"origin"`
		Destination *string `db:"destination"`
		Incoterms   *string `db:"incoterms"`
	}

	query := `
		SELECT id, org_id, rfq_number, customer_id, stage, status, origin, destination, incoterms
		FROM rfqs
		WHERE org_id = ? AND (id = ? OR rfq_number = ?)
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &rfqRow, query, orgID, rfqID, rfqID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying rfq for context: %w", err)
	}

	result["id"] = rfqRow.ID
	result["rfq_id"] = rfqRow.ID
	result["org_id"] = rfqRow.OrgID
	result["rfq_number"] = rfqRow.RFQNumber
	result["stage"] = rfqRow.Stage
	result["status"] = rfqRow.Status
	if rfqRow.CustomerID != nil {
		result["customer_id"] = *rfqRow.CustomerID
	}
	if rfqRow.Origin != nil {
		result["origin"] = *rfqRow.Origin
	}
	if rfqRow.Destination != nil {
		result["destination"] = *rfqRow.Destination
	}
	if rfqRow.Incoterms != nil {
		result["incoterms"] = *rfqRow.Incoterms
	}

	// Fetch RFQ items / cargo specs if available
	var items []struct {
		Commodity   *string  `db:"commodity"`
		WeightKg    *float64 `db:"weight_kg"`
		VolumeCbm   *float64 `db:"volume_cbm"`
		TargetPrice *float64 `db:"target_price"`
	}
	itemQuery := `
		SELECT commodity, weight_kg, volume_cbm, target_price
		FROM rfq_items
		WHERE rfq_id = ?
		ORDER BY id ASC LIMIT 5
	`
	_ = r.db.SelectContext(ctx, &items, itemQuery, rfqRow.ID)
	if len(items) > 0 {
		itemList := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			itemMap := make(map[string]interface{})
			if it.Commodity != nil {
				itemMap["commodity"] = *it.Commodity
			}
			if it.WeightKg != nil {
				itemMap["weight_kg"] = *it.WeightKg
			}
			if it.VolumeCbm != nil {
				itemMap["volume_cbm"] = *it.VolumeCbm
			}
			if it.TargetPrice != nil {
				itemMap["target_price"] = *it.TargetPrice
				result["target_rate"] = *it.TargetPrice
			}
			itemList = append(itemList, itemMap)
		}
		result["items"] = itemList
	}

	return result, nil
}

func (r *MySQLRepository) GetInvoiceContext(ctx context.Context, orgID int64, customerID, invoiceID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var invRow struct {
		ID             int64   `db:"id"`
		OrgID          int64   `db:"org_id"`
		InvoiceNumber  string  `db:"invoice_number"`
		CustomerID     int64   `db:"customer_id"`
		CustomerName   *string `db:"customer_name"`
		TotalAmount    float64 `db:"total_amount"`
		PaidAmount     float64 `db:"paid_amount"`
		BalanceDue     float64 `db:"balance_due"`
		Currency       string  `db:"currency"`
		Status         string  `db:"status"`
		DueDate        *string `db:"due_date"`
		ShipmentNumber *string `db:"shipment_number"`
	}

	query := `
		SELECT id, org_id, invoice_number, customer_id, customer_name, total_amount, paid_amount, balance_due, currency, status,
		       DATE_FORMAT(due_date, '%Y-%m-%d') as due_date, shipment_number
		FROM customer_invoices
		WHERE org_id = ? AND (id = ? OR invoice_number = ? OR (? != '' AND customer_id = ?))
		ORDER BY id DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &invRow, query, orgID, invoiceID, invoiceID, customerID, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying invoice for context: %w", err)
	}

	result["id"] = invRow.ID
	result["invoice_id"] = invRow.ID
	result["org_id"] = invRow.OrgID
	result["invoice_number"] = invRow.InvoiceNumber
	result["customer_id"] = invRow.CustomerID
	if invRow.CustomerName != nil {
		result["customer_name"] = *invRow.CustomerName
	}
	result["total_amount"] = invRow.TotalAmount
	result["paid_amount"] = invRow.PaidAmount
	result["balance_due"] = invRow.BalanceDue
	result["currency"] = invRow.Currency
	result["status"] = invRow.Status
	if invRow.DueDate != nil {
		result["due_date"] = *invRow.DueDate
	}
	if invRow.ShipmentNumber != nil {
		result["shipment_number"] = *invRow.ShipmentNumber
	}

	return result, nil
}

func (r *MySQLRepository) GetContractContext(ctx context.Context, orgID int64, contractID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var ctrRow struct {
		ID                int64    `db:"id"`
		OrgID             int64    `db:"org_id"`
		ContractReference string   `db:"contract_reference"`
		ContractName      string   `db:"contract_name"`
		ContractType      string   `db:"contract_type"`
		PartyID           *int64   `db:"party_id"`
		PartyName         string   `db:"party_name"`
		TransportMode     string   `db:"transport_mode"`
		Status            string   `db:"status"`
		Currency          string   `db:"currency"`
		ContractValue     *float64 `db:"contract_value"`
		EffectiveDate     *string  `db:"effective_date"`
		ExpiryDate        *string  `db:"expiry_date"`
		Description       *string  `db:"description"`
		Notes             *string  `db:"notes"`
	}

	query := `
		SELECT id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
		       transport_mode, status, currency, contract_value,
		       DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
		       DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date,
		       description, notes
		FROM contracts
		WHERE org_id = ? AND (id = ? OR contract_reference = ?)
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &ctrRow, query, orgID, contractID, contractID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("failed querying contract for context: %w", err)
	}

	result["id"] = ctrRow.ID
	result["contract_id"] = ctrRow.ID
	result["org_id"] = ctrRow.OrgID
	result["contract_reference"] = ctrRow.ContractReference
	result["contract_name"] = ctrRow.ContractName
	result["contract_type"] = ctrRow.ContractType
	result["party_name"] = ctrRow.PartyName
	result["transport_mode"] = ctrRow.TransportMode
	result["status"] = ctrRow.Status
	result["currency"] = ctrRow.Currency
	if ctrRow.ContractValue != nil {
		result["contract_value"] = *ctrRow.ContractValue
	}
	if ctrRow.EffectiveDate != nil {
		result["effective_date"] = *ctrRow.EffectiveDate
	}
	if ctrRow.ExpiryDate != nil {
		result["expiry_date"] = *ctrRow.ExpiryDate
	}
	if ctrRow.Description != nil {
		result["description"] = *ctrRow.Description
	}
	if ctrRow.Notes != nil {
		result["notes"] = *ctrRow.Notes
	}

	result["destination_free_days"] = 14
	result["detention_rate_per_day"] = 120.0

	return result, nil
}

func (r *MySQLRepository) GetComplianceContext(ctx context.Context, orgID int64, shipmentID string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["shipment_id"] = shipmentID
	result["org_id"] = orgID

	type docRow struct {
		ID           int64   `db:"id"`
		DocType      string  `db:"doc_type"`
		DocumentName string  `db:"document_name"`
		Status       string  `db:"status"`
	}
	var docs []docRow
	docQuery := `
		SELECT id, doc_type, document_name, status
		FROM shipment_documents
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY id DESC
	`
	_ = r.db.SelectContext(ctx, &docs, docQuery, orgID, shipmentID)
	docList := make([]map[string]interface{}, 0, len(docs))
	hasMissingDocs := len(docs) == 0
	for _, d := range docs {
		docList = append(docList, map[string]interface{}{
			"id":            d.ID,
			"doc_type":      d.DocType,
			"document_name": d.DocumentName,
			"status":        d.Status,
		})
		if d.Status == "PENDING_REVIEW" || d.Status == "DISCREPANCY" {
			hasMissingDocs = true
		}
	}
	result["documents"] = docList
	result["missing_documents"] = hasMissingDocs

	type discRow struct {
		ID            int64   `db:"id"`
		FieldName     string  `db:"field_name"`
		ExpectedValue *string `db:"expected_value"`
		ActualValue   *string `db:"actual_value"`
		Status        string  `db:"status"`
	}
	var discrepancies []discRow
	discQuery := `
		SELECT id, field_name, expected_value, actual_value, status
		FROM shipment_document_discrepancies
		WHERE org_id = ? AND shipment_id = ? AND status = 'OPEN'
	`
	_ = r.db.SelectContext(ctx, &discrepancies, discQuery, orgID, shipmentID)
	discList := make([]string, 0, len(discrepancies))
	for _, dc := range discrepancies {
		exp := ""
		act := ""
		if dc.ExpectedValue != nil {
			exp = *dc.ExpectedValue
		}
		if dc.ActualValue != nil {
			act = *dc.ActualValue
		}
		discList = append(discList, fmt.Sprintf("%s discrepancy: expected '%s', actual '%s'", dc.FieldName, exp, act))
	}
	result["discrepancies"] = discList

	var count int
	_ = r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM shipment_exceptions
		WHERE org_id = ? AND shipment_id = ? AND exception_type = 'CUSTOMS_HOLD' AND status != 'RESOLVED'
	`, orgID, shipmentID)
	result["customs_hold"] = count > 0

	return result, nil
}

// ----------------------------------------------------------------------
// Phase 6.8: Memory & Learning from Outcomes
// ----------------------------------------------------------------------

func (r *MySQLRepository) GetMemoryContext(ctx context.Context, orgID int64, domain, entityType, entityID string, limit int) ([]MemoryItemDTO, error) {
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	whereSQL := "org_id = ? AND is_stale = 0 AND status = 'ACTIVE'"
	args := []interface{}{orgID}

	if domain != "" {
		whereSQL += " AND (category = ? OR memory_type = ? OR title LIKE ?)"
		args = append(args, domain, domain, "%"+domain+"%")
	}
	if entityType != "" {
		whereSQL += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if entityID != "" {
		whereSQL += " AND entity_id = ?"
		args = append(args, entityID)
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, memory_type, category, title, content, confidence, status,
		       COALESCE(entity_type, '') as entity_type, COALESCE(entity_id, '') as entity_id,
		       times_used, success_count, failure_count, is_stale, created_at
		FROM ai_memory_items
		WHERE %s
		ORDER BY recency_weight DESC, id DESC
		LIMIT %d
	`, whereSQL, limit)

	type memRow struct {
		ID           int64     `db:"id"`
		OrgID        int64     `db:"org_id"`
		MemoryType   string    `db:"memory_type"`
		Category     string    `db:"category"`
		Title        string    `db:"title"`
		Content      string    `db:"content"`
		Confidence   float64   `db:"confidence"`
		Status       string    `db:"status"`
		EntityType   string    `db:"entity_type"`
		EntityID     string    `db:"entity_id"`
		TimesUsed    int       `db:"times_used"`
		SuccessCount int       `db:"success_count"`
		FailureCount int       `db:"failure_count"`
		IsStale      bool      `db:"is_stale"`
		CreatedAt    time.Time `db:"created_at"`
	}

	var rows []memRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query memory items: %w", err)
	}

	items := make([]MemoryItemDTO, 0, len(rows))
	for _, rw := range rows {
		items = append(items, MemoryItemDTO{
			ID:           rw.ID,
			OrgID:        rw.OrgID,
			MemoryType:   rw.MemoryType,
			Category:     rw.Category,
			Title:        rw.Title,
			Content:      rw.Content,
			Confidence:   rw.Confidence,
			Status:       rw.Status,
			EntityType:   rw.EntityType,
			EntityID:     rw.EntityID,
			TimesUsed:    rw.TimesUsed,
			SuccessCount: rw.SuccessCount,
			FailureCount: rw.FailureCount,
			IsStale:      rw.IsStale,
			CreatedAt:    rw.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}

func (r *MySQLRepository) RecordOutcome(ctx context.Context, outcome *AgentOutcome) error {
	if outcome.OutcomeID == "" {
		outcome.OutcomeID = fmt.Sprintf("out_%d_%d", outcome.TenantID, time.Now().UnixNano())
	}
	if outcome.CorrelationID == "" {
		outcome.CorrelationID = fmt.Sprintf("corr_%d_%d", outcome.TenantID, time.Now().UnixNano())
	}

	var metaJSON sql.NullString
	if outcome.Metadata != nil {
		if b, err := json.Marshal(outcome.Metadata); err == nil {
			metaJSON = sql.NullString{String: string(b), Valid: true}
		}
	}

	outcomeType := "RECOMMENDATION_OUTCOME"
	if outcome.ActionType != "" {
		outcomeType = "PLAN_EXECUTION_OUTCOME"
	}

	var recJSON string
	if outcome.Recommendation != nil {
		if b, err := json.Marshal(outcome.Recommendation); err == nil {
			recJSON = string(b)
		}
	}

	humanInvolvement := "NONE"
	if outcome.HumanDecision != "" {
		humanInvolvement = outcome.HumanDecision
	}

	query := `
		INSERT INTO ai_agent_outcomes (
			org_id, outcome_id, source_entity_type, source_entity_id,
			workflow_id, plan_id, action_type, outcome_type,
			expected_result, actual_result, status, is_verified,
			human_involvement, reason, confidence_score, metadata, correlation_id
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		outcome.TenantID, outcome.OutcomeID, outcome.SourceEntityType, outcome.SourceEntityID,
		outcome.WorkflowID, outcome.PlanID, outcome.ActionType, outcomeType,
		recJSON, outcome.ActualOutcome, outcome.Status, outcome.IsVerified,
		humanInvolvement, outcome.Lesson, outcome.Confidence, metaJSON, outcome.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert agent outcome: %w", err)
	}
	return nil
}

func (r *MySQLRepository) ListOutcomes(ctx context.Context, orgID int64, entityType, entityID string, limit int) ([]AgentOutcome, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	whereSQL := "org_id = ?"
	args := []interface{}{orgID}
	if entityType != "" {
		whereSQL += " AND source_entity_type = ?"
		args = append(args, entityType)
	}
	if entityID != "" {
		whereSQL += " AND source_entity_id = ?"
		args = append(args, entityID)
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, outcome_id, source_entity_type, source_entity_id,
		       COALESCE(workflow_id, '') as workflow_id, COALESCE(plan_id, '') as plan_id,
		       COALESCE(action_type, '') as action_type, COALESCE(expected_result, '') as expected_result,
		       COALESCE(actual_result, '') as actual_result, status, is_verified,
		       COALESCE(human_involvement, 'NONE') as human_involvement,
		       COALESCE(reason, '') as reason, confidence_score, COALESCE(correlation_id, '') as correlation_id,
		       created_at
		FROM ai_agent_outcomes
		WHERE %s
		ORDER BY id DESC
		LIMIT %d
	`, whereSQL, limit)

	type outRow struct {
		ID               int64     `db:"id"`
		OrgID            int64     `db:"org_id"`
		OutcomeID        string    `db:"outcome_id"`
		SourceEntityType string    `db:"source_entity_type"`
		SourceEntityID   string    `db:"source_entity_id"`
		WorkflowID       string    `db:"workflow_id"`
		PlanID           string    `db:"plan_id"`
		ActionType       string    `db:"action_type"`
		ExpectedResult   string    `db:"expected_result"`
		ActualResult     string    `db:"actual_result"`
		Status           string    `db:"status"`
		IsVerified       bool      `db:"is_verified"`
		HumanInvolvement string    `db:"human_involvement"`
		Reason           string    `db:"reason"`
		ConfidenceScore  float64   `db:"confidence_score"`
		CorrelationID    string    `db:"correlation_id"`
		CreatedAt        time.Time `db:"created_at"`
	}

	var rows []outRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query agent outcomes: %w", err)
	}

	outcomes := make([]AgentOutcome, 0, len(rows))
	for _, rw := range rows {
		outcomes = append(outcomes, AgentOutcome{
			ID:               rw.ID,
			TenantID:         rw.OrgID,
			OutcomeID:        rw.OutcomeID,
			SourceEntityType: rw.SourceEntityType,
			SourceEntityID:   rw.SourceEntityID,
			WorkflowID:       rw.WorkflowID,
			PlanID:           rw.PlanID,
			ActionType:       rw.ActionType,
			ActualOutcome:    rw.ActualResult,
			Status:           rw.Status,
			IsVerified:       rw.IsVerified,
			SuccessIndicator: rw.Status == "SUCCESS" || rw.Status == "PARTIAL_SUCCESS",
			HumanDecision:    rw.HumanInvolvement,
			Lesson:           rw.Reason,
			Confidence:       rw.ConfidenceScore,
			CorrelationID:    rw.CorrelationID,
			CreatedAt:        rw.CreatedAt.Format(time.RFC3339),
		})
	}
	return outcomes, nil
}

func (r *MySQLRepository) GetOutcomesSummary(ctx context.Context, orgID int64) (map[string]interface{}, error) {
	summary := map[string]interface{}{
		"org_id": orgID,
	}
	var total, verified, success, failed int
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM ai_agent_outcomes WHERE org_id = ?`, orgID)
	_ = r.db.GetContext(ctx, &verified, `SELECT COUNT(*) FROM ai_agent_outcomes WHERE org_id = ? AND is_verified = 1`, orgID)
	_ = r.db.GetContext(ctx, &success, `SELECT COUNT(*) FROM ai_agent_outcomes WHERE org_id = ? AND status IN ('SUCCESS', 'PARTIAL_SUCCESS')`, orgID)
	_ = r.db.GetContext(ctx, &failed, `SELECT COUNT(*) FROM ai_agent_outcomes WHERE org_id = ? AND status = 'FAILED'`, orgID)

	summary["total_outcomes"] = total
	summary["verified_outcomes"] = verified
	summary["successful_outcomes"] = success
	summary["failed_outcomes"] = failed
	return summary, nil
}

// ----------------------------------------------------------------------
// Phase 6.9: Governed Autonomy & Command Center Repository Implementation
// ----------------------------------------------------------------------

func (r *MySQLRepository) GetWorkforceHealth(ctx context.Context, orgID int64) (*WorkforceHealthSummary, error) {
	agents, err := r.ListAgents(ctx, orgID, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching agents for workforce health: %w", err)
	}

	activeCount := 0
	pausedCount := 0
	disabledCount := 0

	for _, a := range agents {
		if !a.IsEnabled || a.HealthStatus == HealthStatusDisabled {
			disabledCount++
		} else if a.HealthStatus == AgentStatusPaused {
			pausedCount++
		} else {
			activeCount++
		}
	}

	type statusCount struct {
		Status string `db:"status"`
		Count  int    `db:"cnt"`
	}
	var counts []statusCount
	q := `SELECT status, COUNT(*) as cnt FROM workforce_tasks WHERE org_id = ? GROUP BY status`
	_ = r.db.SelectContext(ctx, &counts, q, orgID)

	summary := &WorkforceHealthSummary{
		ActiveAgentsCount:   activeCount,
		PausedAgentsCount:   pausedCount,
		DisabledAgentsCount: disabledCount,
	}

	for _, sc := range counts {
		summary.TotalTasks += sc.Count
		switch TaskStatus(sc.Status) {
		case TaskStatusPending, TaskStatusAssigned:
			summary.PendingTasks += sc.Count
		case TaskStatusRunning:
			summary.RunningTasks += sc.Count
		case TaskStatusWaiting:
			summary.WaitingTasks += sc.Count
			summary.ApprovalBacklog += sc.Count
		case TaskStatusBlocked:
			summary.BlockedTasks += sc.Count
		case TaskStatusFailed, TaskStatusCancelled:
			summary.FailedTasks += sc.Count
		case TaskStatusEscalated:
			summary.EscalationBacklog += sc.Count
		case TaskStatusCompleted:
			summary.CompletedTasks += sc.Count
		}
	}

	if summary.FailedTasks > 5 || summary.EscalationBacklog > 3 {
		summary.OverallStatus = HealthStatusDegraded
	} else if summary.ActiveAgentsCount == 0 || summary.FailedTasks > 15 {
		summary.OverallStatus = "CRITICAL"
	} else {
		summary.OverallStatus = HealthStatusHealthy
	}

	return summary, nil
}

func (r *MySQLRepository) GetAgentWorkload(ctx context.Context, orgID int64) ([]AgentWorkloadMetrics, error) {
	agents, err := r.ListAgents(ctx, orgID, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching agents for workload: %w", err)
	}

	type agentStatusCount struct {
		AgentID string `db:"assigned_agent_id"`
		Status  string `db:"status"`
		Count   int    `db:"cnt"`
	}
	var counts []agentStatusCount
	q := `SELECT assigned_agent_id, status, COUNT(*) as cnt FROM workforce_tasks WHERE org_id = ? GROUP BY assigned_agent_id, status`
	_ = r.db.SelectContext(ctx, &counts, q, orgID)

	agentTaskMap := make(map[string]map[string]int)
	for _, c := range counts {
		if _, ok := agentTaskMap[c.AgentID]; !ok {
			agentTaskMap[c.AgentID] = make(map[string]int)
		}
		agentTaskMap[c.AgentID][c.Status] = c.Count
	}

	type activeTaskRow struct {
		AgentID string `db:"assigned_agent_id"`
		TaskID  string `db:"task_id"`
	}
	var activeTasks []activeTaskRow
	qActive := `SELECT assigned_agent_id, task_id FROM workforce_tasks WHERE org_id = ? AND status = 'RUNNING' ORDER BY id DESC`
	_ = r.db.SelectContext(ctx, &activeTasks, qActive, orgID)
	activeTaskMap := make(map[string]string)
	for _, at := range activeTasks {
		if _, exists := activeTaskMap[at.AgentID]; !exists {
			activeTaskMap[at.AgentID] = at.TaskID
		}
	}

	res := make([]AgentWorkloadMetrics, 0, len(agents))
	for _, a := range agents {
		tStats := agentTaskMap[a.AgentID]
		pending := tStats[string(TaskStatusPending)] + tStats[string(TaskStatusAssigned)]
		running := tStats[string(TaskStatusRunning)]
		waiting := tStats[string(TaskStatusWaiting)]
		completed := tStats[string(TaskStatusCompleted)]
		failed := tStats[string(TaskStatusFailed)] + tStats[string(TaskStatusCancelled)]

		opStatus := AgentStatusActive
		if !a.IsEnabled || a.HealthStatus == HealthStatusDisabled {
			opStatus = AgentStatusDisabled
		} else if a.HealthStatus == AgentStatusPaused {
			opStatus = AgentStatusPaused
		}

		m := AgentWorkloadMetrics{
			AgentID:           a.AgentID,
			Name:              a.Name,
			AgentType:         string(a.AgentType),
			AutonomyLevel:     a.AutonomyLevel,
			OperationalStatus: opStatus,
			HealthStatus:      a.HealthStatus,
			PendingTasks:      pending,
			RunningTasks:      running,
			WaitingTasks:      waiting,
			CompletedTasks:    completed,
			FailedTasks:       failed,
			FailureCount:      failed,
			CurrentTaskID:     activeTaskMap[a.AgentID],
			LastExecutionAt:   a.LastExecutionAt,
			IsBottleneck:      waiting >= 3,
			IsOverloaded:      (pending + running) >= 5,
		}
		res = append(res, m)
	}

	return res, nil
}

func (r *MySQLRepository) UpdateAgentControl(ctx context.Context, orgID int64, agentID string, isEnabled bool, healthStatus string, autonomyLevel string) error {
	q := `
		UPDATE workforce_agents
		SET is_enabled = ?, health_status = ?, autonomy_level = ?, updated_at = CURRENT_TIMESTAMP
		WHERE (org_id = ? OR org_id = 0) AND agent_id = ?
	`
	res, err := r.db.ExecContext(ctx, q, isEnabled, healthStatus, autonomyLevel, orgID, agentID)
	if err != nil {
		return fmt.Errorf("failed updating agent control: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrAgentNotFound
	}
	return nil
}





package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultSidecarURL     = "http://127.0.0.1:8090"
	DefaultServiceKey     = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	DefaultSidecarTimeout = 20 * time.Second
)

type Service interface {
	Chat(ctx context.Context, orgID, userID int64, userName, userRole string, input ChatInput) (*SidecarChatResponse, error)
	ListSessions(ctx context.Context, orgID, userID int64, limit int) ([]CopilotSession, error)
	GetSession(ctx context.Context, orgID int64, sessionID string) (*CopilotSession, error)
	ListMessages(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotMessage, error)
	ArchiveSession(ctx context.Context, orgID int64, sessionID string) error
	ExecuteAction(ctx context.Context, orgID, userID int64, userName string, input ExecuteActionInput) (*CopilotActionHistory, error)
	ListActions(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotActionHistory, error)
}

type serviceImpl struct {
	repo       Repository
	sidecarURL string
	serviceKey string
	httpClient *http.Client
}

type ServiceOption func(*serviceImpl)

func WithSidecar(url, key string) ServiceOption {
	return func(s *serviceImpl) {
		if url != "" {
			s.sidecarURL = url
		}
		if key != "" {
			s.serviceKey = key
		}
	}
}

func NewService(repo Repository, opts ...ServiceOption) Service {
	sidecar := os.Getenv("AI_SIDECAR_URL")
	if sidecar == "" {
		sidecar = DefaultSidecarURL
	}
	key := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if key == "" {
		key = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if key == "" {
		key = os.Getenv("AI_SIDECAR_SERVICE_KEY")
	}
	if key == "" {
		rawEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		}
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
		}
		if rawEnv == "development" || rawEnv == "test" {
			key = DefaultServiceKey
		}
	}

	s := &serviceImpl{
		repo:       repo,
		sidecarURL: sidecar,
		serviceKey: key,
		httpClient: &http.Client{
			Timeout: DefaultSidecarTimeout,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *serviceImpl) Chat(ctx context.Context, orgID, userID int64, userName, userRole string, input ChatInput) (*SidecarChatResponse, error) {
	if strings.TrimSpace(input.Query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	corrID := fmt.Sprintf("copilot-%s", uuid.New().String()[:12])

	// 1. Session resolution
	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("ses-%s", uuid.New().String()[:16])
		title := input.Query
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		newSession := &CopilotSession{
			OrgID:           orgID,
			UserID:          userID,
			SessionID:       sessionID,
			Title:           title,
			CurrentModule:   input.CurrentModule,
			CurrentRoute:    input.CurrentRoute,
			CurrentRecordID: input.CurrentRecordID,
			IsArchived:      false,
		}
		if err := s.repo.CreateSession(ctx, newSession); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	} else {
		// Update session context
		_ = s.repo.UpdateSessionContext(ctx, orgID, sessionID, input.CurrentModule, input.CurrentRoute, input.CurrentRecordID)
	}

	// 2. Query authorized context from MariaDB strictly scoped by orgID
	records, metrics, err := s.repo.RetrieveModuleContext(ctx, orgID, input.CurrentModule, input.CurrentRecordID)
	if err != nil {
		records = []map[string]interface{}{}
		metrics = map[string]interface{}{}
	}

	// 3. Load recent conversation history (last 6 messages)
	prevMsgs, _ := s.repo.ListMessages(ctx, orgID, sessionID, 6)
	history := make([]SidecarChatMessage, 0, len(prevMsgs))
	for _, m := range prevMsgs {
		history = append(history, SidecarChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// 4. Save the user message in DB
	userMsg := &CopilotMessage{
		SessionID:     sessionID,
		OrgID:         orgID,
		UserID:        userID,
		Role:          "user",
		Content:       input.Query,
		CorrelationID: corrID,
	}
	_ = s.repo.SaveMessage(ctx, userMsg)

	// 5. Construct Python sidecar request payload
	sidecarReq := SidecarChatRequest{
		Context: SidecarContextPayload{
			OrgID:             orgID,
			UserID:            userID,
			UserRole:          userRole,
			CurrentRoute:      input.CurrentRoute,
			CurrentModule:     input.CurrentModule,
			CurrentRecordID:   input.CurrentRecordID,
			ActiveFilters:     input.ActiveFilters,
			AuthorizedRecords: records,
			SummaryMetrics:    metrics,
		},
		Query:               input.Query,
		ConversationHistory: history,
		CorrelationID:       corrID,
	}

	bodyBytes, err := json.Marshal(sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize sidecar request: %w", err)
	}

	// 6. Call Python AI Sidecar (POST /copilot/chat)
	reqURL := fmt.Sprintf("%s/copilot/chat", s.sidecarURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI sidecar copilot service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI sidecar returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp SidecarChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI copilot response: %w", err)
	}

	// 7. Persist AI assistant message with verified facts, sources, and action proposals
	factsJSON := ToJSONString(chatResp.ConfirmedFacts)
	sourcesJSON := ToJSONString(chatResp.SourceReferences)
	proposalsJSON := ToJSONString(chatResp.ActionProposals)
	confScore := chatResp.Confidence

	aiMsg := &CopilotMessage{
		SessionID:        sessionID,
		OrgID:            orgID,
		UserID:           userID,
		Role:             "assistant",
		Content:          chatResp.Answer,
		ConfirmedFacts:   factsJSON,
		SourceReferences: sourcesJSON,
		ActionProposals:  proposalsJSON,
		DraftContent:     chatResp.DraftContent,
		DraftType:        chatResp.DraftType,
		ConfidenceScore:  &confScore,
		CorrelationID:    corrID,
	}
	_ = s.repo.SaveMessage(ctx, aiMsg)

	// Ensure correlationID is set in response
	if chatResp.CorrelationID == "" {
		chatResp.CorrelationID = corrID
	}

	return &chatResp, nil
}

func (s *serviceImpl) ListSessions(ctx context.Context, orgID, userID int64, limit int) ([]CopilotSession, error) {
	return s.repo.ListSessions(ctx, orgID, userID, limit)
}

func (s *serviceImpl) GetSession(ctx context.Context, orgID int64, sessionID string) (*CopilotSession, error) {
	return s.repo.GetSession(ctx, orgID, sessionID)
}

func (s *serviceImpl) ListMessages(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotMessage, error) {
	return s.repo.ListMessages(ctx, orgID, sessionID, limit)
}

func (s *serviceImpl) ArchiveSession(ctx context.Context, orgID int64, sessionID string) error {
	return s.repo.ArchiveSession(ctx, orgID, sessionID)
}

func (s *serviceImpl) ExecuteAction(ctx context.Context, orgID, userID int64, userName string, input ExecuteActionInput) (*CopilotActionHistory, error) {
	if strings.TrimSpace(input.ActionType) == "" {
		return nil, fmt.Errorf("action_type cannot be empty")
	}

	corrID := fmt.Sprintf("act-%s", uuid.New().String()[:12])
	payloadJSON, _ := json.Marshal(input.ActionPayload)
	payloadStr := string(payloadJSON)

	actionTitle := input.ActionTitle
	if actionTitle == "" {
		actionTitle = fmt.Sprintf("Execute %s", input.ActionType)
	}

	var approvalID *int64
	var recID *int64
	status := "EXECUTED"
	resultSummary := fmt.Sprintf("Action '%s' completed successfully.", actionTitle)

	// Determine if consequential action requires Human Approval Gate (HITL)
	actionUpper := strings.ToUpper(input.ActionType)
	switch {
	case actionUpper == "REQUEST_HUMAN_APPROVAL" ||
		actionUpper == "REQUEST_ESCALATION" ||
		actionUpper == "REQUEST_COMPLIANCE_REVIEW" ||
		actionUpper == "REQUEST_DOCUMENT_REVIEW":
		// Consequential operational actions must be gated by approval_requests
		category := "OPERATIONS"
		if strings.Contains(actionUpper, "COMPLIANCE") || strings.Contains(actionUpper, "DOCUMENT") {
			category = "DOCUMENTS"
		}
		createdID, err := s.repo.CreateApprovalRequest(ctx, orgID, userID, userName, actionTitle, category, actionUpper, input.ActionPayload, corrID)
		if err != nil {
			return nil, fmt.Errorf("failed to create approval request for action: %w", err)
		}
		approvalID = &createdID
		status = "PENDING_APPROVAL"
		resultSummary = fmt.Sprintf("Submitted to Human Approval Center (Request #%d). Execution pending approval.", createdID)

	case actionUpper == "CREATE_RECOMMENDATION":
		category := "operations"
		if cat, ok := input.ActionPayload["category"].(string); ok && cat != "" {
			category = cat
		}
		desc := actionTitle
		if d, ok := input.ActionPayload["description"].(string); ok && d != "" {
			desc = d
		}
		createdRecID, err := s.repo.CreateRecommendation(ctx, orgID, actionTitle, category, desc, "COPILOT", input.ActionPayload, corrID)
		if err != nil {
			return nil, fmt.Errorf("failed to create recommendation: %w", err)
		}
		recID = &createdRecID
		status = "EXECUTED"
		resultSummary = fmt.Sprintf("Saved as active recommendation #%d in Recommendation Center.", createdRecID)

	case actionUpper == "CREATE_INTERNAL_TASK":
		status = "EXECUTED"
		resultSummary = fmt.Sprintf("Internal task created and logged: %s", actionTitle)

	case actionUpper == "NAVIGATE_TO_RECORD":
		status = "EXECUTED"
		resultSummary = fmt.Sprintf("Navigated to destination record.")

	default:
		// Safe draft generation or view action
		status = "EXECUTED"
		resultSummary = fmt.Sprintf("Processed %s successfully.", actionTitle)
	}

	actionRecord := &CopilotActionHistory{
		OrgID:            orgID,
		UserID:           userID,
		SessionID:        input.SessionID,
		ActionType:       input.ActionType,
		ActionTitle:      actionTitle,
		ActionPayload:    payloadStr,
		ApprovalID:       approvalID,
		RecommendationID: recID,
		Status:           status,
		ResultSummary:    &resultSummary,
		CorrelationID:    corrID,
	}

	if err := s.repo.LogAction(ctx, actionRecord); err != nil {
		return nil, fmt.Errorf("failed to log action history: %w", err)
	}

	return actionRecord, nil
}

func (s *serviceImpl) ListActions(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotActionHistory, error) {
	return s.repo.ListActions(ctx, orgID, sessionID, limit)
}

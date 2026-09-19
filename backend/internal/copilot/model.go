package copilot

import (
	"encoding/json"
	"time"
)

// CopilotSession represents an ongoing conversational session within an organization.
type CopilotSession struct {
	ID              int64     `db:"id" json:"id"`
	OrgID           int64     `db:"org_id" json:"org_id"`
	UserID          int64     `db:"user_id" json:"user_id"`
	SessionID       string    `db:"session_id" json:"session_id"`
	Title           string    `db:"title" json:"title"`
	CurrentModule   string    `db:"current_module" json:"current_module"`
	CurrentRoute    string    `db:"current_route" json:"current_route"`
	CurrentRecordID *string   `db:"current_record_id" json:"current_record_id,omitempty"`
	IsArchived      bool      `db:"is_archived" json:"is_archived"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// CopilotMessage represents a single conversational turn.
type CopilotMessage struct {
	ID               int64     `db:"id" json:"id"`
	SessionID        string    `db:"session_id" json:"session_id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	UserID           int64     `db:"user_id" json:"user_id"`
	Role             string    `db:"role" json:"role"` // user, assistant, system
	Content          string    `db:"content" json:"content"`
	ConfirmedFacts   *string   `db:"confirmed_facts" json:"confirmed_facts,omitempty"`
	SourceReferences *string   `db:"source_references" json:"source_references,omitempty"`
	ActionProposals  *string   `db:"action_proposals" json:"action_proposals,omitempty"`
	DraftContent     *string   `db:"draft_content" json:"draft_content,omitempty"`
	DraftType        *string   `db:"draft_type" json:"draft_type,omitempty"`
	ConfidenceScore  *float64  `db:"confidence_score" json:"confidence_score,omitempty"`
	CorrelationID    string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// CopilotActionHistory tracks proposals and executed/approved actions.
type CopilotActionHistory struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	UserID           int64     `db:"user_id" json:"user_id"`
	SessionID        string    `db:"session_id" json:"session_id"`
	ActionType       string    `db:"action_type" json:"action_type"`
	ActionTitle      string    `db:"action_title" json:"action_title"`
	ActionPayload    string    `db:"action_payload" json:"action_payload"`
	ApprovalID       *int64    `db:"approval_id" json:"approval_id,omitempty"`
	RecommendationID *int64    `db:"recommendation_id" json:"recommendation_id,omitempty"`
	Status           string    `db:"status" json:"status"` // PROPOSED, PENDING_APPROVAL, EXECUTED, REJECTED
	ResultSummary    *string   `db:"result_summary" json:"result_summary,omitempty"`
	CorrelationID    string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// DTOs for Sidecar Communication

type SidecarSourceRef struct {
	RecordType string  `json:"record_type"`
	RecordID   string  `json:"record_id"`
	Title      string  `json:"title"`
	URL        *string `json:"url,omitempty"`
	Snippet    *string `json:"snippet,omitempty"`
}

type SidecarActionProposal struct {
	ActionType       string                 `json:"action_type"`
	ActionTitle      string                 `json:"action_title"`
	Description      string                 `json:"description"`
	Payload          map[string]interface{} `json:"payload"`
	RequiresApproval bool                   `json:"requires_approval"`
}

type SidecarContextPayload struct {
	OrgID             int64                    `json:"org_id"`
	UserID            int64                    `json:"user_id"`
	UserRole          string                   `json:"user_role"`
	CurrentRoute      string                   `json:"current_route"`
	CurrentModule     string                   `json:"current_module"`
	CurrentRecordID   *string                  `json:"current_record_id,omitempty"`
	ActiveFilters     map[string]interface{}   `json:"active_filters,omitempty"`
	AuthorizedRecords []map[string]interface{} `json:"authorized_records"`
	SummaryMetrics    map[string]interface{}   `json:"summary_metrics,omitempty"`
}

type SidecarChatMessage struct {
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Timestamp *string `json:"timestamp,omitempty"`
}

type SidecarChatRequest struct {
	Context             SidecarContextPayload `json:"context"`
	Query               string                `json:"query"`
	ConversationHistory []SidecarChatMessage  `json:"conversation_history"`
	CorrelationID       string                `json:"correlation_id"`
}

type SidecarChatResponse struct {
	Answer              string                  `json:"answer"`
	ConfirmedFacts      []string                `json:"confirmed_facts"`
	SourceReferences    []SidecarSourceRef      `json:"source_references"`
	Signals             []string                `json:"signals"`
	AIInterpretation    string                  `json:"ai_interpretation"`
	Recommendations     []string                `json:"recommendations"`
	SuggestedFollowups  []string                `json:"suggested_followups"`
	DraftContent        *string                 `json:"draft_content,omitempty"`
	DraftType           *string                 `json:"draft_type,omitempty"`
	ActionProposals     []SidecarActionProposal `json:"action_proposals"`
	Confidence          float64                 `json:"confidence"`
	MissingInformation []string                `json:"missing_information"`
	RequiresApproval    bool                    `json:"requires_approval"`
	SafetyRestrictions  []string                `json:"safety_restrictions"`
	CorrelationID       string                  `json:"correlation_id"`
}

// User Request DTOs
type ChatInput struct {
	SessionID       string                 `json:"session_id,omitempty"`
	CurrentModule   string                 `json:"current_module"`
	CurrentRoute    string                 `json:"current_route"`
	CurrentRecordID *string                `json:"current_record_id,omitempty"`
	Query           string                 `json:"query"`
	ActiveFilters   map[string]interface{} `json:"active_filters,omitempty"`
}

type ExecuteActionInput struct {
	ActionType    string                 `json:"action_type"`
	ActionTitle   string                 `json:"action_title"`
	ActionPayload map[string]interface{} `json:"action_payload"`
	SessionID     string                 `json:"session_id"`
	Reason        string                 `json:"reason,omitempty"`
}

func ToJSONString(v interface{}) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

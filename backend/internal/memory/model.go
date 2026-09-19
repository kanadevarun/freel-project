package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Memory Scopes
const (
	ScopeUser         = "USER"
	ScopeOrganization = "ORGANIZATION"
)

// Memory Statuses
const (
	MemoryStatusActive        = "ACTIVE"
	MemoryStatusDisabled      = "DISABLED"
	MemoryStatusExpired       = "EXPIRED"
	MemoryStatusDeleted       = "DELETED"
	MemoryStatusPendingReview = "PENDING_REVIEW"
)

// Memory Types
const (
	MemoryTypeResponseStyle     = "RESPONSE_STYLE"
	MemoryTypeSummaryPreference = "SUMMARY_PREFERENCE"
	MemoryTypeTerminology       = "TERMINOLOGY"
	MemoryTypeBusinessFact      = "BUSINESS_FACT"
	MemoryTypeModulePreference  = "MODULE_PREFERENCE"
	MemoryTypeExplanationDepth  = "EXPLANATION_DEPTH"
	MemoryTypeWorkingFocus      = "WORKING_FOCUS"
)

// Value Types for Preferences
const (
	ValueTypeString  = "STRING"
	ValueTypeBoolean = "BOOLEAN"
	ValueTypeNumber  = "NUMBER"
	ValueTypeJSON    = "JSON"
)

// Source Types
const (
	SourceTypeExplicitUser      = "EXPLICIT_USER"
	SourceTypeAssistantProposed = "ASSISTANT_PROPOSED"
	SourceTypeOrgPolicy         = "ORGANIZATION_POLICY"
	SourceTypeUserSetting       = "USER_SETTING"
)

// Memory Audit Events
const (
	AuditEventProposed          = "PROPOSED"
	AuditEventConfirmed         = "CONFIRMED"
	AuditEventCreated           = "CREATED"
	AuditEventUpdated           = "UPDATED"
	AuditEventDisabled          = "DISABLED"
	AuditEventReenabled         = "REENABLED"
	AuditEventDeleted           = "DELETED"
	AuditEventCleared           = "CLEARED"
	AuditEventRejectedSensitive = "REJECTED_SENSITIVE"
	AuditEventBlockedAuth       = "BLOCKED_AUTH"
	AuditEventUsedInRuntime     = "USED_IN_RUNTIME"
)

// MemoryItem represents a persistent memory entry
type MemoryItem struct {
	ID                  int64           `db:"id" json:"id"`
	OrgID               int64           `db:"org_id" json:"org_id"`
	UserID              int64           `db:"user_id" json:"user_id"`
	Scope               string          `db:"scope" json:"scope"`
	MemoryType          string          `db:"memory_type" json:"memory_type"`
	Title               string          `db:"title" json:"title"`
	Content             string          `db:"content" json:"content"`
	StructuredValueJSON *string         `db:"structured_value" json:"-"`
	StructuredValue     json.RawMessage `db:"-" json:"structured_value,omitempty"`
	SourceType          string          `db:"source_type" json:"source_type"`
	SourceReference     *string         `db:"source_reference" json:"source_reference,omitempty"`
	Evidence            *string         `db:"evidence" json:"evidence,omitempty"`
	Confidence          float64         `db:"confidence" json:"confidence"`
	ExplicitlyConfirmed bool            `db:"explicitly_confirmed" json:"explicitly_confirmed"`
	Status              string          `db:"status" json:"status"`
	ReviewAt            *time.Time      `db:"review_at" json:"review_at,omitempty"`
	ExpiresAt           *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	LastUsedAt          *time.Time      `db:"last_used_at" json:"last_used_at,omitempty"`
	CreatedBy           *string         `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy           *string         `db:"updated_by" json:"updated_by,omitempty"`
	CorrelationID       *string         `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
}

// Preference represents a key-value operational configuration preference
type Preference struct {
	ID                  int64      `db:"id" json:"id"`
	OrgID               int64      `db:"org_id" json:"org_id"`
	UserID              int64      `db:"user_id" json:"user_id"`
	Scope               string     `db:"scope" json:"scope"`
	PreferenceKey       string     `db:"preference_key" json:"preference_key"`
	PreferenceValue     string     `db:"preference_value" json:"preference_value"`
	ValueType           string     `db:"value_type" json:"value_type"`
	Description         *string    `db:"description" json:"description,omitempty"`
	Source              string     `db:"source" json:"source"`
	ExplicitlyConfirmed bool       `db:"explicitly_confirmed" json:"explicitly_confirmed"`
	IsDisabled          bool       `db:"is_disabled" json:"is_disabled"`
	DisabledAt          *time.Time `db:"disabled_at" json:"disabled_at,omitempty"`
	LastUsedAt          *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	CreatedBy           *string    `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy           *string    `db:"updated_by" json:"updated_by,omitempty"`
	CorrelationID       *string    `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// UserPersonalizationSettings captures user-specific master preferences
type UserPersonalizationSettings struct {
	ID                     int64     `db:"id" json:"id"`
	OrgID                  int64     `db:"org_id" json:"org_id"`
	UserID                 int64     `db:"user_id" json:"user_id"`
	PersonalizationEnabled bool      `db:"personalization_enabled" json:"personalization_enabled"`
	PreferredResponseStyle string    `db:"preferred_response_style" json:"preferred_response_style"`
	PreferredSummaryDepth  string    `db:"preferred_summary_depth" json:"preferred_summary_depth"`
	PreferredCurrency      string    `db:"preferred_currency" json:"preferred_currency"`
	PreferredTimezone      string    `db:"preferred_timezone" json:"preferred_timezone"`
	PreferredDateFormat    string    `db:"preferred_date_format" json:"preferred_date_format"`
	PreferredDefaultModule string    `db:"preferred_default_module" json:"preferred_default_module"`
	ExplanationLevel       string    `db:"explanation_level" json:"explanation_level"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`
}

// MemoryAuditEvent records immutable memory lifecycle events
type MemoryAuditEvent struct {
	ID            int64     `db:"id" json:"id"`
	OrgID         int64     `db:"org_id" json:"org_id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	MemoryItemID  *int64    `db:"memory_item_id" json:"memory_item_id,omitempty"`
	EventType     string    `db:"event_type" json:"event_type"`
	Scope         string    `db:"scope" json:"scope"`
	ActorName     string    `db:"actor_name" json:"actor_name"`
	ActorID       *int64    `db:"actor_id" json:"actor_id,omitempty"`
	Details       *string   `db:"details" json:"details,omitempty"`
	CorrelationID *string   `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// MemoryStats summarizes memory items for dashboards & Workforce views
type MemoryStats struct {
	PersonalizationEnabled bool  `json:"personalization_enabled"`
	ActivePersonalCount    int64 `json:"active_personal_count"`
	ActiveOrgCount         int64 `json:"active_org_count"`
	PendingReviewCount     int64 `json:"pending_review_count"`
	ExpiredCount           int64 `json:"expired_count"`
	TotalAuditEvents       int64 `json:"total_audit_events"`
}

// ProposeMemoryInput is submitted when the AI or user wants to propose a memory item
type ProposeMemoryInput struct {
	Scope           string          `json:"scope"` // "USER" or "ORGANIZATION"
	MemoryType      string          `json:"memory_type"`
	Title           string          `json:"title"`
	Content         string          `json:"content"`
	StructuredValue json.RawMessage `json:"structured_value,omitempty"`
	SourceReference *string         `json:"source_reference,omitempty"`
	Evidence        *string         `json:"evidence,omitempty"`
	WhyUseful       string          `json:"why_useful"`
	ExpiresInDays   *int            `json:"expires_in_days,omitempty"`
	CorrelationID   *string         `json:"correlation_id,omitempty"`
}

// ProposeMemoryOutput represents the safe, unpersisted proposal presented to the user
type ProposeMemoryOutput struct {
	ProposedMemory  MemoryItem `json:"proposed_memory"`
	WhyUseful       string     `json:"why_useful"`
	TargetAudience  string     `json:"target_audience"`
	RequiresConfirm bool       `json:"requires_confirmation"`
	CanEdit         bool       `json:"can_edit"`
}

// CreateMemoryInput is used to explicitly save a memory item
type CreateMemoryInput struct {
	Scope               string          `json:"scope"` // "USER" or "ORGANIZATION"
	MemoryType          string          `json:"memory_type"`
	Title               string          `json:"title"`
	Content             string          `json:"content"`
	StructuredValue     json.RawMessage `json:"structured_value,omitempty"`
	SourceType          string          `json:"source_type"`
	SourceReference     *string         `json:"source_reference,omitempty"`
	Evidence            *string         `json:"evidence,omitempty"`
	Confidence          *float64        `json:"confidence,omitempty"`
	ExplicitlyConfirmed bool            `json:"explicitly_confirmed"`
	ExpiresAt           *time.Time      `json:"expires_at,omitempty"`
	ReviewAt            *time.Time      `json:"review_at,omitempty"`
	CorrelationID       *string         `json:"correlation_id,omitempty"`
}

// UpdateMemoryInput is used to update an existing memory item
type UpdateMemoryInput struct {
	Title           *string         `json:"title,omitempty"`
	Content         *string         `json:"content,omitempty"`
	StructuredValue json.RawMessage `json:"structured_value,omitempty"`
	Status          *string         `json:"status,omitempty"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	ReviewAt        *time.Time      `json:"review_at,omitempty"`
}

// MemoryFilter holds query filters
type MemoryFilter struct {
	Scope      string `json:"scope"`
	MemoryType string `json:"memory_type"`
	Status     string `json:"status"`
	Search     string `json:"search"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// RuntimeContextRequest asks for personalized instructions to inject into AI calls
type RuntimeContextRequest struct {
	Module        string  `json:"module"` // e.g. "SHIPMENTS", "INVOICES", "PRICING", "SALES", "ASSISTANT"
	Intent        string  `json:"intent"` // e.g. "SUMMARIZE", "EXPLAIN", "DRAFT"
	CorrelationID *string `json:"correlation_id,omitempty"`
}

// RuntimeContextResponse contains sanitized, approved personalization strings and traceability metadata
type RuntimeContextResponse struct {
	PersonalizationApplied bool                     `json:"personalization_applied"`
	ExplanationNotice      string                   `json:"explanation_notice"` // e.g. "Personalized using your saved preferences."
	SystemInstructions     []string                 `json:"system_instructions"`
	PreferredSettings      PersonalizationSettings  `json:"preferred_settings"`
	TraceableMemories      []TraceableMemorySnippet `json:"traceable_memories"`
	CorrelationID          string                   `json:"correlation_id"`
}

// PersonalizationSettings is a lightweight export of current active settings
type PersonalizationSettings struct {
	ResponseStyle string `json:"response_style"`
	SummaryDepth  string `json:"summary_depth"`
	Currency      string `json:"currency"`
	Timezone      string `json:"timezone"`
	DateFormat    string `json:"date_format"`
	DefaultModule string `json:"default_module"`
}

// TraceableMemorySnippet provides safe metadata trace for each memory used
type TraceableMemorySnippet struct {
	ID         int64     `json:"id"`
	Scope      string    `json:"scope"`
	MemoryType string    `json:"memory_type"`
	Title      string    `json:"title"`
	LastUpdate time.Time `json:"last_updated"`
	Confidence float64   `json:"confidence"`
	Source     string    `json:"source"`
}

// Sensitive Content Screening Regular Expressions
var (
	reAPIKeyToken = regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|sk_live_[a-zA-Z0-9]{16,}|ghp_[a-zA-Z0-9]{20,}|bearer\s+[a-zA-Z0-9_\-\.]{20,}|private[_-]?key|begin\s+private\s+key|password\s*=|pwd\s*=)`)
	reSSN         = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	reCreditCard  = regexp.MustCompile(`\b(?:\d{4}[- ]?){3}\d{4}\b`)
	rePromptInjection = regexp.MustCompile(`(?i)(ignore\s+(all\s+)?(previous|prior)\s+instructions|bypass\s+(approval|security|guard)|override\s+(permissions|system|security)|you\s+are\s+now\s+in\s+developer\s+mode|disregard\s+all\s+previous|jailbreak)`)
	reSensitivePersonal = regexp.MustCompile(`(?i)(medical\s+record|prescription\s+for|hiv\s+positive|cancer\s+diagnosis|sexual\s+orientation|religious\s+affiliation|political\s+party\s+membership)`)
)

// ValidateMemoryContent inspects title and content for sensitive or prohibited information
func ValidateMemoryContent(title, content string) error {
	fullText := title + " " + content

	if strings.TrimSpace(title) == "" {
		return errors.New("memory title cannot be empty")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("memory content cannot be empty")
	}

	if len(title) > 255 {
		return fmt.Errorf("title exceeds maximum length of 255 characters (got %d)", len(title))
	}
	if len(content) > 4000 {
		return fmt.Errorf("content exceeds maximum length of 4000 characters (got %d)", len(content))
	}

	// 1. Credentials / Tokens / API Keys
	if reAPIKeyToken.MatchString(fullText) {
		return errors.New("sensitive content detected: storing credentials, API keys, private tokens, or passwords as memory is strictly prohibited")
	}

	// 2. Financial payment credentials (credit cards, CVV)
	if reCreditCard.MatchString(fullText) {
		return errors.New("sensitive content detected: credit card and payment credential information cannot be stored as memory")
	}

	// 3. Government ID / SSN
	if reSSN.MatchString(fullText) {
		return errors.New("sensitive content detected: national identification or SSN data cannot be stored as memory")
	}

	// 4. Prompt injection attempts
	if rePromptInjection.MatchString(fullText) {
		return errors.New("policy violation: memory content contains prompt injection or system override instructions")
	}

	// 5. Sensitive personal / health / religious / political attributes
	if reSensitivePersonal.MatchString(fullText) {
		return errors.New("sensitive content detected: medical, political, religious, or private personal data cannot be stored as memory")
	}

	return nil
}

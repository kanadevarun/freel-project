package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines business operations for AI memory and personalization
type Service interface {
	// Memory Items
	ProposeMemory(ctx context.Context, orgID int64, userID int64, userRole string, input ProposeMemoryInput) (*ProposeMemoryOutput, error)
	CreateMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, input CreateMemoryInput) (*MemoryItem, error)
	GetMemory(ctx context.Context, orgID int64, userID int64, id int64) (*MemoryItem, error)
	ListMemories(ctx context.Context, orgID int64, userID int64, filter MemoryFilter) ([]*MemoryItem, int64, error)
	UpdateMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64, input UpdateMemoryInput) (*MemoryItem, error)
	DeleteMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) error
	DisableMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) (*MemoryItem, error)
	EnableMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) (*MemoryItem, error)
	ClearPersonalMemories(ctx context.Context, orgID int64, userID int64, userName string) (int64, error)

	// Preferences
	ListPreferences(ctx context.Context, orgID int64, userID int64, scope string) ([]*Preference, error)
	SetPreference(ctx context.Context, orgID int64, userID int64, userName string, userRole string, pref *Preference) (*Preference, error)
	DeletePreference(ctx context.Context, orgID int64, userID int64, userName string, userRole string, scope string, key string) error

	// Personalization Settings
	GetUserSettings(ctx context.Context, orgID int64, userID int64) (*UserPersonalizationSettings, error)
	UpdateUserSettings(ctx context.Context, orgID int64, userID int64, userName string, settings *UserPersonalizationSettings) (*UserPersonalizationSettings, error)
	TogglePersonalization(ctx context.Context, orgID int64, userID int64, userName string, enabled bool) (*UserPersonalizationSettings, error)

	// Runtime Context Synthesis
	GetRuntimeContext(ctx context.Context, orgID int64, userID int64, req RuntimeContextRequest) (*RuntimeContextResponse, error)

	// Observability & Stats
	GetStats(ctx context.Context, orgID int64, userID int64) (*MemoryStats, error)
	ListAuditEvents(ctx context.Context, orgID int64, userID int64, limit int, offset int) ([]*MemoryAuditEvent, error)
}

type service struct {
	repo Repository
}

// NewService instantiates the memory business service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// canManageOrgScope checks whether the user's role permits creating or modifying organization-scoped items
func canManageOrgScope(role string) bool {
	upper := strings.ToUpper(strings.TrimSpace(role))
	return upper == "SUPER_ADMIN" || upper == "ADMIN" || upper == "ORGANIZATION_ADMIN" || upper == "MANAGER"
}

// ProposeMemory evaluates and returns a safe, unpersisted proposal for user confirmation
func (s *service) ProposeMemory(ctx context.Context, orgID int64, userID int64, userRole string, input ProposeMemoryInput) (*ProposeMemoryOutput, error) {
	scope := ScopeUser
	if strings.EqualFold(input.Scope, ScopeOrganization) {
		if !canManageOrgScope(userRole) {
			return nil, errors.New("insufficient permissions: only organization administrators can propose organization-scoped memory")
		}
		scope = ScopeOrganization
	}

	// Validate content safety
	if err := ValidateMemoryContent(input.Title, input.Content); err != nil {
		// Log rejection audit event safely without storing the forbidden payload
		corr := ""
		if input.CorrelationID != nil {
			corr = *input.CorrelationID
		}
		_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
			OrgID:         orgID,
			UserID:        userID,
			EventType:     AuditEventRejectedSensitive,
			Scope:         scope,
			ActorName:     "System Validation",
			Details:       ptrStr(fmt.Sprintf("Memory proposal rejected by safety screen: %v", err)),
			CorrelationID: &corr,
		})
		return nil, fmt.Errorf("safety validation failed: %w", err)
	}

	whyUseful := input.WhyUseful
	if whyUseful == "" {
		whyUseful = "Helps AI assistants format and prioritize operations according to your explicit preferences."
	}

	var exp *time.Time
	if input.ExpiresInDays != nil && *input.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, *input.ExpiresInDays)
		exp = &t
	}

	memType := input.MemoryType
	if memType == "" {
		memType = MemoryTypeResponseStyle
	}

	audience := "Only you (Personal Scope)"
	effUserID := userID
	if scope == ScopeOrganization {
		audience = "All users in this organization"
		effUserID = 0
	}

	proposed := MemoryItem{
		OrgID:               orgID,
		UserID:              effUserID,
		Scope:               scope,
		MemoryType:          memType,
		Title:               input.Title,
		Content:             input.Content,
		StructuredValue:     input.StructuredValue,
		SourceType:          SourceTypeAssistantProposed,
		SourceReference:     input.SourceReference,
		Evidence:            input.Evidence,
		Confidence:          0.95,
		ExplicitlyConfirmed: false,
		Status:              MemoryStatusPendingReview,
		ExpiresAt:           exp,
		CorrelationID:       input.CorrelationID,
	}

	// Record audit event for proposal
	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:         orgID,
		UserID:        userID,
		EventType:     AuditEventProposed,
		Scope:         scope,
		ActorName:     "AI Assistant",
		Details:       ptrStr(fmt.Sprintf("Proposed memory item: '%s'", input.Title)),
		CorrelationID: input.CorrelationID,
	})

	return &ProposeMemoryOutput{
		ProposedMemory:  proposed,
		WhyUseful:       whyUseful,
		TargetAudience:  audience,
		RequiresConfirm: true,
		CanEdit:         true,
	}, nil
}

// CreateMemory persists an explicitly confirmed memory item
func (s *service) CreateMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, input CreateMemoryInput) (*MemoryItem, error) {
	if !input.ExplicitlyConfirmed {
		return nil, errors.New("explicit confirmation required: memory cannot be created without explicit user confirmation")
	}

	scope := ScopeUser
	effectiveUserID := userID
	if strings.EqualFold(input.Scope, ScopeOrganization) {
		if !canManageOrgScope(userRole) {
			return nil, errors.New("insufficient permissions: only organization administrators can create organization-scoped memory")
		}
		scope = ScopeOrganization
		effectiveUserID = 0
	}

	if err := ValidateMemoryContent(input.Title, input.Content); err != nil {
		corr := ""
		if input.CorrelationID != nil {
			corr = *input.CorrelationID
		}
		_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
			OrgID:         orgID,
			UserID:        userID,
			EventType:     AuditEventRejectedSensitive,
			Scope:         scope,
			ActorName:     userName,
			Details:       ptrStr(fmt.Sprintf("Memory creation blocked: %v", err)),
			CorrelationID: &corr,
		})
		return nil, fmt.Errorf("safety validation failed: %w", err)
	}

	sourceType := input.SourceType
	if sourceType == "" {
		sourceType = SourceTypeExplicitUser
	}

	conf := 1.00
	if input.Confidence != nil && *input.Confidence > 0 && *input.Confidence <= 1.0 {
		conf = *input.Confidence
	}

	memType := input.MemoryType
	if memType == "" {
		memType = MemoryTypeResponseStyle
	}

	item := &MemoryItem{
		OrgID:               orgID,
		UserID:              effectiveUserID,
		Scope:               scope,
		MemoryType:          memType,
		Title:               input.Title,
		Content:             input.Content,
		StructuredValue:     input.StructuredValue,
		SourceType:          sourceType,
		SourceReference:     input.SourceReference,
		Evidence:            input.Evidence,
		Confidence:          conf,
		ExplicitlyConfirmed: true,
		Status:              MemoryStatusActive,
		ReviewAt:            input.ReviewAt,
		ExpiresAt:           input.ExpiresAt,
		CreatedBy:           &userName,
		UpdatedBy:           &userName,
		CorrelationID:       input.CorrelationID,
	}

	saved, err := s.repo.CreateMemoryItem(ctx, item)
	if err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:         orgID,
		UserID:        userID,
		MemoryItemID:  &saved.ID,
		EventType:     AuditEventCreated,
		Scope:         scope,
		ActorName:     userName,
		Details:       ptrStr(fmt.Sprintf("Created %s memory: '%s'", scope, saved.Title)),
		CorrelationID: input.CorrelationID,
	})

	return saved, nil
}

// GetMemory fetches a single memory item ensuring tenant and ownership isolation
func (s *service) GetMemory(ctx context.Context, orgID int64, userID int64, id int64) (*MemoryItem, error) {
	item, err := s.repo.GetMemoryItemByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("memory item not found")
	}

	// Isolation check: if user scope, must match authenticated user
	if item.Scope == ScopeUser && item.UserID != userID {
		return nil, errors.New("unauthorized: cannot access personal memory of another user")
	}

	return item, nil
}

// ListMemories lists accessible memories
func (s *service) ListMemories(ctx context.Context, orgID int64, userID int64, filter MemoryFilter) ([]*MemoryItem, int64, error) {
	return s.repo.ListMemoryItems(ctx, orgID, userID, filter)
}

// UpdateMemory modifies a memory item
func (s *service) UpdateMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64, input UpdateMemoryInput) (*MemoryItem, error) {
	item, err := s.GetMemory(ctx, orgID, userID, id)
	if err != nil {
		return nil, err
	}

	if item.Scope == ScopeOrganization && !canManageOrgScope(userRole) {
		return nil, errors.New("insufficient permissions: only organization administrators can update organization-scoped memory")
	}

	newTitle := item.Title
	if input.Title != nil {
		newTitle = *input.Title
	}
	newContent := item.Content
	if input.Content != nil {
		newContent = *input.Content
	}

	if err := ValidateMemoryContent(newTitle, newContent); err != nil {
		return nil, fmt.Errorf("safety validation failed: %w", err)
	}

	item.Title = newTitle
	item.Content = newContent
	if len(input.StructuredValue) > 0 {
		item.StructuredValue = input.StructuredValue
	}
	if input.Status != nil && *input.Status != "" {
		item.Status = *input.Status
	}
	if input.ExpiresAt != nil {
		item.ExpiresAt = input.ExpiresAt
	}
	if input.ReviewAt != nil {
		item.ReviewAt = input.ReviewAt
	}
	item.UpdatedBy = &userName

	updated, err := s.repo.UpdateMemoryItem(ctx, item)
	if err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:        orgID,
		UserID:       userID,
		MemoryItemID: &id,
		EventType:    AuditEventUpdated,
		Scope:        item.Scope,
		ActorName:    userName,
		Details:      ptrStr(fmt.Sprintf("Updated memory '%s'", updated.Title)),
	})

	return updated, nil
}

// DeleteMemory deletes a memory item
func (s *service) DeleteMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) error {
	item, err := s.GetMemory(ctx, orgID, userID, id)
	if err != nil {
		return err
	}

	if item.Scope == ScopeOrganization && !canManageOrgScope(userRole) {
		return errors.New("insufficient permissions: only organization administrators can delete organization-scoped memory")
	}

	if err := s.repo.DeleteMemoryItem(ctx, orgID, id); err != nil {
		return err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:        orgID,
		UserID:       userID,
		MemoryItemID: &id,
		EventType:    AuditEventDeleted,
		Scope:        item.Scope,
		ActorName:    userName,
		Details:      ptrStr(fmt.Sprintf("Deleted memory item #%d ('%s')", id, item.Title)),
	})

	return nil
}

// DisableMemory sets status to DISABLED
func (s *service) DisableMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) (*MemoryItem, error) {
	item, err := s.GetMemory(ctx, orgID, userID, id)
	if err != nil {
		return nil, err
	}

	if item.Scope == ScopeOrganization && !canManageOrgScope(userRole) {
		return nil, errors.New("insufficient permissions to disable organization memory")
	}

	if err := s.repo.SetMemoryStatus(ctx, orgID, id, MemoryStatusDisabled, userName); err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:        orgID,
		UserID:       userID,
		MemoryItemID: &id,
		EventType:    AuditEventDisabled,
		Scope:        item.Scope,
		ActorName:    userName,
		Details:      ptrStr(fmt.Sprintf("Disabled memory '%s'", item.Title)),
	})

	return s.repo.GetMemoryItemByID(ctx, orgID, id)
}

// EnableMemory sets status to ACTIVE
func (s *service) EnableMemory(ctx context.Context, orgID int64, userID int64, userName string, userRole string, id int64) (*MemoryItem, error) {
	item, err := s.GetMemory(ctx, orgID, userID, id)
	if err != nil {
		return nil, err
	}

	if item.Scope == ScopeOrganization && !canManageOrgScope(userRole) {
		return nil, errors.New("insufficient permissions to enable organization memory")
	}

	if err := s.repo.SetMemoryStatus(ctx, orgID, id, MemoryStatusActive, userName); err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:        orgID,
		UserID:       userID,
		MemoryItemID: &id,
		EventType:    AuditEventReenabled,
		Scope:        item.Scope,
		ActorName:    userName,
		Details:      ptrStr(fmt.Sprintf("Re-enabled memory '%s'", item.Title)),
	})

	return s.repo.GetMemoryItemByID(ctx, orgID, id)
}

// ClearPersonalMemories clears all personal memories for authenticated user
func (s *service) ClearPersonalMemories(ctx context.Context, orgID int64, userID int64, userName string) (int64, error) {
	cleared, err := s.repo.ClearUserMemories(ctx, orgID, userID)
	if err != nil {
		return 0, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:     orgID,
		UserID:    userID,
		EventType: AuditEventCleared,
		Scope:     ScopeUser,
		ActorName: userName,
		Details:   ptrStr(fmt.Sprintf("Cleared %d personal memory items", cleared)),
	})

	return cleared, nil
}

// ListPreferences lists preferences for scope
func (s *service) ListPreferences(ctx context.Context, orgID int64, userID int64, scope string) ([]*Preference, error) {
	if strings.EqualFold(scope, ScopeOrganization) {
		return s.repo.ListPreferences(ctx, orgID, userID, ScopeOrganization)
	}
	return s.repo.ListPreferences(ctx, orgID, userID, ScopeUser)
}

// SetPreference saves or updates a key-value preference
func (s *service) SetPreference(ctx context.Context, orgID int64, userID int64, userName string, userRole string, pref *Preference) (*Preference, error) {
	scope := ScopeUser
	effUserID := userID
	if strings.EqualFold(pref.Scope, ScopeOrganization) {
		if !canManageOrgScope(userRole) {
			return nil, errors.New("insufficient permissions: only organization administrators can set organization preferences")
		}
		scope = ScopeOrganization
		effUserID = 0
	}

	if err := ValidateMemoryContent(pref.PreferenceKey, pref.PreferenceValue); err != nil {
		return nil, fmt.Errorf("safety validation failed: %w", err)
	}

	pref.OrgID = orgID
	pref.UserID = effUserID
	pref.Scope = scope
	pref.CreatedBy = &userName
	pref.UpdatedBy = &userName
	pref.ExplicitlyConfirmed = true

	saved, err := s.repo.SetPreference(ctx, pref)
	if err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:     orgID,
		UserID:    userID,
		EventType: AuditEventUpdated,
		Scope:     scope,
		ActorName: userName,
		Details:   ptrStr(fmt.Sprintf("Updated preference '%s' = '%s'", pref.PreferenceKey, pref.PreferenceValue)),
	})

	return saved, nil
}

// DeletePreference removes a preference
func (s *service) DeletePreference(ctx context.Context, orgID int64, userID int64, userName string, userRole string, scope string, key string) error {
	effScope := ScopeUser
	if strings.EqualFold(scope, ScopeOrganization) {
		if !canManageOrgScope(userRole) {
			return errors.New("insufficient permissions: only organization administrators can delete organization preferences")
		}
		effScope = ScopeOrganization
	}

	if err := s.repo.DeletePreference(ctx, orgID, userID, effScope, key); err != nil {
		return err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:     orgID,
		UserID:    userID,
		EventType: AuditEventDeleted,
		Scope:     effScope,
		ActorName: userName,
		Details:   ptrStr(fmt.Sprintf("Deleted preference key '%s'", key)),
	})

	return nil
}

// GetUserSettings gets personalization settings
func (s *service) GetUserSettings(ctx context.Context, orgID int64, userID int64) (*UserPersonalizationSettings, error) {
	return s.repo.GetUserSettings(ctx, orgID, userID)
}

// UpdateUserSettings updates user personalization settings
func (s *service) UpdateUserSettings(ctx context.Context, orgID int64, userID int64, userName string, settings *UserPersonalizationSettings) (*UserPersonalizationSettings, error) {
	settings.OrgID = orgID
	settings.UserID = userID

	updated, err := s.repo.UpsertUserSettings(ctx, settings)
	if err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:     orgID,
		UserID:    userID,
		EventType: AuditEventUpdated,
		Scope:     ScopeUser,
		ActorName: userName,
		Details:   ptrStr("Updated user personalization settings"),
	})

	return updated, nil
}

// TogglePersonalization toggles the master switch
func (s *service) TogglePersonalization(ctx context.Context, orgID int64, userID int64, userName string, enabled bool) (*UserPersonalizationSettings, error) {
	settings, err := s.repo.GetUserSettings(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	settings.PersonalizationEnabled = enabled
	updated, err := s.repo.UpsertUserSettings(ctx, settings)
	if err != nil {
		return nil, err
	}

	evtType := AuditEventReenabled
	action := "enabled"
	if !enabled {
		evtType = AuditEventDisabled
		action = "disabled"
	}

	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:     orgID,
		UserID:    userID,
		EventType: evtType,
		Scope:     ScopeUser,
		ActorName: userName,
		Details:   ptrStr(fmt.Sprintf("Master personalization %s", action)),
	})

	return updated, nil
}

// GetRuntimeContext synthesizes personalized instructions for AI calls
func (s *service) GetRuntimeContext(ctx context.Context, orgID int64, userID int64, req RuntimeContextRequest) (*RuntimeContextResponse, error) {
	corrID := uuid.New().String()
	if req.CorrelationID != nil && *req.CorrelationID != "" {
		corrID = *req.CorrelationID
	}

	userSettings, err := s.repo.GetUserSettings(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	// If personalization is disabled, return clean default instructions
	if !userSettings.PersonalizationEnabled {
		return &RuntimeContextResponse{
			PersonalizationApplied: false,
			ExplanationNotice:      "Personalization is currently disabled in your settings.",
			SystemInstructions:     []string{},
			PreferredSettings: PersonalizationSettings{
				ResponseStyle: "STANDARD",
				SummaryDepth:  "STANDARD",
				Currency:      "USD",
				Timezone:      "UTC",
				DateFormat:    "YYYY-MM-DD",
				DefaultModule: "DASHBOARD",
			},
			TraceableMemories: []TraceableMemorySnippet{},
			CorrelationID:     corrID,
		}, nil
	}

	// Fetch active memories
	activeMemories, err := s.repo.GetActiveMemoriesForRuntime(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	var instructions []string
	var snippets []TraceableMemorySnippet

	// 1. Settings instructions
	if userSettings.PreferredResponseStyle != "" {
		instructions = append(instructions, fmt.Sprintf("Response Style: Formulate responses in a %s manner.", strings.ToLower(userSettings.PreferredResponseStyle)))
	}
	if userSettings.PreferredSummaryDepth != "" {
		instructions = append(instructions, fmt.Sprintf("Summary Depth: Provide %s summaries for data overviews.", strings.ToLower(userSettings.PreferredSummaryDepth)))
	}
	if userSettings.ExplanationLevel != "" {
		instructions = append(instructions, fmt.Sprintf("Explanation Level: Provide %s explanations without redundant boilerplate.", strings.ToLower(userSettings.ExplanationLevel)))
	}
	if userSettings.PreferredCurrency != "" {
		instructions = append(instructions, fmt.Sprintf("Currency: Default currency display preference is %s.", userSettings.PreferredCurrency))
	}
	if userSettings.PreferredDateFormat != "" {
		instructions = append(instructions, fmt.Sprintf("Date Format: Format dates as %s when presenting to user.", userSettings.PreferredDateFormat))
	}

	// 2. Active memory instructions
	for _, m := range activeMemories {
		// Strictly filter out any expired or non-active items
		if m.ExpiresAt != nil && m.ExpiresAt.Before(time.Now()) {
			continue
		}
		if m.Status != MemoryStatusActive {
			continue
		}

		instruction := fmt.Sprintf("[%s Preference] %s: %s", m.Scope, m.Title, m.Content)
		instructions = append(instructions, instruction)

		snippets = append(snippets, TraceableMemorySnippet{
			ID:         m.ID,
			Scope:      m.Scope,
			MemoryType: m.MemoryType,
			Title:      m.Title,
			LastUpdate: m.UpdatedAt,
			Confidence: m.Confidence,
			Source:     m.SourceType,
		})
	}

	// Record audit event for runtime context consumption
	_ = s.repo.RecordAuditEvent(ctx, &MemoryAuditEvent{
		OrgID:         orgID,
		UserID:        userID,
		EventType:     AuditEventUsedInRuntime,
		Scope:         ScopeUser,
		ActorName:     "AI Runtime",
		Details:       ptrStr(fmt.Sprintf("Personalization context generated for module=%s intent=%s (%d memories applied)", req.Module, req.Intent, len(snippets))),
		CorrelationID: &corrID,
	})

	return &RuntimeContextResponse{
		PersonalizationApplied: true,
		ExplanationNotice:      "Personalized using your saved preferences.",
		SystemInstructions:     instructions,
		PreferredSettings: PersonalizationSettings{
			ResponseStyle: userSettings.PreferredResponseStyle,
			SummaryDepth:  userSettings.PreferredSummaryDepth,
			Currency:      userSettings.PreferredCurrency,
			Timezone:      userSettings.PreferredTimezone,
			DateFormat:    userSettings.PreferredDateFormat,
			DefaultModule: userSettings.PreferredDefaultModule,
		},
		TraceableMemories: snippets,
		CorrelationID:     corrID,
	}, nil
}

// GetStats returns summary metrics
func (s *service) GetStats(ctx context.Context, orgID int64, userID int64) (*MemoryStats, error) {
	return s.repo.GetMemoryStats(ctx, orgID, userID)
}

// ListAuditEvents returns audit events
func (s *service) ListAuditEvents(ctx context.Context, orgID int64, userID int64, limit int, offset int) ([]*MemoryAuditEvent, error) {
	return s.repo.ListAuditEvents(ctx, orgID, userID, limit, offset)
}

func ptrStr(s string) *string {
	return &s
}

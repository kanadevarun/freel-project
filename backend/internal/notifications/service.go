package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/common/events"
)

const (
	DefaultSidecarURL     = "http://localhost:8090"
	DefaultServiceKey     = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	DefaultSidecarTimeout = 15 * time.Second
)

type serviceImpl struct {
	repo         Repository
	engine       Engine
	eventBus     events.Bus
	emailService Service // Optional fallback for outbound email (e.g. SMTP/SES)
	approvalsSvc approvals.Service
	sidecarURL   string
	serviceKey   string
	environment  string
	httpClient   *http.Client
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

func WithApprovals(appr approvals.Service) ServiceOption {
	return func(s *serviceImpl) {
		s.approvalsSvc = appr
	}
}

func WithEnvironment(env string) ServiceOption {
	return func(s *serviceImpl) {
		s.environment = strings.ToLower(strings.TrimSpace(env))
	}
}

func WithServiceKey(key string) ServiceOption {
	return func(s *serviceImpl) {
		s.serviceKey = key
	}
}

// ResolveServiceKey resolves the service key securely.
// DefaultServiceKey is ONLY permitted in explicit "development" or "test" environments.
// In staging, production, unknown, or missing environments, it fails closed returning empty string.
func ResolveServiceKey(explicitKey, env string) string {
	if explicitKey != "" {
		return explicitKey
	}
	key := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if key == "" {
		key = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if key == "" {
		key = os.Getenv("AI_SIDECAR_SERVICE_KEY")
	}
	if key != "" {
		return key
	}

	normEnv := strings.ToLower(strings.TrimSpace(env))
	if normEnv == "" {
		normEnv = strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	}
	if normEnv == "" {
		normEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	}
	if normEnv == "" {
		normEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	}

	if normEnv == "development" || normEnv == "test" {
		return DefaultServiceKey
	}

	// Fail closed: no fallback for staging, production, unknown, or missing environment
	return ""
}

func NewService(repo Repository, engine Engine, eb events.Bus, emailService Service, opts ...ServiceOption) Service {
	sidecar := os.Getenv("AI_SIDECAR_URL")
	if sidecar == "" {
		sidecar = DefaultSidecarURL
	}

	s := &serviceImpl{
		repo:         repo,
		engine:       engine,
		eventBus:     eb,
		emailService: emailService,
		sidecarURL:   sidecar,
		httpClient:   &http.Client{Timeout: DefaultSidecarTimeout},
	}

	for _, opt := range opts {
		opt(s)
	}

	// Resolve service key using fail-closed environment resolution if not explicitly set
	if s.serviceKey == "" {
		s.serviceKey = ResolveServiceKey("", s.environment)
	}

	if eb != nil {
		// Listen for real operational events to trigger evaluation
		eb.Subscribe(events.EventRFQCreated, func(e events.Event) {
			if payload, ok := e.Payload.(map[string]interface{}); ok {
				var orgID int64
				if oid, ok := payload["org_id"].(float64); ok {
					orgID = int64(oid)
				} else if oid, ok := payload["org_id"].(int64); ok {
					orgID = oid
				}
				if orgID > 0 {
					_, _ = s.EvaluateNotifications(context.Background(), orgID)
				}
			}
		})
	}

	return s
}

func (s *serviceImpl) SetApprovalsService(appr approvals.Service) {
	s.approvalsSvc = appr
}

// ── Outbound Email / Messaging (Delegated or Safe No-op) ──

func (s *serviceImpl) SendEmail(ctx context.Context, to string, subject string, body string) error {
	if s.emailService != nil {
		return s.emailService.SendEmail(ctx, to, subject, body)
	}
	log.Printf("[Notifications Service] Outbound email to %s suppressed for in-app safety", to)
	return nil
}

func (s *serviceImpl) SendWhatsApp(ctx context.Context, phone string, message string) error {
	if s.emailService != nil {
		return s.emailService.SendWhatsApp(ctx, phone, message)
	}
	log.Printf("[Notifications Service] Outbound WhatsApp to %s suppressed for in-app safety", phone)
	return nil
}

func (s *serviceImpl) SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error {
	if s.emailService != nil {
		return s.emailService.SendInviteEmail(ctx, toEmail, token, orgName)
	}
	log.Printf("[Notifications Service] Outbound Invite to %s suppressed for in-app safety", toEmail)
	return nil
}

// ── In-App Notifications & Escalations ──

func (s *serviceImpl) GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error) {
	filter := NotificationFilter{
		OrgID:  int64(orgID),
		IsRead: boolPtr(false),
		Limit:  20,
	}
	items, _, err := s.repo.List(ctx, filter)
	return items, err
}

func (s *serviceImpl) ListNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, int, error) {
	// Auto-evaluate notifications for the org to ensure real-time currency
	if filter.OrgID > 0 {
		_, _ = s.engine.Evaluate(ctx, filter.OrgID)
	}
	return s.repo.List(ctx, filter)
}

func (s *serviceImpl) GetNotification(ctx context.Context, orgID int64, id int64) (*Notification, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

func (s *serviceImpl) GetUnreadCount(ctx context.Context, orgID int64, userID int64, role string) (int, error) {
	return s.repo.GetUnreadCount(ctx, orgID, userID, role)
}

func (s *serviceImpl) GetStats(ctx context.Context, orgID int64, userID int64, role string) (*NotificationStats, error) {
	// Auto-evaluate notifications to make stats accurate
	_, _ = s.engine.Evaluate(ctx, orgID)
	return s.repo.GetStats(ctx, orgID, userID, role)
}

func (s *serviceImpl) MarkAsRead(ctx context.Context, orgID int32, notifID int32) error {
	return s.repo.MarkRead(ctx, int64(orgID), int64(notifID), true)
}

func (s *serviceImpl) MarkRead(ctx context.Context, orgID int64, id int64, read bool) error {
	return s.repo.MarkRead(ctx, orgID, id, read)
}

func (s *serviceImpl) MarkAllAsRead(ctx context.Context, orgID int64, userID int64, role string) error {
	return s.repo.MarkAllRead(ctx, orgID, userID, role)
}

func (s *serviceImpl) Dismiss(ctx context.Context, orgID int64, id int64) error {
	return s.repo.Dismiss(ctx, orgID, id)
}

func (s *serviceImpl) Acknowledge(ctx context.Context, orgID int64, id int64, userID int64) error {
	if orgID <= 0 || id <= 0 {
		return fmt.Errorf("invalid org_id or notification id")
	}
	return s.repo.Acknowledge(ctx, orgID, id, userID)
}

func (s *serviceImpl) Snooze(ctx context.Context, orgID int64, id int64, durationMinutes int) error {
	if orgID <= 0 || id <= 0 {
		return fmt.Errorf("invalid org_id or notification id")
	}
	if durationMinutes <= 0 {
		durationMinutes = 60
	}
	snoozedUntil := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	return s.repo.Snooze(ctx, orgID, id, snoozedUntil)
}

func (s *serviceImpl) AnalyzeWithAI(ctx context.Context, orgID int64, id int64) (*NotificationAnalysisResponseDTO, error) {
	if s.serviceKey == "" {
		return nil, fmt.Errorf("sidecar communication disabled: internal service key is not configured")
	}

	notif, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil || notif == nil {
		return nil, fmt.Errorf("notification not found")
	}

	ageHours := time.Since(notif.CreatedAt).Hours()
	var contextPayload map[string]interface{}
	if notif.Metadata != nil && *notif.Metadata != "" {
		_ = json.Unmarshal([]byte(*notif.Metadata), &contextPayload)
	}
	if contextPayload == nil {
		contextPayload = make(map[string]interface{})
	}

	reqDTO := NotificationAnalysisRequestDTO{
		OrgID:                   orgID,
		NotificationID:          &notif.ID,
		SourceModule:            notif.SourceModule,
		SourceRecordType:        notif.SourceRecordType,
		SourceRecordID:          notif.SourceRecordID,
		NotificationType:        notif.NotificationType,
		Title:                   notif.Title,
		Message:                 notif.Message,
		Severity:                notif.Severity,
		CurrentPriority:         notif.Priority,
		AgeHours:                ageHours,
		EscalationLevel:         notif.EscalationLevel,
		UnresolvedDurationHours: ageHours,
		ContextPayload:          contextPayload,
		CorrelationID:           notif.CorrelationID,
	}

	bodyBytes, err := json.Marshal(reqDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to encode AI request: %w", err)
	}

	url := fmt.Sprintf("%s/notifications-escalations/analyze-and-prioritize", s.sidecarURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI sidecar call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI sidecar returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var analysis NotificationAnalysisResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		return nil, fmt.Errorf("failed to decode AI analysis: %w", err)
	}

	// Update notification with AI insights
	_ = s.repo.UpdateAIDetails(ctx, orgID, id, analysis.AISummary, analysis.AIEscalationReason, analysis.GroupKey, analysis.PriorityScore)

	return &analysis, nil
}

func (s *serviceImpl) EscalateWithAI(ctx context.Context, orgID int64, id int64, userID int64, reason string) (*EscalationEvent, error) {
	notif, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil || notif == nil {
		return nil, fmt.Errorf("notification not found for escalation")
	}

	// 1. Analyze issue with AI to get ground-truth escalation summary
	analysis, err := s.AnalyzeWithAI(ctx, orgID, id)
	if err != nil {
		log.Printf("[Notifications Service] Sidecar analyze failed for escalation, proceeding with deterministic fallback: %v", err)
		analysis = &NotificationAnalysisResponseDTO{
			AISummary:          fmt.Sprintf("Escalation for %s #%s: %s", notif.SourceRecordType, notif.SourceRecordID, notif.Title),
			AIEscalationReason: reason,
			SuggestedAction:    "Operational review required",
			RiskLevel:          "HIGH",
			RequiresApproval:   true,
			CorrelationID:      notif.CorrelationID,
		}
	}

	nextLevel := notif.EscalationLevel + 1
	var nextSeverity string
	switch strings.ToUpper(notif.Severity) {
	case "INFORMATIONAL":
		nextSeverity = "MEDIUM"
	case "LOW":
		nextSeverity = "HIGH"
	case "MEDIUM":
		nextSeverity = "HIGH"
	default:
		nextSeverity = "CRITICAL"
	}

	// 2. HITL Approval gating for high-risk escalations
	var approvalID *int64
	if (analysis.RequiresApproval || analysis.RiskLevel == "HIGH" || analysis.RiskLevel == "CRITICAL") && s.approvalsSvc != nil {
		entityID, _ := strconv.ParseInt(notif.SourceRecordID, 10, 64)
		appReq, appErr := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:              fmt.Sprintf("Operational Escalation (Tier %d): %s #%s", nextLevel, notif.SourceRecordType, notif.SourceRecordID),
			Category:           "NOTIFICATION_ESCALATION",
			Type:               "ESCALATION_APPROVAL",
			Priority:           analysis.RiskLevel,
			RelatedEntityType:  notif.SourceRecordType,
			RelatedEntityID:    entityID,
			RelatedRef:         fmt.Sprintf("%s-%s", notif.SourceRecordType, notif.SourceRecordID),
			RequestedByID:     userID,
			Description:        fmt.Sprintf("Issue on %s #%s escalated: %s. Recommended Action: %s", notif.SourceRecordType, notif.SourceRecordID, reason, analysis.SuggestedAction),
			ActionName:         "notifications.escalate_to_management",
			RiskLevel:          analysis.RiskLevel,
			RequiredPermission: "operations:write",
			ProposedPayload:    fmt.Sprintf(`{"notification_id": %d, "escalation_level": %d, "reason": %q}`, id, nextLevel, reason),
			CorrelationID:      notif.CorrelationID,
			ActorType:          "AI",
		}, "Notification Escalation Engine")

		if appErr == nil && appReq != nil {
			approvalID = &appReq.ID
		}
	}

	// 3. Persist escalation audit event
	escEvent := &EscalationEvent{
		NotificationID:      id,
		OrgID:               orgID,
		EscalationLevel:     nextLevel,
		EscalationReason:    reason,
		PreviousSeverity:    notif.Severity,
		NewSeverity:         nextSeverity,
		TriggerType:         "AI_ASSISTED_MANUAL",
		ThresholdHours:      int(time.Since(notif.CreatedAt).Hours()),
		AIEscalationSummary: &analysis.AIEscalationReason,
		RecommendedAction:   &analysis.SuggestedAction,
		ApprovalID:          approvalID,
		CorrelationID:       notif.CorrelationID,
	}

	createdEsc, err := s.repo.CreateEscalationEvent(ctx, escEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to persist escalation event: %w", err)
	}

	// 4. Update notification state
	_ = s.repo.Escalate(ctx, orgID, id, nextSeverity, reason, "AI_ASSISTED_MANUAL", int(time.Since(notif.CreatedAt).Hours()), notif.CorrelationID)

	return createdEsc, nil
}

func (s *serviceImpl) GenerateDraftWithAI(ctx context.Context, orgID int64, id int64, draftType string) (*EscalationDraftResponseDTO, error) {
	if s.serviceKey == "" {
		return nil, fmt.Errorf("sidecar communication disabled: internal service key is not configured")
	}

	notif, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil || notif == nil {
		return nil, fmt.Errorf("notification not found")
	}

	roleTarget := "OPERATIONS"
	if notif.RoleTarget != nil && *notif.RoleTarget != "" {
		roleTarget = *notif.RoleTarget
	}

	reqDTO := EscalationDraftRequestDTO{
		OrgID:           orgID,
		NotificationID:  notif.ID,
		DraftType:       draftType,
		SourceModule:    notif.SourceModule,
		SourceRecordID:  notif.SourceRecordID,
		RecipientRole:   roleTarget,
		EscalationLevel: notif.EscalationLevel,
		KeyFindings:     []string{notif.Title, notif.Message},
		CorrelationID:   notif.CorrelationID,
	}

	bodyBytes, err := json.Marshal(reqDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to encode draft request: %w", err)
	}

	url := fmt.Sprintf("%s/notifications-escalations/generate-escalation-draft", s.sidecarURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar draft request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI sidecar draft call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI sidecar returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var draftResp EscalationDraftResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&draftResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI draft: %w", err)
	}

	return &draftResp, nil
}

func (s *serviceImpl) ListEscalations(ctx context.Context, orgID int64, limit int) ([]EscalationEvent, error) {
	return s.repo.ListEscalations(ctx, orgID, limit)
}

func (s *serviceImpl) GetPreferences(ctx context.Context, orgID int64, userID int64) (*UserNotificationPreferences, error) {
	return s.repo.GetPreferences(ctx, orgID, userID)
}

func (s *serviceImpl) UpdatePreferences(ctx context.Context, prefs *UserNotificationPreferences) error {
	if prefs == nil {
		return fmt.Errorf("preferences cannot be nil")
	}
	return s.repo.UpdatePreferences(ctx, prefs)
}

func (s *serviceImpl) EvaluateNotifications(ctx context.Context, orgID int64) (int, error) {
	return s.engine.Evaluate(ctx, orgID)
}

func boolPtr(b bool) *bool {
	return &b
}

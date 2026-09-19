package notifications

import (
	"context"

	"github.com/freel/backend/internal/approvals"
)

// UnimplementedInAppService provides default no-op implementations of in-app notification methods
// for pure outbound providers like SMTP and SES.
type UnimplementedInAppService struct{}

func (UnimplementedInAppService) ListNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, int, error) {
	return nil, 0, nil
}

func (UnimplementedInAppService) GetNotification(ctx context.Context, orgID int64, id int64) (*Notification, error) {
	return nil, nil
}

func (UnimplementedInAppService) GetUnreadCount(ctx context.Context, orgID int64, userID int64, role string) (int, error) {
	return 0, nil
}

func (UnimplementedInAppService) GetStats(ctx context.Context, orgID int64, userID int64, role string) (*NotificationStats, error) {
	return &NotificationStats{}, nil
}

func (UnimplementedInAppService) MarkRead(ctx context.Context, orgID int64, id int64, read bool) error {
	return nil
}

func (UnimplementedInAppService) MarkAllAsRead(ctx context.Context, orgID int64, userID int64, role string) error {
	return nil
}

func (UnimplementedInAppService) Dismiss(ctx context.Context, orgID int64, id int64) error {
	return nil
}

func (UnimplementedInAppService) ListEscalations(ctx context.Context, orgID int64, limit int) ([]EscalationEvent, error) {
	return nil, nil
}

func (UnimplementedInAppService) GetPreferences(ctx context.Context, orgID int64, userID int64) (*UserNotificationPreferences, error) {
	return &UserNotificationPreferences{
		UserID:                 userID,
		OrgID:                  orgID,
		MinSeverity:            SeverityInformational,
		InAppEnabled:           true,
		AssignedOnly:           false,
		ApprovalsEnabled:       true,
		AutomationsEnabled:     true,
		RecommendationsEnabled: true,
		FinanceEnabled:         true,
		OperationsEnabled:      true,
		ComplianceEnabled:      true,
	}, nil
}

func (UnimplementedInAppService) UpdatePreferences(ctx context.Context, prefs *UserNotificationPreferences) error {
	return nil
}

func (UnimplementedInAppService) Acknowledge(ctx context.Context, orgID int64, id int64, userID int64) error {
	return nil
}

func (UnimplementedInAppService) Snooze(ctx context.Context, orgID int64, id int64, durationMinutes int) error {
	return nil
}

func (UnimplementedInAppService) EscalateWithAI(ctx context.Context, orgID int64, id int64, userID int64, reason string) (*EscalationEvent, error) {
	return nil, nil
}

func (UnimplementedInAppService) GenerateDraftWithAI(ctx context.Context, orgID int64, id int64, draftType string) (*EscalationDraftResponseDTO, error) {
	return nil, nil
}

func (UnimplementedInAppService) AnalyzeWithAI(ctx context.Context, orgID int64, id int64) (*NotificationAnalysisResponseDTO, error) {
	return nil, nil
}

func (UnimplementedInAppService) EvaluateNotifications(ctx context.Context, orgID int64) (int, error) {
	return 0, nil
}

func (UnimplementedInAppService) SetApprovalsService(appr approvals.Service) {
}


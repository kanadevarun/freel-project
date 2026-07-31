package notifications

import (
	"context"
	"time"

	"github.com/freel/backend/internal/common/events"
)

type mockInAppServiceImpl struct {
	eventBus events.Bus
}

// In-memory mock store for notifications for MVP
var mockNotifications = []Notification{
	{
		ID:        1,
		OrgID:     1, // Assuming org 1
		Title:     "New Lead Created",
		Message:   "Acme Corp was added to the CRM.",
		Type:      "INFO",
		IsRead:    false,
		CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        2,
		OrgID:     1,
		Title:     "Pricing Agent Drafted Quote",
		Message:   "A new quote for RFQ #12 requires your approval.",
		Type:      "SUCCESS",
		IsRead:    false,
		CreatedAt: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
	},
}

func NewMockInAppService(eb events.Bus) Service {
	s := &mockInAppServiceImpl{
		eventBus: eb,
	}

	// Subscribe to events that generate notifications
	s.eventBus.Subscribe(events.EventRFQAssigned, func(e events.Event) {
		// Mock inserting a notification
	})

	return s
}

func (s *mockInAppServiceImpl) SendEmail(ctx context.Context, to string, subject string, body string) error {
	return nil
}
func (s *mockInAppServiceImpl) SendWhatsApp(ctx context.Context, phone string, message string) error {
	return nil
}
func (s *mockInAppServiceImpl) SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error {
	return nil
}

func (s *mockInAppServiceImpl) GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error) {
	var result []Notification
	for _, n := range mockNotifications {
		if n.OrgID == orgID && !n.IsRead {
			result = append(result, n)
		}
	}
	return result, nil
}

func (s *mockInAppServiceImpl) MarkAsRead(ctx context.Context, orgID int32, notifID int32) error {
	for i, n := range mockNotifications {
		if n.OrgID == orgID && n.ID == notifID {
			mockNotifications[i].IsRead = true
			break
		}
	}
	return nil
}

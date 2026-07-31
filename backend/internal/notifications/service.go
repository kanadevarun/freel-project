package notifications

import "context"

// Service handles sending notifications across different channels.
type Service interface {
	// SendEmail sends an email to a recipient.
	// Simple meaning: It shoots off an email behind the scenes using AWS SES or another provider.
	// Example: err := notifSvc.SendEmail(ctx, "client@example.com", "Your Quote", "Here is your quote...")
	SendEmail(ctx context.Context, to string, subject string, body string) error

	// SendWhatsApp sends a WhatsApp message.
	// Simple meaning: It sends a quick text message to a user's WhatsApp number.
	// Example: err := notifSvc.SendWhatsApp(ctx, "+123456789", "Your cargo has arrived!")
	SendWhatsApp(ctx context.Context, phone string, message string) error

	// SendInviteEmail sends a styled HTML email inviting a user to join an organization.
	// It uses the invite.html template and injects the dynamic token and org name.
	SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error

	// In-App Notifications
	GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error)
	MarkAsRead(ctx context.Context, orgID int32, notifID int32) error
}

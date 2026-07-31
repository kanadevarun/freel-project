package activity

import "context"

// TimelineService handles creating and retrieving audit logs and activity timelines.
type TimelineService interface {
	// RecordActivity logs an event that happened to a specific entity.
	// Simple meaning: It writes down history, like "John sent an email to this Lead at 3 PM".
	// Example: err := timelineSvc.RecordActivity(ctx, "LEAD", 123, "EMAIL_SENT", "Sent introductory email.")
	RecordActivity(ctx context.Context, entityType string, entityID int64, action string, details string) error

	// GetTimeline gets the history of what happened to an entity.
	// Simple meaning: It reads back the history book for a specific Lead or RFQ so we can show it on the UI.
	// Example: events, err := timelineSvc.GetTimeline(ctx, "LEAD", 123)
	GetTimeline(ctx context.Context, entityType string, entityID int64) ([]interface{}, error)
}

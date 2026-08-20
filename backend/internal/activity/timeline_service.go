package activity

import (
	"context"
	"fmt"
	"log"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/middleware"
	"github.com/jmoiron/sqlx"
)

type timelineService struct {
	db       *sqlx.DB
	eventBus events.Bus
}

// NewTimelineService creates a new Activity Timeline service and starts event listeners.
func NewTimelineService(db *sqlx.DB, eb events.Bus) TimelineService {
	svc := &timelineService{
		db:       db,
		eventBus: eb,
	}

	// Subscribe to relevant events
	svc.eventBus.Subscribe(events.EventLeadCreated, svc.handleLeadCreatedEvent)
	svc.eventBus.Subscribe(events.EventLeadEnriched, svc.handleLeadEnrichedEvent)

	return svc
}

func (s *timelineService) RecordActivity(ctx context.Context, entityType string, entityID int64, action string, details string) error {
	var orgID int64 = 1
	userCtx, ok := ctx.Value(middleware.UserContextKey).(middleware.UserContext)
	if ok && userCtx.OrgID > 0 {
		orgID = userCtx.OrgID
	}
	
	query := `
		INSERT INTO activities (org_id, entity_type, entity_id, action, description) 
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, orgID, entityType, entityID, action, details)
	return err
}

func (s *timelineService) GetTimeline(ctx context.Context, entityType string, entityID int64) ([]interface{}, error) {
	// Not fully implemented for this backend sprint, just returning empty.
	return []interface{}{}, nil
}

// handleLeadCreatedEvent logs that a lead was created.
func (s *timelineService) handleLeadCreatedEvent(event events.Event) {
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return
	}

	var leadID int64
	if idFloat, ok := payload["lead_id"].(float64); ok {
		leadID = int64(idFloat)
	} else if idInt, ok := payload["lead_id"].(int32); ok {
		leadID = int64(idInt)
	} else if idInt2, ok := payload["lead_id"].(int); ok {
		leadID = int64(idInt2)
	} else {
		return
	}

	err := s.RecordActivity(context.Background(), "LEAD", leadID, "CREATED", "Lead was manually added to the system.")
	if err != nil {
		log.Printf("Timeline Error: failed to log LeadCreated activity: %v", err)
	}
}

// handleLeadEnrichedEvent logs that a lead was automatically enriched by AI.
func (s *timelineService) handleLeadEnrichedEvent(event events.Event) {
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return
	}

	var leadID int64
	if idFloat, ok := payload["lead_id"].(float64); ok {
		leadID = int64(idFloat)
	} else if idInt, ok := payload["lead_id"].(int32); ok {
		leadID = int64(idInt)
	} else if idInt2, ok := payload["lead_id"].(int); ok {
		leadID = int64(idInt2)
	} else {
		return
	}

	var score int32
	if scFloat, ok := payload["score"].(float64); ok {
		score = int32(scFloat)
	} else if scInt, ok := payload["score"].(int32); ok {
		score = scInt
	}

	details := fmt.Sprintf("Lead was automatically enriched and scored by AI. Score: %d/100", score)
	err := s.RecordActivity(context.Background(), "LEAD", leadID, "AI_ENRICHED", details)
	if err != nil {
		log.Printf("Timeline Error: failed to log LeadEnriched activity: %v", err)
	}
}

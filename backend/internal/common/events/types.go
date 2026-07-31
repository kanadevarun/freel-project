package events

// Pre-defined event types
const (
	EventOrgCreated     EventType = "org.created"
	EventUserInvited    EventType = "user.invited"
	EventLeadCreated    EventType = "lead.created"
	EventLeadEnriched   EventType = "lead.enriched"
	EventRFQCreated     EventType = "rfq.created"
	EventRFQUpdated     EventType = "rfq.updated"
	EventRFQAssigned    EventType = "rfq.assigned"
	EventQuoteGenerated EventType = "quote.generated"
	EventQuoteApproved  EventType = "quote.approved"
	EventQuoteSent      EventType = "quote.sent"
	EventRFQWon         EventType = "rfq.won"
	EventRFQLost        EventType = "rfq.lost"
	
	EventEmailReceived  EventType = "email.received"
	
	EventAgentPricingStarted   EventType = "agent.pricing.started"
	EventAgentPricingDraftReady EventType = "agent.pricing.draft_ready"
)

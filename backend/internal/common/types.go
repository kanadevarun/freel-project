package common

// AIRecommendation represents a standardized output from an AI analysis.
// This generic structure ensures that AI suggestions can be used universally
// across Pricing, Operations, Finance, and Sales.
type AIRecommendation struct {
	Type            string `json:"type"`             // e.g., "Margin", "Carrier", "Compliance", "Task"
	Priority        string `json:"priority"`         // "High", "Medium", "Low"
	Confidence      int    `json:"confidence"`       // 0-100
	Reason          string `json:"reason"`           // Why is this recommended?
	SuggestedAction string `json:"suggested_action"` // Actionable instruction or task
}

// TimelineEvent represents a standardized, chronologically ordered event.
// Used for displaying HubSpot-style unified timelines (System Events, Notes, Emails, AI Tasks).
type TimelineEvent struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "SYSTEM_EVENT", "NOTE", "EMAIL", "AI_TASK"
	Time        string `json:"time"` // Formatted timestamp (e.g., 09:15)
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"` // Optional hex color for UI dot
	Icon        string `json:"icon"`  // Optional emoji or icon key
}

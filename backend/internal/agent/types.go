package agent

// Agent States representing the State Machine of the AI Employee
const (
	StateIdle                  = "IDLE"
	StateCollectingInformation = "COLLECTING_INFORMATION"
	StateAnalyzing             = "ANALYZING_DATA"
	StateWaitingForLLM         = "WAITING_FOR_LLM"
	StateGeneratingDraft       = "GENERATING_DRAFT"
	StateWaitingForHuman       = "WAITING_FOR_HUMAN_REVIEW"
	StateCompleted             = "COMPLETED"
	StateError                 = "ERROR"
)

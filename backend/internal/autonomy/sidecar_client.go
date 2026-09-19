package autonomy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type StepConditionPredicateDTO struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type StepVerificationCriteriaDTO struct {
	CheckType     string      `json:"check_type"`
	TargetField   string      `json:"target_field"`
	ExpectedValue interface{} `json:"expected_value"`
	Description   string      `json:"description,omitempty"`
}

type SidecarPlanGenRequest struct {
	OrgID             int64                  `json:"org_id"`
	UserID            *int64                 `json:"user_id,omitempty"`
	GoalID            string                 `json:"goal_id,omitempty"`
	Goal              string                 `json:"goal"`
	Module            string                 `json:"module"`
	RelatedEntityType string                 `json:"related_entity_type"`
	RelatedEntityID   string                 `json:"related_entity_id"`
	Objective         string                 `json:"objective,omitempty"`
	Priority          string                 `json:"priority,omitempty"`
	RiskTolerance     string                 `json:"risk_tolerance,omitempty"`
	CurrentState      map[string]interface{} `json:"current_state"`
	Constraints       []string               `json:"constraints"`
	HardConstraints   []ConstraintDTO        `json:"hard_constraints"`
	SoftConstraints   []ConstraintDTO        `json:"soft_constraints"`
	AutonomyLevel     AutonomyLevel          `json:"autonomy_level"`
	CorrelationID     string                 `json:"correlation_id"`
}

type SidecarPlanStep struct {
	StepID               string                       `json:"step_id"`
	StepNumber           int                          `json:"step_number"`
	ActionType           string                       `json:"action_type"`
	Title                string                       `json:"title"`
	Description          string                       `json:"description"`
	Parameters           map[string]interface{}       `json:"parameters"`
	Dependencies         []string                     `json:"dependencies"`
	ConditionPredicate   *StepConditionPredicateDTO   `json:"condition_predicate,omitempty"`
	ExpectedOutcome      string                       `json:"expected_outcome"`
	VerificationCriteria *StepVerificationCriteriaDTO `json:"verification_criteria,omitempty"`
	RiskLevel            string                       `json:"risk_level"`
	RequiresApproval     bool                         `json:"requires_approval"`
	Reversibility        string                       `json:"reversibility"`
	FallbackAction       map[string]interface{}       `json:"fallback_action,omitempty"`
	TimeoutSeconds       int                          `json:"timeout_seconds"`
	IdempotencyKey       string                       `json:"idempotency_key"`
}

type SidecarStopCondition struct {
	ConditionType    string      `json:"condition_type"`
	Threshold        interface{} `json:"threshold"`
	EscalationTarget string      `json:"escalation_target"`
	Description      string      `json:"description"`
}

type SidecarPlanGenResponse struct {
	PlanID              string                 `json:"plan_id"`
	GoalID              *string                `json:"goal_id"`
	Version             int                    `json:"version"`
	ParentPlanID        *string                `json:"parent_plan_id"`
	Goal                string                 `json:"goal"`
	Module              string                 `json:"module"`
	RelatedEntityType   string                 `json:"related_entity_type"`
	RelatedEntityID     string                 `json:"related_entity_id"`
	CurrentStateSumm    string                 `json:"current_state_summary"`
	Constraints         []string               `json:"constraints"`
	HardConstraints     []ConstraintDTO        `json:"hard_constraints"`
	SoftConstraints     []ConstraintDTO        `json:"soft_constraints"`
	Assumptions         []string               `json:"assumptions"`
	Risks               []interface{}          `json:"risks"`
	Candidates          []map[string]interface{} `json:"candidates"`
	SelectedCandidateID string                 `json:"selected_candidate_id"`
	OrderedSteps        []SidecarPlanStep      `json:"ordered_steps"`
	EvaluationSummary   map[string]interface{} `json:"evaluation_summary"`
	ConfidenceScore     float64                `json:"confidence_score"`
	DataSufficiency     bool                   `json:"data_sufficiency"`
	EstimatedImpact     string                 `json:"estimated_impact"`
	RiskLevel           string                 `json:"risk_level"`
	AutonomyLevel       AutonomyLevel          `json:"autonomy_level"`
	StopConditions      []SidecarStopCondition `json:"stop_conditions"`
	StalenessStatus     string                 `json:"staleness_status"`
	WaitingState        string                 `json:"waiting_state,omitempty"`
	WaitingUntil        string                 `json:"waiting_until,omitempty"`
	CustomerCommitmentDate string              `json:"customer_commitment_date,omitempty"`
	PredictedETA        string                 `json:"predicted_eta,omitempty"`
	ETADeviationHours   float64                `json:"eta_deviation_hours,omitempty"`
	CommitmentRiskSeverity string              `json:"commitment_risk_severity,omitempty"`
	Status              PlanStatus             `json:"status"`
	CorrelationID       string                 `json:"correlation_id"`
	Explanation         string                 `json:"explanation"`
}

type SidecarShipmentEventEvalRequest struct {
	OrgID                  int64                  `json:"org_id"`
	ShipmentID             int64                  `json:"shipment_id"`
	EventType              string                 `json:"event_type"`
	Severity               string                 `json:"severity"`
	CurrentState           map[string]interface{} `json:"current_state"`
	PreviousState          map[string]interface{} `json:"previous_state,omitempty"`
	ActivePlan             map[string]interface{} `json:"active_plan,omitempty"`
	CustomerCommitmentDate *string                `json:"customer_commitment_date,omitempty"`
	PredictedETA           *string                `json:"predicted_eta,omitempty"`
	RawEventPayload        map[string]interface{} `json:"raw_event_payload,omitempty"`
	CorrelationID          string                 `json:"correlation_id"`
}

type SidecarShipmentEventEvalResponse struct {
	ShipmentID             int64   `json:"shipment_id"`
	EventType              string  `json:"event_type"`
	Decision               string  `json:"decision"`
	DecisionReason         string  `json:"decision_reason"`
	IsMeaningfulChange     bool    `json:"is_meaningful_change"`
	ETADeviationHours      float64 `json:"eta_deviation_hours"`
	CommitmentRiskSeverity string  `json:"commitment_risk_severity"`
	RecommendedActionType  string  `json:"recommended_action_type,omitempty"`
	EscalationReason       string  `json:"escalation_reason,omitempty"`
	RequiresNewPlan        bool    `json:"requires_new_plan"`
	RequiresReplan         bool    `json:"requires_replan"`
	ActivePlanID           string  `json:"active_plan_id,omitempty"`
	WaitingState           string  `json:"waiting_state,omitempty"`
	CorrelationID          string  `json:"correlation_id"`
}

type SidecarPlanEvalRequest struct {
	Plan   *SidecarPlanGenResponse `json:"plan"`
	Policy map[string]interface{}  `json:"policy"`
}

type SidecarPlanEvalResponse struct {
	IsPermitted       bool     `json:"is_permitted"`
	PolicyDecision    string   `json:"policy_decision"`
	RequiresApproval  bool     `json:"requires_approval"`
	PolicyReason      string   `json:"policy_reason"`
	ViolatedRules     []string `json:"violated_rules"`
	ConfidenceAccept  bool     `json:"confidence_acceptable"`
	DataSufficiencyOk bool     `json:"data_sufficiency_acceptable"`
}

type SidecarReplanRequest struct {
	OriginalPlanID      string                 `json:"original_plan_id"`
	CurrentVersion      int                    `json:"current_version"`
	ExecutedSteps       []interface{}          `json:"executed_steps"`
	FailedOrDriftedStep interface{}            `json:"failed_or_drifted_step"`
	ObservedNewState    map[string]interface{} `json:"observed_new_state"`
	ReplanReason        string                 `json:"replan_reason"`
	TriggeringEvent     string                 `json:"triggering_event"`
	CorrelationID       string                 `json:"correlation_id"`
}

type SidecarReplanResponse struct {
	RevisedPlanID   string            `json:"revised_plan_id"`
	NewVersion      int               `json:"new_version"`
	ParentPlanID    string            `json:"parent_plan_id"`
	ReplanReason    string            `json:"replan_reason"`
	ChangesSummary  string            `json:"changes_summary"`
	UpdatedSteps    []SidecarPlanStep `json:"updated_steps"`
	ConfidenceScore float64           `json:"confidence_score"`
	RiskLevel       string            `json:"risk_level"`
	Status          PlanStatus        `json:"status"`
}

type SidecarClient interface {
	GeneratePlan(ctx context.Context, req *SidecarPlanGenRequest) (*SidecarPlanGenResponse, error)
	EvaluatePlan(ctx context.Context, req *SidecarPlanEvalRequest) (*SidecarPlanEvalResponse, error)
	Replan(ctx context.Context, req *SidecarReplanRequest) (*SidecarReplanResponse, error)
	EvaluateShipmentEvent(ctx context.Context, req *SidecarShipmentEventEvalRequest) (*SidecarShipmentEventEvalResponse, error)
	GenerateAdaptiveShipmentPlan(ctx context.Context, req *SidecarPlanGenRequest) (*SidecarPlanGenResponse, error)
	EvaluateCustomerFollowup(ctx context.Context, req *SidecarCustomerFollowupEvalRequest) (*SidecarCustomerFollowupEvalResponse, error)
	ClassifyCustomerResponse(ctx context.Context, req *SidecarClassifyCustomerResponseRequest) (*SidecarClassifyCustomerResponseResponse, error)
	EvaluateRfqPricing(ctx context.Context, req *SidecarEvaluateRfqPricingRequest) (*SidecarEvaluateRfqPricingResponse, error)
	ReplanRfqPricing(ctx context.Context, req *SidecarReplanPricingRequest) (*SidecarEvaluateRfqPricingResponse, error)
	EvaluateFinanceCollection(ctx context.Context, req *FinanceCollectionEvaluationRequestDTO) (*FinanceCollectionEvaluationResponseDTO, error)
	ReplanFinanceCollection(ctx context.Context, req *FinanceCollectionReplanningRequestDTO) (*FinanceCollectionEvaluationResponseDTO, error)
	EvaluateContractCompliance(ctx context.Context, req *ContractComplianceEvaluationRequestDTO) (*ContractComplianceEvaluationResponseDTO, error)
	ReplanContractCompliance(ctx context.Context, req *ContractComplianceReplanningRequestDTO) (*ContractComplianceEvaluationResponseDTO, error)
	EvaluateExceptionResolution(ctx context.Context, req *ExceptionEvaluationRequestDTO) (*ExceptionEvaluationResponseDTO, error)
	ReplanExceptionResolution(ctx context.Context, req *ExceptionReplanningRequestDTO) (*ExceptionEvaluationResponseDTO, error)
	ValidatePlanGraph(ctx context.Context, req *SidecarPlanValidationRequest) (*SidecarPlanValidationResponse, error)
	GenerateCrossModulePlan(ctx context.Context, req *SidecarCrossModulePlanRequest) (*SidecarCrossModulePlanResponse, error)
	EvaluateStateChange(ctx context.Context, req *SidecarStateChangeEvaluationRequest) (*SidecarStateChangeEvaluationResponse, error)
	GenerateAdaptiveReplan(ctx context.Context, req *SidecarContinuousReplanningRequest) (*SidecarContinuousReplanningResponse, error)
	AnalyzeDecisionPoint(ctx context.Context, req *SidecarDecisionAnalysisRequest) (*SidecarDecisionAnalysisResponse, error)
	AnalyzeHumanFeedback(ctx context.Context, req *SidecarFeedbackAnalysisRequest) (*SidecarFeedbackAnalysisResponse, error)
	PrioritizeCommandCenterItems(ctx context.Context, req *SidecarPrioritizeRequest) (*SidecarPrioritizeResponse, error)
	EvaluateOutcome(ctx context.Context, req *SidecarOutcomeEvaluationRequest) (*SidecarOutcomeEvaluationResponse, error)
	RetrieveRelevantMemory(ctx context.Context, req *SidecarMemoryRetrievalRequest) (*SidecarMemoryRetrievalResponse, error)
	DetectLearnedPatterns(ctx context.Context, req *SidecarPatternDetectionRequest) (*SidecarPatternDetectionResponse, error)
	ResolveMemoryConflicts(ctx context.Context, req *SidecarMemoryConflictRequest) (*SidecarMemoryConflictResponse, error)
	EvaluateGovernanceContext(ctx context.Context, req SidecarGovernanceEvalRequest) (*SidecarGovernanceEvalResponse, error)
	PreviewGovernancePlan(ctx context.Context, req SidecarPlanPreviewRequest) (*SidecarPlanPreviewResponse, error)
}


type httpSidecarClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSidecarClient(baseURL, apiKey string) SidecarClient {
	if baseURL == "" {
		baseURL = os.Getenv("AI_SIDECAR_URL")
		if baseURL == "" {
			baseURL = "http://127.0.0.1:8090"
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if apiKey == "" {
		apiKey = os.Getenv("INTERNAL_SERVICE_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("INTERNAL_SERVICE_TOKEN")
			if apiKey == "" {
				apiKey = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
			}
		}
	}

	return &httpSidecarClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *httpSidecarClient) GeneratePlan(ctx context.Context, req *SidecarPlanGenRequest) (*SidecarPlanGenResponse, error) {
	url := fmt.Sprintf("%s/autonomy/plan/generate", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling generate plan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar generate plan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar generate plan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var planResp SidecarPlanGenResponse
	if err := json.Unmarshal(respBody, &planResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar plan response: %w", err)
	}
	return &planResp, nil
}

func (c *httpSidecarClient) EvaluatePlan(ctx context.Context, req *SidecarPlanEvalRequest) (*SidecarPlanEvalResponse, error) {
	url := fmt.Sprintf("%s/autonomy/plan/evaluate", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate plan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate plan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate plan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp SidecarPlanEvalResponse
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar eval response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) Replan(ctx context.Context, req *SidecarReplanRequest) (*SidecarReplanResponse, error) {
	url := fmt.Sprintf("%s/autonomy/plan/replan", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling replan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar replan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar replan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp SidecarReplanResponse
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar replan response: %w", err)
	}
	return &replanResp, nil
}

func (c *httpSidecarClient) EvaluateShipmentEvent(ctx context.Context, req *SidecarShipmentEventEvalRequest) (*SidecarShipmentEventEvalResponse, error) {
	url := fmt.Sprintf("%s/autonomy/shipments/evaluate-event", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate shipment event request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate shipment event call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate shipment event returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp SidecarShipmentEventEvalResponse
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar shipment eval response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) GenerateAdaptiveShipmentPlan(ctx context.Context, req *SidecarPlanGenRequest) (*SidecarPlanGenResponse, error) {
	url := fmt.Sprintf("%s/autonomy/shipments/adaptive-plan", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling generate adaptive shipment plan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar adaptive shipment plan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar adaptive shipment plan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var planResp SidecarPlanGenResponse
	if err := json.Unmarshal(respBody, &planResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar adaptive plan response: %w", err)
	}
	return &planResp, nil
}

// -----------------------------------------------------------------------------
// Customer Follow-Up Sidecar Integration
// -----------------------------------------------------------------------------

type SidecarCustomerFollowupDraft struct {
	Subject         string   `json:"subject"`
	ActualFacts     []string `json:"actual_facts"`
	Predictions     []string `json:"predictions"`
	Recommendations []string `json:"recommendations"`
	FullBody        string   `json:"full_body"`
	Channel         string   `json:"channel"`
}

type SidecarCustomerFollowupEvalRequest struct {
	OrgID                int64                             `json:"org_id"`
	CustomerID           int64                             `json:"customer_id"`
	CustomerName         string                            `json:"customer_name"`
	AccountTier          string                            `json:"account_tier"`
	EventType            string                            `json:"event_type"`
	EventPayload         map[string]interface{}            `json:"event_payload"`
	Preferences          *CustomerCommunicationPreferences `json:"preferences,omitempty"`
	Contact              *VerifiedContact                  `json:"contact,omitempty"`
	RecentFollowupsCount int                               `json:"recent_followups_count"`
	LastFollowupHoursAgo *float64                          `json:"last_followup_hours_ago,omitempty"`
	ActivePlan           map[string]interface{}            `json:"active_plan,omitempty"`
	CorrelationID        string                            `json:"correlation_id"`
}

type SidecarCustomerFollowupEvalResponse struct {
	CustomerID            int64                         `json:"customer_id"`
	EventType             string                        `json:"event_type"`
	Decision              string                        `json:"decision"`
	DecisionReason        string                        `json:"decision_reason"`
	Urgency               string                        `json:"urgency"`
	Draft                 *SidecarCustomerFollowupDraft `json:"draft,omitempty"`
	Channel               string                        `json:"channel"`
	RequiresApproval      bool                          `json:"requires_approval"`
	ApprovalReason        string                        `json:"approval_reason,omitempty"`
	RecommendedActionType string                        `json:"recommended_action_type"`
	StopConditions        []string                      `json:"stop_conditions"`
	Confidence            float64                       `json:"confidence"`
	CorrelationID         string                        `json:"correlation_id"`
}

type SidecarClassifyCustomerResponseRequest struct {
	OrgID           int64                  `json:"org_id"`
	CustomerID      int64                  `json:"customer_id"`
	ContactName     string                 `json:"contact_name"`
	MessageText     string                 `json:"message_text"`
	FollowupContext map[string]interface{} `json:"followup_context,omitempty"`
	CorrelationID   string                 `json:"correlation_id"`
}

type SidecarClassifyCustomerResponseResponse struct {
	CustomerID          int64   `json:"customer_id"`
	Classification      string  `json:"classification"`
	Sentiment           string  `json:"sentiment"`
	ActionRequested     string  `json:"action_requested,omitempty"`
	RecommendedNextStep string  `json:"recommended_next_step"`
	Confidence          float64 `json:"confidence"`
	CorrelationID       string  `json:"correlation_id"`
}

func (c *httpSidecarClient) EvaluateCustomerFollowup(ctx context.Context, req *SidecarCustomerFollowupEvalRequest) (*SidecarCustomerFollowupEvalResponse, error) {
	url := fmt.Sprintf("%s/autonomy/customers/evaluate-followup", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling customer followup eval request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar customer followup eval call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar customer followup eval returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp SidecarCustomerFollowupEvalResponse
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar customer followup eval response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) ClassifyCustomerResponse(ctx context.Context, req *SidecarClassifyCustomerResponseRequest) (*SidecarClassifyCustomerResponseResponse, error) {
	url := fmt.Sprintf("%s/autonomy/customers/classify-response", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling classify customer response request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar classify customer response call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar classify customer response returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var classResp SidecarClassifyCustomerResponseResponse
	if err := json.Unmarshal(respBody, &classResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar classify response: %w", err)
	}
	return &classResp, nil
}

func (c *httpSidecarClient) EvaluateRfqPricing(ctx context.Context, req *SidecarEvaluateRfqPricingRequest) (*SidecarEvaluateRfqPricingResponse, error) {
	url := fmt.Sprintf("%s/autonomy/pricing/evaluate-rfq", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate rfq pricing request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate rfq pricing call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate rfq pricing returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp SidecarEvaluateRfqPricingResponse
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate rfq pricing response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) ReplanRfqPricing(ctx context.Context, req *SidecarReplanPricingRequest) (*SidecarEvaluateRfqPricingResponse, error) {
	url := fmt.Sprintf("%s/autonomy/pricing/replan", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling replan rfq pricing request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar replan rfq pricing call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar replan rfq pricing returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp SidecarEvaluateRfqPricingResponse
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar replan rfq pricing response: %w", err)
	}
	return &replanResp, nil
}

func (c *httpSidecarClient) EvaluateFinanceCollection(ctx context.Context, req *FinanceCollectionEvaluationRequestDTO) (*FinanceCollectionEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/finance/evaluate-collection", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate finance collection request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate finance collection call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate finance collection returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp FinanceCollectionEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate finance collection response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) ReplanFinanceCollection(ctx context.Context, req *FinanceCollectionReplanningRequestDTO) (*FinanceCollectionEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/finance/replan-collection", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling replan finance collection request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar replan finance collection call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar replan finance collection returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp FinanceCollectionEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar replan finance collection response: %w", err)
	}
	return &replanResp, nil
}

func (c *httpSidecarClient) EvaluateContractCompliance(ctx context.Context, req *ContractComplianceEvaluationRequestDTO) (*ContractComplianceEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/compliance/evaluate-contract", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate contract compliance request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate contract compliance call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate contract compliance returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp ContractComplianceEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate contract compliance response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) ReplanContractCompliance(ctx context.Context, req *ContractComplianceReplanningRequestDTO) (*ContractComplianceEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/compliance/replan-contract", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling replan contract compliance request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar replan contract compliance call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar replan contract compliance returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp ContractComplianceEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar replan contract compliance response: %w", err)
	}
	return &replanResp, nil
}

// -------------------------------------------------------------------------
// Phase 5 Task 5.8: Autonomous Exception Resolution Sidecar DTOs and Methods
// -------------------------------------------------------------------------

type ExceptionResolutionContextDTO struct {
	OrgID             int64                    `json:"org_id"`
	ExceptionID       int64                    `json:"exception_id"`
	ShipmentID        int64                    `json:"shipment_id"`
	ExceptionType     string                   `json:"exception_type"`
	Severity          string                   `json:"severity"`
	Title             string                   `json:"title"`
	Description       *string                  `json:"description,omitempty"`
	Status            string                   `json:"status"`
	SourceEventID     *string                  `json:"source_event_id,omitempty"`
	ShipmentDetails   map[string]interface{}   `json:"shipment_details,omitempty"`
	CustomerDetails   map[string]interface{}   `json:"customer_details,omitempty"`
	ActiveMilestones  []map[string]interface{} `json:"active_milestones,omitempty"`
	RelatedExceptions []map[string]interface{} `json:"related_exceptions,omitempty"`
	Documents         []map[string]interface{} `json:"documents,omitempty"`
	Notes             *string                  `json:"notes,omitempty"`
}

type ExceptionEvaluationRequestDTO struct {
	Context       ExceptionResolutionContextDTO `json:"context"`
	CorrelationID string                        `json:"correlation_id"`
}

type ExceptionReplanningRequestDTO struct {
	Context            ExceptionResolutionContextDTO `json:"context"`
	CurrentPlanVersion int                           `json:"current_plan_version"`
	TriggerEvent       string                        `json:"trigger_event"`
	EventPayload       map[string]interface{}        `json:"event_payload,omitempty"`
	CorrelationID      string                        `json:"correlation_id"`
}

type ExceptionEvaluationResponseDTO struct {
	ExceptionID          int64                                   `json:"exception_id"`
	ShipmentID           int64                                   `json:"shipment_id"`
	ExceptionType        string                                  `json:"exception_type"`
	Severity             string                                  `json:"severity"`
	LifecycleStatus      string                                  `json:"lifecycle_status"`
	WaitingState         *string                                 `json:"waiting_state,omitempty"`
	Symptom              string                                  `json:"symptom"`
	LikelyRootCause      string                                  `json:"likely_root_cause"`
	ContributingFactors  []string                                `json:"contributing_factors,omitempty"`
	Evidence             []string                                `json:"evidence,omitempty"`
	ConfidenceScore      float64                                 `json:"confidence_score"`
	UnknownFactors       []string                                `json:"unknown_factors,omitempty"`
	ImpactAssessment     map[string]interface{}                  `json:"impact_assessment,omitempty"`
	HardConstraints      []string                                `json:"hard_constraints,omitempty"`
	CandidateStrategies  []ExceptionCandidateRecoveryStrategyDTO `json:"candidate_strategies,omitempty"`
	SelectedStrategyID   string                                  `json:"selected_strategy_id"`
	SelectedStrategyName string                                  `json:"selected_strategy_name"`
	RecoveryPlanSteps    []map[string]interface{}                `json:"recovery_plan_steps,omitempty"`
	VerificationCriteria map[string]interface{}                  `json:"verification_criteria,omitempty"`
	RequiresApproval     bool                                    `json:"requires_approval"`
	ApprovalReason       *string                                 `json:"approval_reason,omitempty"`
	StopReason           *string                                 `json:"stop_reason,omitempty"`
	EscalationReason     *string                                 `json:"escalation_reason,omitempty"`
	DataSufficiency      string                                  `json:"data_sufficiency"`
	CorrelationID        string                                  `json:"correlation_id"`
}

func (c *httpSidecarClient) EvaluateExceptionResolution(ctx context.Context, req *ExceptionEvaluationRequestDTO) (*ExceptionEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/exceptions/evaluate-exception", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate exception request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate exception call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate exception returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp ExceptionEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate exception response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) ReplanExceptionResolution(ctx context.Context, req *ExceptionReplanningRequestDTO) (*ExceptionEvaluationResponseDTO, error) {
	url := fmt.Sprintf("%s/autonomy/exceptions/replan-exception", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling replan exception request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar replan exception call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar replan exception returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp ExceptionEvaluationResponseDTO
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar replan exception response: %w", err)
	}
	return &replanResp, nil
}

func (c *httpSidecarClient) ValidatePlanGraph(ctx context.Context, req *SidecarPlanValidationRequest) (*SidecarPlanValidationResponse, error) {
	url := fmt.Sprintf("%s/autonomy/plan/validate-graph", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling validate plan graph request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar validate plan graph call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar validate plan graph returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var valResp SidecarPlanValidationResponse
	if err := json.Unmarshal(respBody, &valResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar validate plan graph response: %w", err)
	}
	return &valResp, nil
}

func (c *httpSidecarClient) GenerateCrossModulePlan(ctx context.Context, req *SidecarCrossModulePlanRequest) (*SidecarCrossModulePlanResponse, error) {
	url := fmt.Sprintf("%s/autonomy/plan/cross-module", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling cross-module plan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar cross-module plan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar cross-module plan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var cmResp SidecarCrossModulePlanResponse
	if err := json.Unmarshal(respBody, &cmResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar cross-module plan response: %w", err)
	}
	return &cmResp, nil
}

func (c *httpSidecarClient) EvaluateStateChange(ctx context.Context, req *SidecarStateChangeEvaluationRequest) (*SidecarStateChangeEvaluationResponse, error) {
	url := fmt.Sprintf("%s/autonomy/monitoring/evaluate-state-change", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate state change request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate state change call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate state change returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var evalResp SidecarStateChangeEvaluationResponse
	if err := json.Unmarshal(respBody, &evalResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate state change response: %w", err)
	}
	return &evalResp, nil
}

func (c *httpSidecarClient) GenerateAdaptiveReplan(ctx context.Context, req *SidecarContinuousReplanningRequest) (*SidecarContinuousReplanningResponse, error) {
	url := fmt.Sprintf("%s/autonomy/monitoring/replan", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling adaptive replan request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar adaptive replan call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar adaptive replan returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var replanResp SidecarContinuousReplanningResponse
	if err := json.Unmarshal(respBody, &replanResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar adaptive replan response: %w", err)
	}
	return &replanResp, nil
}

func (c *httpSidecarClient) AnalyzeDecisionPoint(ctx context.Context, req *SidecarDecisionAnalysisRequest) (*SidecarDecisionAnalysisResponse, error) {
	url := fmt.Sprintf("%s/autonomy/human-ai/analyze-decision", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling decision analysis request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar analyze decision call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar analyze decision returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var decResp SidecarDecisionAnalysisResponse
	if err := json.Unmarshal(respBody, &decResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar analyze decision response: %w", err)
	}
	return &decResp, nil
}

func (c *httpSidecarClient) AnalyzeHumanFeedback(ctx context.Context, req *SidecarFeedbackAnalysisRequest) (*SidecarFeedbackAnalysisResponse, error) {
	url := fmt.Sprintf("%s/autonomy/human-ai/analyze-feedback", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling feedback analysis request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar analyze feedback call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar analyze feedback returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var fbResp SidecarFeedbackAnalysisResponse
	if err := json.Unmarshal(respBody, &fbResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar analyze feedback response: %w", err)
	}
	return &fbResp, nil
}

func (c *httpSidecarClient) PrioritizeCommandCenterItems(ctx context.Context, req *SidecarPrioritizeRequest) (*SidecarPrioritizeResponse, error) {
	url := fmt.Sprintf("%s/autonomy/command-center/prioritize", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling prioritize command center request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar command center prioritize call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar command center prioritize returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var pResp SidecarPrioritizeResponse
	if err := json.Unmarshal(respBody, &pResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar command center prioritize response: %w", err)
	}
	return &pResp, nil
}

func (c *httpSidecarClient) EvaluateOutcome(ctx context.Context, req *SidecarOutcomeEvaluationRequest) (*SidecarOutcomeEvaluationResponse, error) {
	url := fmt.Sprintf("%s/autonomy/memory/evaluate-outcome", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling evaluate outcome request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar evaluate outcome call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar evaluate outcome returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var outResp SidecarOutcomeEvaluationResponse
	if err := json.Unmarshal(respBody, &outResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar evaluate outcome response: %w", err)
	}
	return &outResp, nil
}

func (c *httpSidecarClient) RetrieveRelevantMemory(ctx context.Context, req *SidecarMemoryRetrievalRequest) (*SidecarMemoryRetrievalResponse, error) {
	url := fmt.Sprintf("%s/autonomy/memory/retrieve", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling retrieve memory request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar retrieve memory call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar retrieve memory returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var retResp SidecarMemoryRetrievalResponse
	if err := json.Unmarshal(respBody, &retResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar retrieve memory response: %w", err)
	}
	return &retResp, nil
}

func (c *httpSidecarClient) DetectLearnedPatterns(ctx context.Context, req *SidecarPatternDetectionRequest) (*SidecarPatternDetectionResponse, error) {
	url := fmt.Sprintf("%s/autonomy/memory/detect-patterns", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling detect patterns request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar detect patterns call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar detect patterns returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var patResp SidecarPatternDetectionResponse
	if err := json.Unmarshal(respBody, &patResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar detect patterns response: %w", err)
	}
	return &patResp, nil
}

func (c *httpSidecarClient) ResolveMemoryConflicts(ctx context.Context, req *SidecarMemoryConflictRequest) (*SidecarMemoryConflictResponse, error) {
	url := fmt.Sprintf("%s/autonomy/memory/resolve-conflicts", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling resolve conflicts request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar resolve conflicts call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar resolve conflicts returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var confResp SidecarMemoryConflictResponse
	if err := json.Unmarshal(respBody, &confResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar resolve conflicts response: %w", err)
	}
	return &confResp, nil
}

func (c *httpSidecarClient) EvaluateGovernanceContext(ctx context.Context, req SidecarGovernanceEvalRequest) (*SidecarGovernanceEvalResponse, error) {
	url := fmt.Sprintf("%s/autonomy/governance/evaluate-context", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling governance eval request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar governance eval call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar governance eval returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var govResp SidecarGovernanceEvalResponse
	if err := json.Unmarshal(respBody, &govResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar governance eval response: %w", err)
	}
	return &govResp, nil
}

func (c *httpSidecarClient) PreviewGovernancePlan(ctx context.Context, req SidecarPlanPreviewRequest) (*SidecarPlanPreviewResponse, error) {
	url := fmt.Sprintf("%s/autonomy/governance/preview-plan", c.baseURL)
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling governance plan preview request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.apiKey)
	httpReq.Header.Set("X-Internal-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed executing sidecar governance plan preview call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar governance plan preview returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var prevResp SidecarPlanPreviewResponse
	if err := json.Unmarshal(respBody, &prevResp); err != nil {
		return nil, fmt.Errorf("failed unmarshaling sidecar governance plan preview response: %w", err)
	}
	return &prevResp, nil
}






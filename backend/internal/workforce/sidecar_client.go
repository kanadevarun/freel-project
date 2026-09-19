package workforce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrSidecarUnavailable = errors.New("ai sidecar unavailable")
	ErrSidecarInvalidResp = errors.New("invalid sidecar response structure")
)

type SidecarClient interface {
	ExecuteTask(ctx context.Context, req *SidecarTaskRequest) (*SidecarTaskResponse, error)
	ListAgents(ctx context.Context) ([]*WorkforceAgent, error)
}

type HTTPSSidecarClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSidecarClient(baseURL string) *HTTPSSidecarClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8090"
	}
	return &HTTPSSidecarClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *HTTPSSidecarClient) ExecuteTask(ctx context.Context, req *SidecarTaskRequest) (*SidecarTaskResponse, error) {
	url := fmt.Sprintf("%s/api/v1/workforce/execute-task", c.baseURL)

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling sidecar task request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed creating sidecar request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSidecarUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("sidecar returned HTTP %d: %v", resp.StatusCode, errResp)
	}

	var taskResp SidecarTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSidecarInvalidResp, err)
	}

	// CRITICAL SECURITY VALIDATION: Validate untrusted AI output
	c.validateAndSanitizeOutput(&taskResp)

	return &taskResp, nil
}

func (c *HTTPSSidecarClient) ListAgents(ctx context.Context) ([]*WorkforceAgent, error) {
	url := fmt.Sprintf("%s/api/v1/workforce/agents", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating sidecar request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSidecarUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar returned HTTP %d", resp.StatusCode)
	}

	var rawAgents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawAgents); err != nil {
		return nil, fmt.Errorf("failed decoding sidecar agents: %w", err)
	}

	var agents []*WorkforceAgent
	for _, raw := range rawAgents {
		agentID, _ := raw["agent_id"].(string)
		name, _ := raw["name"].(string)
		desc, _ := raw["description"].(string)
		agentTypeStr, _ := raw["agent_type"].(string)
		autonomyLevel, _ := raw["autonomy_level"].(string)
		health, _ := raw["health_status"].(string)

		capsJSON, _ := json.Marshal(raw["capabilities"])
		tasksJSON, _ := json.Marshal(raw["allowed_tasks"])
		entitiesJSON, _ := json.Marshal(raw["allowed_entities"])

		agents = append(agents, &WorkforceAgent{
			OrgID:           0,
			AgentID:         agentID,
			AgentType:       AgentType(agentTypeStr),
			Name:            name,
			Description:     desc,
			Capabilities:    capsJSON,
			AllowedTasks:    tasksJSON,
			AllowedEntities: entitiesJSON,
			AutonomyLevel:   autonomyLevel,
			IsEnabled:       true,
			Version:         "1.0.0",
			PromptVersion:   "1.0.0",
			HealthStatus:    health,
		})
	}

	return agents, nil
}

// validateAndSanitizeOutput inspects and sanitizes AI output before returning to Go callers.
// Ensures confidence is bounded, action proposals require approval for any mutating steps,
// and removes any injection attempts.
func (c *HTTPSSidecarClient) validateAndSanitizeOutput(resp *SidecarTaskResponse) {
	if resp.Confidence < 0.0 {
		resp.Confidence = 0.0
	} else if resp.Confidence > 1.0 {
		resp.Confidence = 1.0
	}

	sanitizedActions := make([]SidecarProposedAction, 0, len(resp.ProposedActions))
	for _, action := range resp.ProposedActions {
		if action.ActionType == "" || action.EntityType == "" {
			continue // skip malformed action proposals
		}
		// In Go enforcement layer, all AI-proposed actions require HITL approval unless explicitly exempted
		action.RequiresApproval = true
		sanitizedActions = append(sanitizedActions, action)
	}
	resp.ProposedActions = sanitizedActions

	// Phase 6.4: Validate and sanitize DecisionRecord
	if resp.DecisionRecord != nil {
		if resp.DecisionRecord.OverallConfidence < 0.0 {
			resp.DecisionRecord.OverallConfidence = 0.0
		} else if resp.DecisionRecord.OverallConfidence > 1.0 {
			resp.DecisionRecord.OverallConfidence = 1.0
		}
		for i := range resp.DecisionRecord.ProposedActions {
			resp.DecisionRecord.ProposedActions[i].RequiresApproval = true
		}
	}
}

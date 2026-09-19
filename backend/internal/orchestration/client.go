package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type PythonProposalResponse struct {
	Status            string          `json:"status"`
	Proposal          *ActionProposal `json:"proposal"`
	EvaluationSummary string          `json:"evaluation_summary"`
	RejectionReason   *string         `json:"rejection_reason"`
	DurationMs        int             `json:"duration_ms"`
}

type SidecarClient interface {
	ProposeAction(ctx context.Context, payload map[string]interface{}) (*PythonProposalResponse, error)
}

type sidecarClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewSidecarClient() SidecarClient {
	baseURL := os.Getenv("AI_SIDECAR_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	token := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if token == "" {
		token = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}

	return &sidecarClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (c *sidecarClient) ProposeAction(ctx context.Context, payload map[string]interface{}) (*PythonProposalResponse, error) {
	url := fmt.Sprintf("%s/orchestrator/propose-action", c.baseURL)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize orchestrator context: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar connection failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sidecar response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var pyResp PythonProposalResponse
	if err := json.Unmarshal(respBody, &pyResp); err != nil {
		return nil, fmt.Errorf("invalid sidecar JSON response: %w (raw: %s)", err, string(respBody))
	}

	return &pyResp, nil
}

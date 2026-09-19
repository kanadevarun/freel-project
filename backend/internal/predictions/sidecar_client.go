package predictions

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

type SidecarClient interface {
	GeneratePrediction(ctx context.Context, req *GenerateRequest) (*SidecarPredictionResponse, error)
}

type httpSidecarClient struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

func NewSidecarClient(baseURL, serviceKey string) SidecarClient {
	if baseURL == "" {
		baseURL = os.Getenv("PYTHON_SIDECAR_URL")
		if baseURL == "" {
			baseURL = "http://127.0.0.1:8090"
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if serviceKey == "" {
		serviceKey = os.Getenv("INTERNAL_SERVICE_TOKEN")
		if serviceKey == "" {
			serviceKey = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
		}
	}

	return &httpSidecarClient{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *httpSidecarClient) GeneratePrediction(ctx context.Context, req *GenerateRequest) (*SidecarPredictionResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prediction request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/predictions/generate", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", c.serviceKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to python sidecar: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sidecar response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python sidecar returned error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var predResp SidecarPredictionResponse
	if err := json.Unmarshal(bodyBytes, &predResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sidecar prediction response: %w", err)
	}

	return &predResp, nil
}

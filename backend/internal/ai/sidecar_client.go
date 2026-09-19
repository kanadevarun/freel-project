package ai

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

const (
	DefaultSidecarURL     = "http://127.0.0.1:8090"
	DefaultServiceKey     = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	DefaultSidecarTimeout = 25 * time.Second
)

// ScoreLeadRequest defines the contract for lead scoring sent to the Python AI sidecar.
type ScoreLeadRequest struct {
	LeadID                int32   `json:"lead_id"`
	OrgID                 int32   `json:"org_id"`
	CompanyName           string  `json:"company_name"`
	ContactEmail          *string `json:"contact_email,omitempty"`
	ContactName           *string `json:"contact_name,omitempty"`
	Origin                *string `json:"origin,omitempty"`
	Destination           *string `json:"destination,omitempty"`
	EstimatedVolume       *string `json:"estimated_volume,omitempty"`
	Industry              *string `json:"industry,omitempty"`
	Location              *string `json:"location,omitempty"`
	EmployeeCount         *int    `json:"employee_count,omitempty"`
	EstimatedRevenue      *string `json:"estimated_revenue,omitempty"`
	MonthlyShippingVolume *string `json:"monthly_shipping_volume,omitempty"`
	TopSuppliers          *string `json:"top_suppliers,omitempty"`
	IsExporter            *bool   `json:"is_exporter,omitempty"`
	Notes                 *string `json:"notes,omitempty"`
	CorrelationID         string  `json:"correlation_id"`
}

// ScoreLeadResponse defines the structured response received from the Python AI sidecar.
type ScoreLeadResponse struct {
	Score           int32    `json:"score"`
	ResearchReport  string   `json:"research_report"`
	Confidence      float64  `json:"confidence"`
	RecommendedTier string   `json:"recommended_tier"`
	KeyStrengths    []string `json:"key_strengths"`
	RiskFactors     []string `json:"risk_factors"`
	CorrelationID   string   `json:"correlation_id"`
}

// ClassifyEmailRequest defines the contract for email relevance & intent classification.
type ClassifyEmailRequest struct {
	InteractionID *int64 `json:"interaction_id,omitempty"`
	OrgID         int32  `json:"org_id"`
	UserID        *int64 `json:"user_id,omitempty"`
	FromEmail     string `json:"from_email"`
	Subject       string `json:"subject"`
	Body          string `json:"body"`
	CorrelationID string `json:"correlation_id"`
}

// ClassifyEmailResponse defines the structured classification returned by Python.
type ClassifyEmailResponse struct {
	IsLogisticsRelated bool                   `json:"is_logistics_related"`
	Intent             string                 `json:"intent"`
	Sentiment          string                 `json:"sentiment"`
	Reasoning          string                 `json:"reasoning"`
	Confidence         float64                `json:"confidence"`
	ExtractedEntities  map[string]interface{} `json:"extracted_entities,omitempty"`
	CorrelationID      string                 `json:"correlation_id"`
}

// ParseShipmentRequestContract defines the contract for parsing unstructured RFQs.
type ParseShipmentRequestContract struct {
	OrgID         int64  `json:"org_id"`
	UserID        *int64 `json:"user_id,omitempty"`
	Text          string `json:"text"`
	CorrelationID string `json:"correlation_id"`
}

// ExtractedShipmentDataContract represents the extracted shipment fields.
type ExtractedShipmentDataContract struct {
	Origin      *string `json:"origin"`
	Destination *string `json:"destination"`
	Incoterms   *string `json:"incoterms"`
	Weight      *string `json:"weight"`
	Volume      *string `json:"volume"`
}

// ParseShipmentResponseContract defines the response from Python RFQ parsing.
type ParseShipmentResponseContract struct {
	Data            ExtractedShipmentDataContract `json:"data"`
	ConfidenceScore int                           `json:"confidence_score"`
	MissingFields   []string                      `json:"missing_fields"`
	CorrelationID   string                        `json:"correlation_id"`
}

// CompletionRequest defines a general LLM completion request to the Python sidecar.
type CompletionRequest struct {
	Prompt        string  `json:"prompt"`
	JSONMode      bool    `json:"json_mode,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
	MaxTokens     int     `json:"max_tokens,omitempty"`
	CorrelationID string  `json:"correlation_id"`
	OrgID         int64   `json:"org_id,omitempty"`
}

// CompletionResponse defines a general completion response from Python.
type CompletionResponse struct {
	Content       string  `json:"content"`
	Model         string  `json:"model"`
	Confidence    float64 `json:"confidence"`
	CorrelationID string  `json:"correlation_id"`
	Status        string  `json:"status"`
}

// SidecarClient defines the interface for interacting with the Python AI sidecar.
type SidecarClient interface {
	ScoreLead(ctx context.Context, req *ScoreLeadRequest) (*ScoreLeadResponse, error)
	ClassifyEmail(ctx context.Context, req *ClassifyEmailRequest) (*ClassifyEmailResponse, error)
	ParseShipmentRequest(ctx context.Context, req *ParseShipmentRequestContract) (*ParseShipmentResponseContract, error)
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

type sidecarClientImpl struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

// NewSidecarClient instantiates a client communicating with the Python AI Sidecar.
func NewSidecarClient(baseURL, serviceKey string) SidecarClient {
	if baseURL == "" {
		baseURL = os.Getenv("AI_SIDECAR_URL")
		if baseURL == "" {
			baseURL = DefaultSidecarURL
		}
	}
	if serviceKey == "" {
		serviceKey = os.Getenv("INTERNAL_SERVICE_TOKEN")
		if serviceKey == "" {
			serviceKey = os.Getenv("INTERNAL_SERVICE_KEY")
		}
		if serviceKey == "" {
			serviceKey = os.Getenv("AI_SIDECAR_SERVICE_KEY")
		}
		if serviceKey == "" {
			rawEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
			if rawEnv == "" {
				rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
			}
			if rawEnv == "" {
				rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
			}
			if rawEnv == "development" || rawEnv == "test" {
				serviceKey = DefaultServiceKey
			}
		}
	}

	return &sidecarClientImpl{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{
			Timeout: DefaultSidecarTimeout,
		},
	}
}

func (c *sidecarClientImpl) post(ctx context.Context, endpoint string, reqPayload interface{}, respDest interface{}) error {
	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", c.serviceKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sidecar connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sidecar returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(respDest); err != nil {
		return fmt.Errorf("failed to decode sidecar response: %w", err)
	}

	return nil
}

func (c *sidecarClientImpl) ScoreLead(ctx context.Context, req *ScoreLeadRequest) (*ScoreLeadResponse, error) {
	var resp ScoreLeadResponse
	if err := c.post(ctx, "/leads/score-lead", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *sidecarClientImpl) ClassifyEmail(ctx context.Context, req *ClassifyEmailRequest) (*ClassifyEmailResponse, error) {
	var resp ClassifyEmailResponse
	if err := c.post(ctx, "/leads/classify-email", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *sidecarClientImpl) ParseShipmentRequest(ctx context.Context, req *ParseShipmentRequestContract) (*ParseShipmentResponseContract, error) {
	var resp ParseShipmentResponseContract
	if err := c.post(ctx, "/rfq/parse-shipment-request", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *sidecarClientImpl) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	var resp CompletionResponse
	if err := c.post(ctx, "/ai/completion", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SidecarProvider adapts SidecarClient to the legacy Provider interface.
type SidecarProvider struct {
	client SidecarClient
}

// NewSidecarProvider creates a Provider implementation backed by the Python AI sidecar.
func NewSidecarProvider(client SidecarClient) Provider {
	return &SidecarProvider{client: client}
}

func (p *SidecarProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := p.client.Complete(ctx, &CompletionRequest{
		Prompt:        prompt,
		CorrelationID: fmt.Sprintf("sidecar-comp-%d", time.Now().UnixNano()),
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

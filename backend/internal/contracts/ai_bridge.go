package contracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AIBridge defines the interface for communicating with the external Python AI processing sidecar.
type AIBridge interface {
	// TriggerProcessing dispatches a request to the Python sidecar to start parsing the document asynchronously.
	TriggerProcessing(ctx context.Context, req ProcessingRequest) error
	// TriggerResumption dispatches a resume instruction to the sidecar to continue a paused LangGraph execution thread.
	TriggerResumption(ctx context.Context, req ResumptionRequest) error
}

// ProcessingRequest defines the JSON payload shape for triggering the AI sidecar document pipeline.
type ProcessingRequest struct {
	// DocumentID identifies the unique document record in contract_documents.
	DocumentID  string `json:"document_id"`
	// OrgID is the multi-tenant identifier of the user's organization.
	OrgID       int64  `json:"org_id"`
	// S3Key is the filename reference in the upload service storage.
	S3Key       string `json:"s3_key"`
	// FileType indicates the format (e.g. PDF, XLSX).
	FileType    string `json:"file_type"`
	// CallbackURL is the REST path the sidecar POSTs to upon completing layout analysis.
	CallbackURL string `json:"callback_url"`
}

// ResumptionRequest defines the JSON payload shape for resuming a paused LangGraph workflow.
type ResumptionRequest struct {
	// DocumentID identifies the unique document record in contract_documents.
	DocumentID     string      `json:"document_id"`
	// OrgID is the multi-tenant identifier of the user's organization.
	OrgID          int64       `json:"org_id"`
	// Action indicates the operator decision ("APPROVE" or "REJECT").
	Action         string      `json:"action"`
	// CorrectedRates contains optional manually corrected port rates.
	CorrectedRates interface{} `json:"corrected_rates,omitempty"`
	// Notes holds operator annotations/feedback.
	Notes          string      `json:"notes,omitempty"`
	// CallbackURL is the callback webhook target.
	CallbackURL    string      `json:"callback_url"`
}

// aiBridge implements the AIBridge interface using a standard Go HTTP client.
type aiBridge struct {
	// sidecarURL is the target server address (e.g., http://localhost:8090).
	sidecarURL string
	// client is the reusable HTTP client.
	client     *http.Client
}

// NewAIBridge initializes and returns a concrete instance of AIBridge.
// If the sidecarURL parameter is empty, it defaults to localhost:8090.
func NewAIBridge(sidecarURL string) AIBridge {
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8090"
	}
	return &aiBridge{
		sidecarURL: sidecarURL,
		client: &http.Client{
			// Prevent hanging HTTP calls by setting a strict connection timeout.
			Timeout: 10 * time.Second,
		},
	}
}

// TriggerProcessing sends an HTTP POST request to the Python sidecar.
// The sidecar queues the request internally and processes it asynchronously.
func (b *aiBridge) TriggerProcessing(ctx context.Context, req ProcessingRequest) error {
	// Serialize the request payload to JSON.
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Build the target endpoint URL.
	url := fmt.Sprintf("%s/process", b.sidecarURL)
	
	// Create an outgoing POST request associated with the incoming context.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create HTTP request: %w", err)
	}
	// Declare the payload encoding type.
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the HTTP round-trip.
	resp, err := b.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	// Safely close the response stream to free resources.
	defer resp.Body.Close()

	// Check if the sidecar accepted the payload.
	// Sidecar responds with 202 Accepted for background execution.
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sidecar returned non-success code: %d", resp.StatusCode)
	}

	return nil
}

// TriggerResumption sends an HTTP POST request to the Python AI sidecar's '/resume' endpoint.
//
// Simple meaning:
//   When a human operator approves or corrects a flagged rate discrepancy in the React UI,
//   the Go backend calls this function. It sends a message to the Python sidecar telling it:
//   "Hey, the human resolved this flagged item. Here are the approved rates (with corrections if any).
//    Please wake up your paused LangGraph thread and finish processing the document."
//
// Example payload structure sent to the Python sidecar:
//   {
//     "document_id": "3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3",
//     "org_id": 5,
//     "action": "APPROVE",
//     "corrected_rates": [ { "origin_port": "INNSA", "destination_port": "DEHAM", "ocean_freight": 2800.0, "total_buy_price": 3200.0 } ],
//     "notes": "Corrected ocean freight rate",
//     "callback_url": "http://localhost:8080/internal/contracts/callback"
//   }
func (b *aiBridge) TriggerResumption(ctx context.Context, req ResumptionRequest) error {
	// Serialize the Go struct payload into raw JSON bytes.
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// target resumption route url: e.g. http://localhost:8090/resume
	url := fmt.Sprintf("%s/resume", b.sidecarURL)
	
	// Construct the outgoing POST request with context tracking.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Dispatch request to the Python server.
	resp, err := b.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Sidecar responds with 202 Accepted, indicating the thread is resumed in the background.
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sidecar returned non-success code: %d", resp.StatusCode)
	}

	return nil
}




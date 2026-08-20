package ai

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

// geminiProvider communicates with the Google Gemini API using raw HTTP REST requests.
// This bypasses the heavy Vertex/Google cloud SDK, keeping compilation and dev iterations lightning fast.
type geminiProvider struct {
	apiKey string
	client *http.Client
}

// NewGeminiProvider creates a new instance of the Gemini AI provider.
func NewGeminiProvider(apiKey string) Provider {
	return &geminiProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request structure for Gemini generateContent endpoint
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// Response structure for Gemini generateContent endpoint
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *geminiProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// ── STEP 1: CONSTRUCT URL ─────────────────────────────────────────────────
	// We build the HTTP API endpoint URL, appending your Google AI Studio API Key.
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-3.1-flash-lite"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", geminiModel, g.apiKey)



	// ── STEP 2: BUILD JSON REQUEST BODY ───────────────────────────────────────
	// Format the prompt text into the nested structure Gemini API expects.
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// ── STEP 3: EXECUTE HTTP POST REQUEST ─────────────────────────────────────
	// Send the JSON payload to the Gemini API server using a POST request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http execute call: %w", err)
	}
	defer resp.Body.Close()

	// ── STEP 4: CHECK RESPONSE STATUS ─────────────────────────────────────────
	// If the server did not return HTTP 200 (Success), read the error description.
	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	// ── STEP 5: DECODE GENERATED TEXT ─────────────────────────────────────────
	// Parse the successful response body, extraction candidate text, and return it.
	var respData geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(respData.Candidates) == 0 || len(respData.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini response returned empty candidates")
	}

	return respData.Candidates[0].Content.Parts[0].Text, nil
}

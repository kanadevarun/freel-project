package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// openAIProvider communicates with the OpenAI ChatGPT API using standard HTTP REST requests.
type openAIProvider struct {
	apiKey string
	client *http.Client
}

// NewOpenAIProvider creates a new instance of the OpenAI ChatGPT provider.
func NewOpenAIProvider(apiKey string) Provider {
	return &openAIProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request structure for OpenAI chat completions endpoint
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response structure for OpenAI chat completions endpoint
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (o *openAIProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// ── STEP 1: CONSTRUCT URL ─────────────────────────────────────────────────
	// We build the HTTP API endpoint URL for OpenAI's Chat Completions.
	url := "https://api.openai.com/v1/chat/completions"

	// ── STEP 2: BUILD JSON REQUEST BODY ───────────────────────────────────────
	// Format the prompt text into the structure OpenAI's chat model expects.
	// We default to the state-of-the-art "gpt-4o" model.
	reqBody := openAIRequest{
		Model: "gpt-4o",
		Messages: []openAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// ── STEP 3: EXECUTE HTTP POST REQUEST ─────────────────────────────────────
	// Send the JSON payload to the OpenAI server. We pass the API key in the
	// "Authorization" header as a "Bearer" token.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http execute call: %w", err)
	}
	defer resp.Body.Close()

	// ── STEP 4: CHECK RESPONSE STATUS ─────────────────────────────────────────
	// If the server did not return HTTP 200 (Success), read the error description.
	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	// ── STEP 5: DECODE GENERATED TEXT ─────────────────────────────────────────
	// Parse the successful response body, extract choice message text, and return it.
	var respData openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("openai response returned zero choices")
	}

	return respData.Choices[0].Message.Content, nil
}

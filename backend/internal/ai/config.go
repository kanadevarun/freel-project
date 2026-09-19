package ai

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrInvalidConfig       = errors.New("invalid ai runtime configuration")
	ErrProductionMockMode  = errors.New("production environment prohibits mock mode or unverified fallback")
	ErrMissingProviderKey  = errors.New("required ai provider api key is missing")
	ErrUnsupportedModel    = errors.New("unsupported or unverified model identifier")
)

// Supported and verified models
var VerifiedModels = map[string]bool{
	"gemini-1.5-flash":    true,
	"gemini-1.5-pro":      true,
	"gemini-2.0-flash":    true,
	"gemini-3.1-flash-lite": true, // Accepted alias in current Google SDK
	"gpt-4o":              true,
	"gpt-4o-mini":         true,
	"gpt-4-turbo":         true,
	"mock-evaluator":      true,
}

// RuntimeConfig defines the authoritative single source of truth for AI execution.
type RuntimeConfig struct {
	Environment            string  `json:"environment"`
	PrimaryProvider        string  `json:"primary_provider"`
	PrimaryModel           string  `json:"primary_model"`
	FailoverProvider       string  `json:"failover_provider"`
	FailoverModel          string  `json:"failover_model"`
	AllowMockFallback      bool    `json:"allow_mock_fallback"`
	MaxRetries             int     `json:"max_retries"`
	ProviderTimeoutSeconds int     `json:"provider_timeout_seconds"`
	ToolCallTimeoutSeconds int     `json:"tool_call_timeout_seconds"`
	MaxOutputTokens        int     `json:"max_output_tokens"`
	Temperature            float64 `json:"temperature"`
	StreamingEnabled       bool    `json:"streaming_enabled"`
	DefaultPromptVersion   string  `json:"default_prompt_version"`
}

// LoadRuntimeConfigFromEnv loads configuration from environment variables with safe defaults.
func LoadRuntimeConfigFromEnv() (*RuntimeConfig, error) {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = "development"
	}

	primaryProv := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PRIMARY_PROVIDER")))
	if primaryProv == "" {
		primaryProv = "gemini"
	}

	primaryModel := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if primaryModel == "" {
		primaryModel = "gemini-1.5-flash"
	}

	failoverProv := strings.ToLower(strings.TrimSpace(os.Getenv("AI_FAILOVER_PROVIDER")))
	if failoverProv == "" {
		failoverProv = "openai"
	}

	failoverModel := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if failoverModel == "" {
		failoverModel = "gpt-4o-mini"
	}

	// Mock mode policy: permitted by default ONLY in non-production environments
	mockAllowed := false
	if env != "production" {
		mockAllowed = true
	}
	if rawMock := os.Getenv("ALLOW_MOCK_FALLBACK"); rawMock != "" {
		mockAllowed, _ = strconv.ParseBool(rawMock)
	}

	maxRetries := 2
	if mr := os.Getenv("AI_MAX_RETRIES"); mr != "" {
		if val, err := strconv.Atoi(mr); err == nil && val >= 0 {
			maxRetries = val
		}
	}

	timeoutSec := 30
	if to := os.Getenv("AI_PROVIDER_TIMEOUT_SECONDS"); to != "" {
		if val, err := strconv.Atoi(to); err == nil && val > 0 {
			timeoutSec = val
		}
	}

	temp := 0.1
	if t := os.Getenv("AI_TEMPERATURE"); t != "" {
		if val, err := strconv.ParseFloat(t, 64); err == nil && val >= 0 && val <= 2.0 {
			temp = val
		}
	}

	streaming := false
	if str := os.Getenv("AI_STREAMING_ENABLED"); str != "" {
		streaming, _ = strconv.ParseBool(str)
	}

	promptVer := os.Getenv("AI_DEFAULT_PROMPT_VERSION")
	if promptVer == "" {
		promptVer = "1.0.0"
	}

	cfg := &RuntimeConfig{
		Environment:            env,
		PrimaryProvider:        primaryProv,
		PrimaryModel:           primaryModel,
		FailoverProvider:       failoverProv,
		FailoverModel:          failoverModel,
		AllowMockFallback:      mockAllowed,
		MaxRetries:             maxRetries,
		ProviderTimeoutSeconds: timeoutSec,
		ToolCallTimeoutSeconds: timeoutSec,
		MaxOutputTokens:        2048,
		Temperature:            temp,
		StreamingEnabled:       streaming,
		DefaultPromptVersion:   promptVer,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate ensures configuration complies with environment rules and security standards.
func (c *RuntimeConfig) Validate() error {
	if c.Environment == "production" {
		if c.AllowMockFallback {
			return fmt.Errorf("%w: allow_mock_fallback cannot be enabled in production", ErrProductionMockMode)
		}
		// In production, primary provider key is mandatory
		if c.PrimaryProvider == "gemini" && os.Getenv("GOOGLE_API_KEY") == "" && os.Getenv("GEMINI_API_KEY") == "" {
			return fmt.Errorf("%w: primary provider 'gemini' requires GOOGLE_API_KEY in production", ErrMissingProviderKey)
		}
		if c.PrimaryProvider == "openai" && os.Getenv("OPENAI_API_KEY") == "" {
			return fmt.Errorf("%w: primary provider 'openai' requires OPENAI_API_KEY in production", ErrMissingProviderKey)
		}
	}

	// Verify model identifiers
	if c.PrimaryModel != "" && !VerifiedModels[c.PrimaryModel] {
		return fmt.Errorf("%w: primary model '%s' is not in verified model registry", ErrUnsupportedModel, c.PrimaryModel)
	}
	if c.FailoverModel != "" && !VerifiedModels[c.FailoverModel] {
		return fmt.Errorf("%w: failover model '%s' is not in verified model registry", ErrUnsupportedModel, c.FailoverModel)
	}

	return nil
}

// SanitizeForLogging returns a copy of the config suitable for safe logging without credentials.
func (c *RuntimeConfig) SanitizeForLogging() map[string]interface{} {
	return map[string]interface{}{
		"environment":               c.Environment,
		"primary_provider":          c.PrimaryProvider,
		"primary_model":             c.PrimaryModel,
		"failover_provider":         c.FailoverProvider,
		"failover_model":            c.FailoverModel,
		"allow_mock_fallback":       c.AllowMockFallback,
		"max_retries":               c.MaxRetries,
		"provider_timeout_seconds":  c.ProviderTimeoutSeconds,
		"temperature":               c.Temperature,
		"streaming_enabled":         c.StreamingEnabled,
		"default_prompt_version":    c.DefaultPromptVersion,
	}
}

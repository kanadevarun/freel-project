package token

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/carrier/domain"
)

// CachedToken stores access token data in memory.
type CachedToken struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}

// TokenManager provides thread-safe OAuth2 client-credentials token caching.
type TokenManager struct {
	mu     sync.RWMutex
	tokens map[string]*CachedToken
	client *http.Client
}

var (
	defaultTokenManager *TokenManager
	onceTokenMgr        sync.Once
)

// GetDefaultTokenManager returns the singleton token manager.
func GetDefaultTokenManager() *TokenManager {
	onceTokenMgr.Do(func() {
		defaultTokenManager = &TokenManager{
			tokens: make(map[string]*CachedToken),
			client: &http.Client{Timeout: 15 * time.Second},
		}
	})
	return defaultTokenManager
}

// TokenRequest contains OAuth2 client credentials parameters.
type TokenRequest struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	ProviderCode string
	Environment  domain.Environment
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// GetToken returns an active cached token or negotiates a fresh OAuth2 bearer token.
func (m *TokenManager) GetToken(ctx context.Context, req TokenRequest) (string, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s", req.ProviderCode, req.ClientID, req.Environment)

	// Check existing cached token
	m.mu.RLock()
	cached, exists := m.tokens[cacheKey]
	if exists && cached != nil && time.Now().Add(60*time.Second).Before(cached.ExpiresAt) {
		m.mu.RUnlock()
		return cached.AccessToken, nil
	}
	m.mu.RUnlock()

	// Acquire lock to request new token
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check cache under write lock
	if cached, exists := m.tokens[cacheKey]; exists && cached != nil && time.Now().Add(60*time.Second).Before(cached.ExpiresAt) {
		return cached.AccessToken, nil
	}

	if req.TokenURL == "" {
		return "", domain.NewIntegrationError(
			req.ProviderCode,
			"OAUTH_TOKEN",
			domain.ErrCodeInvalidConfig,
			"OAuth2 token endpoint URL is not configured",
			http.StatusBadRequest,
			false,
			fmt.Errorf("missing token url"),
		)
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", req.ClientID)
	form.Set("client_secret", req.ClientSecret)
	if req.Scope != "" {
		form.Set("scope", req.Scope)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", domain.NewIntegrationError(
			req.ProviderCode,
			"OAUTH_TOKEN",
			domain.ErrCodeInvalidRequest,
			"Failed to construct OAuth token request",
			http.StatusBadRequest,
			false,
			err,
		)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return "", domain.NewIntegrationError(
			req.ProviderCode,
			"OAUTH_TOKEN",
			domain.ErrCodeUnavailable,
			"Failed to reach carrier OAuth2 authorization server",
			http.StatusGatewayTimeout,
			true,
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", domain.NewIntegrationError(
			req.ProviderCode,
			"OAUTH_TOKEN",
			domain.ErrCodeAuthFailed,
			"Carrier OAuth2 client authentication rejected. Verify Client ID and Secret.",
			resp.StatusCode,
			false,
			fmt.Errorf("oauth failed with status %d", resp.StatusCode),
		)
	}

	var tokenResp oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", domain.NewIntegrationError(
			req.ProviderCode,
			"OAUTH_TOKEN",
			domain.ErrCodeInternal,
			"Invalid JSON response from carrier OAuth2 server",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600 // Default 1 hour
	}

	m.tokens[cacheKey] = &CachedToken{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	return tokenResp.AccessToken, nil
}

// Invalidate removes cached tokens for a specific client key.
func (m *TokenManager) Invalidate(providerCode, clientID string, env domain.Environment) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cacheKey := fmt.Sprintf("%s:%s:%s", providerCode, clientID, env)
	delete(m.tokens, cacheKey)
}

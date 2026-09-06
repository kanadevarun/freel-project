package organization

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/config"
)

// SyncEmail represents normalized email details for leads ingestion.
type SyncEmail struct {
	RawEmailID       string
	RFCMessageID     string
	ThreadID         string
	From             string
	To               string
	Subject          string
	Body             string
	MessageID        string // Keep for compatibility
	InReplyTo        string
	ReferencesHeader string
	Sender           string
	Recipients       string
	CCRecipients     string
	ReceivedAt       time.Time
}

// SyncEmailHandler is the callback signature for processing synchronized emails.
type SyncEmailHandler func(ctx context.Context, email SyncEmail) error

// TokenUpdateFunc is the callback signature for updating persisted encrypted tokens.
type TokenUpdateFunc func(ctx context.Context, accessToken string, expiry time.Time) error

// MailboxProvider defines the interface that all email sync providers must implement.
type MailboxProvider interface {
	SyncMailbox(ctx context.Context, mailbox *ConnectedMailbox, handler SyncEmailHandler, onTokenUpdate TokenUpdateFunc) error
}

// OutboundMailProvider defines the interface for outbound email sending.
type OutboundMailProvider interface {
	SendEmail(ctx context.Context, mailbox *ConnectedMailbox, to string, subject string, body string, threadID string, inReplyTo string, references string, onTokenUpdate TokenUpdateFunc) (string, string, string, error)
}

// GmailProvider implements MailboxProvider for Gmail accounts.
type GmailProvider struct {
	HTTPClient *http.Client
}

func NewGmailProvider() *GmailProvider {
	return &GmailProvider{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type GmailMessagePart struct {
	MimeType string `json:"mimeType"`
	Body     struct {
		Data string `json:"data"`
		Size int    `json:"size"`
	} `json:"body"`
	Parts []GmailMessagePart `json:"parts"`
}

type GmailMessagePayload struct {
	Headers []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"headers"`
	MimeType string             `json:"mimeType"`
	Body     struct {
		Data string `json:"data"`
	} `json:"body"`
	Parts []GmailMessagePart `json:"parts"`
}

type GmailMessage struct {
	Id           string               `json:"id"`
	ThreadId     string               `json:"threadId"`
	Snippet      string               `json:"snippet"`
	InternalDate string               `json:"internalDate"` // Unix timestamp in ms
	Payload      GmailMessagePayload  `json:"payload"`
}

func (p *GmailProvider) SyncMailbox(ctx context.Context, mailbox *ConnectedMailbox, handler SyncEmailHandler, onTokenUpdate TokenUpdateFunc) error {
	encKey := os.Getenv(config.EnvMailboxEncryptionKey)
	if encKey == "" {
		return errors.New("MAILBOX_ENCRYPTION_KEY is not configured")
	}

	clientID := os.Getenv(config.EnvGoogleClientID)
	clientSecret := os.Getenv(config.EnvGoogleClientSecret)
	if clientID == "" || clientSecret == "" {
		return errors.New("Google client credentials are not configured")
	}

	if mailbox.AccessTokenEncrypted == nil || mailbox.RefreshTokenEncrypted == nil {
		return errors.New("mailbox credentials are not configured")
	}

	accessToken, err := crypto.Decrypt(*mailbox.AccessTokenEncrypted, encKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken, err := crypto.Decrypt(*mailbox.RefreshTokenEncrypted, encKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	expiry := time.Time{}
	if mailbox.TokenExpiry != nil {
		expiry = *mailbox.TokenExpiry
	}

	// Check if token needs refreshing
	if expiry.IsZero() || time.Now().Add(5*time.Minute).After(expiry) {
		log.Printf("[Gmail Sync Provider] Refreshing access token for mailbox: %s", mailbox.Email)
		newAccess, newExpiry, err := p.refreshAccessToken(ctx, clientID, clientSecret, refreshToken)
		if err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}
		accessToken = newAccess
		expiry = newExpiry

		// Persist refreshed credentials encrypted
		if err := onTokenUpdate(ctx, accessToken, expiry); err != nil {
			log.Printf("[Gmail Sync Provider] Warning: failed to persist refreshed token in DB: %v", err)
		}
	}

	// Retrieve new inbound emails since last sync cursor
	cursor := ""
	if mailbox.SyncCursor != nil {
		cursor = *mailbox.SyncCursor
	}

	log.Printf("[Gmail Sync Provider] Fetching messages list for mailbox %s (Cursor: %s)...", mailbox.Email, cursor)
	msgIDs, err := p.listMessages(ctx, accessToken, cursor)
	if err != nil {
		return fmt.Errorf("list messages failed: %w", err)
	}

	log.Printf("[Gmail Sync Provider] Found %d new messages to process for mailbox: %s", len(msgIDs), mailbox.Email)

	var newCursor string
	// Process from oldest to newest (sync list returns newest first)
	for i := len(msgIDs) - 1; i >= 0; i-- {
		msgID := msgIDs[i]

		log.Printf("[Gmail Sync Provider] Fetching message details for ID: %s", msgID)
		msgDetail, err := p.getMessageDetail(ctx, accessToken, msgID)
		if err != nil {
			log.Printf("[Gmail Sync Provider] Failed to fetch message detail for ID %s: %v", msgID, err)
			continue
		}

		// Normalize to SyncEmail format
		email, err := p.normalizeMessage(msgDetail)
		if err != nil {
			log.Printf("[Gmail Sync Provider] Failed to normalize message ID %s: %v", msgID, err)
			continue
		}

		// Skip if sent by the mailbox owner itself
		if strings.EqualFold(strings.TrimSpace(email.From), strings.TrimSpace(mailbox.Email)) {
			log.Printf("[Gmail Sync Provider] Skipping message ID %s (sent by self)", msgID)
			newCursor = msgID
			continue
		}

		// Process inbound email using shared ingestion method
		if err := handler(ctx, email); err != nil {
			log.Printf("[Gmail Sync Provider] Handler failed processing message ID %s: %v", msgID, err)
			continue
		}

		newCursor = msgID
	}

	if newCursor != "" {
		mailbox.SyncCursor = &newCursor
	}

	return nil
}

func (p *GmailProvider) refreshAccessToken(ctx context.Context, clientID, clientSecret, refreshToken string) (string, time.Time, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", time.Time{}, fmt.Errorf("token refresh status %d: %s", res.StatusCode, string(body))
	}

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return "", time.Time{}, err
	}

	expiry := time.Now().Add(time.Duration(tokenRes.ExpiresIn) * time.Second)
	return tokenRes.AccessToken, expiry, nil
}

func (p *GmailProvider) listMessages(ctx context.Context, accessToken string, cursor string) ([]string, error) {
	// Retrieve messages only inside INBOX label (meets item 3: Inbound emails only)
	apiURL := getGmailAPIBase() + "/gmail/v1/users/me/messages?q=label:INBOX&maxResults=50"
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("list messages status %d: %s", res.StatusCode, string(body))
	}

	var listRes struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(res.Body).Decode(&listRes); err != nil {
		return nil, err
	}

	var results []string
	for _, m := range listRes.Messages {
		if m.ID == cursor {
			// Found the latest message processed previously, stop incremental list
			break
		}
		results = append(results, m.ID)
	}

	return results, nil
}

func (p *GmailProvider) getMessageDetail(ctx context.Context, accessToken string, messageID string) (*GmailMessage, error) {
	apiURL := fmt.Sprintf("%s/gmail/v1/users/me/messages/%s?format=full", getGmailAPIBase(), messageID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("get message status %d: %s", res.StatusCode, string(body))
	}

	var msg GmailMessage
	if err := json.NewDecoder(res.Body).Decode(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (p *GmailProvider) normalizeMessage(g *GmailMessage) (SyncEmail, error) {
	var email SyncEmail
	email.RawEmailID = g.Id
	email.MessageID = g.Id // fallback compatibility
	email.ThreadID = g.ThreadId

	// Parse headers
	for _, h := range g.Payload.Headers {
		name := strings.ToLower(h.Name)
		switch name {
		case "subject":
			email.Subject = h.Value
		case "from":
			// Extract plain email from From header (e.g. "Varun Kanade <varun@example.com>")
			email.From = extractEmailAddress(h.Value)
			email.Sender = h.Value
		case "to":
			email.To = h.Value
			email.Recipients = h.Value
		case "cc":
			email.CCRecipients = h.Value
		case "message-id":
			email.RFCMessageID = h.Value
			email.MessageID = h.Value // fallback compatibility
		case "in-reply-to":
			email.InReplyTo = h.Value
		case "references":
			email.ReferencesHeader = h.Value
		}
	}

	// Parse date/timestamp
	if g.InternalDate != "" {
		ms, err := strconv.ParseInt(g.InternalDate, 10, 64)
		if err == nil {
			email.ReceivedAt = time.Unix(ms/1000, (ms%1000)*1000000)
		}
	}
	if email.ReceivedAt.IsZero() {
		email.ReceivedAt = time.Now()
	}

	// Extract body plain text and html
	plain, html := extractBodyText(g.Payload)
	if plain != "" {
		email.Body = plain
	} else if html != "" {
		// Convert HTML email to plain text (Item 3 requirement)
		email.Body = htmlToText(html)
	} else {
		email.Body = g.Snippet // fallback
	}

	return email, nil
}

func extractEmailAddress(s string) string {
	re := regexp.MustCompile(`<([^>]+)>`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(s)
}

func extractBodyText(part GmailMessagePayload) (string, string) {
	var plainText, htmlText string
	if part.MimeType == "text/plain" && part.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
		plainText = string(decoded)
	} else if part.MimeType == "text/html" && part.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
		htmlText = string(decoded)
	}

	for _, subPart := range part.Parts {
		subPlain, subHtml := extractBodyFromPart(subPart)
		if subPlain != "" {
			plainText += "\n" + subPlain
		}
		if subHtml != "" {
			htmlText += "\n" + subHtml
		}
	}
	return plainText, htmlText
}

func extractBodyFromPart(part GmailMessagePart) (string, string) {
	var plainText, htmlText string
	if part.MimeType == "text/plain" && part.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
		plainText = string(decoded)
	} else if part.MimeType == "text/html" && part.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
		htmlText = string(decoded)
	}

	for _, subPart := range part.Parts {
		subPlain, subHtml := extractBodyFromPart(subPart)
		if subPlain != "" {
			plainText += "\n" + subPlain
		}
		if subHtml != "" {
			htmlText += "\n" + subHtml
		}
	}
	return plainText, htmlText
}

func htmlToText(htmlContent string) string {
	// Remove style and script blocks
	reStyle := regexp.MustCompile(`(?s)<style.*?>.*?</style>`)
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")
	reScript := regexp.MustCompile(`(?s)<script.*?>.*?</script>`)
	htmlContent = reScript.ReplaceAllString(htmlContent, "")

	// Replace common block tags with newlines
	reBlocks := regexp.MustCompile(`(?i)</?(p|br|div|tr|h[1-6]|li).*?>`)
	htmlContent = reBlocks.ReplaceAllString(htmlContent, "\n")

	// Strip remaining tags
	reTags := regexp.MustCompile(`<.*?>`)
	plainText := reTags.ReplaceAllString(htmlContent, "")

	// Unescape HTML entities
	plainText = html.UnescapeString(plainText)

	// Clean up multiple newlines/whitespace
	reSpaces := regexp.MustCompile(`\n\s*\n`)
	plainText = reSpaces.ReplaceAllString(plainText, "\n\n")

	return strings.TrimSpace(plainText)
}

// MicrosoftProvider implements MailboxProvider for Outlook/Office365 accounts.
type MicrosoftProvider struct{}

func (p *MicrosoftProvider) SyncMailbox(ctx context.Context, mailbox *ConnectedMailbox, handler SyncEmailHandler, onTokenUpdate TokenUpdateFunc) error {
	return nil
}

// IMAPProvider implements MailboxProvider for generic IMAP accounts.
type IMAPProvider struct{}

func (p *IMAPProvider) SyncMailbox(ctx context.Context, mailbox *ConnectedMailbox, handler SyncEmailHandler, onTokenUpdate TokenUpdateFunc) error {
	return nil
}

// Signed State Helper Functions

func GenerateOAuthState(orgID int64) (string, error) {
	key := os.Getenv(config.EnvMailboxEncryptionKey)
	if key == "" {
		return "", fmt.Errorf("MAILBOX_ENCRYPTION_KEY not set")
	}
	timestamp := time.Now().Unix()
	msg := fmt.Sprintf("%d:%d", orgID, timestamp)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	sig := mac.Sum(nil)
	
	stateStr := fmt.Sprintf("%s:%s", msg, base64.RawURLEncoding.EncodeToString(sig))
	return base64.RawURLEncoding.EncodeToString([]byte(stateStr)), nil
}

func ValidateOAuthState(stateToken string) (int64, error) {
	key := os.Getenv(config.EnvMailboxEncryptionKey)
	if key == "" {
		return 0, fmt.Errorf("MAILBOX_ENCRYPTION_KEY not set")
	}
	
	decodedBytes, err := base64.RawURLEncoding.DecodeString(stateToken)
	if err != nil {
		return 0, fmt.Errorf("invalid state encoding")
	}
	
	parts := strings.Split(string(decodedBytes), ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid state format")
	}
	
	orgIDStr, timestampStr, sigStr := parts[0], parts[1], parts[2]
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid org ID in state")
	}
	
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp in state")
	}
	
	// Check expiry (15 minutes)
	if time.Now().Unix()-timestamp > 900 {
		return 0, fmt.Errorf("state token has expired")
	}
	
	// Verify signature
	msg := fmt.Sprintf("%d:%d", orgID, timestamp)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	expectedSig := mac.Sum(nil)
	expectedSigStr := base64.RawURLEncoding.EncodeToString(expectedSig)
	
	if !hmac.Equal([]byte(sigStr), []byte(expectedSigStr)) {
		return 0, fmt.Errorf("invalid state signature")
	}
	
	return orgID, nil
}

func (p *GmailProvider) SendEmail(ctx context.Context, mailbox *ConnectedMailbox, to string, subject string, body string, threadID string, inReplyTo string, references string, onTokenUpdate TokenUpdateFunc) (string, string, string, error) {
	encKey := os.Getenv(config.EnvMailboxEncryptionKey)
	if encKey == "" {
		return "", "", "", fmt.Errorf("MAILBOX_ENCRYPTION_KEY not set")
	}

	if mailbox.AccessTokenEncrypted == nil || mailbox.RefreshTokenEncrypted == nil {
		return "", "", "", fmt.Errorf("mailbox credentials not connected")
	}

	// 1. Decrypt token in memory
	accessToken, err := crypto.Decrypt(*mailbox.AccessTokenEncrypted, encKey)
	if err != nil {
		return "", "", "", fmt.Errorf("decrypt access token: %w", err)
	}

	refreshToken, err := crypto.Decrypt(*mailbox.RefreshTokenEncrypted, encKey)
	if err != nil {
		return "", "", "", fmt.Errorf("decrypt refresh token: %w", err)
	}

	clientID := os.Getenv(config.EnvGoogleClientID)
	clientSecret := os.Getenv(config.EnvGoogleClientSecret)
	if clientID == "" || clientSecret == "" {
		return "", "", "", fmt.Errorf("Google client credentials are not configured")
	}

	// 2. Check token expiry, refresh if expired
	if mailbox.TokenExpiry != nil && time.Now().After(mailbox.TokenExpiry.Add(-5*time.Minute)) {
		log.Printf("[Gmail Sync Provider] Access token expired or close to expiry. Refreshing...")
		newAccess, newExpiry, err := p.refreshAccessToken(ctx, clientID, clientSecret, refreshToken)
		if err != nil {
			return "", "", "", fmt.Errorf("refresh token failed: %w", err)
		}
		accessToken = newAccess
		mailbox.TokenExpiry = &newExpiry

		// Persist refreshed credentials encrypted
		if err := onTokenUpdate(ctx, accessToken, newExpiry); err != nil {
			log.Printf("[Gmail Sync Provider] Warning: failed to persist refreshed token: %v", err)
		}
	}

	// 3. Format "Re: " Subject
	emailSubject := subject
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(emailSubject)), "re:") {
		emailSubject = "Re: " + emailSubject
	}

	// 4. Generate unique RFC Message-ID for tracking outbound reply
	outboundMessageID := fmt.Sprintf("<%d.%s@freel-platform.local>", time.Now().UnixNano(), mailbox.Email)

	// 5. Construct Raw MIME message (RFC 822)
	var rawMsg bytes.Buffer
	rawMsg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", mailbox.OwnerName, mailbox.Email))
	rawMsg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	rawMsg.WriteString(fmt.Sprintf("Subject: %s\r\n", emailSubject))
	rawMsg.WriteString(fmt.Sprintf("Message-ID: %s\r\n", outboundMessageID))
	if inReplyTo != "" {
		rawMsg.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	if references != "" {
		rawMsg.WriteString(fmt.Sprintf("References: %s\r\n", references))
	}
	rawMsg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	rawMsg.WriteString("\r\n")
	rawMsg.WriteString(body)

	rawBase64 := base64.RawURLEncoding.EncodeToString(rawMsg.Bytes())

	// 6. Invoke Google Send API
	apiURL := getGmailAPIBase() + "/gmail/v1/users/me/messages/send"
	sendPayload := map[string]interface{}{
		"raw": rawBase64,
	}
	if threadID != "" {
		sendPayload["threadId"] = threadID
	}

	payloadBytes, err := json.Marshal(sendPayload)
	if err != nil {
		return "", "", "", fmt.Errorf("marshal send payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("execute send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		log.Printf("[Gmail Sync Provider] Gmail Send API failed (%d): %s. Attempting SMTP fallback...", res.StatusCode, string(bodyBytes))

		// Fallback to SMTP sending if SMTP credentials exist in environment
		smtpErr := p.sendViaSMTP(ctx, mailbox, to, emailSubject, body, outboundMessageID, inReplyTo, references)
		if smtpErr == nil {
			log.Printf("[Gmail Sync Provider] SMTP fallback successfully delivered outbound email to %s", to)
			return outboundMessageID, outboundMessageID, threadID, nil
		}
		log.Printf("[Gmail Sync Provider] SMTP fallback error: %v", smtpErr)

		return "", "", "", fmt.Errorf("send email failed with status %d: %s", res.StatusCode, string(bodyBytes))
	}

	var sendRes struct {
		ID       string `json:"id"`
		ThreadID string `json:"threadId"`
	}
	if err := json.NewDecoder(res.Body).Decode(&sendRes); err != nil {
		return "", "", "", fmt.Errorf("decode send response: %w", err)
	}

	return sendRes.ID, outboundMessageID, sendRes.ThreadID, nil
}

func (p *GmailProvider) sendViaSMTP(ctx context.Context, mailbox *ConnectedMailbox, to string, emailSubject string, body string, outboundMessageID string, inReplyTo string, references string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP credentials not configured in environment")
	}

	if smtpPort == "" {
		smtpPort = "587"
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	fromAddr := mailbox.Email
	if fromAddr == "" {
		fromAddr = smtpUser
	}

	var rawMsg bytes.Buffer
	rawMsg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", mailbox.OwnerName, fromAddr))
	rawMsg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	rawMsg.WriteString(fmt.Sprintf("Subject: %s\r\n", emailSubject))
	rawMsg.WriteString(fmt.Sprintf("Message-ID: %s\r\n", outboundMessageID))
	if inReplyTo != "" {
		rawMsg.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	if references != "" {
		rawMsg.WriteString(fmt.Sprintf("References: %s\r\n", references))
	}
	rawMsg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	rawMsg.WriteString("\r\n")
	rawMsg.WriteString(body)

	addr := net.JoinHostPort(smtpHost, smtpPort)

	// Extract target email address if wrapped in <...>
	recipients := []string{to}
	if strings.Contains(to, "<") && strings.Contains(to, ">") {
		start := strings.Index(to, "<")
		end := strings.Index(to, ">")
		if start < end {
			recipients = []string{to[start+1 : end]}
		}
	}

	err := smtp.SendMail(addr, auth, smtpUser, recipients, rawMsg.Bytes())
	if err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

func getGmailAPIBase() string {
	if override := os.Getenv("GMAIL_API_URL"); override != "" {
		return override
	}
	return "https://gmail.googleapis.com"
}

package organization

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/files"
)

type Service interface {
	CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error)
	UpdateOrganization(ctx context.Context, id int64, req UpdateOrgRequest) (*Organization, error)
	GetProfile(ctx context.Context, id int64) (*Organization, error)
	UpdateProfile(ctx context.Context, id int64, req CompanyProfileUpdateRequest) (*Organization, error)
	UploadLogo(ctx context.Context, id int64, filename string, reader io.Reader) (*Organization, error)
	InviteUser(ctx context.Context, orgID int64, req InviteRequest) error
	GetNotificationPreferences(ctx context.Context, id int64) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, id int64, req UpdateNotificationPreferencesRequest) (*NotificationPreferences, error)
	GetEmailSettings(ctx context.Context, orgID int64) (*EmailSettings, error)
	UpdateEmailSettings(ctx context.Context, orgID int64, req UpdateEmailSettingsRequest) error
	GetConnectedMailboxes(ctx context.Context, orgID int64) ([]ConnectedMailbox, error)
	GetConnectedMailboxByID(ctx context.Context, id int64, orgID int64) (*ConnectedMailbox, error)
	ConnectMailbox(ctx context.Context, orgID int64, req ConnectMailboxRequest) (*ConnectedMailbox, error)
	UpdateMailbox(ctx context.Context, id int64, orgID int64, req UpdateMailboxRequest) error
	RemoveMailbox(ctx context.Context, id int64, orgID int64) error
	DisconnectMailbox(ctx context.Context, id int64, orgID int64) error
	SyncMailbox(ctx context.Context, id int64, orgID int64) error
	ToggleMailboxProcessing(ctx context.Context, id int64, orgID int64, pause bool) error
	StartGmailOAuth(ctx context.Context, orgID int64) (string, error)
	CompleteGmailOAuth(ctx context.Context, code string, state string) (string, error)
	SetSyncNowFunc(f func(ctx context.Context, id int64, orgID int64) error)
	GetCarrierIntegrations(ctx context.Context, orgID int64) ([]CarrierIntegration, error)
	ConnectCarrier(ctx context.Context, orgID int64, req ConnectCarrierRequest) (*CarrierIntegration, error)
	UpdateCarrier(ctx context.Context, id int64, orgID int64, req UpdateCarrierRequest) error
	RemoveCarrier(ctx context.Context, id int64, orgID int64) error
	SyncCarrier(ctx context.Context, id int64, orgID int64) error
	TestCarrierConnection(ctx context.Context, id int64, orgID int64) error
}

type service struct {
	repo        Repository
	eventBus    events.Bus
	fileSvc     files.Service
	syncNowFunc func(ctx context.Context, id int64, orgID int64) error
}

func NewService(repo Repository, eventBus events.Bus, fileSvc files.Service) Service {
	return &service{repo: repo, eventBus: eventBus, fileSvc: fileSvc, syncNowFunc: nil}
}

func (s *service) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
	org := &Organization{Name: req.Name}
	err := s.repo.Create(ctx, org)
	return org, err
}

func (s *service) UpdateOrganization(ctx context.Context, id int64, req UpdateOrgRequest) (*Organization, error) {
	org := &Organization{ID: id, Name: req.Name}
	err := s.repo.Update(ctx, org)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetProfile(ctx context.Context, id int64) (*Organization, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) UpdateProfile(ctx context.Context, id int64, req CompanyProfileUpdateRequest) (*Organization, error) {
	org := &Organization{
		ID:                 id,
		Name:               req.Name,
		LegalName:          req.LegalName,
		RegistrationNumber: req.RegistrationNumber,
		TaxNumber:          req.TaxNumber,
		Website:            req.Website,
		PrimaryEmail:       req.PrimaryEmail,
		PhoneNumber:        req.PhoneNumber,
		SupportEmail:       req.SupportEmail,
		Address:            req.Address,
		City:               req.City,
		State:              req.State,
		Country:            req.Country,
		PostalCode:         req.PostalCode,
		Industry:           req.Industry,
		CompanyType:        req.CompanyType,
		DefaultCurrency:    req.DefaultCurrency,
		DefaultTimezone:    req.DefaultTimezone,
		DateFormat:         req.DateFormat,
		DefaultLanguage:    req.DefaultLanguage,
		MeasurementSystem:  req.MeasurementSystem,
		WeightUnit:         req.WeightUnit,
		DimensionUnit:      req.DimensionUnit,
		VolumeUnit:         req.VolumeUnit,
		TimeFormat:         req.TimeFormat,
	}
	err := s.repo.Update(ctx, org)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) UploadLogo(ctx context.Context, id int64, filename string, reader io.Reader) (*Organization, error) {
	url, err := s.fileSvc.UploadFile(ctx, fmt.Sprintf("org_%d_logo_%d_%s", id, time.Now().Unix(), filename), reader)
	if err != nil {
		return nil, err
	}
	err = s.repo.UpdateLogo(ctx, id, url)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) InviteUser(ctx context.Context, orgID int64, req InviteRequest) error {
	// 1. Create a pending invite record in DB (Users module dependency, or handled here temporarily)
	// 2. Publish UserInvited event
	s.eventBus.Publish(events.Event{
		ID:        "evt-user-invited",
		Type:      events.EventUserInvited,
		Payload: map[string]interface{}{
			"org_id": orgID,
			"email":  req.Email,
			"role":   req.Role,
		},
		Timestamp: time.Now(),
	})
	return nil
}

func (s *service) GetNotificationPreferences(ctx context.Context, id int64) (*NotificationPreferences, error) {
	return s.repo.GetNotificationPreferences(ctx, id)
}

func (s *service) UpdateNotificationPreferences(ctx context.Context, id int64, req UpdateNotificationPreferencesRequest) (*NotificationPreferences, error) {
	prefs := &NotificationPreferences{
		OrgID:                 id,
		NewRFQReceived:        req.NewRFQReceived,
		NewQuoteReceived:      req.NewQuoteReceived,
		ShipmentStatusUpdates: req.ShipmentStatusUpdates,
		ShipmentExceptions:    req.ShipmentExceptions,
		InvitationAccepted:    req.InvitationAccepted,
		InvoicePaymentEvents:  req.InvoicePaymentEvents,
		SystemSecurityAlerts:  req.SystemSecurityAlerts,
	}
	err := s.repo.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		return nil, err
	}
	return s.repo.GetNotificationPreferences(ctx, id)
}

func (s *service) GetEmailSettings(ctx context.Context, orgID int64) (*EmailSettings, error) {
	return s.repo.GetEmailSettings(ctx, orgID)
}

func (s *service) UpdateEmailSettings(ctx context.Context, orgID int64, req UpdateEmailSettingsRequest) error {
	settings := &EmailSettings{
		OrgID:                     orgID,
		ProcessLogisticsInquiries: req.ProcessLogisticsInquiries,
		TrackEmailThreads:         req.TrackEmailThreads,
		SmartFiltering:            req.SmartFiltering,
	}
	return s.repo.UpdateEmailSettings(ctx, settings)
}

func (s *service) GetConnectedMailboxes(ctx context.Context, orgID int64) ([]ConnectedMailbox, error) {
	return s.repo.GetConnectedMailboxes(ctx, orgID)
}

func (s *service) GetConnectedMailboxByID(ctx context.Context, id int64, orgID int64) (*ConnectedMailbox, error) {
	// Org isolation is validated inside the repo call (WHERE id = ? AND org_id = ?)
	return s.repo.GetConnectedMailboxByID(ctx, id, orgID)
}

func (s *service) ConnectMailbox(ctx context.Context, orgID int64, req ConnectMailboxRequest) (*ConnectedMailbox, error) {
	encKey := os.Getenv(config.EnvMailboxEncryptionKey)
	if encKey == "" {
		return nil, errors.New("MAILBOX_ENCRYPTION_KEY is not configured")
	}

	provider := strings.ToUpper(req.Provider)
	if provider != "GMAIL" && provider != "MICROSOFT" && provider != "IMAP" {
		return nil, fmt.Errorf("invalid provider: %s", req.Provider)
	}

	var accessTokenEncrypted *string
	var refreshTokenEncrypted *string

	if req.AccessToken != "" {
		encrypted, err := crypto.Encrypt(req.AccessToken, encKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt access token: %w", err)
		}
		accessTokenEncrypted = &encrypted
	}

	if req.RefreshToken != "" {
		encrypted, err := crypto.Encrypt(req.RefreshToken, encKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		refreshTokenEncrypted = &encrypted
	}

	var scopesStr *string
	if len(req.OAuthScopes) > 0 {
		joined := strings.Join(req.OAuthScopes, " ")
		scopesStr = &joined
	}

	status := "CONNECTED"
	if provider != "IMAP" && (req.AccessToken == "" || req.RefreshToken == "") {
		status = "PENDING"
	}

	mailbox := &ConnectedMailbox{
		OrgID:                 orgID,
		Email:                 req.Email,
		OwnerName:             req.OwnerName,
		MailboxType:           req.MailboxType,
		Status:                status,
		SyncFrequency:         "Real-time",
		ProcessingEnabled:     true,
		Provider:              provider,
		AccessTokenEncrypted:  accessTokenEncrypted,
		RefreshTokenEncrypted: refreshTokenEncrypted,
		TokenExpiry:           req.TokenExpiry,
		OAuthScopes:           scopesStr,
	}

	log.Printf("[Mailbox Service] Connecting mailbox %s for OrgID %d with provider %s. Setting status: %s", req.Email, orgID, provider, status)

	err := s.repo.AddConnectedMailbox(ctx, mailbox)
	return mailbox, err
}

func (s *service) UpdateMailbox(ctx context.Context, id int64, orgID int64, req UpdateMailboxRequest) error {
	mailbox, err := s.repo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to find mailbox (isolation check): %w", err)
	}

	mailbox.OwnerName = req.OwnerName
	mailbox.MailboxType = req.MailboxType
	mailbox.SyncFrequency = req.SyncFrequency
	mailbox.ProcessingEnabled = req.ProcessingEnabled

	log.Printf("[Mailbox Service] Updating mailbox settings for mailbox ID %d (OrgID %d)", id, orgID)
	return s.repo.UpdateConnectedMailbox(ctx, mailbox)
}

func (s *service) RemoveMailbox(ctx context.Context, id int64, orgID int64) error {
	// Verify org isolation first
	_, err := s.repo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("mailbox not found or unauthorized: %w", err)
	}
	log.Printf("[Mailbox Service] Removing mailbox ID %d for OrgID %d", id, orgID)
	return s.repo.DeleteConnectedMailbox(ctx, id, orgID)
}

func (s *service) DisconnectMailbox(ctx context.Context, id int64, orgID int64) error {
	mailbox, err := s.repo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("mailbox not found or unauthorized: %w", err)
	}

	// Safely clear credentials and OAuth/sync states (idempotent operation)
	mailbox.AccessTokenEncrypted = nil
	mailbox.RefreshTokenEncrypted = nil
	mailbox.SyncCursor = nil
	mailbox.LastSyncError = nil
	
	oldStatus := mailbox.Status
	mailbox.Status = "DISCONNECTED"

	log.Printf("[Mailbox Service] Disconnecting mailbox ID %d (OrgID %d): Status transition %s -> DISCONNECTED. Credentials cleared.", id, orgID, oldStatus)

	return s.repo.UpdateConnectedMailbox(ctx, mailbox)
}

func (s *service) SyncMailbox(ctx context.Context, id int64, orgID int64) error {
	_, err := s.repo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("mailbox not found or unauthorized: %w", err)
	}
	log.Printf("[Mailbox Service] Syncing mailbox ID %d for OrgID %d", id, orgID)
	if s.syncNowFunc != nil {
		return s.syncNowFunc(ctx, id, orgID)
	}
	return s.repo.UpdateMailboxStatus(ctx, id, orgID, "SYNCING", true)
}

func (s *service) SetSyncNowFunc(f func(ctx context.Context, id int64, orgID int64) error) {
	s.syncNowFunc = f
}

func (s *service) ToggleMailboxProcessing(ctx context.Context, id int64, orgID int64, pause bool) error {
	mailbox, err := s.repo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("mailbox not found or unauthorized: %w", err)
	}

	mailbox.ProcessingEnabled = !pause
	log.Printf("[Mailbox Service] Toggling mailbox processing enabled=%t for mailbox ID %d", !pause, id)
	return s.repo.UpdateConnectedMailbox(ctx, mailbox)
}

func (s *service) GetCarrierIntegrations(ctx context.Context, orgID int64) ([]CarrierIntegration, error) {
	return s.repo.GetCarrierIntegrations(ctx, orgID)
}

func (s *service) ConnectCarrier(ctx context.Context, orgID int64, req ConnectCarrierRequest) (*CarrierIntegration, error) {
	if req.CarrierSCAC == "" {
		return nil, errors.New("CarrierSCAC is required")
	}
	ci := &CarrierIntegration{
		OrgID:            orgID,
		CarrierSCAC:      req.CarrierSCAC,
		ConnectionMethod: req.ConnectionMethod,
		Environment:      req.Environment,
		CredentialsJSON:  &req.CredentialsJSON,
		Capabilities:     req.Capabilities,
		ConnectionStatus: "Connected",
		IsActive:         true,
	}
	if err := s.repo.AddCarrierIntegration(ctx, ci); err != nil {
		return nil, err
	}
	return ci, nil
}

func (s *service) UpdateCarrier(ctx context.Context, id int64, orgID int64, req UpdateCarrierRequest) error {
	var creds *string
	if req.CredentialsJSON != "" {
		creds = &req.CredentialsJSON
	}
	ci := &CarrierIntegration{
		ID:               id,
		OrgID:            orgID,
		ConnectionMethod: req.ConnectionMethod,
		Environment:      req.Environment,
		CredentialsJSON:  creds,
		Capabilities:     req.Capabilities,
	}
	return s.repo.UpdateCarrierIntegration(ctx, ci)
}

func (s *service) RemoveCarrier(ctx context.Context, id int64, orgID int64) error {
	return s.repo.DeleteCarrierIntegration(ctx, id, orgID)
}

func (s *service) SyncCarrier(ctx context.Context, id int64, orgID int64) error {
	ci, err := s.repo.GetCarrierIntegrationByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	syncer := NewSyncer(s.repo.GetDB())
	err = syncer.Sync(ctx, orgID, ci.CarrierSCAC)
	
	if err != nil {
		s.repo.GetDB().ExecContext(ctx, `
			UPDATE carrier_integrations 
			SET failed_attempts = failed_attempts + 1, 
				error_details = ?, 
				next_retry_time = DATE_ADD(NOW(), INTERVAL LEAST(1440, 5 * (1 << failed_attempts)) MINUTE),
				sync_status = 'Failed',
				last_synced_at = NOW()
			WHERE id = ? AND org_id = ?
		`, err.Error(), id, orgID)
		return err
	}

	s.repo.GetDB().ExecContext(ctx, `
		UPDATE carrier_integrations 
		SET failed_attempts = 0, 
			error_details = NULL, 
			next_retry_time = DATE_ADD(NOW(), INTERVAL 4 HOUR),
			sync_status = 'Completed',
			last_synced_at = NOW()
		WHERE id = ? AND org_id = ?
	`, id, orgID)

	// Emit audit log for manual sync success
	s.eventBus.Publish(events.Event{
		ID:        fmt.Sprintf("evt-carrier-sync-%d", time.Now().Unix()),
		Type:      "CARRIER_SYNC_COMPLETED",
		Payload: map[string]interface{}{
			"org_id": orgID,
			"carrier_scac": ci.CarrierSCAC,
		},
		Timestamp: time.Now(),
	})

	return nil
}

func (s *service) TestCarrierConnection(ctx context.Context, id int64, orgID int64) error {
	ci, err := s.repo.GetCarrierIntegrationByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	provider, err := adapters.GetTrackingProvider(s.repo.GetDB(), orgID, ci.CarrierSCAC)
	if err != nil {
		return err
	}

	// Make a dummy request to test authentication and reachability
	_, err = provider.GetTrackingEvents(ctx, carrier.TrackingRequest{
		CarrierSCAC: ci.CarrierSCAC,
		BookingNumber: "TEST-PING",
	})
	
	if err != nil {
		s.repo.UpdateCarrierIntegrationStatus(ctx, id, orgID, "Error", "Test connection failed: " + err.Error(), false)
		return err
	}
	
	s.repo.UpdateCarrierIntegrationStatus(ctx, id, orgID, "Connected", "Connection test successful", false)
	return nil
}

func (s *service) StartGmailOAuth(ctx context.Context, orgID int64) (string, error) {
	clientID := os.Getenv(config.EnvGoogleClientID)
	redirectURI := os.Getenv(config.EnvGoogleRedirectURI)
	if clientID == "" || redirectURI == "" {
		return "", errors.New("Google OAuth is not configured on the server")
	}

	state, err := GenerateOAuthState(orgID)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure state: %w", err)
	}

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent&state=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("https://www.googleapis.com/auth/gmail.readonly https://www.googleapis.com/auth/gmail.send https://www.googleapis.com/auth/userinfo.email"),
		url.QueryEscape(state),
	)

	return authURL, nil
}

func (s *service) CompleteGmailOAuth(ctx context.Context, code string, state string) (string, error) {
	orgID, err := ValidateOAuthState(state)
	if err != nil {
		return "", fmt.Errorf("invalid state: %w", err)
	}

	clientID := os.Getenv(config.EnvGoogleClientID)
	clientSecret := os.Getenv(config.EnvGoogleClientSecret)
	redirectURI := os.Getenv(config.EnvGoogleRedirectURI)
	encKey := os.Getenv(config.EnvMailboxEncryptionKey)
	if clientID == "" || clientSecret == "" || redirectURI == "" || encKey == "" {
		return "", errors.New("server is misconfigured (missing Google or Mailbox Encryption settings)")
	}

	// 1. Exchange authorization code for tokens
	tokenExchangeData := url.Values{}
	tokenExchangeData.Set("client_id", clientID)
	tokenExchangeData.Set("client_secret", clientSecret)
	tokenExchangeData.Set("code", code)
	tokenExchangeData.Set("grant_type", "authorization_code")
	tokenExchangeData.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(tokenExchangeData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("token exchange status %d: %s", res.StatusCode, string(body))
	}

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return "", fmt.Errorf("decode token response failed: %w", err)
	}

	// 2. Retrieve connected Gmail address profile
	profileReq, err := http.NewRequestWithContext(ctx, "GET", "https://gmail.googleapis.com/gmail/v1/users/me/profile", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create profile request: %w", err)
	}
	profileReq.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	profileRes, err := client.Do(profileReq)
	if err != nil {
		return "", fmt.Errorf("profile request failed: %w", err)
	}
	defer profileRes.Body.Close()

	if profileRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(profileRes.Body)
		return "", fmt.Errorf("profile request status %d: %s", profileRes.StatusCode, string(body))
	}

	var profile struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.NewDecoder(profileRes.Body).Decode(&profile); err != nil {
		return "", fmt.Errorf("decode profile failed: %w", err)
	}

	if profile.EmailAddress == "" {
		return "", errors.New("retrieved empty email address from Google profile")
	}

	// 3. Encrypt access and refresh tokens using Task 7 encryption service
	accessTokenEnc, err := crypto.Encrypt(tokenRes.AccessToken, encKey)
	if err != nil {
		return "", fmt.Errorf("encrypt access token failed: %w", err)
	}

	var refreshTokenEnc *string
	if tokenRes.RefreshToken != "" {
		rtEnc, err := crypto.Encrypt(tokenRes.RefreshToken, encKey)
		if err != nil {
			return "", fmt.Errorf("encrypt refresh token failed: %w", err)
		}
		refreshTokenEnc = &rtEnc
	}

	tokenExpiry := time.Now().Add(time.Duration(tokenRes.ExpiresIn) * time.Second)

	// 4. Save connected mailbox in database
	mailboxes, err := s.repo.GetConnectedMailboxes(ctx, orgID)
	var existing *ConnectedMailbox
	if err == nil {
		for _, m := range mailboxes {
			if strings.EqualFold(m.Email, profile.EmailAddress) {
				existing = &m
				break
			}
		}
	}

	if existing != nil {
		// Update existing mailbox
		existing.AccessTokenEncrypted = &accessTokenEnc
		if refreshTokenEnc != nil {
			existing.RefreshTokenEncrypted = refreshTokenEnc
		}
		existing.TokenExpiry = &tokenExpiry
		existing.OAuthScopes = &tokenRes.Scope
		existing.Status = "CONNECTED"
		existing.Provider = "GMAIL"
		
		err = s.repo.UpdateConnectedMailbox(ctx, existing)
		if err != nil {
			return "", fmt.Errorf("failed to update connected mailbox: %w", err)
		}
		log.Printf("[Gmail OAuth] Reconnected existing mailbox: %s for OrgID %d", profile.EmailAddress, orgID)
	} else {
		// Create new mailbox connection configuration
		mailbox := &ConnectedMailbox{
			OrgID:                 orgID,
			Email:                 profile.EmailAddress,
			OwnerName:             "Gmail Account",
			MailboxType:           "Individual",
			IsPrimary:             false,
			Status:                "CONNECTED",
			SyncFrequency:         "Real-time",
			ProcessingEnabled:     true,
			Provider:              "GMAIL",
			AccessTokenEncrypted:  &accessTokenEnc,
			RefreshTokenEncrypted: refreshTokenEnc,
			TokenExpiry:           &tokenExpiry,
			OAuthScopes:           &tokenRes.Scope,
		}
		err = s.repo.AddConnectedMailbox(ctx, mailbox)
		if err != nil {
			return "", fmt.Errorf("failed to save connected mailbox: %w", err)
		}
		log.Printf("[Gmail OAuth] Connected new mailbox: %s for OrgID %d", profile.EmailAddress, orgID)
	}

	return profile.EmailAddress, nil
}


// Replace the existing SyncCarrier


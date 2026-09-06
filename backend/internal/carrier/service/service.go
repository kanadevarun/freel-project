package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/audit"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/repository"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/jmoiron/sqlx"
)

var (
	ErrCarrierNotFound        = errors.New("carrier provider not found")
	ErrIntegrationNotFound    = errors.New("carrier integration not found")
	ErrDuplicateIntegration   = errors.New("an integration already exists for this carrier and environment")
	ErrMissingCredentials     = errors.New("credentials must be provided to connect carrier")
	ErrDecryptionFailed       = errors.New("failed to decrypt carrier credentials")
	ErrCapabilityDisabled     = errors.New("requested capability is not enabled for this carrier integration")
	ErrIntegrationDisabled    = errors.New("carrier integration is currently disabled")
)

type CarrierService interface {
	GetProviders(ctx context.Context) ([]domain.CarrierProvider, error)
	GetIntegrations(ctx context.Context, orgID int64) ([]domain.CarrierIntegrationView, error)
	GetIntegration(ctx context.Context, orgID int64, id int64) (*domain.CarrierIntegrationView, error)
	ConnectCarrier(ctx context.Context, orgID int64, userID int64, req domain.ConnectCarrierRequest) (*domain.CarrierIntegrationView, error)
	UpdateCarrier(ctx context.Context, orgID int64, userID int64, id int64, req domain.UpdateCarrierRequest) (*domain.CarrierIntegrationView, error)
	ToggleCarrier(ctx context.Context, orgID int64, userID int64, id int64, enable bool) (*domain.CarrierIntegrationView, error)
	DisconnectCarrier(ctx context.Context, orgID int64, userID int64, id int64) error
	TestConnection(ctx context.Context, orgID int64, id int64) (*domain.TestConnectionResult, error)
	TestDirectConnection(ctx context.Context, orgID int64, req domain.TestDirectRequest) (*domain.TestConnectionResult, error)
	SyncCarrier(ctx context.Context, orgID int64, id int64) error

	// Task 6: Synchronization, History, Health & Webhook Processing
	SyncNow(ctx context.Context, orgID int64, integrationID int64, req domain.SyncNowRequest) (*domain.IntegrationSyncJobView, error)
	GetSyncHistory(ctx context.Context, orgID int64, integrationID int64, limit, offset int) ([]domain.IntegrationSyncJobView, int, error)
	GetSyncJob(ctx context.Context, orgID int64, jobID int64) (*domain.IntegrationSyncJobView, error)
	GetIntegrationHealth(ctx context.Context, orgID int64, integrationID int64) (*domain.IntegrationHealthDetail, error)
	ProcessWebhook(ctx context.Context, providerCode string, rawBody []byte, headers map[string]string) (*domain.CarrierWebhookEvent, error)
	SetTrackingSyncer(handler TrackingSyncHandler)
	SetBookingSyncer(handler BookingSyncHandler)
	SetDB(db *sqlx.DB)

	// Adapter Operations with Capability Enforcement & Normalization
	GetTracking(ctx context.Context, orgID int64, integrationID int64, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error)
	GetRates(ctx context.Context, orgID int64, integrationID int64, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error)
	GetContractRates(ctx context.Context, orgID int64, integrationID int64, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error)
	GetSpotRates(ctx context.Context, orgID int64, integrationID int64, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error)
	CreateBooking(ctx context.Context, orgID int64, integrationID int64, req domain.BookingRequest) (*domain.NormalizedBookingResult, error)
	GetBooking(ctx context.Context, orgID int64, integrationID int64, bookingRef string) (*domain.NormalizedBookingResult, error)
	GetDocuments(ctx context.Context, orgID int64, integrationID int64, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error)

	// Adapter Execution
	GetAdapterForIntegration(ctx context.Context, orgID int64, scac string, env domain.Environment) (adapters.CarrierAdapter, domain.DecryptedCredentials, error)
}

type carrierService struct {
	repo          repository.CarrierRepository
	registry      *adapters.AdapterRegistry
	encryptionKey string
	db            *sqlx.DB
	syncEngine    *CarrierSyncEngine
}

func NewCarrierService(repo repository.CarrierRepository, encryptionKey string) CarrierService {
	if encryptionKey == "" {
		encryptionKey = "LogisticsHQ_Carrier_Secret_Encryption_Key_32B!"
	}
	s := &carrierService{
		repo:          repo,
		registry:      adapters.GetDefaultRegistry(),
		encryptionKey: encryptionKey,
	}
	s.syncEngine = NewCarrierSyncEngine(nil, repo, s)
	return s
}

func (s *carrierService) SetDB(db *sqlx.DB) {
	s.db = db
	s.syncEngine.db = db
}

func (s *carrierService) SetTrackingSyncer(handler TrackingSyncHandler) {
	s.syncEngine.SetTrackingSyncer(handler)
}

func (s *carrierService) SetBookingSyncer(handler BookingSyncHandler) {
	s.syncEngine.SetBookingSyncer(handler)
}

func (s *carrierService) GetProviders(ctx context.Context) ([]domain.CarrierProvider, error) {
	return s.repo.ListProviders(ctx)
}

func (s *carrierService) GetIntegrations(ctx context.Context, orgID int64) ([]domain.CarrierIntegrationView, error) {
	integrations, err := s.repo.ListIntegrations(ctx, orgID)
	if err != nil {
		return nil, err
	}

	views := make([]domain.CarrierIntegrationView, 0, len(integrations))
	for _, ci := range integrations {
		views = append(views, s.toView(&ci))
	}
	return views, nil
}

func (s *carrierService) GetIntegration(ctx context.Context, orgID int64, id int64) (*domain.CarrierIntegrationView, error) {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}
	view := s.toView(ci)
	return &view, nil
}

func (s *carrierService) ConnectCarrier(ctx context.Context, orgID int64, userID int64, req domain.ConnectCarrierRequest) (*domain.CarrierIntegrationView, error) {
	scac := strings.ToUpper(strings.TrimSpace(req.CarrierSCAC))
	if scac == "" {
		return nil, errors.New("carrier SCAC is required")
	}

	env, err := domain.ParseEnvironment(req.Environment)
	if err != nil {
		env = domain.EnvProduction
	}

	method, err := domain.ParseConnectionMethod(req.ConnectionMethod)
	if err != nil {
		method = domain.MethodAPI
	}

	// Check if already exists for this org + scac + env
	existing, _ := s.repo.GetIntegrationBySCAC(ctx, orgID, scac, env)
	if existing != nil {
		return nil, ErrDuplicateIntegration
	}

	// Resolve provider
	var providerID *int64
	var carrierName string
	var supportedCaps []domain.Capability
	provider, _ := s.repo.GetProviderBySCAC(ctx, scac)
	if provider != nil {
		providerID = &provider.ID
		carrierName = provider.Name
		supportedCaps = provider.SupportedCapabilities
	} else {
		carrierName = scac
		supportedCaps = []domain.Capability{domain.CapTracking, domain.CapRates, domain.CapBooking, domain.CapDocuments}
	}

	// Parse credentials
	credsJSON := ""
	if req.Credentials != nil && len(req.Credentials) > 0 {
		if b, err := json.Marshal(req.Credentials); err == nil {
			credsJSON = string(b)
		}
	} else if req.CredentialsJSON != "" {
		credsJSON = req.CredentialsJSON
	}
	if credsJSON == "" {
		credsJSON = "{}"
	}

	// Encrypt credentials at rest
	encryptedCreds, err := crypto.Encrypt(credsJSON, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt carrier credentials: %w", err)
	}

	// Generate safe mask
	maskedMap := domain.MaskCredentials(credsJSON)
	maskedJSONBytes, _ := json.Marshal(maskedMap)
	maskedJSONStr := string(maskedJSONBytes)

	// Capabilities parsing
	caps := s.parseCapabilities(req.Capabilities, supportedCaps)
	capsJSONBytes, _ := json.Marshal(caps)
	capsJSONStr := string(capsJSONBytes)

	// Config options
	var configJSONStr *string
	if req.ConfigOptions != nil {
		if b, err := json.Marshal(req.ConfigOptions); err == nil {
			s := string(b)
			configJSONStr = &s
		}
	}

	ci := &domain.CarrierIntegration{
		OrgID:                orgID,
		CarrierProviderID:    providerID,
		CarrierSCAC:          scac,
		CarrierName:          &carrierName,
		ConnectionMethod:     method,
		Environment:          env,
		ConnectionStatus:     domain.StatusDisconnected,
		IsEnabled:            true,
		CredentialsJSON:      &credsJSON,
		EncryptedCredentials: &encryptedCreds,
		CredentialMaskJSON:   &maskedJSONStr,
		CapabilitiesJSON:     &capsJSONStr,
		ConfigOptionsJSON:    configJSONStr,
		Capabilities:         caps,
	}

	if err := s.repo.CreateIntegration(ctx, ci); err != nil {
		return nil, err
	}

	// Perform initial connection test in background / synchronously
	testRes, _ := s.testIntegrationConnection(ctx, ci)
	if testRes != nil && testRes.Success {
		ci.ConnectionStatus = domain.StatusConnected
		_ = s.repo.UpdateStatus(ctx, orgID, ci.ID, domain.StatusConnected, nil)
	}

	_ = s.repo.RecordAuditLog(ctx, orgID, &userID, "CARRIER_CONNECTED", ci.ID, map[string]interface{}{
		"carrier_scac":      scac,
		"carrier_name":      carrierName,
		"environment":       string(env),
		"connection_method": string(method),
		"capabilities":      caps,
	})

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &userID,
		Action:       auditDomain.ActionConnect,
		Module:       auditDomain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER",
		ResourceID:   fmt.Sprintf("%d", ci.ID),
		ResourceName: carrierName,
		Description:  fmt.Sprintf("Connected carrier integration for %s (%s)", carrierName, env),
		After: map[string]interface{}{
			"carrier_scac": scac,
			"environment":  string(env),
		},
		Result: auditDomain.ResultSuccess,
	})

	view := s.toView(ci)
	return &view, nil
}

func (s *carrierService) UpdateCarrier(ctx context.Context, orgID int64, userID int64, id int64, req domain.UpdateCarrierRequest) (*domain.CarrierIntegrationView, error) {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	if req.Environment != nil {
		if env, err := domain.ParseEnvironment(*req.Environment); err == nil {
			ci.Environment = env
		}
	}

	if req.ConnectionMethod != nil {
		if method, err := domain.ParseConnectionMethod(*req.ConnectionMethod); err == nil {
			ci.ConnectionMethod = method
		}
	}

	if req.IsEnabled != nil {
		ci.IsEnabled = *req.IsEnabled
		if !ci.IsEnabled {
			ci.ConnectionStatus = domain.StatusDisabled
		}
	}

	// Update credentials if provided
	credsJSON := ""
	if req.CredentialsJSON != nil && *req.CredentialsJSON != "" {
		credsJSON = *req.CredentialsJSON
	} else if req.Credentials != nil && len(req.Credentials) > 0 {
		if b, err := json.Marshal(req.Credentials); err == nil {
			credsJSON = string(b)
		}
	}

	if credsJSON != "" {
		encryptedCreds, err := crypto.Encrypt(credsJSON, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt carrier credentials: %w", err)
		}
		ci.CredentialsJSON = &credsJSON
		ci.EncryptedCredentials = &encryptedCreds

		maskedMap := domain.MaskCredentials(credsJSON)
		maskedBytes, _ := json.Marshal(maskedMap)
		maskedStr := string(maskedBytes)
		ci.CredentialMaskJSON = &maskedStr
	}

	// Update capabilities if provided
	if req.Capabilities != nil {
		caps := s.parseCapabilities(req.Capabilities, nil)
		ci.Capabilities = caps
		capsBytes, _ := json.Marshal(caps)
		capsStr := string(capsBytes)
		ci.CapabilitiesJSON = &capsStr
		if b, err := json.Marshal(caps); err == nil {
			s := string(b)
			ci.CapabilitiesJSON = &s
		}
	}

	if req.ConfigOptions != nil {
		if b, err := json.Marshal(req.ConfigOptions); err == nil {
			s := string(b)
			ci.ConfigOptionsJSON = &s
		}
	}

	if req.IsEnabled != nil {
		ci.IsEnabled = *req.IsEnabled
		if !ci.IsEnabled {
			ci.ConnectionStatus = domain.StatusDisabled
		}
	}

	if err := s.repo.UpdateIntegration(ctx, ci); err != nil {
		return nil, err
	}

	_ = s.repo.RecordAuditLog(ctx, orgID, &userID, "CARRIER_UPDATED", ci.ID, map[string]interface{}{
		"carrier_scac":      ci.CarrierSCAC,
		"environment":       string(ci.Environment),
		"connection_method": string(ci.ConnectionMethod),
		"capabilities":      ci.Capabilities,
		"is_active":         ci.IsEnabled,
	})

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &userID,
		Action:       auditDomain.ActionUpdate,
		Module:       auditDomain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER",
		ResourceID:   fmt.Sprintf("%d", ci.ID),
		ResourceName: ci.CarrierSCAC,
		Description:  fmt.Sprintf("Updated carrier integration configuration for %s", ci.CarrierSCAC),
		Result:       auditDomain.ResultSuccess,
		Metadata: map[string]interface{}{
			"carrier_scac":      ci.CarrierSCAC,
			"environment":       string(ci.Environment),
			"connection_method": string(ci.ConnectionMethod),
			"capabilities":      ci.Capabilities,
			"is_active":         ci.IsEnabled,
		},
	})

	view := s.toView(ci)
	return &view, nil
}

func (s *carrierService) ToggleCarrier(ctx context.Context, orgID int64, userID int64, id int64, enable bool) (*domain.CarrierIntegrationView, error) {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	ci.IsEnabled = enable
	if !enable {
		ci.ConnectionStatus = domain.StatusDisabled
	} else {
		ci.ConnectionStatus = domain.StatusConnected
	}

	if err := s.repo.UpdateIntegration(ctx, ci); err != nil {
		return nil, err
	}

	action := "CARRIER_ENABLED"
	auditAction := auditDomain.ActionEnable
	desc := fmt.Sprintf("Enabled carrier integration for %s", ci.CarrierSCAC)
	if !enable {
		action = "CARRIER_DISABLED"
		auditAction = auditDomain.ActionDisable
		desc = fmt.Sprintf("Disabled carrier integration for %s", ci.CarrierSCAC)
	}

	_ = s.repo.RecordAuditLog(ctx, orgID, &userID, action, ci.ID, map[string]interface{}{
		"carrier_scac": ci.CarrierSCAC,
		"is_active":    ci.IsEnabled,
	})

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &userID,
		Action:       auditAction,
		Module:       auditDomain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: ci.CarrierSCAC,
		Description:  desc,
		Result:       auditDomain.ResultSuccess,
	})

	view := s.toView(ci)
	return &view, nil
}

func (s *carrierService) DisconnectCarrier(ctx context.Context, orgID int64, userID int64, id int64) error {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrIntegrationNotFound
		}
		return err
	}

	if err := s.repo.DeleteIntegration(ctx, orgID, id); err != nil {
		return err
	}

	_ = s.repo.RecordAuditLog(ctx, orgID, &userID, "CARRIER_DISCONNECTED", id, map[string]interface{}{
		"carrier_scac": ci.CarrierSCAC,
		"environment":  string(ci.Environment),
	})

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &userID,
		Action:       auditDomain.ActionDisconnect,
		Module:       auditDomain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: ci.CarrierSCAC,
		Description:  fmt.Sprintf("Disconnected carrier integration for %s", ci.CarrierSCAC),
		Result:       auditDomain.ResultSuccess,
	})

	return nil
}

func (s *carrierService) TestConnection(ctx context.Context, orgID int64, id int64) (*domain.TestConnectionResult, error) {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	res, err := s.testIntegrationConnection(ctx, ci)
	if err != nil {
		errMsg := domain.SanitizeError(err)
		_ = s.repo.UpdateStatus(ctx, orgID, id, domain.StatusError, &errMsg)
		return nil, err
	}

	if res.Success {
		_ = s.repo.UpdateStatus(ctx, orgID, id, domain.StatusConnected, nil)
	} else {
		_ = s.repo.UpdateStatus(ctx, orgID, id, domain.StatusError, &res.Message)
	}

	return res, nil
}

func (s *carrierService) TestDirectConnection(ctx context.Context, orgID int64, req domain.TestDirectRequest) (*domain.TestConnectionResult, error) {
	scac := strings.ToUpper(strings.TrimSpace(req.CarrierSCAC))
	if scac == "" {
		return nil, errors.New("carrier SCAC is required")
	}

	env, _ := domain.ParseEnvironment(req.Environment)
	if env == "" {
		env = domain.EnvProduction
	}

	credsJSON := req.CredentialsJSON
	if credsJSON == "" && req.Credentials != nil {
		if b, err := json.Marshal(req.Credentials); err == nil {
			credsJSON = string(b)
		}
	}

	var decryptedCreds domain.DecryptedCredentials
	if credsJSON != "" {
		_ = json.Unmarshal([]byte(credsJSON), &decryptedCreds)
	}

	adapter := s.registry.GetAdapterBySCAC(scac)
	return adapter.TestConnection(ctx, decryptedCreds, env)
}

func (s *carrierService) SyncCarrier(ctx context.Context, orgID int64, id int64) error {
	_, err := s.SyncNow(ctx, orgID, id, domain.SyncNowRequest{Operation: string(domain.SyncOpFullSync)})
	return err
}

func (s *carrierService) SyncNow(ctx context.Context, orgID int64, integrationID int64, req domain.SyncNowRequest) (*domain.IntegrationSyncJobView, error) {
	return s.syncEngine.SyncNow(ctx, orgID, integrationID, req)
}

func (s *carrierService) GetSyncHistory(ctx context.Context, orgID int64, integrationID int64, limit, offset int) ([]domain.IntegrationSyncJobView, int, error) {
	return s.syncEngine.GetSyncHistory(ctx, orgID, integrationID, limit, offset)
}

func (s *carrierService) GetSyncJob(ctx context.Context, orgID int64, jobID int64) (*domain.IntegrationSyncJobView, error) {
	return s.syncEngine.GetSyncJob(ctx, orgID, jobID)
}

func (s *carrierService) GetIntegrationHealth(ctx context.Context, orgID int64, integrationID int64) (*domain.IntegrationHealthDetail, error) {
	return s.syncEngine.GetIntegrationHealth(ctx, orgID, integrationID)
}

func (s *carrierService) ProcessWebhook(ctx context.Context, providerCode string, rawBody []byte, headers map[string]string) (*domain.CarrierWebhookEvent, error) {
	return s.syncEngine.ProcessWebhook(ctx, providerCode, rawBody, headers)
}

func (s *carrierService) GetAdapterForIntegration(ctx context.Context, orgID int64, scac string, env domain.Environment) (adapters.CarrierAdapter, domain.DecryptedCredentials, error) {
	ci, err := s.repo.GetIntegrationBySCAC(ctx, orgID, scac, env)
	if err != nil {
		return nil, domain.DecryptedCredentials{}, ErrIntegrationNotFound
	}

	if !ci.IsEnabled {
		return nil, domain.DecryptedCredentials{}, errors.New("carrier integration is disabled")
	}

	decryptedCreds, err := s.decryptCredentials(ci)
	if err != nil {
		return nil, domain.DecryptedCredentials{}, err
	}

	adapter := s.registry.GetAdapterBySCAC(scac)
	return adapter, decryptedCreds, nil
}

func (s *carrierService) testIntegrationConnection(ctx context.Context, ci *domain.CarrierIntegration) (*domain.TestConnectionResult, error) {
	decryptedCreds, err := s.decryptCredentials(ci)
	if err != nil {
		return &domain.TestConnectionResult{
			Success:   false,
			Message:   "Failed to decrypt credentials: " + err.Error(),
			ErrorCode: "DECRYPTION_ERROR",
		}, nil
	}

	adapter := s.registry.GetAdapterBySCAC(ci.CarrierSCAC)
	return adapter.TestConnection(ctx, decryptedCreds, ci.Environment)
}

func (s *carrierService) decryptCredentials(ci *domain.CarrierIntegration) (domain.DecryptedCredentials, error) {
	var rawJSON string

	if ci.EncryptedCredentials != nil && *ci.EncryptedCredentials != "" {
		decrypted, err := crypto.Decrypt(*ci.EncryptedCredentials, s.encryptionKey)
		if err != nil {
			// Fallback to cleartext if migration is in transition
			if ci.CredentialsJSON != nil {
				rawJSON = *ci.CredentialsJSON
			} else {
				return domain.DecryptedCredentials{}, ErrDecryptionFailed
			}
		} else {
			rawJSON = decrypted
		}
	} else if ci.CredentialsJSON != nil {
		rawJSON = *ci.CredentialsJSON
	}

	var creds domain.DecryptedCredentials
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &creds); err != nil {
			return domain.DecryptedCredentials{}, err
		}
	}
	return creds, nil
}

func (s *carrierService) parseCapabilities(input interface{}, supported []domain.Capability) []domain.Capability {
	if input == nil {
		if len(supported) > 0 {
			return supported
		}
		return []domain.Capability{domain.CapTracking, domain.CapRates, domain.CapBooking, domain.CapDocuments}
	}

	var result []domain.Capability

	switch v := input.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				if capEnum, valid := domain.ParseCapability(str); valid {
					result = append(result, capEnum)
				}
			}
		}
	case []string:
		for _, str := range v {
			if capEnum, valid := domain.ParseCapability(str); valid {
				result = append(result, capEnum)
			}
		}
	case map[string]interface{}:
		for k, val := range v {
			if b, ok := val.(bool); ok && b {
				if capEnum, valid := domain.ParseCapability(k); valid {
					result = append(result, capEnum)
				}
			}
		}
	case map[string]bool:
		for k, b := range v {
			if b {
				if capEnum, valid := domain.ParseCapability(k); valid {
					result = append(result, capEnum)
				}
			}
		}
	}

	if len(result) == 0 {
		result = []domain.Capability{domain.CapTracking}
	}
	return result
}

func (s *carrierService) resolveIntegrationAndAdapter(ctx context.Context, orgID int64, integrationID int64, requiredCap domain.Capability) (*domain.CarrierIntegration, adapters.CarrierAdapter, domain.DecryptedCredentials, error) {
	ci, err := s.repo.GetIntegrationByID(ctx, orgID, integrationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, domain.DecryptedCredentials{}, ErrIntegrationNotFound
		}
		return nil, nil, domain.DecryptedCredentials{}, err
	}

	if !ci.IsEnabled || ci.ConnectionStatus == domain.StatusDisabled {
		return nil, nil, domain.DecryptedCredentials{}, ErrIntegrationDisabled
	}

	// Capability enforcement
	hasCap := false
	for _, c := range ci.Capabilities {
		if c == requiredCap {
			hasCap = true
			break
		}
	}
	if !hasCap {
		return nil, nil, domain.DecryptedCredentials{}, ErrCapabilityDisabled
	}

	// Decrypt credentials
	creds, err := s.decryptCredentials(ci)
	if err != nil {
		return nil, nil, domain.DecryptedCredentials{}, err
	}

	adapter := s.registry.GetAdapterBySCAC(ci.CarrierSCAC)
	return ci, adapter, creds, nil
}

func (s *carrierService) GetTracking(ctx context.Context, orgID int64, integrationID int64, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapTracking)
	if err != nil {
		return nil, err
	}

	return adapter.GetTracking(ctx, creds, ci.Environment, req)
}

func (s *carrierService) GetRates(ctx context.Context, orgID int64, integrationID int64, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapRates)
	if err != nil {
		return nil, err
	}

	return adapter.GetRates(ctx, creds, ci.Environment, req)
}

func (s *carrierService) GetContractRates(ctx context.Context, orgID int64, integrationID int64, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapContractRates)
	if err != nil {
		return nil, err
	}

	return adapter.GetContractRates(ctx, creds, ci.Environment, req)
}

func (s *carrierService) GetSpotRates(ctx context.Context, orgID int64, integrationID int64, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapSpotRates)
	if err != nil {
		return nil, err
	}

	return adapter.GetSpotRates(ctx, creds, ci.Environment, req)
}

func (s *carrierService) CreateBooking(ctx context.Context, orgID int64, integrationID int64, req domain.BookingRequest) (*domain.NormalizedBookingResult, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapBooking)
	if err != nil {
		return nil, err
	}

	result, err := adapter.CreateBooking(ctx, creds, ci.Environment, req)
	if err != nil {
		_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
			OrgID:        orgID,
			Action:       auditDomain.ActionCreate,
			Module:       auditDomain.ModuleCarrierIntegrations,
			ResourceType: "CARRIER_BOOKING",
			ResourceID:   fmt.Sprintf("%d", integrationID),
			ResourceName: ci.CarrierSCAC,
			Description:  fmt.Sprintf("Failed carrier booking creation on %s: %s", ci.CarrierSCAC, err.Error()),
			Result:       auditDomain.ResultFailed,
			ErrorMessage: err.Error(),
		})
		return nil, err
	}

	if result != nil {
		_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
			OrgID:        orgID,
			Action:       auditDomain.ActionCreate,
			Module:       auditDomain.ModuleCarrierIntegrations,
			ResourceType: "CARRIER_BOOKING",
			ResourceID:   result.BookingNumber,
			ResourceName: ci.CarrierSCAC,
			Description:  fmt.Sprintf("Created carrier booking %s with %s", result.BookingNumber, ci.CarrierSCAC),
			Result:       auditDomain.ResultSuccess,
			Metadata: map[string]interface{}{
				"booking_number": result.BookingNumber,
				"carrier_scac":   ci.CarrierSCAC,
				"status":         result.Status,
			},
		})
	}

	return result, nil
}

func (s *carrierService) GetBooking(ctx context.Context, orgID int64, integrationID int64, bookingRef string) (*domain.NormalizedBookingResult, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapBooking)
	if err != nil {
		return nil, err
	}

	return adapter.GetBooking(ctx, creds, ci.Environment, bookingRef)
}

func (s *carrierService) GetDocuments(ctx context.Context, orgID int64, integrationID int64, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error) {
	ci, adapter, creds, err := s.resolveIntegrationAndAdapter(ctx, orgID, integrationID, domain.CapDocuments)
	if err != nil {
		return nil, err
	}

	return adapter.GetDocuments(ctx, creds, ci.Environment, req)
}

func (s *carrierService) toView(ci *domain.CarrierIntegration) domain.CarrierIntegrationView {
	carrierName := ci.CarrierSCAC
	if ci.CarrierName != nil && *ci.CarrierName != "" {
		carrierName = *ci.CarrierName
	}

	maskedMap := make(map[string]string)
	if ci.CredentialMaskJSON != nil && *ci.CredentialMaskJSON != "" {
		_ = json.Unmarshal([]byte(*ci.CredentialMaskJSON), &maskedMap)
	}

	hasCreds := len(maskedMap) > 0 || (ci.EncryptedCredentials != nil && *ci.EncryptedCredentials != "") || (ci.CredentialsJSON != nil && *ci.CredentialsJSON != "")

	syncStatus := "Idle"
	if ci.SyncStatus != nil && *ci.SyncStatus != "" {
		syncStatus = *ci.SyncStatus
	}

	healthState, healthReason := s.syncEngine.computeHealth(ci)
	isSyncing := ci.SyncStatus != nil && (*ci.SyncStatus == "Syncing" || *ci.SyncStatus == "Running")

	adapter := s.registry.GetAdapterBySCAC(ci.CarrierSCAC)
	supportedCaps := adapter.SupportedCapabilities()

	return domain.CarrierIntegrationView{
		ID:                  ci.ID,
		OrgID:               ci.OrgID,
		CarrierSCAC:         ci.CarrierSCAC,
		CarrierName:         carrierName,
		Environment:         ci.Environment,
		ConnectionMethod:    ci.ConnectionMethod,
		ConnectionStatus:    ci.ConnectionStatus,
		HealthState:         healthState,
		HealthReason:        healthReason,
		IsEnabled:           ci.IsEnabled,
		Capabilities:        ci.Capabilities,
		SupportedCaps:       supportedCaps,
		CredentialMask:      maskedMap,
		HasCredentials:      hasCreds,
		SyncStatus:          syncStatus,
		IsSyncing:           isSyncing,
		LastSyncedAt:        ci.LastSyncedAt,
		LastSuccessAt:       ci.LastSuccessAt,
		LastFailureAt:       ci.LastFailureAt,
		LastError:           ci.LastError,
		FailedAttempts:      ci.FailedAttempts,
		ConsecutiveFailures: ci.FailedAttempts,
		CreatedAt:           ci.CreatedAt,
		UpdatedAt:           ci.UpdatedAt,
	}
}


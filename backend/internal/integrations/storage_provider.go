package integrations

import (
	"context"
	"time"
)

// StorageProvider defines the vendor-agnostic contract for object/document storage (e.g. AWS S3).
type StorageProvider interface {
	ProviderName() ProviderName
	Upload(ctx context.Context, orgID int64, key string, data []byte, contentType string) (*UploadResponse, error)
	Download(ctx context.Context, orgID int64, key string) ([]byte, string, error)
	Exists(ctx context.Context, orgID int64, key string) (bool, error)
	Delete(ctx context.Context, orgID int64, key string) error
	Status(ctx context.Context, orgID int64) (*ProviderHealth, error)
}

// UnconfiguredStorageProvider provides an honest, fail-safe implementation when S3 is disabled or unconfigured.
// It strictly returns ErrCodeProviderNotConfigured and NEVER generates fake upload locations.
type UnconfiguredStorageProvider struct {
	Name ProviderName
}

func NewUnconfiguredStorageProvider(name ProviderName) StorageProvider {
	if name == "" {
		name = ProviderUnconfigured
	}
	return &UnconfiguredStorageProvider{Name: name}
}

func (p *UnconfiguredStorageProvider) ProviderName() ProviderName {
	return p.Name
}

func (p *UnconfiguredStorageProvider) Upload(ctx context.Context, orgID int64, key string, data []byte, contentType string) (*UploadResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "Storage")
}

func (p *UnconfiguredStorageProvider) Download(ctx context.Context, orgID int64, key string) ([]byte, string, error) {
	return nil, "", NewProviderNotConfiguredError(string(p.Name), "Storage")
}

func (p *UnconfiguredStorageProvider) Exists(ctx context.Context, orgID int64, key string) (bool, error) {
	return false, NewProviderNotConfiguredError(string(p.Name), "Storage")
}

func (p *UnconfiguredStorageProvider) Delete(ctx context.Context, orgID int64, key string) error {
	return NewProviderNotConfiguredError(string(p.Name), "Storage")
}

func (p *UnconfiguredStorageProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	return &ProviderHealth{
		Provider:       p.Name,
		Type:           TypeStorage,
		Status:         StatusNotConfigured,
		Message:        "Storage provider is not configured. Live AWS S3 object store is disabled.",
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}

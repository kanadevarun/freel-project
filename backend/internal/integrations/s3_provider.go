package integrations

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	EnvS3Bucket        = "S3_BUCKET"
	EnvS3Region        = "AWS_REGION"
	EnvS3Prefix        = "S3_PREFIX"
	MaxDocumentSize    = 25 * 1024 * 1024 // 25 MB
	DefaultPresignMins = 15
	MaxPresignMins     = 60
)

// Allowed MIME types for business document uploads
var allowedMIMETypes = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/tiff":      true,
	"text/plain":      true,
	"text/csv":        true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-excel": true,
}

// Executable binary magic byte prefixes to reject unconditionally
var dangerousExecutableSignatures = [][]byte{
	{0x4D, 0x5A},             // Windows PE / DOS MZ ("MZ")
	{0x7F, 0x45, 0x4C, 0x46}, // Linux ELF ("\x7fELF")
	{0xCA, 0xFE, 0xBA, 0xBE}, // Mach-O / Java bytecode
	{0xCF, 0xFA, 0xED, 0xFE}, // Mach-O 64-bit
	{0xFE, 0xED, 0xFA, 0xCF}, // Mach-O 64-bit
	{0x23, 0x21},             // Shebang ("#!")
}

// S3ResolvedConfig holds resolved AWS S3 bucket and credential settings.
type S3ResolvedConfig struct {
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	Prefix          string `json:"prefix"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	IsEnabled       bool   `json:"is_enabled"`
	TimeoutSec      int    `json:"timeout_sec"`
}

// S3StorageProvider implements StorageProvider for AWS S3 with strict multi-tenant isolation.
type S3StorageProvider struct {
	db         *sqlx.DB
	configRepo ConfigRepository
	httpClient *ResilientHTTPClient
}

// NewS3StorageProvider constructs an AWS S3 storage provider adapter.
func NewS3StorageProvider(db *sqlx.DB, configRepo ConfigRepository, httpClient *ResilientHTTPClient) *S3StorageProvider {
	return &S3StorageProvider{
		db:         db,
		configRepo: configRepo,
		httpClient: httpClient,
	}
}

func (p *S3StorageProvider) ProviderName() ProviderName {
	return ProviderAWSS3
}

// ResolveConfig resolves S3 configuration prioritizing tenant database settings over environment fallbacks.
func (p *S3StorageProvider) ResolveConfig(ctx context.Context, orgID int64) (*S3ResolvedConfig, error) {
	cfg := &S3ResolvedConfig{
		Region:     "ap-south-1",
		TimeoutSec: 15,
	}

	// 1. Try tenant database configuration if present
	if p.configRepo != nil && orgID > 0 {
		tenantCfg, err := p.configRepo.GetConfig(ctx, orgID, TypeStorage, ProviderAWSS3)
		if err == nil && tenantCfg != nil {
			var nonSec struct {
				Bucket     string `json:"endpoint_url"` // or bucket in config JSON
				Region     string `json:"region"`
				TimeoutSec int    `json:"timeout_sec"`
			}
			var sec struct {
				AccessKeyID     string `json:"access_key_id"`
				SecretAccessKey string `json:"secret_value"`
			}
			_ = json.Unmarshal([]byte(tenantCfg.ConfigJSON), &nonSec)
			_ = json.Unmarshal([]byte(tenantCfg.EncryptedSecret), &sec)

			if nonSec.Bucket != "" {
				cfg.Bucket = nonSec.Bucket
			}
			if nonSec.Region != "" {
				cfg.Region = nonSec.Region
			}
			if nonSec.TimeoutSec > 0 {
				cfg.TimeoutSec = nonSec.TimeoutSec
			}
			if sec.AccessKeyID != "" && sec.AccessKeyID != "••••••••" {
				cfg.AccessKeyID = sec.AccessKeyID
			}
			if sec.SecretAccessKey != "" && sec.SecretAccessKey != "••••••••" {
				cfg.SecretAccessKey = sec.SecretAccessKey
			}
			cfg.IsEnabled = tenantCfg.IsEnabled
		}
	}

	// 2. Fall back to environment configuration
	if cfg.Bucket == "" {
		cfg.Bucket = os.Getenv(EnvS3Bucket)
	}
	if cfg.Region == "" || cfg.Region == "ap-south-1" {
		if envReg := os.Getenv(EnvS3Region); envReg != "" {
			cfg.Region = envReg
		}
	}
	if cfg.Prefix == "" {
		cfg.Prefix = os.Getenv(EnvS3Prefix)
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = os.Getenv(EnvAWSAccessKeyID)
	}
	if cfg.SecretAccessKey == "" {
		cfg.SecretAccessKey = os.Getenv(EnvAWSSecretKey)
	}

	// Environment feature flag
	envEnabled := os.Getenv(EnvS3Enabled)
	if strings.EqualFold(envEnabled, "true") || envEnabled == "1" {
		cfg.IsEnabled = true
	}

	return cfg, nil
}

// BuildTenantKey validates, cleans, and namespaces an S3 object key strictly by tenant organization.
func BuildTenantKey(orgID int64, key string) (string, error) {
	if orgID <= 0 {
		return "", NewInvalidRequestError("AWS_S3", "valid organization ID is required for storage operations")
	}

	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", NewInvalidRequestError("AWS_S3", "storage key cannot be empty")
	}

	// Reject null bytes, backslashes, and path traversal attempts
	if strings.Contains(trimmed, "\x00") || strings.Contains(trimmed, "\\") {
		return "", NewInvalidRequestError("AWS_S3", "invalid characters in storage key")
	}

	// Clean path
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." || strings.HasPrefix(cleaned, "/") {
		return "", NewInvalidRequestError("AWS_S3", "path traversal attempt detected in storage key")
	}

	expectedPrefix := fmt.Sprintf("orgs/%d/", orgID)

	// Check if already prefixed
	if strings.HasPrefix(cleaned, "orgs/") {
		if strings.HasPrefix(cleaned, expectedPrefix) {
			return cleaned, nil
		}
		// Attempting to access another tenant's namespace
		return "", NewAuthorizationFailedError("AWS_S3", fmt.Sprintf("cross-tenant storage access forbidden for org %d", orgID))
	}

	// Clean legacy uploads prefix if present
	cleaned = strings.TrimPrefix(cleaned, "uploads/")
	cleaned = strings.TrimPrefix(cleaned, "/")

	return expectedPrefix + cleaned, nil
}

// ValidateFileContent validates size, detects MIME type, and guards against malicious files.
func ValidateFileContent(data []byte, filename string) (string, error) {
	if len(data) == 0 {
		return "", NewInvalidRequestError("AWS_S3", "uploaded file cannot be empty (0 bytes)")
	}

	if len(data) > MaxDocumentSize {
		return "", NewInvalidRequestError("AWS_S3", fmt.Sprintf("file size exceeds maximum allowed limit of %d MB", MaxDocumentSize/(1024*1024)))
	}

	// Guard against executable binaries
	for _, sig := range dangerousExecutableSignatures {
		if len(data) >= len(sig) && bytes.Equal(data[:len(sig)], sig) {
			return "", NewInvalidRequestError("AWS_S3", "file rejected: executable binaries and scripts are not permitted")
		}
	}

	// Sniff MIME type from initial 512 bytes
	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffedMIME := http.DetectContentType(data[:sniffLen])
	baseMIME := strings.Split(sniffedMIME, ";")[0]
	baseMIME = strings.TrimSpace(strings.ToLower(baseMIME))

	// Check against extension if sniffed is generic octet-stream
	ext := strings.ToLower(filepath.Ext(filename))
	if baseMIME == "application/octet-stream" {
		switch ext {
		case ".pdf":
			baseMIME = "application/pdf"
		case ".csv":
			baseMIME = "text/csv"
		case ".xlsx":
			baseMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case ".xls":
			baseMIME = "application/vnd.ms-excel"
		case ".png":
			baseMIME = "image/png"
		case ".jpg", ".jpeg":
			baseMIME = "image/jpeg"
		case ".tiff", ".tif":
			baseMIME = "image/tiff"
		}
	}

	if !allowedMIMETypes[baseMIME] {
		return "", NewUnsupportedDocumentError("AWS_S3", fmt.Sprintf("unsupported MIME type %q (extension %s)", baseMIME, ext))
	}

	return baseMIME, nil
}

// buildS3Client initializes an AWS S3 client using resolved configuration.
func (p *S3StorageProvider) buildS3Client(ctx context.Context, cfg *S3ResolvedConfig) (*s3.Client, error) {
	if cfg.Bucket == "" || !cfg.IsEnabled {
		return nil, NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	var optFns []func(*awsConfig.LoadOptions) error
	if cfg.Region != "" {
		optFns = append(optFns, awsConfig.WithRegion(cfg.Region))
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		optFns = append(optFns, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, NewConnectionFailedError("AWS_S3", fmt.Sprintf("failed to load AWS configuration: %v", err))
	}

	return s3.NewFromConfig(awsCfg), nil
}

// Upload stores an object in S3 with strict tenant prefixing and file validation.
func (p *S3StorageProvider) Upload(ctx context.Context, orgID int64, key string, data []byte, contentType string) (*UploadResponse, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !cfg.IsEnabled || cfg.Bucket == "" {
		return nil, NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	// Validate content
	detectedMIME, err := ValidateFileContent(data, key)
	if err != nil {
		return nil, err
	}
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedMIME
	}

	// Scope tenant key
	tenantKey, err := BuildTenantKey(orgID, key)
	if err != nil {
		return nil, err
	}

	client, err := p.buildS3Client(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Compute SHA-256 checksum for audit & integrity
	hash := sha256.Sum256(data)
	etag := hex.EncodeToString(hash[:])

	// PutObject with AES-256 server-side encryption
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	uploadCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err = client.PutObject(uploadCtx, &s3.PutObjectInput{
		Bucket:               aws.String(cfg.Bucket),
		Key:                  aws.String(tenantKey),
		Body:                 bytes.NewReader(data),
		ContentType:          aws.String(contentType),
		ContentLength:        aws.Int64(int64(len(data))),
		ServerSideEncryption: s3types.ServerSideEncryptionAes256,
		Metadata: map[string]string{
			"organization_id": fmt.Sprintf("%d", orgID),
			"sha256":          etag,
		},
	})
	if err != nil {
		return nil, normalizeS3Error(err, cfg.Bucket, tenantKey)
	}

	return &UploadResponse{
		Key:           tenantKey,
		Location:      fmt.Sprintf("s3://%s/%s", cfg.Bucket, tenantKey),
		ETag:          etag,
		Size:          int64(len(data)),
		UploadedAt:    time.Now().UTC(),
		CorrelationID: uuid.NewString(),
	}, nil
}

// Download retrieves an object from S3 after verifying tenant ownership.
func (p *S3StorageProvider) Download(ctx context.Context, orgID int64, key string) ([]byte, string, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, "", err
	}

	if !cfg.IsEnabled || cfg.Bucket == "" {
		return nil, "", NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	tenantKey, err := BuildTenantKey(orgID, key)
	if err != nil {
		return nil, "", err
	}

	client, err := p.buildS3Client(ctx, cfg)
	if err != nil {
		return nil, "", err
	}

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	downloadCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	out, err := client.GetObject(downloadCtx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(tenantKey),
	})
	if err != nil {
		return nil, "", normalizeS3Error(err, cfg.Bucket, tenantKey)
	}
	defer out.Body.Close()

	buf, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read S3 object body: %w", err)
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil && *out.ContentType != "" {
		contentType = *out.ContentType
	}

	return buf, contentType, nil
}

// Exists checks if an object exists in the tenant's S3 namespace.
func (p *S3StorageProvider) Exists(ctx context.Context, orgID int64, key string) (bool, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return false, err
	}

	if !cfg.IsEnabled || cfg.Bucket == "" {
		return false, NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	tenantKey, err := BuildTenantKey(orgID, key)
	if err != nil {
		return false, err
	}

	client, err := p.buildS3Client(ctx, cfg)
	if err != nil {
		return false, err
	}

	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(tenantKey),
	})
	if err != nil {
		var nsk *s3types.NoSuchKey
		var nf *s3types.NotFound
		if errors.As(err, &nsk) || errors.As(err, &nf) || strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, normalizeS3Error(err, cfg.Bucket, tenantKey)
	}

	return true, nil
}

// Delete removes an object from S3 within the tenant's namespace.
func (p *S3StorageProvider) Delete(ctx context.Context, orgID int64, key string) error {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return err
	}

	if !cfg.IsEnabled || cfg.Bucket == "" {
		return NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	tenantKey, err := BuildTenantKey(orgID, key)
	if err != nil {
		return err
	}

	client, err := p.buildS3Client(ctx, cfg)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(tenantKey),
	})
	if err != nil {
		return normalizeS3Error(err, cfg.Bucket, tenantKey)
	}

	return nil
}

// GetPresignedURL generates a secure, short-lived presigned URL for downloading an authorized document.
func (p *S3StorageProvider) GetPresignedURL(ctx context.Context, orgID int64, key string, durationMinutes int) (string, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return "", err
	}

	if !cfg.IsEnabled || cfg.Bucket == "" {
		return "", NewProviderNotConfiguredError("AWS_S3", "Storage")
	}

	tenantKey, err := BuildTenantKey(orgID, key)
	if err != nil {
		return "", err
	}

	client, err := p.buildS3Client(ctx, cfg)
	if err != nil {
		return "", err
	}

	if durationMinutes <= 0 {
		durationMinutes = DefaultPresignMins
	}
	if durationMinutes > MaxPresignMins {
		durationMinutes = MaxPresignMins
	}

	presignClient := s3.NewPresignClient(client)
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(tenantKey),
	}, s3.WithPresignExpires(time.Duration(durationMinutes)*time.Minute))
	if err != nil {
		return "", normalizeS3Error(err, cfg.Bucket, tenantKey)
	}

	return presignedReq.URL, nil
}

// Status returns runtime health for the S3 integration.
func (p *S3StorageProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return &ProviderHealth{
			Provider:       ProviderAWSS3,
			Type:           TypeStorage,
			Status:         StatusConfigInvalid,
			Message:        err.Error(),
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	if !cfg.IsEnabled {
		return &ProviderHealth{
			Provider:       ProviderAWSS3,
			Type:           TypeStorage,
			Status:         StatusDisabled,
			Message:        "AWS S3 document storage is disabled by configuration.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	if cfg.Bucket == "" {
		return &ProviderHealth{
			Provider:       ProviderAWSS3,
			Type:           TypeStorage,
			Status:         StatusNotConfigured,
			Message:        "AWS S3 bucket is not configured. Live S3 storage is disabled.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	return &ProviderHealth{
		Provider:       ProviderAWSS3,
		Type:           TypeStorage,
		Status:         StatusHealthy,
		Message:        fmt.Sprintf("AWS S3 bucket %q configured in region %s", cfg.Bucket, cfg.Region),
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}

// normalizeS3Error converts raw AWS S3 SDK errors into normalized integration errors.
func normalizeS3Error(err error, bucket, key string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	var nsk *s3types.NoSuchKey
	var nf *s3types.NotFound
	if errors.As(err, &nsk) || errors.As(err, &nf) || strings.Contains(errMsg, "NoSuchKey") || strings.Contains(errMsg, "404") {
		return NewObjectNotFoundError("AWS_S3", key)
	}

	if strings.Contains(errMsg, "AccessDenied") || strings.Contains(errMsg, "403") {
		return NewAuthorizationFailedError("AWS_S3", fmt.Sprintf("access denied for bucket %s / object %s", bucket, key))
	}

	if strings.Contains(errMsg, "SlowDown") || strings.Contains(errMsg, "503") || strings.Contains(errMsg, "RateLimit") {
		return NewRateLimitedError("AWS_S3", 5)
	}

	if strings.Contains(errMsg, "context deadline exceeded") {
		return NewTimeoutError("AWS_S3", "15s")
	}

	return NewConnectionFailedError("AWS_S3", fmt.Sprintf("S3 request failed: %s", errMsg))
}

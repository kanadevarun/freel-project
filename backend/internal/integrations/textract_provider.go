package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textracttypes "github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Supported document MIME types for AWS Textract
var textractSupportedFormats = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/tiff":      true,
}

// TextractRequest encapsulates an OCR / document extraction task.
type TextractRequest struct {
	DocumentBytes  []byte            `json:"-"`
	S3Bucket       string            `json:"s3_bucket,omitempty"`
	S3Key          string            `json:"s3_key,omitempty"`
	DocumentType   string            `json:"document_type"`
	MIMEType       string            `json:"mime_type,omitempty"`
	FeatureTypes   []string          `json:"feature_types,omitempty"` // TABLES, FORMS
	CorrelationID  string            `json:"correlation_id"`
	IdempotencyKey string            `json:"idempotency_key"`
}

// TextractResponse holds normalized OCR and structured extraction data.
type TextractResponse struct {
	Status        string                 `json:"status"` // COMPLETED, FAILED, NOT_CONFIGURED, UNSUPPORTED
	RawText       string                 `json:"raw_text"`
	Lines         []string               `json:"lines"`
	Confidence    float64                `json:"confidence"`
	KeyValues     map[string]string      `json:"key_values,omitempty"`
	Tables        [][]string             `json:"tables,omitempty"`
	CorrelationID string                 `json:"correlation_id"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TextractProvider defines the vendor-agnostic interface for OCR document extraction.
type TextractProvider interface {
	ProviderName() ProviderName
	ExtractDocumentText(ctx context.Context, orgID int64, req TextractRequest) (*TextractResponse, error)
	Status(ctx context.Context, orgID int64) (*ProviderHealth, error)
}

// TextractResolvedConfig holds resolved credentials and operational settings for Textract.
type TextractResolvedConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	IsEnabled       bool   `json:"is_enabled"`
	TimeoutSec      int    `json:"timeout_sec"`
}

// AWSTextractProvider implements TextractProvider using AWS SDK v2.
type AWSTextractProvider struct {
	db         *sqlx.DB
	configRepo ConfigRepository
	httpClient *ResilientHTTPClient
}

// NewAWSTextractProvider constructs an AWS Textract provider adapter.
func NewAWSTextractProvider(db *sqlx.DB, configRepo ConfigRepository, httpClient *ResilientHTTPClient) *AWSTextractProvider {
	return &AWSTextractProvider{
		db:         db,
		configRepo: configRepo,
		httpClient: httpClient,
	}
}

func (p *AWSTextractProvider) ProviderName() ProviderName {
	return ProviderAWSTextract
}

// ResolveConfig resolves Textract configuration prioritizing tenant DB over environment variables.
func (p *AWSTextractProvider) ResolveConfig(ctx context.Context, orgID int64) (*TextractResolvedConfig, error) {
	cfg := &TextractResolvedConfig{
		Region:     "ap-south-1",
		TimeoutSec: 20,
	}

	// 1. Try tenant database configuration if present
	if p.configRepo != nil && orgID > 0 {
		tenantCfg, err := p.configRepo.GetConfig(ctx, orgID, TypeTextract, ProviderAWSTextract)
		if err == nil && tenantCfg != nil {
			var nonSec struct {
				Region     string `json:"region"`
				TimeoutSec int    `json:"timeout_sec"`
			}
			var sec struct {
				AccessKeyID     string `json:"access_key_id"`
				SecretAccessKey string `json:"secret_value"`
			}
			_ = json.Unmarshal([]byte(tenantCfg.ConfigJSON), &nonSec)
			_ = json.Unmarshal([]byte(tenantCfg.EncryptedSecret), &sec)

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
	if cfg.Region == "" || cfg.Region == "ap-south-1" {
		if envReg := os.Getenv(EnvAWSRegion); envReg != "" {
			cfg.Region = envReg
		}
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = os.Getenv(EnvAWSAccessKeyID)
	}
	if cfg.SecretAccessKey == "" {
		cfg.SecretAccessKey = os.Getenv(EnvAWSSecretKey)
	}

	envEnabled := os.Getenv(EnvTextractEnabled)
	if strings.EqualFold(envEnabled, "true") || envEnabled == "1" {
		cfg.IsEnabled = true
	}

	return cfg, nil
}

// buildClient initializes AWS Textract client using resolved configuration.
func (p *AWSTextractProvider) buildClient(ctx context.Context, cfg *TextractResolvedConfig) (*textract.Client, error) {
	if !cfg.IsEnabled || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, NewProviderNotConfiguredError("AWS_TEXTRACT", "OCR")
	}

	var optFns []func(*awsConfig.LoadOptions) error
	if cfg.Region != "" {
		optFns = append(optFns, awsConfig.WithRegion(cfg.Region))
	}
	optFns = append(optFns, awsConfig.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	))

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, NewConnectionFailedError("AWS_TEXTRACT", fmt.Sprintf("failed to load AWS configuration: %v", err))
	}

	return textract.NewFromConfig(awsCfg), nil
}

// ExtractDocumentText executes synchronous text detection and structure extraction.
func (p *AWSTextractProvider) ExtractDocumentText(ctx context.Context, orgID int64, req TextractRequest) (*TextractResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("AWS_TEXTRACT", "valid organization ID is required for OCR operations")
	}

	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !cfg.IsEnabled {
		return nil, NewProviderNotConfiguredError("AWS_TEXTRACT", "OCR")
	}

	// Validate document input
	if len(req.DocumentBytes) == 0 && (req.S3Bucket == "" || req.S3Key == "") {
		return nil, NewInvalidRequestError("AWS_TEXTRACT", "either document bytes or S3 bucket/key reference must be provided")
	}

	// Format validation
	mime := strings.ToLower(strings.TrimSpace(req.MIMEType))
	if mime != "" && !textractSupportedFormats[mime] {
		return nil, NewUnsupportedDocumentError("AWS_TEXTRACT", fmt.Sprintf("unsupported MIME type %q (AWS Textract supports PDF, PNG, JPEG, TIFF)", mime))
	}

	client, err := p.buildClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = uuid.NewString()
	}

	// Prepare Document payload
	docInput := &textracttypes.Document{}
	if len(req.DocumentBytes) > 0 {
		docInput.Bytes = req.DocumentBytes
	} else {
		// Scoped tenant S3 key validation
		scopedKey, err := BuildTenantKey(orgID, req.S3Key)
		if err != nil {
			return nil, err
		}
		docInput.S3Object = &textracttypes.S3Object{
			Bucket: aws.String(req.S3Bucket),
			Name:   aws.String(scopedKey),
		}
	}

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// If feature types requested (FORMS or TABLES), use AnalyzeDocument, otherwise DetectDocumentText
	if len(req.FeatureTypes) > 0 {
		var features []textracttypes.FeatureType
		for _, f := range req.FeatureTypes {
			switch strings.ToUpper(f) {
			case "TABLES":
				features = append(features, textracttypes.FeatureTypeTables)
			case "FORMS":
				features = append(features, textracttypes.FeatureTypeForms)
			}
		}

		out, err := client.AnalyzeDocument(execCtx, &textract.AnalyzeDocumentInput{
			Document:     docInput,
			FeatureTypes: features,
		})
		if err != nil {
			return nil, normalizeTextractError(err)
		}
		return parseTextractBlocks(out.Blocks, corrID), nil
	}

	out, err := client.DetectDocumentText(execCtx, &textract.DetectDocumentTextInput{
		Document: docInput,
	})
	if err != nil {
		return nil, normalizeTextractError(err)
	}

	return parseTextractBlocks(out.Blocks, corrID), nil
}

// parseTextractBlocks transforms AWS Textract blocks into a clean, normalized response.
func parseTextractBlocks(blocks []textracttypes.Block, corrID string) *TextractResponse {
	var lines []string
	var totalConfidence float64
	var lineCount int

	blockMap := make(map[string]textracttypes.Block)
	keyMap := make(map[string]textracttypes.Block)
	valMap := make(map[string]textracttypes.Block)

	for _, b := range blocks {
		if b.Id != nil {
			blockMap[*b.Id] = b
		}

		if b.BlockType == textracttypes.BlockTypeLine && b.Text != nil {
			lines = append(lines, *b.Text)
			if b.Confidence != nil {
				totalConfidence += float64(*b.Confidence)
				lineCount++
			}
		}

		if b.BlockType == textracttypes.BlockTypeKeyValueSet {
			for _, et := range b.EntityTypes {
				if et == textracttypes.EntityTypeKey && b.Id != nil {
					keyMap[*b.Id] = b
				} else if et == textracttypes.EntityTypeValue && b.Id != nil {
					valMap[*b.Id] = b
				}
			}
		}
	}

	avgConfidence := 0.0
	if lineCount > 0 {
		avgConfidence = totalConfidence / float64(lineCount)
	}

	// Reconstruct KEY -> VALUE relationships
	keyValues := make(map[string]string)
	for _, keyBlock := range keyMap {
		keyText := getTextFromRelationships(keyBlock, blockMap)
		valText := ""
		for _, rel := range keyBlock.Relationships {
			if rel.Type == textracttypes.RelationshipTypeValue {
				for _, valID := range rel.Ids {
					if vb, ok := valMap[valID]; ok {
						valText = getTextFromRelationships(vb, blockMap)
					}
				}
			}
		}
		if keyText != "" && valText != "" {
			keyValues[strings.TrimSpace(keyText)] = strings.TrimSpace(valText)
		}
	}

	return &TextractResponse{
		Status:        "COMPLETED",
		RawText:       strings.Join(lines, "\n"),
		Lines:         lines,
		Confidence:    avgConfidence,
		KeyValues:     keyValues,
		CorrelationID: corrID,
		Metadata: map[string]interface{}{
			"line_count":  len(lines),
			"block_count": len(blocks),
		},
	}
}

// getTextFromRelationships extracts concatenated word tokens for a block.
func getTextFromRelationships(b textracttypes.Block, blockMap map[string]textracttypes.Block) string {
	var words []string
	for _, rel := range b.Relationships {
		if rel.Type == textracttypes.RelationshipTypeChild {
			for _, childID := range rel.Ids {
				if child, ok := blockMap[childID]; ok && child.BlockType == textracttypes.BlockTypeWord && child.Text != nil {
					words = append(words, *child.Text)
				}
			}
		}
	}
	return strings.Join(words, " ")
}

// Status returns runtime health for AWS Textract.
func (p *AWSTextractProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return &ProviderHealth{
			Provider:       ProviderAWSTextract,
			Type:           TypeTextract,
			Status:         StatusConfigInvalid,
			Message:        err.Error(),
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	if !cfg.IsEnabled {
		return &ProviderHealth{
			Provider:       ProviderAWSTextract,
			Type:           TypeTextract,
			Status:         StatusDisabled,
			Message:        "AWS Textract OCR is disabled by configuration.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return &ProviderHealth{
			Provider:       ProviderAWSTextract,
			Type:           TypeTextract,
			Status:         StatusNotConfigured,
			Message:        "AWS Textract credentials are not configured. OCR extraction is inactive.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	return &ProviderHealth{
		Provider:       ProviderAWSTextract,
		Type:           TypeTextract,
		Status:         StatusHealthy,
		Message:        fmt.Sprintf("AWS Textract configured in region %s", cfg.Region),
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}

// normalizeTextractError maps AWS Textract SDK errors to normalized IntegrationErrors.
func normalizeTextractError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	if strings.Contains(errMsg, "UnsupportedDocumentException") {
		return NewUnsupportedDocumentError("AWS_TEXTRACT", "document format is not supported by Textract")
	}
	if strings.Contains(errMsg, "DocumentTooLargeException") {
		return NewInvalidRequestError("AWS_TEXTRACT", "document exceeds maximum allowed size for AWS Textract")
	}
	if strings.Contains(errMsg, "BadDocumentException") || strings.Contains(errMsg, "InvalidParameterException") {
		return NewInvalidRequestError("AWS_TEXTRACT", fmt.Sprintf("invalid document parameters: %s", errMsg))
	}
	if strings.Contains(errMsg, "AccessDeniedException") || strings.Contains(errMsg, "403") {
		return NewAuthorizationFailedError("AWS_TEXTRACT", "AWS Textract access denied: check IAM permissions")
	}
	if strings.Contains(errMsg, "ProvisionedThroughputExceededException") || strings.Contains(errMsg, "ThrottlingException") {
		return NewRateLimitedError("AWS_TEXTRACT", 5)
	}
	if strings.Contains(errMsg, "context deadline exceeded") {
		return NewTimeoutError("AWS_TEXTRACT", "20s")
	}

	return NewConnectionFailedError("AWS_TEXTRACT", fmt.Sprintf("Textract OCR request failed: %s", errMsg))
}

// UnconfiguredTextractProvider provides an honest, fail-safe implementation when Textract is unconfigured.
type UnconfiguredTextractProvider struct {
	Name ProviderName
}

func NewUnconfiguredTextractProvider(name ProviderName) TextractProvider {
	if name == "" {
		name = ProviderAWSTextract
	}
	return &UnconfiguredTextractProvider{Name: name}
}

func (p *UnconfiguredTextractProvider) ProviderName() ProviderName {
	return p.Name
}

func (p *UnconfiguredTextractProvider) ExtractDocumentText(ctx context.Context, orgID int64, req TextractRequest) (*TextractResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "OCR")
}

func (p *UnconfiguredTextractProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	return &ProviderHealth{
		Provider:       p.Name,
		Type:           TypeTextract,
		Status:         StatusNotConfigured,
		Message:        "AWS Textract OCR provider is not configured. Live OCR extraction is disabled.",
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}

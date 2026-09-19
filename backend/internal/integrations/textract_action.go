package integrations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

// ExtractTextActionInput defines input for textract.extract_text.
type ExtractTextActionInput struct {
	S3Bucket       string   `json:"s3_bucket,omitempty"`
	S3Key          string   `json:"s3_key,omitempty"`
	DataBase64     string   `json:"data_base64,omitempty"`
	DocumentType   string   `json:"document_type"`
	MIMEType       string   `json:"mime_type,omitempty"`
	FeatureTypes   []string `json:"feature_types,omitempty"`
	CorrelationID  string   `json:"correlation_id,omitempty"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
}

// ExtractTextAction integrates AWS Textract OCR into the LogisticsHQ Action System.
type ExtractTextAction struct {
	service GatewayService
}

func NewExtractTextAction(service GatewayService) actions.Action {
	return &ExtractTextAction{service: service}
}

func (a *ExtractTextAction) Name() string {
	return "textract.extract_text"
}

func (a *ExtractTextAction) Module() string {
	return "textract"
}

func (a *ExtractTextAction) Description() string {
	return "Extract normalized text, key-value pairs, and tables from commercial documents via AWS Textract."
}

func (a *ExtractTextAction) Category() actions.ActionCategory {
	return actions.ActionCategoryWrite
}

func (a *ExtractTextAction) InputSchema() interface{} {
	return &ExtractTextActionInput{}
}

func (a *ExtractTextAction) RequiresConfirmation() bool {
	return false
}

func (a *ExtractTextAction) RequiredPermission() (string, string) {
	return rbac.ResourceDocuments, rbac.ActionUpdate
}

func (a *ExtractTextAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in ExtractTextActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid textract action payload: %v", err),
			},
		}, nil
	}

	var docBytes []byte
	if in.DataBase64 != "" {
		b, err := base64.StdEncoding.DecodeString(in.DataBase64)
		if err != nil {
			return &actions.ActionResult{
				Success: false,
				Action:  a.Name(),
				Error: &actions.ActionError{
					Type:    "Validation",
					Message: fmt.Sprintf("invalid base64 document content: %v", err),
				},
			}, nil
		}
		docBytes = b
	}

	if len(docBytes) == 0 && in.S3Key == "" {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: "either data_base64 or s3_key must be provided",
			},
		}, nil
	}

	var orgID int64
	if ctx != nil {
		orgID = ctx.OrganizationID
	}

	req := TextractRequest{
		DocumentBytes:  docBytes,
		S3Bucket:       in.S3Bucket,
		S3Key:          in.S3Key,
		DocumentType:   in.DocumentType,
		MIMEType:       in.MIMEType,
		FeatureTypes:   in.FeatureTypes,
		CorrelationID:  in.CorrelationID,
		IdempotencyKey: in.IdempotencyKey,
	}

	resp, err := a.service.ExtractDocumentText(ctx.Context, orgID, req)
	if err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "ExecutionFailed",
				Message: err.Error(),
			},
		}, nil
	}

	return &actions.ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "TextractOCR",
		ResourceID:   in.S3Key,
		Summary:      fmt.Sprintf("Textract OCR completed: extracted %d lines (confidence: %.2f%%)", len(resp.Lines), resp.Confidence),
		Data: map[string]interface{}{
			"status":          resp.Status,
			"line_count":      len(resp.Lines),
			"confidence":      resp.Confidence,
			"key_value_count": len(resp.KeyValues),
			"raw_text":        resp.RawText,
			"key_values":      resp.KeyValues,
		},
	}, nil
}

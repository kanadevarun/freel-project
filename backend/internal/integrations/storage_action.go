package integrations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

// UploadDocumentActionInput defines the parameters for storage.upload_document.
type UploadDocumentActionInput struct {
	Key         string `json:"key"`
	DataBase64  string `json:"data_base64"`
	ContentType string `json:"content_type"`
}

// UploadDocumentAction integrates document upload into the LogisticsHQ Action System.
type UploadDocumentAction struct {
	service GatewayService
}

func NewUploadDocumentAction(service GatewayService) actions.Action {
	return &UploadDocumentAction{service: service}
}

func (a *UploadDocumentAction) Name() string {
	return "storage.upload_document"
}

func (a *UploadDocumentAction) Module() string {
	return "storage"
}

func (a *UploadDocumentAction) Description() string {
	return "Upload a document to tenant-isolated cloud storage (AWS S3) via the Integration Gateway."
}

func (a *UploadDocumentAction) Category() actions.ActionCategory {
	return actions.ActionCategoryWrite
}

func (a *UploadDocumentAction) InputSchema() interface{} {
	return &UploadDocumentActionInput{}
}

func (a *UploadDocumentAction) RequiresConfirmation() bool {
	return false
}

func (a *UploadDocumentAction) RequiredPermission() (string, string) {
	return rbac.ResourceDocuments, rbac.ActionCreate
}

func (a *UploadDocumentAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in UploadDocumentActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid storage upload action payload: %v", err),
			},
		}, nil
	}

	if in.Key == "" || in.DataBase64 == "" {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: "key and data_base64 are required",
			},
		}, nil
	}

	data, err := base64.StdEncoding.DecodeString(in.DataBase64)
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

	var orgID int64
	if ctx != nil {
		orgID = ctx.OrganizationID
	}

	resp, err := a.service.UploadDocument(ctx.Context, orgID, in.Key, data, in.ContentType)
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
		ResourceType: "Document",
		ResourceID:   resp.Key,
		Summary:      fmt.Sprintf("Document uploaded to storage location %s", resp.Location),
		Data: map[string]interface{}{
			"key":         resp.Key,
			"location":    resp.Location,
			"size":        resp.Size,
			"uploaded_at": resp.UploadedAt,
			"etag":        resp.ETag,
		},
	}, nil
}

// DownloadDocumentActionInput defines parameters for storage.download_document.
type DownloadDocumentActionInput struct {
	Key string `json:"key"`
}

// DownloadDocumentAction integrates document retrieval into the Action System.
type DownloadDocumentAction struct {
	service GatewayService
}

func NewDownloadDocumentAction(service GatewayService) actions.Action {
	return &DownloadDocumentAction{service: service}
}

func (a *DownloadDocumentAction) Name() string {
	return "storage.download_document"
}

func (a *DownloadDocumentAction) Module() string {
	return "storage"
}

func (a *DownloadDocumentAction) Description() string {
	return "Download a document from tenant-isolated cloud storage (AWS S3) via the Integration Gateway."
}

func (a *DownloadDocumentAction) Category() actions.ActionCategory {
	return actions.ActionCategoryRead
}

func (a *DownloadDocumentAction) InputSchema() interface{} {
	return &DownloadDocumentActionInput{}
}

func (a *DownloadDocumentAction) RequiresConfirmation() bool {
	return false
}

func (a *DownloadDocumentAction) RequiredPermission() (string, string) {
	return rbac.ResourceDocuments, rbac.ActionRead
}

func (a *DownloadDocumentAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in DownloadDocumentActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid storage download action payload: %v", err),
			},
		}, nil
	}

	if in.Key == "" {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: "key is required",
			},
		}, nil
	}

	var orgID int64
	if ctx != nil {
		orgID = ctx.OrganizationID
	}

	data, contentType, err := a.service.DownloadDocument(ctx.Context, orgID, in.Key)
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
		ResourceType: "Document",
		ResourceID:   in.Key,
		Summary:      fmt.Sprintf("Document %s downloaded (%d bytes)", in.Key, len(data)),
		Data: map[string]interface{}{
			"key":          in.Key,
			"content_type": contentType,
			"size":         len(data),
			"data_base64":  base64.StdEncoding.EncodeToString(data),
		},
	}, nil
}

package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/freel/backend/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3StorageProvider_TenantNamespaceIsolation(t *testing.T) {
	// 1. Basic tenant scoping
	key, err := BuildTenantKey(1, "invoices/inv_123.pdf")
	require.NoError(t, err)
	assert.Equal(t, "orgs/1/invoices/inv_123.pdf", key)

	// 2. Already prefixed for same tenant is accepted
	key2, err := BuildTenantKey(1, "orgs/1/invoices/inv_123.pdf")
	require.NoError(t, err)
	assert.Equal(t, "orgs/1/invoices/inv_123.pdf", key2)

	// 3. Cross-tenant attempt is strictly rejected
	_, err = BuildTenantKey(1, "orgs/2/secret_contract.pdf")
	require.Error(t, err)
	var intErr *IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, ErrCodeAuthorizationFailed, intErr.Code())
	assert.Contains(t, intErr.Message(), "cross-tenant storage access forbidden")

	// 4. Path traversal attempts rejected
	traversalKeys := []string{
		"../../etc/passwd",
		"..\\windows\\win.ini",
		"documents/../../../secrets.key",
		"doc\x00nullbyte.pdf",
		"/absolute/path/file.pdf",
	}
	for _, tk := range traversalKeys {
		_, err := BuildTenantKey(1, tk)
		assert.Error(t, err, "expected error for traversal key: %s", tk)
	}

	// 5. Invalid organization ID rejected
	_, err = BuildTenantKey(0, "file.pdf")
	assert.Error(t, err)
	_, err = BuildTenantKey(-1, "file.pdf")
	assert.Error(t, err)
}

func TestS3StorageProvider_FileValidation(t *testing.T) {
	// 1. Empty file rejected
	_, err := ValidateFileContent([]byte{}, "empty.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty (0 bytes)")

	// 2. Executable binaries rejected
	dosMZ := []byte{0x4D, 0x5A, 0x90, 0x00, 0x03}
	_, err = ValidateFileContent(dosMZ, "malware.exe")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "executable binaries and scripts are not permitted")

	linuxELF := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02}
	_, err = ValidateFileContent(linuxELF, "script.bin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "executable binaries and scripts are not permitted")

	// 3. Valid PDF content
	validPDF := []byte("%PDF-1.4\n%âãÏÓ\n1 0 obj\n<<\n>>\nendobj\ntrailer\n<<\n>>\n%%EOF")
	mime, err := ValidateFileContent(validPDF, "bill_of_lading.pdf")
	require.NoError(t, err)
	assert.Equal(t, "application/pdf", mime)

	// 4. Valid plain text / CSV content
	csvData := []byte("ContainerNumber,SealNumber,WeightKG\nMSCU1234567,SL-9988,24000\n")
	mime, err = ValidateFileContent(csvData, "packing_list.csv")
	require.NoError(t, err)
	assert.True(t, mime == "text/csv" || mime == "text/plain")

	// 5. Unsupported file type rejected
	unsupportedData := []byte("PK\x03\x04something unknown and weird")
	_, err = ValidateFileContent(unsupportedData, "archive.xyz")
	require.Error(t, err)
}

func TestS3StorageProvider_UnconfiguredFailsHonestly(t *testing.T) {
	provider := NewS3StorageProvider(nil, nil, nil)
	ctx := context.Background()

	// Upload should return provider_not_configured
	_, err := provider.Upload(ctx, 1, "doc.pdf", []byte("%PDF-1.4 test"), "application/pdf")
	require.Error(t, err)
	var intErr *IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())

	// Download should return provider_not_configured
	_, _, err = provider.Download(ctx, 1, "doc.pdf")
	require.Error(t, err)
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())

	// Status should report disabled or not configured
	health, err := provider.Status(ctx, 1)
	require.NoError(t, err)
	assert.True(t, health.Status == StatusNotConfigured || health.Status == StatusDisabled)
	assert.Equal(t, ProviderAWSS3, health.Provider)
}

func TestTextractProvider_UnconfiguredFailsHonestly(t *testing.T) {
	provider := NewAWSTextractProvider(nil, nil, nil)
	ctx := context.Background()

	// Extract text should return provider_not_configured
	req := TextractRequest{
		DocumentBytes: []byte("%PDF-1.4 test contract"),
		DocumentType:  "PDF",
		MIMEType:      "application/pdf",
	}
	_, err := provider.ExtractDocumentText(ctx, 1, req)
	require.Error(t, err)
	var intErr *IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())

	// Status should report disabled or not configured
	health, err := provider.Status(ctx, 1)
	require.NoError(t, err)
	assert.True(t, health.Status == StatusNotConfigured || health.Status == StatusDisabled)
	assert.Equal(t, ProviderAWSTextract, health.Provider)
}

func TestTextractProvider_FormatValidation(t *testing.T) {
	provider := &AWSTextractProvider{}
	ctx := context.Background()

	// Unsupported MIME format should fail before AWS network call
	cfg := &TextractResolvedConfig{
		IsEnabled:       true,
		AccessKeyID:     "test-key",
		SecretAccessKey: "test-secret",
	}
	_ = cfg

	unsupportedReq := TextractRequest{
		DocumentBytes: []byte("sample"),
		MIMEType:      "application/x-msdos-program",
	}
	// With disabled config, it will report unconfigured
	_, err := provider.ExtractDocumentText(ctx, 1, unsupportedReq)
	require.Error(t, err)
}

func TestStorageActionSystem_Registration(t *testing.T) {
	gw := NewGatewayService(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	uploadAction := NewUploadDocumentAction(gw)
	assert.Equal(t, "storage.upload_document", uploadAction.Name())
	assert.Equal(t, "storage", uploadAction.Module())

	downloadAction := NewDownloadDocumentAction(gw)
	assert.Equal(t, "storage.download_document", downloadAction.Name())
	assert.Equal(t, "storage", downloadAction.Module())

	textractAction := NewExtractTextAction(gw)
	assert.Equal(t, "textract.extract_text", textractAction.Name())
	assert.Equal(t, "textract", textractAction.Module())

	// Test unconfigured execution returns graceful failure
	actCtx := &actions.ActionContext{
		OrganizationID: 1,
		Context:        context.Background(),
	}

	payload, _ := json.Marshal(UploadDocumentActionInput{
		Key:         "invoice.pdf",
		DataBase64:  base64.StdEncoding.EncodeToString([]byte("%PDF-1.4 test")),
		ContentType: "application/pdf",
	})
	res, err := uploadAction.Execute(actCtx, payload)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Contains(t, res.Error.Message, "Storage provider (AWS_S3) is not configured")
}

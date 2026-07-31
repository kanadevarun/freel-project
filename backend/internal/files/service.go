package files

import (
	"context"
	"io"
)

// Service handles uploading and retrieving files (e.g., invoices, quotes).
type Service interface {
	// UploadFile saves a file to storage (like AWS S3).
	// Simple meaning: It takes a file you upload and saves it securely in the cloud.
	// Example: url, err := fileSvc.UploadFile(ctx, "invoice-123.pdf", fileStream)
	UploadFile(ctx context.Context, filename string, reader io.Reader) (string, error)

	// GetFileURL gets a temporary link to download a file.
	// Simple meaning: It generates a secure, clickable link so a user can download their document.
	// Example: link, err := fileSvc.GetFileURL(ctx, "invoice-123.pdf")
	GetFileURL(ctx context.Context, filename string) (string, error)
}

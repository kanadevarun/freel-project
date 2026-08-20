package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type localService struct {
	uploadDir string
	baseURL   string
}

// NewLocalService creates a files.Service that stores files on the local disk.
// Perfect for development, local testing, and mock setups.
func NewLocalService(uploadDir string, baseURL string) Service {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080/uploads"
	}
	return &localService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (s *localService) UploadFile(ctx context.Context, filename string, reader io.Reader) (string, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}

	// Clean/sanitize filename to prevent path traversal
	safeName := filepath.Base(filename)
	filePath := filepath.Join(s.uploadDir, safeName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("write file data: %w", err)
	}

	// For local mode, we return the relative path or base URL path
	return safeName, nil
}

func (s *localService) GetFileURL(ctx context.Context, filename string) (string, error) {
	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

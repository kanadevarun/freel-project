package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	// Clean/sanitize filename. If nested path is provided (e.g. organizations/1/branding/logo/...), preserve structured path.
	cleanName := filepath.Clean(filepath.ToSlash(filename))
	cleanName = strings.TrimPrefix(cleanName, "/")
	for strings.HasPrefix(cleanName, "../") || cleanName == ".." {
		cleanName = strings.TrimPrefix(cleanName, "../")
	}

	filePath := filepath.Join(s.uploadDir, filepath.FromSlash(cleanName))
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("create upload subdirectories: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("write file data: %w", err)
	}

	// For local mode, return the relative path
	return filepath.ToSlash(cleanName), nil
}

func (s *localService) GetFileURL(ctx context.Context, filename string) (string, error) {
	cleanName := filepath.ToSlash(filepath.Clean(filename))
	cleanName = strings.TrimPrefix(cleanName, "/")
	return fmt.Sprintf("%s/%s", s.baseURL, cleanName), nil
}

func (s *localService) DownloadFile(ctx context.Context, filename string) ([]byte, string, error) {
	cleanName := filepath.Clean(filename)
	safePath := filepath.Join(s.uploadDir, cleanName)

	// Ensure the path stays within uploadDir (path traversal defense)
	rel, err := filepath.Rel(s.uploadDir, safePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, "", fmt.Errorf("invalid path traversal attempt")
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, "", fmt.Errorf("file not found on disk: %w", err)
	}

	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}
	mimeType := http.DetectContentType(data[:sniffLen])

	return data, mimeType, nil
}

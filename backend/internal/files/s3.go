package files

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Service struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

// NewS3Service creates a files.Service that uploads files to an AWS S3 bucket and generates pre-signed download URLs.
func NewS3Service(client *s3.Client, bucketName string) Service {
	presignClient := s3.NewPresignClient(client)
	return &s3Service{
		client:        client,
		presignClient: presignClient,
		bucketName:    bucketName,
	}
}

func (s *s3Service) UploadFile(ctx context.Context, filename string, reader io.Reader) (string, error) {
	var key string
	cleanName := filepath.ToSlash(filepath.Clean(filename))
	cleanName = strings.TrimPrefix(cleanName, "/")
	for strings.HasPrefix(cleanName, "../") || cleanName == ".." {
		cleanName = strings.TrimPrefix(cleanName, "../")
	}

	if strings.HasPrefix(cleanName, "organizations/") {
		key = cleanName
	} else {
		safeName := filepath.Base(cleanName)
		key = fmt.Sprintf("uploads/%d-%s", time.Now().UnixNano(), safeName)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return key, nil
}

func (s *s3Service) GetFileURL(ctx context.Context, filename string) (string, error) {
	// Generate a secure pre-signed download URL valid for 15 minutes.
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filename),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return req.URL, nil
}

func (s *s3Service) DownloadFile(ctx context.Context, filename string) ([]byte, string, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read object body from S3: %w", err)
	}

	mimeType := "application/octet-stream"
	if out.ContentType != nil && *out.ContentType != "" {
		mimeType = *out.ContentType
	}

	return data, mimeType, nil
}

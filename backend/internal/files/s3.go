package files

import (
	"context"
	"fmt"
	"io"
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
	// Clean/sanitize filename to prevent path traversal. Save under unique timestamped key.
	key := fmt.Sprintf("uploads/%d-%s", time.Now().UnixNano(), filename)

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

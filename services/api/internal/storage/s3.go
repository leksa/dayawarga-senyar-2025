package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage handles S3-compatible storage operations
type S3Storage struct {
	client     *s3.Client
	bucket     string
	baseURL    string // Public URL for serving files
	pathPrefix string // Optional prefix for all keys
}

// S3Config holds S3 configuration
type S3Config struct {
	Endpoint        string // S3-compatible endpoint (e.g., is3.cloudhost.id)
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string // Default: auto
	PathPrefix      string // Optional: prefix for all keys (e.g., "photos/")
	UsePathStyle    bool   // For S3-compatible services, usually true
}

// NewS3Storage creates a new S3 storage client
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	if cfg.Region == "" {
		cfg.Region = "auto"
	}

	// Create custom resolver for S3-compatible endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s", cfg.Endpoint),
			SigningRegion:     cfg.Region,
			HostnameImmutable: true,
		}, nil
	})

	// Load AWS config with custom credentials and endpoint
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with path-style addressing for S3-compatible services
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	// Construct base URL for public access
	baseURL := fmt.Sprintf("https://%s.%s", cfg.Bucket, cfg.Endpoint)
	if cfg.UsePathStyle {
		baseURL = fmt.Sprintf("https://%s/%s", cfg.Endpoint, cfg.Bucket)
	}

	return &S3Storage{
		client:     client,
		bucket:     cfg.Bucket,
		baseURL:    baseURL,
		pathPrefix: cfg.PathPrefix,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	fullKey := s.buildKey(key)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fullKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         "public-read", // Make publicly readable
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s.GetPublicURL(key), nil
}

// UploadFromReader uploads from an io.Reader to S3
func (s *S3Storage) UploadFromReader(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	// Read all data (S3 SDK requires knowing content length or using multipart)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	return s.Upload(ctx, key, data, contentType)
}

// Download downloads a file from S3
func (s *S3Storage) Download(ctx context.Context, key string) ([]byte, error) {
	fullKey := s.buildKey(key)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// GetReader returns a reader for streaming download
func (s *S3Storage) GetReader(ctx context.Context, key string) (io.ReadCloser, string, error) {
	fullKey := s.buildKey(key)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from S3: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return result.Body, contentType, nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	fullKey := s.buildKey(key)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.buildKey(key)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetPublicURL returns the public URL for a key
func (s *S3Storage) GetPublicURL(key string) string {
	fullKey := s.buildKey(key)
	return fmt.Sprintf("%s/%s", s.baseURL, fullKey)
}

// GetSignedURL returns a pre-signed URL valid for the specified duration
func (s *S3Storage) GetSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	fullKey := s.buildKey(key)

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}, s3.WithPresignExpires(duration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// buildKey constructs the full S3 key with optional prefix
func (s *S3Storage) buildKey(key string) string {
	if s.pathPrefix == "" {
		return key
	}
	return filepath.Join(s.pathPrefix, key)
}

// GetBucket returns the bucket name
func (s *S3Storage) GetBucket() string {
	return s.bucket
}

// GetBaseURL returns the base URL
func (s *S3Storage) GetBaseURL() string {
	return s.baseURL
}

// DetectContentType returns content type based on file extension
func DetectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

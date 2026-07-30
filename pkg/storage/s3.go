package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edsonmubezi/myapp/pkg/resilience"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg Config) (*S3Storage, error) {
	if cfg.S3Bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}
	if cfg.S3Region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}

	ctx := context.Background()

	// Build AWS config
	var awsCfg aws.Config
	var err error

	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		// Use explicit credentials
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.S3Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.S3AccessKey,
				cfg.S3SecretKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (IAM roles, env vars, etc.)
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.S3Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	var s3Client *s3.Client
	if cfg.S3Endpoint != "" {
		// Custom endpoint (for MinIO compatibility)
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	return &S3Storage{
		client: s3Client,
		bucket: cfg.S3Bucket,
	}, nil
}

// ErrS3Unavailable is returned when S3 circuit breaker is open
var ErrS3Unavailable = errors.New("S3 storage service unavailable")

// Upload uploads a file to S3 with circuit breaker protection
func (s *S3Storage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	cb := resilience.GetBreaker(resilience.ServiceS3)

	operation := func() error {
		input := &s3.PutObjectInput{
			Bucket:               aws.String(s.bucket),
			Key:                  aws.String(path),
			Body:                 reader,
			ContentType:          aws.String(contentType),
			ServerSideEncryption: types.ServerSideEncryptionAes256,
		}

		_, err := s.client.PutObject(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to upload to S3: %w", err)
		}
		return nil
	}

	if cb != nil {
		err := cb.Execute(ctx, operation)
		if errors.Is(err, resilience.ErrCircuitOpen) {
			return ErrS3Unavailable
		}
		return err
	}

	return operation()
}

// Download retrieves a file from S3 with circuit breaker protection
func (s *S3Storage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	cb := resilience.GetBreaker(resilience.ServiceS3)

	var result *s3.GetObjectOutput
	operation := func() error {
		input := &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		}

		var err error
		result, err = s.client.GetObject(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to download from S3: %w", err)
		}
		return nil
	}

	if cb != nil {
		err := cb.Execute(ctx, operation)
		if errors.Is(err, resilience.ErrCircuitOpen) {
			return nil, ErrS3Unavailable
		}
		if err != nil {
			return nil, err
		}
	} else {
		if err := operation(); err != nil {
			return nil, err
		}
	}

	return result.Body, nil
}

// Delete removes a file from S3 with circuit breaker protection
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	cb := resilience.GetBreaker(resilience.ServiceS3)

	operation := func() error {
		input := &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		}

		_, err := s.client.DeleteObject(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to delete from S3: %w", err)
		}
		return nil
	}

	if cb != nil {
		err := cb.Execute(ctx, operation)
		if errors.Is(err, resilience.ErrCircuitOpen) {
			return ErrS3Unavailable
		}
		return err
	}

	return operation()
}

// GetURL returns a presigned URL for temporary access to the file
func (s *S3Storage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = 15 * time.Minute // Default 15 minutes
	}

	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		// AWS SDK v2 doesn't have a simple IsNotFound check
		// We need to check the error message or type
		return false, nil
	}

	return true, nil
}

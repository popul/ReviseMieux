// Package s3 implements the chapter.Storage port using S3/MinIO.
package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage implements chapter.Storage backed by S3/MinIO.
type Storage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// Config holds the S3/MinIO connection parameters.
type Config struct {
	Endpoint  string // e.g. "localhost:9000" or "http://localhost:9000"
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

// New creates a new S3 storage adapter and ensures the bucket exists.
func New(cfg Config) (*Storage, error) {
	// Strip scheme from endpoint if present (minio-go handles SSL separately)
	endpoint := cfg.Endpoint
	useSSL := cfg.UseSSL
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
		useSSL = true
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3.New: create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("s3.New: check bucket %q: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("s3.New: create bucket %q: %w", cfg.Bucket, err)
		}
	}

	return &Storage{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

// Upload stores a file in S3 and returns its URL.
// Implements chapter.Storage.
func (s *Storage) Upload(ctx context.Context, key string, contentType string, body io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("s3.Upload(%s): %w", key, err)
	}

	// Build the URL: scheme://endpoint/bucket/key
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	objectURL := fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, url.PathEscape(key))

	return objectURL, nil
}

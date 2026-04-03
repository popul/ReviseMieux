//go:build integration

package s3

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

const (
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin"
)

func minioAvailable() bool {
	conn, err := net.DialTimeout("tcp", minioEndpoint, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func newTestStorage(t *testing.T, bucket string) *Storage {
	t.Helper()
	if !minioAvailable() {
		t.Skip("MinIO not available at " + minioEndpoint)
	}

	st, err := New(Config{
		Endpoint:  minioEndpoint,
		AccessKey: minioAccessKey,
		SecretKey: minioSecretKey,
		Bucket:    bucket,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return st
}

func TestStorage_Upload_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	st := newTestStorage(t, "test-upload-integ")
	ctx := context.Background()

	content := "hello world"
	url, err := st.Upload(ctx, "test/file.txt", "text/plain", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}

	if url == "" {
		t.Fatal("Upload() returned empty URL")
	}
	if !strings.Contains(url, "test-upload-integ") {
		t.Errorf("URL should contain bucket name, got %q", url)
	}
	if !strings.Contains(url, "test/file.txt") {
		t.Errorf("URL should contain key, got %q", url)
	}
}

func TestStorage_Upload_ContentType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	st := newTestStorage(t, "test-content-type-integ")
	ctx := context.Background()

	// Upload a JPEG-like file with image/jpeg content type
	fakeJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	url, err := st.Upload(ctx, "photo.jpg", "image/jpeg", strings.NewReader(string(fakeJPEG)))
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}

	if !strings.Contains(url, "photo.jpg") {
		t.Errorf("URL should contain filename, got %q", url)
	}

	// Download via minio client to verify content
	obj, err := st.client.GetObject(ctx, st.bucket, "photo.jpg", minio.GetObjectOptions{})
	if err != nil {
		t.Fatalf("GetObject() error: %v", err)
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	if info.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %q, want %q", info.ContentType, "image/jpeg")
	}

	data, _ := io.ReadAll(obj)
	if len(data) != len(fakeJPEG) {
		t.Errorf("downloaded size = %d, want %d", len(data), len(fakeJPEG))
	}
}

func TestStorage_New_CreatesBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if !minioAvailable() {
		t.Skip("MinIO not available at " + minioEndpoint)
	}

	bucketName := "auto-created-bucket-test"

	// First call should create the bucket
	st1, err := New(Config{
		Endpoint:  minioEndpoint,
		AccessKey: minioAccessKey,
		SecretKey: minioSecretKey,
		Bucket:    bucketName,
	})
	if err != nil {
		t.Fatalf("New() should create bucket: %v", err)
	}

	// Second call should succeed (bucket already exists)
	_, err = New(Config{
		Endpoint:  minioEndpoint,
		AccessKey: minioAccessKey,
		SecretKey: minioSecretKey,
		Bucket:    bucketName,
	})
	if err != nil {
		t.Fatalf("New() on existing bucket should succeed: %v", err)
	}

	// Verify bucket is functional by uploading
	ctx := context.Background()
	_, err = st1.Upload(ctx, "verify.txt", "text/plain", strings.NewReader("ok"))
	if err != nil {
		t.Fatalf("Upload to auto-created bucket failed: %v", err)
	}
}

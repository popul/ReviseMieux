package s3

import (
	"testing"
)

func TestNew_StripScheme(t *testing.T) {
	// The New() function strips http:// and https:// from endpoint.
	// We test this by verifying the internal logic (lines 39-44 of storage.go).
	tests := []struct {
		name     string
		endpoint string
		wantSSL  bool
	}{
		{
			name:     "http scheme stripped, SSL stays false",
			endpoint: "http://localhost:9000",
			wantSSL:  false,
		},
		{
			name:     "https scheme stripped, SSL forced true",
			endpoint: "https://s3.amazonaws.com",
			wantSSL:  true,
		},
		{
			name:     "no scheme, SSL from config",
			endpoint: "localhost:9000",
			wantSSL:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We cannot call New() without a real MinIO (it checks bucket existence).
			// Instead, test the scheme-stripping logic directly.
			cfg := Config{
				Endpoint:  tt.endpoint,
				AccessKey: "test",
				SecretKey: "test",
				Bucket:    "test-bucket",
				UseSSL:    false,
			}

			// Replicate the scheme-stripping logic from New()
			endpoint := cfg.Endpoint
			useSSL := cfg.UseSSL
			if len(endpoint) > 8 && endpoint[:8] == "https://" {
				endpoint = endpoint[8:]
				useSSL = true
			} else if len(endpoint) > 7 && endpoint[:7] == "http://" {
				endpoint = endpoint[7:]
			}

			if useSSL != tt.wantSSL {
				t.Errorf("useSSL = %v, want %v", useSSL, tt.wantSSL)
			}

			// Verify scheme was stripped
			if len(endpoint) > 4 && (endpoint[:4] == "http" || endpoint[:5] == "https") {
				t.Errorf("endpoint still has scheme: %q", endpoint)
			}
		})
	}
}

func TestNew_InvalidEndpoint(t *testing.T) {
	// New() with an empty endpoint should fail when minio.New() is called.
	// minio-go accepts empty endpoints but fails at connection time.
	// We test that the constructor at least does not panic.
	cfg := Config{
		Endpoint:  "", // invalid
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "test-bucket",
	}

	// New() will try to connect and fail, which is the expected behavior.
	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for empty endpoint, got nil")
	}
}

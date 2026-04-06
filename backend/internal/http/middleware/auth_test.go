package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/http/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "test-secret-key-for-jwt"

// newAuthRouter builds a test router with Auth middleware and a handler
// that echoes the extracted userID and role from context.
func newAuthRouter(secret string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Auth(secret))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)
		role := middleware.GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"userID": userID.String(),
			"role":   role,
		})
	})
	return r
}

func mustGenerateToken(t *testing.T, secret string, userID uuid.UUID, role string, ttl time.Duration) string {
	t.Helper()
	token, err := middleware.GenerateToken(secret, userID, role, ttl, time.Now())
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return token
}

// --- Auth middleware tests ---

func TestAuth_ValidToken(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, testSecret, userID, "student", 1*time.Hour)

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["userID"] != userID.String() {
		t.Errorf("userID: got %q, want %q", body["userID"], userID.String())
	}
	if body["role"] != "student" {
		t.Errorf("role: got %q, want %q", body["role"], "student")
	}
}

func TestAuth_ValidToken_ParentRole(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, testSecret, userID, "parent", 1*time.Hour)

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["role"] != "parent" {
		t.Errorf("role: got %q, want %q", body["role"], "parent")
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	// Generate a token that expired 1 hour ago
	token, err := middleware.GenerateToken(testSecret, userID, "student", 1*time.Hour, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_MalformedToken(t *testing.T) {
	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "missing authorization header" {
		t.Errorf("error message: got %q, want %q", body["error"], "missing authorization header")
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, "wrong-secret", userID, "student", 1*time.Hour)

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_NoBearerPrefix(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, testSecret, userID, "student", 1*time.Hour)

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token) // missing "Bearer " prefix
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "invalid authorization format" {
		t.Errorf("error message: got %q, want %q", body["error"], "invalid authorization format")
	}
}

func TestAuth_BasicAuthPrefix(t *testing.T) {
	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_EmptyBearerToken(t *testing.T) {
	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- GenerateToken / claims tests ---

func TestGenerateToken_ClaimsContent(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	ttl := 2 * time.Hour

	tokenStr, err := middleware.GenerateToken(testSecret, userID, "student", ttl, now)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, &middleware.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims: %v", err)
	}

	claims := token.Claims.(*middleware.Claims)

	if claims.Subject != userID.String() {
		t.Errorf("sub: got %q, want %q", claims.Subject, userID.String())
	}
	if claims.Role != "student" {
		t.Errorf("role: got %q, want %q", claims.Role, "student")
	}
	if claims.Issuer != "revisemieux" {
		t.Errorf("iss: got %q, want %q", claims.Issuer, "revisemieux")
	}
	wantExp := now.Add(ttl)
	if !claims.ExpiresAt.Equal(wantExp) {
		t.Errorf("exp: got %v, want %v", claims.ExpiresAt.Time, wantExp)
	}
	if !claims.IssuedAt.Equal(now) {
		t.Errorf("iat: got %v, want %v", claims.IssuedAt.Time, now)
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	// Create a token with "none" algorithm (unsigned) — must be rejected
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":  uuid.Must(uuid.NewV7()).String(),
		"role": "student",
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	r := newAuthRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for none signing method, got %d", w.Code)
	}
}

// --- Helper function tests ---

func TestGetUserID_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	id, ok := middleware.GetUserID(c)
	if ok {
		t.Errorf("expected ok=false, got true with id=%s", id)
	}
	if id != (uuid.UUID{}) {
		t.Errorf("expected zero UUID, got %s", id)
	}
}

func TestGetUserRole_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	role := middleware.GetUserRole(c)
	if role != "" {
		t.Errorf("expected empty role, got %q", role)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, testSecret, userID, "parent", 1*time.Hour)

	r := gin.New()
	r.Use(middleware.Auth(testSecret))
	r.Use(middleware.RequireRole("parent", "admin"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := mustGenerateToken(t, testSecret, userID, "student", 1*time.Hour)

	r := gin.New()
	r.Use(middleware.Auth(testSecret))
	r.Use(middleware.RequireRole("parent", "admin"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "insufficient permissions" {
		t.Errorf("error message: got %q, want %q", body["error"], "insufficient permissions")
	}
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// ContextKeyUserID is the gin context key for the authenticated user ID.
	ContextKeyUserID = "userID"
	// ContextKeyUserRole is the gin context key for the authenticated user role.
	ContextKeyUserRole = "userRole"
)

// Claims represents the JWT claims for Révise Mieux.
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// Auth returns a Gin middleware that validates JWT bearer tokens.
func Auth(secret string) gin.HandlerFunc {
	secretBytes := []byte(secret) // allocate once, reuse per request
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secretBytes, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// GetUserRole extracts the authenticated user role from the gin context.
func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(ContextKeyUserRole)
	role, _ := v.(string)
	return role
}

// RequireRole returns middleware that checks the user has the required role.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

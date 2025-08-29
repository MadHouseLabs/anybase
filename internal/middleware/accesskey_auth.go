package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/accesskey"
)

// AccessKeyAuthMiddleware handles authentication via access keys
type AccessKeyAuthMiddleware struct {
	repo accesskey.Repository
}

// NewAccessKeyAuthMiddleware creates a new access key auth middleware
func NewAccessKeyAuthMiddleware(repo accesskey.Repository) *AccessKeyAuthMiddleware {
	return &AccessKeyAuthMiddleware{repo: repo}
}

// Authenticate validates access keys from Authorization header
func (m *AccessKeyAuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next() // Let other middleware handle if no access key
			return
		}

		// Check if it's an access key (starts with "Bearer ak_")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		if !strings.HasPrefix(token, "ak_") {
			c.Next() // Not an access key, let JWT middleware handle it
			return
		}

		// Validate the access key
		ak, err := m.repo.ValidateKey(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access key"})
			c.Abort()
			return
		}

		// Check if the key is active
		if !ak.Active {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "access key is inactive"})
			c.Abort()
			return
		}

		// Set context values for downstream handlers
		c.Set("auth_type", "access_key")
		c.Set("access_key_id", ak.ID.Hex())
		c.Set("permissions", ak.Permissions)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequireAccessKeyPermission checks if the access key has a specific permission
func (m *AccessKeyAuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, exists := c.Get("auth_type")
		if !exists || authType != "access_key" {
			c.Next() // Not access key auth, let other middleware handle
			return
		}

		perms, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			c.Abort()
			return
		}

		permissions := perms.([]string)
		requiredPerm := resource + ":" + action

		// Check for exact match or wildcard permissions
		hasPermission := false
		for _, perm := range permissions {
			if matchPermission(perm, requiredPerm) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// matchPermission checks if a permission pattern matches the required permission
func matchPermission(pattern, required string) bool {
	// Exact match
	if pattern == required {
		return true
	}

	// Super admin wildcard
	if pattern == "*:*:*" {
		return true
	}

	// Split into parts for pattern matching
	patternParts := strings.Split(pattern, ":")
	requiredParts := strings.Split(required, ":")

	// Must have same number of parts
	if len(patternParts) != len(requiredParts) {
		return false
	}

	// Check each part
	for i, pp := range patternParts {
		if pp != "*" && pp != requiredParts[i] {
			return false
		}
	}

	return true
}
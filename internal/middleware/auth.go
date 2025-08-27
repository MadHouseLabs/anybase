package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/karthik/anybase/internal/auth"
	"github.com/karthik/anybase/internal/config"
	"github.com/karthik/anybase/internal/governance"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthMiddleware struct {
	tokenService auth.TokenService
	rbacService  governance.RBACService
}

func NewAuthMiddleware(config *config.AuthConfig, rbacService governance.RBACService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: auth.NewTokenService(config),
		rbacService:  rbacService,
	}
}

// RequireAuth validates JWT token and sets user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if already authenticated by access key
		if authenticated, exists := c.Get("authenticated"); exists && authenticated.(bool) {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Skip if it's an access key (handled by access key middleware)
		if strings.HasPrefix(token, "ak_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		claims, err := m.tokenService.ValidateToken(token, auth.AccessToken)
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Set user context
		c.Set("userID", claims.UserID)
		c.Set("user_id", claims.UserID) // For consistency with existing code
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("auth_type", "jwt")
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequireRole checks if user has specific role (only for JWT auth)
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Access keys don't have roles, only permissions
		authType, _ := c.Get("auth_type")
		if authType == "access_key" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access keys cannot access role-based endpoints"})
			c.Abort()
			return
		}

		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		userRolesList := userRoles.([]string)
		hasRole := false

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			for _, userRole := range userRolesList {
				if userRole == requiredRole || userRole == "admin" {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission checks if user has specific permission
func (m *AuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		hasPermission, err := m.rbacService.HasPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth validates JWT token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.tokenService.ValidateToken(token, auth.AccessToken)
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)
			c.Set("permissions", claims.Permissions)
		}

		c.Next()
	}
}

// extractToken extracts token from Authorization header or query parameter
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter as fallback
	return c.Query("token")
}
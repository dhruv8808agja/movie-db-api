package middleware

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that checks if the authenticated user has one of the required roles
func RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Fetch user from database to get their role
		var user models.User
		if err := storage.DB.Where("username = ?", username).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
			return
		}

		// Check if user has one of the allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		// Store user role in context for use in handlers
		c.Set("user_role", user.Role)
		c.Set("user_id", user.ID)

		c.Next()
	}
}

// RequireAdmin is a convenience middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireModerator is a convenience middleware that requires moderator or admin role
func RequireModerator() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleModerator)
}

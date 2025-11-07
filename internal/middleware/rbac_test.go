package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRBACTest() {
	testutil.SetTestEnv()
	testutil.InitTestStorage()

	// Seed test users with explicit IsActive values
	testutil.SeedTestUsers(storage.DB)

	// Explicitly set the inactive user to false (in case default overrides it)
	storage.DB.Model(&models.User{}).Where("username = ?", "inactive").Update("is_active", false)

	gin.SetMode(gin.TestMode)
}

func TestRequireRole_Success(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "admin")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestRequireRole_NoUsername(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestRequireRole_UserNotFound(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "nonexistent")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestRequireRole_InactiveUser(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "inactive")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleUser), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "account is inactive")
}

func TestRequireRole_InsufficientPermissions(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	setupRBACTest()

	tests := []struct {
		name           string
		username       string
		allowedRoles   []models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin with admin or moderator roles",
			username:       "admin",
			allowedRoles:   []models.UserRole{models.RoleAdmin, models.RoleModerator},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Moderator with admin or moderator roles",
			username:       "moderator",
			allowedRoles:   []models.UserRole{models.RoleAdmin, models.RoleModerator},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User with admin or moderator roles",
			username:       "user",
			allowedRoles:   []models.UserRole{models.RoleAdmin, models.RoleModerator},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User with user role",
			username:       "user",
			allowedRoles:   []models.UserRole{models.RoleUser},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("username", tt.username)
				c.Next()
			})
			router.GET("/protected", RequireRole(tt.allowedRoles...), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireRole_SetsContextValues(t *testing.T) {
	setupRBACTest()

	var capturedRole models.UserRole
	var capturedUserID uint

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "admin")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		role, _ := c.Get("user_role")
		userID, _ := c.Get("user_id")
		capturedRole = role.(models.UserRole)
		capturedUserID = userID.(uint)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, models.RoleAdmin, capturedRole)
	assert.NotZero(t, capturedUserID)
}

func TestRequireAdmin_Success(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "admin")
		c.Next()
	})
	router.GET("/admin", RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}

func TestRequireAdmin_NonAdminDenied(t *testing.T) {
	setupRBACTest()

	tests := []struct {
		name     string
		username string
	}{
		{"Regular user", "user"},
		{"Moderator", "moderator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("username", tt.username)
				c.Next()
			})
			router.GET("/admin", RequireAdmin(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

			req := httptest.NewRequest("GET", "/admin", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "insufficient permissions")
		})
	}
}

func TestRequireModerator_Success(t *testing.T) {
	setupRBACTest()

	tests := []struct {
		name     string
		username string
	}{
		{"Admin can access moderator route", "admin"},
		{"Moderator can access moderator route", "moderator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("username", tt.username)
				c.Next()
			})
			router.GET("/moderate", RequireModerator(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "moderator access granted"})
			})

			req := httptest.NewRequest("GET", "/moderate", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "moderator access granted")
		})
	}
}

func TestRequireModerator_UserDenied(t *testing.T) {
	setupRBACTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.GET("/moderate", RequireModerator(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "moderator access granted"})
	})

	req := httptest.NewRequest("GET", "/moderate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRequireRole_AbortsMiddlewareChain(t *testing.T) {
	setupRBACTest()

	handlerCalled := false

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.GET("/protected", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, handlerCalled, "Handler should not be called when RBAC check fails")
}

func TestRequireRole_AllRoles(t *testing.T) {
	setupRBACTest()

	tests := []struct {
		name           string
		username       string
		expectedStatus int
	}{
		{"Admin", "admin", http.StatusOK},
		{"Moderator", "moderator", http.StatusOK},
		{"User", "user", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("username", tt.username)
				c.Next()
			})
			router.GET("/protected", RequireRole(models.RoleAdmin, models.RoleModerator, models.RoleUser), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

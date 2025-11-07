package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAdminTest() {
	testutil.SetTestEnv()
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)
	gin.SetMode(gin.TestMode)
}

// ===== LIST USERS TESTS =====

func TestListUsers_Success(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.GET("/admin/users", ListUsers)

	req := httptest.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Users)
	assert.Greater(t, response.Total, int64(0))
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.Limit)
}

func TestListUsers_WithPagination(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.GET("/admin/users", ListUsers)

	// Test page 1, limit 2
	req := httptest.NewRequest("GET", "/admin/users?page=1&limit=2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(response.Users), 2)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 2, response.Limit)
}

func TestListUsers_Page2(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.GET("/admin/users", ListUsers)

	// Test page 2, limit 1
	req := httptest.NewRequest("GET", "/admin/users?page=2&limit=1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 1, response.Limit)
}

func TestListUsers_InvalidPaginationParams(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.GET("/admin/users", ListUsers)

	// Test invalid page param - should use default
	req := httptest.NewRequest("GET", "/admin/users?page=invalid&limit=abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)   // defaults to 1
	assert.Equal(t, 10, response.Limit) // defaults to 10
}

// ===== UPDATE USER ROLE TESTS =====

func TestUpdateUserRole_Success(t *testing.T) {
	setupAdminTest()

	// Get user ID
	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := UpdateRoleRequest{
		Role: models.RoleModerator,
	}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/admin/users/%d/role", user.ID)
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	err := json.Unmarshal(w.Body.Bytes(), &profile)
	assert.NoError(t, err)
	assert.Equal(t, models.RoleModerator, profile.Role)

	// Verify in database
	var updatedUser models.User
	storage.DB.Where("id = ?", user.ID).First(&updatedUser)
	assert.Equal(t, models.RoleModerator, updatedUser.Role)
}

func TestUpdateUserRole_ToAdmin(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := UpdateRoleRequest{
		Role: models.RoleAdmin,
	}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/admin/users/%d/role", user.ID)
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	assert.Equal(t, models.RoleAdmin, profile.Role)
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := map[string]string{
		"role": "superadmin", // Invalid role
	}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/admin/users/%d/role", user.ID)
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid role")
}

func TestUpdateUserRole_UserNotFound(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := UpdateRoleRequest{
		Role: models.RoleModerator,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/admin/users/99999/role", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestUpdateUserRole_InvalidJSON(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	req := httptest.NewRequest("PUT", "/admin/users/1/role", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUserRole_MissingRole(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := map[string]string{}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/admin/users/%d/role", user.ID)
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===== DEACTIVATE USER TESTS =====

func TestDeactivateUser_Success(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)
	assert.True(t, user.IsActive) // Initially active

	router := gin.New()
	router.POST("/admin/users/:id/deactivate", DeactivateUser)

	url := fmt.Sprintf("/admin/users/%d/deactivate", user.ID)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user deactivated successfully")

	// Verify in database
	var updatedUser models.User
	storage.DB.Where("id = ?", user.ID).First(&updatedUser)
	assert.False(t, updatedUser.IsActive)
}

func TestDeactivateUser_AlreadyDeactivated(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	// Deactivate first
	user.IsActive = false
	storage.DB.Save(&user)

	router := gin.New()
	router.POST("/admin/users/:id/deactivate", DeactivateUser)

	url := fmt.Sprintf("/admin/users/%d/deactivate", user.ID)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user deactivated successfully")

	// Verify still deactivated
	var updatedUser models.User
	storage.DB.Where("id = ?", user.ID).First(&updatedUser)
	assert.False(t, updatedUser.IsActive)
}

func TestDeactivateUser_UserNotFound(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.POST("/admin/users/:id/deactivate", DeactivateUser)

	req := httptest.NewRequest("POST", "/admin/users/99999/deactivate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

// ===== ACTIVATE USER TESTS =====

func TestActivateUser_Success(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	// Deactivate first
	user.IsActive = false
	storage.DB.Save(&user)

	router := gin.New()
	router.POST("/admin/users/:id/activate", ActivateUser)

	url := fmt.Sprintf("/admin/users/%d/activate", user.ID)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user activated successfully")

	// Verify in database
	var updatedUser models.User
	storage.DB.Where("id = ?", user.ID).First(&updatedUser)
	assert.True(t, updatedUser.IsActive)
}

func TestActivateUser_AlreadyActive(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)
	assert.True(t, user.IsActive) // Already active

	router := gin.New()
	router.POST("/admin/users/:id/activate", ActivateUser)

	url := fmt.Sprintf("/admin/users/%d/activate", user.ID)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user activated successfully")

	// Verify still active
	var updatedUser models.User
	storage.DB.Where("id = ?", user.ID).First(&updatedUser)
	assert.True(t, updatedUser.IsActive)
}

func TestActivateUser_UserNotFound(t *testing.T) {
	setupAdminTest()

	router := gin.New()
	router.POST("/admin/users/:id/activate", ActivateUser)

	req := httptest.NewRequest("POST", "/admin/users/99999/activate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

// ===== HELPER FUNCTION TESTS =====

func TestParseIntDefault_Success(t *testing.T) {
	val, err := parseIntDefault("42", 10)
	assert.NoError(t, err)
	assert.Equal(t, 42, val)
}

func TestParseIntDefault_InvalidString(t *testing.T) {
	val, err := parseIntDefault("invalid", 10)
	assert.Error(t, err)
	assert.Equal(t, 10, val) // Returns default
}

func TestParseIntDefault_EmptyString(t *testing.T) {
	val, err := parseIntDefault("", 5)
	assert.Error(t, err)
	assert.Equal(t, 5, val) // Returns default
}

// ===== EDGE CASE AND INTEGRATION TESTS =====

func TestListUsers_EmptyDatabase(t *testing.T) {
	testutil.SetTestEnv()
	testutil.InitTestStorage()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin/users", ListUsers)

	req := httptest.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), response.Total)
	assert.Empty(t, response.Users)
}

func TestUpdateUserRole_DowngradeFromAdmin(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "admin").First(&user)
	assert.Equal(t, models.RoleAdmin, user.Role)

	router := gin.New()
	router.PUT("/admin/users/:id/role", UpdateUserRole)

	reqBody := UpdateRoleRequest{
		Role: models.RoleUser,
	}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/admin/users/%d/role", user.ID)
	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	assert.Equal(t, models.RoleUser, profile.Role)
}

func TestDeactivateAndActivateUser_Workflow(t *testing.T) {
	setupAdminTest()

	var user models.User
	storage.DB.Where("username = ?", "user").First(&user)

	router := gin.New()
	router.POST("/admin/users/:id/deactivate", DeactivateUser)
	router.POST("/admin/users/:id/activate", ActivateUser)

	// Deactivate
	url := fmt.Sprintf("/admin/users/%d/deactivate", user.ID)
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deactivated
	storage.DB.Where("id = ?", user.ID).First(&user)
	assert.False(t, user.IsActive)

	// Activate
	url = fmt.Sprintf("/admin/users/%d/activate", user.ID)
	req = httptest.NewRequest("POST", url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify activated
	storage.DB.Where("id = ?", user.ID).First(&user)
	assert.True(t, user.IsActive)
}

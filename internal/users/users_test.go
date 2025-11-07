package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupUserTest() {
	testutil.SetTestEnv()
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)
	gin.SetMode(gin.TestMode)
}

// ===== REGISTRATION TESTS =====

func TestRegister_Success(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.POST("/register", Register)

	reqBody := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@test.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		Bio:       "I'm a new user",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response RegisterResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User registered successfully", response.Message)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "newuser", response.User.Username)
	assert.Equal(t, "newuser@test.com", response.User.Email)
	assert.Equal(t, models.RoleUser, response.User.Role)
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.POST("/register", Register)

	reqBody := RegisterRequest{
		Username: "admin", // Already exists
		Email:    "newemail@test.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "username already exists")
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.POST("/register", Register)

	reqBody := RegisterRequest{
		Username: "brandnewuser",
		Email:    "admin@test.com", // Already exists
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "email already exists")
}

func TestRegister_ValidationErrors(t *testing.T) {
	setupUserTest()

	tests := []struct {
		name    string
		request RegisterRequest
		errMsg  string
	}{
		{
			name: "Missing username",
			request: RegisterRequest{
				Email:    "test@test.com",
				Password: "password123",
			},
			errMsg: "Username",
		},
		{
			name: "Username too short",
			request: RegisterRequest{
				Username: "ab",
				Email:    "test@test.com",
				Password: "password123",
			},
			errMsg: "min",
		},
		{
			name: "Missing email",
			request: RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			errMsg: "Email",
		},
		{
			name: "Invalid email format",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "notanemail",
				Password: "password123",
			},
			errMsg: "email",
		},
		{
			name: "Missing password",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@test.com",
			},
			errMsg: "Password",
		},
		{
			name: "Password too short",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@test.com",
				Password: "short",
			},
			errMsg: "min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/register", Register)

			jsonData, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.errMsg)
		})
	}
}

func TestRegister_UsernameEmailNormalization(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.POST("/register", Register)

	reqBody := RegisterRequest{
		Username: "NewUserTest", // Mixed case (normalization happens after validation)
		Email:    "NewUser@Test.COM",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response RegisterResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "newusertest", response.User.Username)
	assert.Equal(t, "newuser@test.com", response.User.Email)
}

func TestRegister_InvalidJSON(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.POST("/register", Register)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===== PROFILE TESTS =====

func TestGetProfile_Success(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "admin")
		c.Next()
	})
	router.GET("/profile", GetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	assert.Equal(t, "admin", profile.Username)
	assert.Equal(t, "admin@test.com", profile.Email)
}

func TestGetProfile_Unauthorized(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.GET("/profile", GetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestGetProfile_UserNotFound(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "nonexistent")
		c.Next()
	})
	router.GET("/profile", GetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestGetProfileByID_Success(t *testing.T) {
	setupUserTest()

	// Get admin user ID
	var user models.User
	storage.DB.Where("username = ?", "admin").First(&user)

	router := gin.New()
	router.GET("/users/:id", GetProfileByID)

	req := httptest.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	assert.Equal(t, "admin", profile.Username)
}

func TestGetProfileByID_NotFound(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.GET("/users/:id", GetProfileByID)

	req := httptest.NewRequest("GET", "/users/99999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

func TestUpdateProfile_Success(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.PUT("/profile", UpdateProfile)

	reqBody := UpdateProfileRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Bio:       "Updated bio",
		Avatar:    "https://example.com/avatar.jpg",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile models.UserProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	assert.Equal(t, "Updated", profile.FirstName)
	assert.Equal(t, "Name", profile.LastName)
	assert.Equal(t, "Updated bio", profile.Bio)
	assert.Equal(t, "https://example.com/avatar.jpg", profile.Avatar)
}

func TestUpdateProfile_Unauthorized(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.PUT("/profile", UpdateProfile)

	reqBody := UpdateProfileRequest{
		FirstName: "Updated",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUpdateProfile_InvalidJSON(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.PUT("/profile", UpdateProfile)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteProfile_Success(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "user")
		c.Next()
	})
	router.DELETE("/profile", DeleteProfile)

	req := httptest.NewRequest("DELETE", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "account deleted successfully")

	// Verify user is soft deleted
	var user models.User
	err := storage.DB.Where("username = ?", "user").First(&user).Error
	assert.Error(t, err) // Should not find the user
}

func TestDeleteProfile_Unauthorized(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.DELETE("/profile", DeleteProfile)

	req := httptest.NewRequest("DELETE", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestDeleteProfile_UserNotFound(t *testing.T) {
	setupUserTest()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", "nonexistent")
		c.Next()
	})
	router.DELETE("/profile", DeleteProfile)

	req := httptest.NewRequest("DELETE", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}

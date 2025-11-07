package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database and seed users
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	credentials := map[string]string{
		"username": "admin",
		"password": "password",
	}
	jsonData, _ := json.Marshal(credentials)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "admin", response.User.Username)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database and seed users
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"Wrong password", "admin", "wrongpassword"},
		{"Wrong username", "wronguser", "password"},
		{"Both wrong", "wronguser", "wrongpassword"},
		{"Empty username", "", "password"},
		{"Empty password", "admin", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			credentials := map[string]string{
				"username": tc.username,
				"password": tc.password,
			}
			jsonData, _ := json.Marshal(credentials)

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "invalid credentials")
		})
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJWTMiddleware_Integration(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database and seed users
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Public login endpoint
	router.POST("/login", Login)

	// Protected endpoint
	protected := router.Group("/")
	protected.Use(JWTMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"message": "success", "user": username})
	})

	// Step 1: Login to get token
	credentials := map[string]string{
		"username": "admin",
		"password": "password",
	}
	jsonData, _ := json.Marshal(credentials)

	loginReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]string
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]

	// Step 2: Access protected endpoint with token
	protectedReq := httptest.NewRequest("GET", "/protected", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+token)
	protectedW := httptest.NewRecorder()

	router.ServeHTTP(protectedW, protectedReq)

	assert.Equal(t, http.StatusOK, protectedW.Code)

	var protectedResponse map[string]interface{}
	json.Unmarshal(protectedW.Body.Bytes(), &protectedResponse)
	assert.Equal(t, "success", protectedResponse["message"])
	assert.Equal(t, "admin", protectedResponse["user"])
}

func TestJWTMiddleware_Integration_Unauthorized(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/")
	protected.Use(JWTMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Try to access protected endpoint without token
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthFlow_EndToEnd(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Initialize test database and seed users
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes
	router.POST("/login", Login)

	protected := router.Group("/")
	protected.Use(JWTMiddleware())
	protected.GET("/data", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"data": "secret", "user": username})
	})

	// Test 1: Try to access protected route without auth
	req1 := httptest.NewRequest("GET", "/data", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// Test 2: Login with valid credentials
	credentials := map[string]string{
		"username": "admin",
		"password": "password",
	}
	jsonData, _ := json.Marshal(credentials)
	req2 := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var loginResponse map[string]string
	json.Unmarshal(w2.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]
	assert.NotEmpty(t, token)

	// Test 3: Access protected route with valid token
	req3 := httptest.NewRequest("GET", "/data", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	var dataResponse map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &dataResponse)
	assert.Equal(t, "secret", dataResponse["data"])
	assert.Equal(t, "admin", dataResponse["user"])

	// Test 4: Try with invalid token
	req4 := httptest.NewRequest("GET", "/data", nil)
	req4.Header.Set("Authorization", "Bearer invalid_token")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)
}

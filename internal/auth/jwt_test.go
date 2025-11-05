package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key")
	code := m.Run()
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

func TestGenerateToken_Success(t *testing.T) {
	username := "testuser"

	token, err := GenerateToken(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token can be parsed
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Check claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, username, claims["username"])

	// Check expiry is in the future
	exp := int64(claims["exp"].(float64))
	assert.Greater(t, exp, time.Now().Unix())
}

func TestGenerateToken_DifferentUsernames(t *testing.T) {
	token1, err1 := GenerateToken("user1")
	token2, err2 := GenerateToken("user2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Generate a valid token
	token, _ := GenerateToken("testuser")

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	// Create a test handler
	called := false
	testHandler := func(c *gin.Context) {
		called = true
		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)
		c.Status(http.StatusOK)
	}

	// Run middleware and handler
	middleware := JWTMiddleware()
	middleware(c)

	if !c.IsAborted() {
		testHandler(c)
	}

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	middleware := JWTMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing Authorization header")
}

func TestJWTMiddleware_InvalidAuthFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name   string
		header string
	}{
		{"Missing Bearer", "token123"},
		{"Wrong prefix", "Basic token123"},
		{"Only Bearer", "Bearer"},
		{"Empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Request.Header.Set("Authorization", tc.header)

			middleware := JWTMiddleware()
			middleware(c)

			assert.True(t, c.IsAborted())
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	middleware := JWTMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create an expired token
	claims := jwt.MapClaims{
		"username": "testuser",
		"exp":      time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	middleware := JWTMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_TamperedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Generate a valid token
	token, _ := GenerateToken("testuser")

	// Tamper with the token by changing a character
	tamperedToken := token[:len(token)-5] + "xxxxx"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tamperedToken)

	middleware := JWTMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

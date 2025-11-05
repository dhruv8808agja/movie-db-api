package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetRateLimiterConfig_Defaults(t *testing.T) {
	// Clear env vars to test defaults
	os.Unsetenv("RATE_LIMIT_REQUESTS")
	os.Unsetenv("RATE_LIMIT_WINDOW_SECONDS")

	config := GetRateLimiterConfig()

	assert.Equal(t, 100, config.RequestsPerWindow)
	assert.Equal(t, time.Minute, config.WindowDuration)
}

func TestGetRateLimiterConfig_CustomValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_REQUESTS", "50")
	os.Setenv("RATE_LIMIT_WINDOW_SECONDS", "120")
	defer os.Unsetenv("RATE_LIMIT_REQUESTS")
	defer os.Unsetenv("RATE_LIMIT_WINDOW_SECONDS")

	config := GetRateLimiterConfig()

	assert.Equal(t, 50, config.RequestsPerWindow)
	assert.Equal(t, 120*time.Second, config.WindowDuration)
}

func TestGetRateLimiterConfig_InvalidValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_REQUESTS", "invalid")
	os.Setenv("RATE_LIMIT_WINDOW_SECONDS", "invalid")
	defer os.Unsetenv("RATE_LIMIT_REQUESTS")
	defer os.Unsetenv("RATE_LIMIT_WINDOW_SECONDS")

	config := GetRateLimiterConfig()

	// Should use defaults on invalid values
	assert.Equal(t, 100, config.RequestsPerWindow)
	assert.Equal(t, time.Minute, config.WindowDuration)
}

func TestRateLimiter_WithinLimit(t *testing.T) {
	redisClient := testutil.SetupTestRedis()
	if redisClient == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer testutil.CleanupTestRedis(redisClient)

	storage.RedisClient = redisClient
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := RateLimiterConfig{
		RequestsPerWindow: 5,
		WindowDuration:    time.Minute,
	}
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make 5 requests (should all succeed)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234" // Set consistent IP
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	redisClient := testutil.SetupTestRedis()
	if redisClient == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer testutil.CleanupTestRedis(redisClient)

	storage.RedisClient = redisClient
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := RateLimiterConfig{
		RequestsPerWindow: 3,
		WindowDuration:    time.Minute,
	}
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make 3 requests (should succeed)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:1234" // Use RemoteAddr instead of X-Forwarded-For
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 4th request should fail
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate limit exceeded")
	assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	redisClient := testutil.SetupTestRedis()
	if redisClient == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer testutil.CleanupTestRedis(redisClient)

	storage.RedisClient = redisClient
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := RateLimiterConfig{
		RequestsPerWindow: 2,
		WindowDuration:    time.Minute,
	}
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// IP 1: Make 2 requests (should succeed)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.10:1234"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 2: Should still be able to make requests
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.20:1234"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// IP 1: 3rd request should fail
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.10:1234"
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_Headers(t *testing.T) {
	redisClient := testutil.SetupTestRedis()
	if redisClient == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer testutil.CleanupTestRedis(redisClient)

	storage.RedisClient = redisClient
	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := RateLimiterConfig{
		RequestsPerWindow: 10,
		WindowDuration:    time.Minute,
	}
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:1234"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check headers are present
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))

	remaining := w.Header().Get("X-RateLimit-Remaining")
	assert.NotEmpty(t, remaining)

	reset := w.Header().Get("X-RateLimit-Reset")
	assert.NotEmpty(t, reset)
}

func TestRateLimiter_RedisUnavailable(t *testing.T) {
	// Set storage.RedisClient to nil to simulate Redis unavailability
	storage.RedisClient = nil

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := RateLimiterConfig{
		RequestsPerWindow: 2,
		WindowDuration:    time.Minute,
	}
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// When Redis is unavailable, requests should still go through
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not block requests when Redis is down (graceful degradation)
	assert.Equal(t, http.StatusOK, w.Code)
}

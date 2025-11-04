package middleware

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds the configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerWindow int
	WindowDuration    time.Duration
}

// GetRateLimiterConfig reads rate limiter configuration from environment variables
func GetRateLimiterConfig() RateLimiterConfig {
	requestsPerWindow := 100 // default: 100 requests
	windowDuration := time.Minute // default: per minute

	if reqStr := os.Getenv("RATE_LIMIT_REQUESTS"); reqStr != "" {
		if req, err := strconv.Atoi(reqStr); err == nil && req > 0 {
			requestsPerWindow = req
		}
	}

	if winStr := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); winStr != "" {
		if win, err := strconv.Atoi(winStr); err == nil && win > 0 {
			windowDuration = time.Duration(win) * time.Second
		}
	}

	return RateLimiterConfig{
		RequestsPerWindow: requestsPerWindow,
		WindowDuration:    windowDuration,
	}
}

// RateLimiter creates a middleware that limits requests based on client IP
// Uses Redis with sliding window algorithm
func RateLimiter(config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP as the rate limit key
		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		// Current timestamp
		now := time.Now().Unix()
		windowStart := now - int64(config.WindowDuration.Seconds())

		// Use Redis pipeline for atomic operations
		pipe := storage.RedisClient.Pipeline()

		// Remove old entries outside the current window
		pipe.ZRemRangeByScore(storage.Ctx, key, "0", strconv.FormatInt(windowStart, 10))

		// Count requests in current window
		countCmd := pipe.ZCount(storage.Ctx, key, strconv.FormatInt(windowStart, 10), "+inf")

		// Add current request with current timestamp as score
		pipe.ZAdd(storage.Ctx, key, redis.Z{
			Score:  float64(now),
			Member: strconv.FormatInt(now, 10),
		})

		// Set expiry on the key (cleanup)
		pipe.Expire(storage.Ctx, key, config.WindowDuration)

		// Execute pipeline
		_, err := pipe.Exec(storage.Ctx)
		if err != nil {
			// If Redis fails, log but don't block the request
			c.Next()
			return
		}

		// Get the count result
		count, err := countCmd.Result()
		if err != nil {
			// If Redis fails, log but don't block the request
			c.Next()
			return
		}

		// Check if rate limit exceeded
		if count > int64(config.RequestsPerWindow) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(config.WindowDuration.Seconds()), 10))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after_seconds": int64(config.WindowDuration.Seconds()),
			})
			return
		}

		// Set rate limit headers
		remaining := config.RequestsPerWindow - int(count)
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(config.WindowDuration.Seconds()), 10))

		c.Next()
	}
}

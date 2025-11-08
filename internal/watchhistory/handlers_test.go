package watchhistory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupWatchHistoryTest() {
	testutil.SetTestEnv()
	testutil.InitTestStorage()
	testutil.SeedTestUsers(storage.DB)
	testutil.SeedTestMovies(storage.DB)
	gin.SetMode(gin.TestMode)
}

func createAuthContext(userID uint) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)
	return c
}

// ===== RECORD WATCH TESTS =====

func TestRecordWatch_Success(t *testing.T) {
	setupWatchHistoryTest()

	router := gin.New()
	router.POST("/watch-history", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		RecordWatch(c)
	})

	reqBody := RecordWatchRequest{
		MovieID:       1,
		WatchDuration: 3600,
		LastPosition:  2700,
		Progress:      75.0,
		Completed:     false,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/watch-history", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RecordWatchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Watch history recorded successfully", response.Message)
	assert.Equal(t, uint(1), response.WatchHistory.MovieID)
	assert.Equal(t, 3600, response.WatchHistory.WatchDuration)
	assert.Equal(t, float32(75.0), response.WatchHistory.Progress)
}

func TestRecordWatch_UpdateExisting(t *testing.T) {
	setupWatchHistoryTest()

	// Create initial watch history
	watchHistory := models.WatchHistory{
		UserID:        1,
		MovieID:       1,
		WatchedAt:     time.Now(),
		WatchDuration: 1800,
		Progress:      50.0,
		LastPosition:  1800,
		Completed:     false,
		WatchCount:    1,
		Quality:       "720p",
		DeviceType:    "web",
	}
	storage.DB.Create(&watchHistory)

	router := gin.New()
	router.POST("/watch-history", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		RecordWatch(c)
	})

	reqBody := RecordWatchRequest{
		MovieID:       1,
		WatchDuration: 1800, // Additional 30 minutes
		LastPosition:  3600,
		Progress:      100.0,
		Completed:     true,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/watch-history", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RecordWatchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Watch history updated successfully", response.Message)
	assert.Equal(t, 3600, response.WatchHistory.WatchDuration) // 1800 + 1800
	assert.Equal(t, float32(100.0), response.WatchHistory.Progress)
	assert.True(t, response.WatchHistory.Completed)
	assert.Equal(t, 2, response.WatchHistory.WatchCount)
}

func TestRecordWatch_MovieNotFound(t *testing.T) {
	setupWatchHistoryTest()

	router := gin.New()
	router.POST("/watch-history", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		RecordWatch(c)
	})

	reqBody := RecordWatchRequest{
		MovieID:       9999, // Non-existent movie
		WatchDuration: 3600,
		LastPosition:  2700,
		Progress:      75.0,
		Completed:     false,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/watch-history", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRecordWatch_Unauthorized(t *testing.T) {
	setupWatchHistoryTest()

	router := gin.New()
	router.POST("/watch-history", RecordWatch) // No user_id set

	reqBody := RecordWatchRequest{
		MovieID:       1,
		WatchDuration: 3600,
		LastPosition:  2700,
		Progress:      75.0,
		Completed:     false,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/watch-history", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ===== GET WATCH HISTORY TESTS =====

func TestGetWatchHistory_Success(t *testing.T) {
	setupWatchHistoryTest()

	// Create watch history entries
	for i := 1; i <= 5; i++ {
		watchHistory := models.WatchHistory{
			UserID:        1,
			MovieID:       uint(i),
			WatchedAt:     time.Now().Add(-time.Duration(i) * time.Hour),
			WatchDuration: i * 1000,
			Progress:      float32(i * 20),
			LastPosition:  i * 500,
			Completed:     i%2 == 0,
			WatchCount:    1,
			Quality:       "1080p",
			DeviceType:    "web",
		}
		storage.DB.Create(&watchHistory)
	}

	router := gin.New()
	router.GET("/watch-history", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		GetWatchHistory(c)
	})

	req := httptest.NewRequest("GET", "/watch-history?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["pagination"])

	data := response["data"].([]interface{})
	assert.True(t, len(data) > 0)
}

func TestGetWatchHistory_WithFilters(t *testing.T) {
	setupWatchHistoryTest()

	// Create watch history entries
	watchHistory := models.WatchHistory{
		UserID:        1,
		MovieID:       1,
		WatchedAt:     time.Now(),
		WatchDuration: 3600,
		Progress:      100.0,
		LastPosition:  3600,
		Completed:     true,
		WatchCount:    1,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	storage.DB.Create(&watchHistory)

	router := gin.New()
	router.GET("/watch-history", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		GetWatchHistory(c)
	})

	req := httptest.NewRequest("GET", "/watch-history?movie_id=1&completed=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))
}

// ===== GET CONTINUE WATCHING TESTS =====

func TestGetContinueWatching_Success(t *testing.T) {
	setupWatchHistoryTest()

	// Create in-progress watch history
	watchHistory := models.WatchHistory{
		UserID:        1,
		MovieID:       1,
		WatchedAt:     time.Now(),
		WatchDuration: 1800,
		Progress:      50.0,
		LastPosition:  1800,
		Completed:     false,
		WatchCount:    1,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	storage.DB.Create(&watchHistory)

	router := gin.New()
	router.GET("/watch-history/continue-watching", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		GetContinueWatching(c)
	})

	req := httptest.NewRequest("GET", "/watch-history/continue-watching?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.ContinueWatching
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, len(response) > 0)
}

// ===== GET WATCH STATS TESTS =====

func TestGetWatchStats_Success(t *testing.T) {
	setupWatchHistoryTest()

	// Create multiple watch history entries
	for i := 1; i <= 3; i++ {
		watchHistory := models.WatchHistory{
			UserID:        1,
			MovieID:       uint(i),
			WatchedAt:     time.Now().Add(-time.Duration(i) * time.Hour),
			WatchDuration: 3600,
			Progress:      float32(i * 30),
			LastPosition:  3000,
			Completed:     i == 1,
			WatchCount:    1,
			Quality:       "1080p",
			DeviceType:    "web",
		}
		storage.DB.Create(&watchHistory)
	}

	router := gin.New()
	router.GET("/watch-history/stats", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		GetWatchStats(c)
	})

	req := httptest.NewRequest("GET", "/watch-history/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats models.WatchStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.TotalMoviesWatched)
	assert.True(t, stats.TotalWatchTime > 0)
}

// ===== DELETE WATCH HISTORY TESTS =====

func TestDeleteWatchHistory_Success(t *testing.T) {
	setupWatchHistoryTest()

	// Create watch history entry
	watchHistory := models.WatchHistory{
		UserID:        1,
		MovieID:       1,
		WatchedAt:     time.Now(),
		WatchDuration: 3600,
		Progress:      100.0,
		LastPosition:  3600,
		Completed:     true,
		WatchCount:    1,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	storage.DB.Create(&watchHistory)

	router := gin.New()
	router.DELETE("/watch-history/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		DeleteWatchHistory(c)
	})

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/watch-history/%d", watchHistory.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int64
	storage.DB.Model(&models.WatchHistory{}).Where("id = ?", watchHistory.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteWatchHistory_Forbidden(t *testing.T) {
	setupWatchHistoryTest()

	// Create watch history entry for user 1
	watchHistory := models.WatchHistory{
		UserID:        1,
		MovieID:       1,
		WatchedAt:     time.Now(),
		WatchDuration: 3600,
		Progress:      100.0,
		LastPosition:  3600,
		Completed:     true,
		WatchCount:    1,
		Quality:       "1080p",
		DeviceType:    "web",
	}
	storage.DB.Create(&watchHistory)

	router := gin.New()
	router.DELETE("/watch-history/:id", func(c *gin.Context) {
		c.Set("user_id", uint(2)) // Different user trying to delete
		DeleteWatchHistory(c)
	})

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/watch-history/%d", watchHistory.ID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ===== CLEAR WATCH HISTORY TESTS =====

func TestClearWatchHistory_Success(t *testing.T) {
	setupWatchHistoryTest()

	// Create multiple watch history entries
	for i := 1; i <= 3; i++ {
		watchHistory := models.WatchHistory{
			UserID:        1,
			MovieID:       uint(i),
			WatchedAt:     time.Now(),
			WatchDuration: 3600,
			Progress:      100.0,
			LastPosition:  3600,
			Completed:     true,
			WatchCount:    1,
			Quality:       "1080p",
			DeviceType:    "web",
		}
		storage.DB.Create(&watchHistory)
	}

	router := gin.New()
	router.DELETE("/watch-history/clear", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		ClearWatchHistory(c)
	})

	req := httptest.NewRequest("DELETE", "/watch-history/clear", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all entries deleted
	var count int64
	storage.DB.Model(&models.WatchHistory{}).Where("user_id = ?", 1).Count(&count)
	assert.Equal(t, int64(0), count)
}

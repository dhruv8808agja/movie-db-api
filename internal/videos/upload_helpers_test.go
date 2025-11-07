package videos

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

func TestGetUploadStatus_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test upload session
	session := &models.UploadSession{
		ID:             "test-session-123",
		Filename:       "test.mp4",
		FileSize:       1024000,
		ChunkSize:      5242880,
		TotalChunks:    10,
		UploadedChunks: 5,
		Status:         "in_progress",
	}
	storage.DB.Create(session)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/upload/status/:sessionId", GetUploadStatus)

	req := httptest.NewRequest("GET", "/videos/upload/status/test-session-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-session-123")
	assert.Contains(t, w.Body.String(), "in_progress")
}

func TestGetUploadStatus_SessionNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/upload/status/:sessionId", GetUploadStatus)

	req := httptest.NewRequest("GET", "/videos/upload/status/non-existent-session", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "upload session not found")
}

func TestCancelUpload_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test upload session
	session := &models.UploadSession{
		ID:             "test-session-456",
		Filename:       "test.mp4",
		FileSize:       1024000,
		ChunkSize:      5242880,
		TotalChunks:    10,
		UploadedChunks: 5,
		Status:         "in_progress",
		TempDir:        "/tmp/test-session",
	}
	storage.DB.Create(session)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/videos/upload/cancel/:sessionId", CancelUpload)

	req := httptest.NewRequest("DELETE", "/videos/upload/cancel/test-session-456", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cancelled successfully")

	// Verify session status was updated
	var updatedSession models.UploadSession
	storage.DB.First(&updatedSession, "id = ?", "test-session-456")
	assert.Equal(t, "cancelled", updatedSession.Status)
}

func TestCancelUpload_SessionNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/videos/upload/cancel/:sessionId", CancelUpload)

	req := httptest.NewRequest("DELETE", "/videos/upload/cancel/non-existent-session", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "upload session not found")
}

func TestGetUploadStatus_PercentageCalculation(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	testCases := []struct {
		name            string
		uploadedChunks  int
		totalChunks     int
		expectedPercent float64
	}{
		{"0% complete", 0, 10, 0.0},
		{"25% complete", 25, 100, 25.0},
		{"50% complete", 5, 10, 50.0},
		{"75% complete", 75, 100, 75.0},
		{"100% complete", 10, 10, 100.0},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessionID := "test-session-" + string(rune('a'+i))

			session := &models.UploadSession{
				ID:             sessionID,
				Filename:       "test.mp4",
				FileSize:       1024000,
				ChunkSize:      5242880,
				TotalChunks:    tc.totalChunks,
				UploadedChunks: tc.uploadedChunks,
				Status:         "in_progress",
			}
			storage.DB.Create(session)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/videos/upload/status/:sessionId", GetUploadStatus)

			req := httptest.NewRequest("GET", "/videos/upload/status/"+sessionID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

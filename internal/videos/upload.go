package videos

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InitiateUploadRequest represents the request to start a video upload
type InitiateUploadRequest struct {
	Filename string `json:"filename" binding:"required" example:"movie.mp4"`
	FileSize int64  `json:"file_size" binding:"required" example:"1073741824"`
	MovieID  *uint  `json:"movie_id" example:"1"`
}

// InitiateUploadResponse represents the response when upload is initiated
type InitiateUploadResponse struct {
	SessionID   string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ChunkSize   int64  `json:"chunk_size" example:"5242880"`
	TotalChunks int    `json:"total_chunks" example:"205"`
}

// InitiateUpload starts a new video upload session
// @Summary      Initiate video upload
// @Description  Start a new chunked video upload session. Returns session ID and chunk information.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        request  body      InitiateUploadRequest   true  "Upload initiation details"
// @Success      200      {object}  InitiateUploadResponse  "Upload session created"
// @Failure      400      {object}  map[string]string       "Invalid request"
// @Failure      500      {object}  map[string]string       "Internal server error"
// @Security     BearerAuth
// @Router       /videos/upload/initiate [post]
func InitiateUpload(c *gin.Context) {
	var req InitiateUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("failed to bind initiate upload request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate file format
	if err := ValidateVideoFile(req.Filename); err != nil {
		logger.Log.Warn("invalid video format", zap.String("filename", req.Filename), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate file size
	maxSize := storage.GetMaxVideoSize()
	if err := ValidateFileSize(req.FileSize, maxSize); err != nil {
		logger.Log.Warn("invalid file size", zap.Int64("size", req.FileSize), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate movie exists if movieID provided
	if req.MovieID != nil {
		var movie models.Movie
		if err := storage.DB.First(&movie, req.MovieID).Error; err != nil {
			logger.Log.Warn("movie not found", zap.Uint("movie_id", *req.MovieID))
			c.JSON(http.StatusBadRequest, gin.H{"error": "movie not found"})
			return
		}
	}

	// Calculate chunks
	chunkSize := storage.GetUploadChunkSize()
	totalChunks := CalculateTotalChunks(req.FileSize, chunkSize)

	// Generate session ID
	sessionID := uuid.New().String()

	// Create temporary directory for chunks
	tempDir := filepath.Join(getTempUploadDir(), sessionID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Log.Error("failed to create temp directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
		return
	}

	// Create upload session record
	session := models.UploadSession{
		ID:             sessionID,
		MovieID:        req.MovieID,
		Filename:       req.Filename,
		FileSize:       req.FileSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: 0,
		Status:         "in_progress",
		TempDir:        tempDir,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	if err := storage.DB.Create(&session).Error; err != nil {
		logger.Log.Error("failed to create upload session", zap.Error(err))
		// Clean up temp directory
		os.RemoveAll(tempDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload session"})
		return
	}

	logger.Log.Info("upload session initiated",
		zap.String("session_id", sessionID),
		zap.String("filename", req.Filename),
		zap.Int64("file_size", req.FileSize),
		zap.Int("total_chunks", totalChunks))

	c.JSON(http.StatusOK, InitiateUploadResponse{
		SessionID:   sessionID,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
	})
}

// getTempUploadDir returns the temporary upload directory
func getTempUploadDir() string {
	dir := os.Getenv("UPLOAD_TEMP_DIR")
	if dir == "" {
		dir = "./tmp/uploads"
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Log.Error("failed to create temp upload directory", zap.Error(err))
		return "./tmp/uploads"
	}

	return dir
}

// getChunkFilename returns the filename for a specific chunk
func getChunkFilename(sessionID string, chunkIndex int) string {
	return fmt.Sprintf("chunk_%d", chunkIndex)
}

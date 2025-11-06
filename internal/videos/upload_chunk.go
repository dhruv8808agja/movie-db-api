package videos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UploadChunk handles uploading a single chunk
// @Summary      Upload video chunk
// @Description  Upload a single chunk of a video file. Chunks must be uploaded sequentially or in parallel.
// @Tags         videos
// @Accept       multipart/form-data
// @Produce      json
// @Param        X-Session-ID    header    string  true   "Upload session ID"
// @Param        X-Chunk-Index   header    int     true   "Chunk index (0-based)"
// @Param        X-Total-Chunks  header    int     true   "Total number of chunks"
// @Param        chunk           formData  file    true   "Chunk data"
// @Success      200             {object}  models.UploadProgress  "Chunk uploaded successfully"
// @Failure      400             {object}  map[string]string      "Invalid request"
// @Failure      404             {object}  map[string]string      "Session not found"
// @Failure      500             {object}  map[string]string      "Internal server error"
// @Security     BearerAuth
// @Router       /videos/upload/chunk [post]
func UploadChunk(c *gin.Context) {
	// Get headers
	sessionID := c.GetHeader("X-Session-ID")
	chunkIndexStr := c.GetHeader("X-Chunk-Index")
	totalChunksStr := c.GetHeader("X-Total-Chunks")

	if sessionID == "" || chunkIndexStr == "" || totalChunksStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required headers"})
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chunk index"})
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid total chunks"})
		return
	}

	// Validate chunk index
	if err := ValidateChunkIndex(chunkIndex, totalChunks); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get upload session from database
	var session models.UploadSession
	if err := storage.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		logger.Log.Warn("upload session not found", zap.String("session_id", sessionID))
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	// Check if session is still valid
	if session.Status != "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("upload session is %s", session.Status)})
		return
	}

	// Check if session has expired
	if session.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upload session has expired"})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("chunk")
	if err != nil {
		logger.Log.Error("failed to get chunk file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "no chunk file provided"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		logger.Log.Error("failed to open chunk file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process chunk"})
		return
	}
	defer src.Close()

	// Save chunk to temporary directory
	chunkPath := filepath.Join(session.TempDir, getChunkFilename(sessionID, chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		logger.Log.Error("failed to create chunk file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save chunk"})
		return
	}
	defer dst.Close()

	// Copy chunk data
	written, err := io.Copy(dst, src)
	if err != nil {
		logger.Log.Error("failed to write chunk", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save chunk"})
		return
	}

	// Update upload session
	session.UploadedChunks++
	if err := storage.DB.Save(&session).Error; err != nil {
		logger.Log.Error("failed to update session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update upload progress"})
		return
	}

	logger.Log.Info("chunk uploaded",
		zap.String("session_id", sessionID),
		zap.Int("chunk_index", chunkIndex),
		zap.Int64("chunk_size", written),
		zap.Int("uploaded_chunks", session.UploadedChunks),
		zap.Int("total_chunks", session.TotalChunks))

	// Calculate progress
	progress := models.UploadProgress{
		SessionID:       sessionID,
		TotalChunks:     session.TotalChunks,
		UploadedChunks:  session.UploadedChunks,
		PercentComplete: float64(session.UploadedChunks) / float64(session.TotalChunks) * 100,
		Status:          session.Status,
	}

	c.JSON(http.StatusOK, progress)
}

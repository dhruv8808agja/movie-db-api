package videos

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetUploadStatus returns the current status of an upload session
// @Summary      Get upload status
// @Description  Get the current progress and status of a video upload session
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string                 true  "Upload session ID"
// @Success      200        {object}  models.UploadProgress  "Upload progress"
// @Failure      404        {object}  map[string]string      "Session not found"
// @Security     BearerAuth
// @Router       /videos/upload/status/{sessionId} [get]
func GetUploadStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var session models.UploadSession
	if err := storage.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	progress := models.UploadProgress{
		SessionID:       sessionID,
		TotalChunks:     session.TotalChunks,
		UploadedChunks:  session.UploadedChunks,
		PercentComplete: float64(session.UploadedChunks) / float64(session.TotalChunks) * 100,
		Status:          session.Status,
	}

	c.JSON(http.StatusOK, progress)
}

// CancelUpload cancels an ongoing upload session
// @Summary      Cancel upload
// @Description  Cancel an ongoing video upload session and clean up temporary files
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        sessionId  path      string             true  "Upload session ID"
// @Success      200        {object}  map[string]string  "Upload cancelled"
// @Failure      404        {object}  map[string]string  "Session not found"
// @Failure      500        {object}  map[string]string  "Failed to cancel"
// @Security     BearerAuth
// @Router       /videos/upload/cancel/{sessionId} [delete]
func CancelUpload(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var session models.UploadSession
	if err := storage.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	// Update session status
	session.Status = "cancelled"
	if err := storage.DB.Save(&session).Error; err != nil {
		logger.Log.Error("failed to update session status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel upload"})
		return
	}

	// Clean up temp files asynchronously
	go cleanupSessionFiles(&session)

	logger.Log.Info("upload cancelled", zap.String("session_id", sessionID))

	c.JSON(http.StatusOK, gin.H{"message": "upload cancelled successfully"})
}

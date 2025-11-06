package videos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CompleteUploadRequest represents the request to complete an upload
type CompleteUploadRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CompleteUpload merges chunks and finalizes the video upload
// @Summary      Complete video upload
// @Description  Merge all uploaded chunks, upload to storage, and create video record
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        request  body      CompleteUploadRequest  true  "Complete upload request"
// @Success      200      {object}  models.Video           "Upload completed successfully"
// @Failure      400      {object}  map[string]string      "Invalid request or incomplete upload"
// @Failure      404      {object}  map[string]string      "Session not found"
// @Failure      500      {object}  map[string]string      "Internal server error"
// @Security     BearerAuth
// @Router       /videos/upload/complete [post]
func CompleteUpload(c *gin.Context) {
	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get upload session
	var session models.UploadSession
	if err := storage.DB.Where("id = ?", req.SessionID).First(&session).Error; err != nil {
		logger.Log.Warn("upload session not found", zap.String("session_id", req.SessionID))
		c.JSON(http.StatusNotFound, gin.H{"error": "upload session not found"})
		return
	}

	// Verify all chunks are uploaded
	if session.UploadedChunks != session.TotalChunks {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("incomplete upload: %d/%d chunks uploaded",
				session.UploadedChunks, session.TotalChunks),
		})
		return
	}

	// Check session status
	if session.Status == "completed" {
		// Already completed, return existing video
		var video models.Video
		if err := storage.DB.Where("id = ?", session.VideoID).First(&video).Error; err == nil {
			c.JSON(http.StatusOK, video)
			return
		}
	}

	logger.Log.Info("starting upload completion", zap.String("session_id", req.SessionID))

	// Merge chunks into single file
	mergedFilePath, err := mergeChunks(&session)
	if err != nil {
		logger.Log.Error("failed to merge chunks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to merge chunks"})
		return
	}
	defer os.Remove(mergedFilePath) // Clean up merged file after upload

	// Detect MIME type
	mimeType, err := DetectVideoMimeType(mergedFilePath)
	if err != nil {
		logger.Log.Warn("failed to detect MIME type, using default", zap.Error(err))
		mimeType = "video/mp4"
	}

	// Generate storage path
	storagePath := generateStoragePath(&session)

	// Upload to MinIO
	ctx := context.Background()
	if err := storage.UploadFile(ctx, storagePath, mergedFilePath, mimeType); err != nil {
		logger.Log.Error("failed to upload to MinIO", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video"})
		return
	}

	logger.Log.Info("video uploaded to storage", zap.String("path", storagePath))

	// Extract metadata from merged file
	metadata, err := ExtractMetadata(mergedFilePath)
	if err != nil {
		logger.Log.Warn("failed to extract metadata, continuing without it",
			zap.Error(err),
			zap.String("path", mergedFilePath))
		metadata = &VideoMetadata{} // Use empty metadata
	}

	// Generate thumbnail
	thumbnailPath := ""
	thumbnailStoragePath := ""
	if metadata.Duration > 0 { // Only generate if we have duration info
		// Create local thumbnail path
		thumbnailDir := filepath.Join(os.Getenv("UPLOAD_TEMP_DIR"), "thumbnails")
		thumbnailFilename := fmt.Sprintf("%s.jpg", req.SessionID)
		thumbnailLocalPath := filepath.Join(thumbnailDir, thumbnailFilename)

		// Generate thumbnail
		if err := GenerateThumbnailWithDefault(mergedFilePath, thumbnailLocalPath); err != nil {
			logger.Log.Warn("failed to generate thumbnail, continuing without it",
				zap.Error(err),
				zap.String("video_path", mergedFilePath))
		} else {
			// Upload thumbnail to MinIO
			thumbnailStoragePath = generateThumbnailStoragePath(&session, thumbnailFilename)
			contentType := "image/jpeg"

			if err := storage.UploadFile(ctx, thumbnailStoragePath, thumbnailLocalPath, contentType); err != nil {
				logger.Log.Warn("failed to upload thumbnail to storage",
					zap.Error(err),
					zap.String("path", thumbnailStoragePath))
			} else {
				// Generate public URL for thumbnail
				thumbnailURL, err := storage.GetPresignedURL(ctx, thumbnailStoragePath, 604800) // 7 days
				if err == nil {
					thumbnailPath = thumbnailURL
				}
				logger.Log.Info("thumbnail uploaded",
					zap.String("storage_path", thumbnailStoragePath),
					zap.String("url", thumbnailPath))
			}

			// Clean up local thumbnail file
			os.Remove(thumbnailLocalPath)
		}
	}

	// Create video record with metadata
	video := models.Video{
		MovieID:      session.MovieID,
		Title:        session.Filename,
		Filename:     session.Filename,
		OriginalName: session.Filename,
		FileSize:     session.FileSize,
		Duration:     metadata.Duration,
		Width:        metadata.Width,
		Height:       metadata.Height,
		Codec:        metadata.Codec,
		Format:       metadata.Format,
		Bitrate:      metadata.Bitrate,
		FPS:          metadata.FPS,
		StoragePath:  storagePath,
		ThumbnailURL: thumbnailPath,
		UploadStatus: "completed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := storage.DB.Create(&video).Error; err != nil {
		logger.Log.Error("failed to create video record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create video record"})
		return
	}

	// Update session
	session.VideoID = &video.ID
	session.Status = "completed"
	if err := storage.DB.Save(&session).Error; err != nil {
		logger.Log.Error("failed to update session", zap.Error(err))
	}

	// Clean up temporary files
	go cleanupSessionFiles(&session)

	logger.Log.Info("upload completed successfully",
		zap.String("session_id", req.SessionID),
		zap.Uint("video_id", video.ID),
		zap.String("storage_path", storagePath))

	c.JSON(http.StatusOK, video)
}

// mergeChunks merges all chunks into a single file
func mergeChunks(session *models.UploadSession) (string, error) {
	// Create merged file path
	mergedPath := filepath.Join(session.TempDir, "merged_"+session.Filename)

	// Create output file
	outFile, err := os.Create(mergedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create merged file: %w", err)
	}
	defer outFile.Close()

	// Merge chunks in order
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, getChunkFilename(session.ID, i))

		// Open chunk file
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return "", fmt.Errorf("failed to open chunk %d: %w", i, err)
		}

		// Copy chunk to merged file
		_, err = io.Copy(outFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return "", fmt.Errorf("failed to copy chunk %d: %w", i, err)
		}
	}

	return mergedPath, nil
}

// generateStoragePath generates the storage path for a video
func generateStoragePath(session *models.UploadSession) string {
	if session.MovieID != nil {
		return fmt.Sprintf("videos/movie-%d/original/%s-%s",
			*session.MovieID, session.ID, session.Filename)
	}
	return fmt.Sprintf("videos/standalone/%s-%s", session.ID, session.Filename)
}

// generateThumbnailStoragePath generates the MinIO storage path for thumbnails
func generateThumbnailStoragePath(session *models.UploadSession, filename string) string {
	if session.MovieID != nil {
		return fmt.Sprintf("videos/movie-%d/thumbnails/%s",
			*session.MovieID, filename)
	}
	return fmt.Sprintf("videos/standalone/thumbnails/%s", filename)
}

// cleanupSessionFiles removes temporary files for a session
func cleanupSessionFiles(session *models.UploadSession) {
	if session.TempDir != "" {
		if err := os.RemoveAll(session.TempDir); err != nil {
			logger.Log.Error("failed to clean up temp files",
				zap.String("session_id", session.ID),
				zap.Error(err))
		} else {
			logger.Log.Info("cleaned up temp files", zap.String("session_id", session.ID))
		}
	}
}

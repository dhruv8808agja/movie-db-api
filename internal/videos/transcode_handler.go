package videos

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// StartTranscodingRequest represents the request to start transcoding
type StartTranscodingRequest struct {
	VideoID   uint     `json:"video_id" binding:"required"`
	Qualities []string `json:"qualities"` // Optional: if not provided, will auto-select based on source
}

// TranscodingJobResponse represents the response when starting a transcoding job
type TranscodingJobResponse struct {
	JobID           uint     `json:"job_id"`
	VideoID         uint     `json:"video_id"`
	TargetQualities []string `json:"target_qualities"`
	Status          string   `json:"status"`
	Message         string   `json:"message"`
}

// StartTranscoding initiates transcoding for a video
func StartTranscoding(c *gin.Context) {
	var req StartTranscodingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get video from database
	var video models.Video
	if err := storage.DB.First(&video, req.VideoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Check if video upload is completed
	if video.UploadStatus != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video upload not completed"})
		return
	}

	// Determine target qualities
	targetQualities := req.Qualities
	if len(targetQualities) == 0 {
		targetQualities = GetOptimalQualities(video.Width, video.Height)
	}

	// Validate qualities
	for _, quality := range targetQualities {
		if _, exists := QualityPresets[quality]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid quality preset: %s", quality),
			})
			return
		}
	}

	// Create transcoding job
	job := models.TranscodingJob{
		VideoID:         video.ID,
		TargetQualities: strings.Join(targetQualities, ","),
		Status:          "pending",
		Progress:        0,
		CreatedAt:       time.Now(),
	}

	if err := storage.DB.Create(&job).Error; err != nil {
		logger.Log.Error("failed to create transcoding job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transcoding job"})
		return
	}

	logger.Log.Info("transcoding job created",
		zap.Uint("job_id", job.ID),
		zap.Uint("video_id", video.ID),
		zap.Strings("qualities", targetQualities))

	// Start transcoding asynchronously
	go processTranscodingJob(job.ID, video, targetQualities)

	c.JSON(http.StatusOK, TranscodingJobResponse{
		JobID:           job.ID,
		VideoID:         video.ID,
		TargetQualities: targetQualities,
		Status:          "pending",
		Message:         "Transcoding job started successfully",
	})
}

// processTranscodingJob processes transcoding job in background
func processTranscodingJob(jobID uint, video models.Video, qualities []string) {
	ctx := context.Background()

	// Update job status to processing
	now := time.Now()
	storage.DB.Model(&models.TranscodingJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     "processing",
		"started_at": &now,
	})

	logger.Log.Info("starting transcoding job",
		zap.Uint("job_id", jobID),
		zap.Uint("video_id", video.ID))

	// Download original video from MinIO to temp directory
	tempDir := filepath.Join(os.Getenv("UPLOAD_TEMP_DIR"), fmt.Sprintf("transcode_%d", jobID))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		updateJobError(jobID, fmt.Sprintf("failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Download original video
	originalPath := filepath.Join(tempDir, "original.mp4")
	if err := downloadFromMinIO(ctx, video.StoragePath, originalPath); err != nil {
		updateJobError(jobID, fmt.Sprintf("failed to download original video: %v", err))
		return
	}

	totalQualities := len(qualities)
	for i, quality := range qualities {
		// Update current quality
		storage.DB.Model(&models.TranscodingJob{}).Where("id = ?", jobID).Update("current_quality", quality)

		logger.Log.Info("transcoding quality",
			zap.Uint("job_id", jobID),
			zap.String("quality", quality),
			zap.Int("index", i+1),
			zap.Int("total", totalQualities))

		// Transcode
		preset := QualityPresets[quality]
		outputFilename := fmt.Sprintf("video_%s.mp4", quality)
		outputPath := filepath.Join(tempDir, outputFilename)

		opts := TranscodeOptions{
			InputPath:  originalPath,
			OutputPath: outputPath,
			Quality:    preset,
			Codec:      "h264",
			Preset:     "medium",
		}

		result, err := TranscodeVideo(ctx, opts)
		if err != nil {
			logger.Log.Error("transcoding failed",
				zap.Uint("job_id", jobID),
				zap.String("quality", quality),
				zap.Error(err))
			// Continue with other qualities
			continue
		}

		// Upload transcoded video to MinIO
		storagePath := GenerateTranscodedStoragePath(video.ID, quality, outputFilename)
		if err := UploadTranscodedVideo(ctx, outputPath, storagePath); err != nil {
			logger.Log.Error("failed to upload transcoded video",
				zap.Uint("job_id", jobID),
				zap.String("quality", quality),
				zap.Error(err))
			continue
		}

		// Create transcoded video record
		transcodedVideo := models.TranscodedVideo{
			VideoID:     video.ID,
			Quality:     quality,
			Width:       result.Width,
			Height:      result.Height,
			Codec:       "h264",
			Bitrate:     result.Bitrate,
			FileSize:    result.FileSize,
			StoragePath: storagePath,
			Status:      "completed",
			CreatedAt:   time.Now(),
		}
		completedAt := time.Now()
		transcodedVideo.CompletedAt = &completedAt

		if err := storage.DB.Create(&transcodedVideo).Error; err != nil {
			logger.Log.Error("failed to create transcoded video record",
				zap.Uint("job_id", jobID),
				zap.String("quality", quality),
				zap.Error(err))
		}

		// Update progress
		progress := float64(i+1) / float64(totalQualities) * 100.0
		storage.DB.Model(&models.TranscodingJob{}).Where("id = ?", jobID).Update("progress", progress)

		logger.Log.Info("quality transcoded successfully",
			zap.Uint("job_id", jobID),
			zap.String("quality", quality),
			zap.Float64("progress", progress))
	}

	// Mark job as completed
	completedAt := time.Now()
	storage.DB.Model(&models.TranscodingJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":       "completed",
		"progress":     100.0,
		"completed_at": &completedAt,
	})

	logger.Log.Info("transcoding job completed",
		zap.Uint("job_id", jobID),
		zap.Uint("video_id", video.ID))
}

// updateJobError updates job with error status
func updateJobError(jobID uint, errorMsg string) {
	logger.Log.Error("transcoding job failed",
		zap.Uint("job_id", jobID),
		zap.String("error", errorMsg))

	storage.DB.Model(&models.TranscodingJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":        "failed",
		"error_message": errorMsg,
	})
}

// downloadFromMinIO downloads a file from MinIO to local path
func downloadFromMinIO(ctx context.Context, storagePath, localPath string) error {
	logger.Log.Info("downloading from MinIO",
		zap.String("storage_path", storagePath),
		zap.String("local_path", localPath))

	// Use MinIO client to download file
	obj, err := storage.MinioClient.GetObject(ctx, storage.MinioBucket, storagePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object from MinIO: %w", err)
	}
	defer obj.Close()

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer outFile.Close()

	_, err = outFile.ReadFrom(obj)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	logger.Log.Info("download from MinIO completed",
		zap.String("storage_path", storagePath))

	return nil
}

// GetTranscodingJobStatus returns the status of a transcoding job
func GetTranscodingJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")

	var job models.TranscodingJob
	if err := storage.DB.First(&job, jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transcoding job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListTranscodedVideos returns all transcoded versions of a video
func ListTranscodedVideos(c *gin.Context) {
	videoID := c.Param("videoId")

	var transcodedVideos []models.TranscodedVideo
	if err := storage.DB.Where("video_id = ?", videoID).Find(&transcodedVideos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transcoded videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id":          videoID,
		"transcoded_videos": transcodedVideos,
		"count":             len(transcodedVideos),
	})
}

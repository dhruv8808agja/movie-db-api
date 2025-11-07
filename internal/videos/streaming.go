package videos

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// GenerateHLSRequest represents the request to generate HLS for a video
type GenerateHLSRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}

// GenerateHLSResponse represents the response when HLS generation starts
type GenerateHLSResponse struct {
	VideoID        uint   `json:"video_id"`
	MasterPlaylist string `json:"master_playlist_url"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// GenerateHLS generates HLS segments and playlists for a video
func GenerateHLS(c *gin.Context) {
	var req GenerateHLSRequest
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

	// Get all transcoded versions
	var transcodedVideos []models.TranscodedVideo
	if err := storage.DB.Where("video_id = ? AND status = ?", req.VideoID, "completed").
		Find(&transcodedVideos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transcoded videos"})
		return
	}

	if len(transcodedVideos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no transcoded videos found. Please transcode the video first",
		})
		return
	}

	logger.Log.Info("generating HLS",
		zap.Uint("video_id", req.VideoID),
		zap.Int("qualities", len(transcodedVideos)))

	// Start HLS generation asynchronously
	go processHLSGeneration(req.VideoID, video, transcodedVideos)

	masterPlaylistURL := GetHLSMasterPlaylistURL(req.VideoID)

	c.JSON(http.StatusOK, GenerateHLSResponse{
		VideoID:        req.VideoID,
		MasterPlaylist: masterPlaylistURL,
		Status:         "processing",
		Message:        "HLS generation started",
	})
}

// processHLSGeneration generates HLS in background
func processHLSGeneration(videoID uint, video models.Video, transcodedVideos []models.TranscodedVideo) {
	ctx := context.Background()

	logger.Log.Info("starting HLS generation",
		zap.Uint("video_id", videoID),
		zap.Int("qualities", len(transcodedVideos)))

	// Create temp directory for HLS
	hlsDir := filepath.Join(os.Getenv("UPLOAD_TEMP_DIR"), fmt.Sprintf("hls_%d", videoID))
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		logger.Log.Error("failed to create HLS directory", zap.Error(err))
		return
	}
	defer os.RemoveAll(hlsDir)

	// Download all transcoded videos
	var videosPaths []TranscodedVideoPath
	for _, tv := range transcodedVideos {
		localPath := filepath.Join(hlsDir, fmt.Sprintf("%s.mp4", tv.Quality))

		// Download from MinIO
		if err := downloadFromMinIO(ctx, tv.StoragePath, localPath); err != nil {
			logger.Log.Error("failed to download transcoded video",
				zap.String("quality", tv.Quality),
				zap.Error(err))
			continue
		}

		videosPaths = append(videosPaths, TranscodedVideoPath{
			Quality:   tv.Quality,
			LocalPath: localPath,
			Width:     tv.Width,
			Height:    tv.Height,
			Bitrate:   tv.Bitrate,
		})
	}

	if len(videosPaths) == 0 {
		logger.Log.Error("no videos downloaded for HLS generation")
		return
	}

	// Generate HLS
	result, err := GenerateAdaptiveHLS(ctx, videoID, videosPaths)
	if err != nil {
		logger.Log.Error("HLS generation failed",
			zap.Uint("video_id", videoID),
			zap.Error(err))
		return
	}

	// Upload to MinIO
	if err := UploadHLSToStorage(ctx, videoID, hlsDir); err != nil {
		logger.Log.Error("failed to upload HLS to storage",
			zap.Uint("video_id", videoID),
			zap.Error(err))
		return
	}

	logger.Log.Info("HLS generation completed",
		zap.Uint("video_id", videoID),
		zap.Int("segments", result.SegmentCount))
}

// StreamMasterPlaylist serves the master HLS playlist
func StreamMasterPlaylist(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
		return
	}

	// Get video from database
	var video models.Video
	if err := storage.DB.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Get master playlist from MinIO
	storagePath := fmt.Sprintf("videos/video-%d/hls/master.m3u8", videoID)

	// Stream file from MinIO
	streamFileFromMinIO(c, storagePath, "application/vnd.apple.mpegurl")
}

// StreamQualityPlaylist serves a quality-specific HLS playlist
func StreamQualityPlaylist(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	quality := c.Param("quality")

	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
		return
	}

	storagePath := fmt.Sprintf("videos/video-%d/hls/%s/playlist.m3u8", videoID, quality)
	streamFileFromMinIO(c, storagePath, "application/vnd.apple.mpegurl")
}

// StreamSegment serves an HLS video segment
func StreamSegment(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	quality := c.Param("quality")
	segment := c.Param("segment")

	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
		return
	}

	// Validate segment filename (security)
	if !strings.HasSuffix(segment, ".ts") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment"})
		return
	}

	storagePath := fmt.Sprintf("videos/video-%d/hls/%s/%s", videoID, quality, segment)
	streamFileFromMinIO(c, storagePath, "video/mp2t")
}

// streamFileFromMinIO streams a file from MinIO to the client
func streamFileFromMinIO(c *gin.Context, storagePath string, contentType string) {
	ctx := c.Request.Context()

	// Get object from MinIO
	obj, err := storage.MinioClient.GetObject(ctx, storage.MinioBucket, storagePath, minio.GetObjectOptions{})
	if err != nil {
		logger.Log.Error("failed to get object from MinIO",
			zap.String("path", storagePath),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer obj.Close()

	// Get object info for content length
	info, err := obj.Stat()
	if err != nil {
		logger.Log.Error("failed to stat object",
			zap.String("path", storagePath),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	// Set headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	c.Header("Access-Control-Allow-Origin", "*")          // Allow CORS

	// Stream the file
	c.DataFromReader(http.StatusOK, info.Size, contentType, obj, nil)
}

// GetVideoStreamInfo returns streaming information for a video
func GetVideoStreamInfo(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
		return
	}

	// Get video
	var video models.Video
	if err := storage.DB.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Get transcoded versions
	var transcodedVideos []models.TranscodedVideo
	storage.DB.Where("video_id = ? AND status = ?", videoID, "completed").
		Find(&transcodedVideos)

	// Check if HLS exists
	hlsAvailable := checkHLSAvailability(uint(videoID))

	qualities := make([]map[string]interface{}, len(transcodedVideos))
	for i, tv := range transcodedVideos {
		qualities[i] = map[string]interface{}{
			"quality":    tv.Quality,
			"width":      tv.Width,
			"height":     tv.Height,
			"bitrate":    tv.Bitrate,
			"file_size":  tv.FileSize,
			"direct_url": fmt.Sprintf("/videos/%d/download/%s", videoID, tv.Quality),
		}
	}

	response := gin.H{
		"video_id":  videoID,
		"title":     video.Title,
		"duration":  video.Duration,
		"hls": gin.H{
			"available":       hlsAvailable,
			"master_playlist": GetHLSMasterPlaylistURL(uint(videoID)),
		},
		"qualities": qualities,
		"thumbnail": video.ThumbnailURL,
	}

	c.JSON(http.StatusOK, response)
}

// checkHLSAvailability checks if HLS files exist for a video
func checkHLSAvailability(videoID uint) bool {
	ctx := context.Background()
	storagePath := fmt.Sprintf("videos/video-%d/hls/master.m3u8", videoID)

	_, err := storage.MinioClient.StatObject(ctx, storage.MinioBucket, storagePath, minio.StatObjectOptions{})
	return err == nil
}

// DownloadQualityVideo allows direct download of a specific quality
func DownloadQualityVideo(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	quality := c.Param("quality")

	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
		return
	}

	// Get transcoded video
	var transcodedVideo models.TranscodedVideo
	if err := storage.DB.Where("video_id = ? AND quality = ?", videoID, quality).
		First(&transcodedVideo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quality not found"})
		return
	}

	// Get presigned URL (24 hour expiry)
	ctx := c.Request.Context()
	url, err := storage.MinioClient.PresignedGetObject(ctx, storage.MinioBucket,
		transcodedVideo.StoragePath, 24*time.Hour, nil)
	if err != nil {
		logger.Log.Error("failed to generate presigned URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url.String(),
		"quality":      quality,
		"file_size":    transcodedVideo.FileSize,
		"expires_in":   "24 hours",
	})
}

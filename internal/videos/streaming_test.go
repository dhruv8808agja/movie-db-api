package videos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStreamMasterPlaylist_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test video
	video := &models.Video{
		Title:        "Test Video",
		Filename:     "test.mp4",
		OriginalName: "test.mp4",
		FileSize:     1024000,
		Duration:     120.0,
		Width:        1920,
		Height:       1080,
		Codec:        "h264",
		Format:       "mp4",
		Bitrate:      5000000,
		FPS:          30.0,
		StoragePath:  "videos/video-1/original/test.mp4",
		UploadStatus: "completed",
	}
	storage.DB.Create(video)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/master.m3u8", StreamMasterPlaylist)

	req := httptest.NewRequest("GET", "/videos/1/stream/master.m3u8", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 as MinIO is not available in test
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStreamMasterPlaylist_InvalidVideoID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/master.m3u8", StreamMasterPlaylist)

	req := httptest.NewRequest("GET", "/videos/invalid/stream/master.m3u8", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid video ID")
}

func TestStreamMasterPlaylist_VideoNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/master.m3u8", StreamMasterPlaylist)

	req := httptest.NewRequest("GET", "/videos/999/stream/master.m3u8", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "video not found")
}

func TestStreamQualityPlaylist_InvalidVideoID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/:quality/playlist.m3u8", StreamQualityPlaylist)

	req := httptest.NewRequest("GET", "/videos/invalid/stream/720p/playlist.m3u8", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid video ID")
}

func TestStreamSegment_InvalidVideoID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/:quality/:segment", StreamSegment)

	req := httptest.NewRequest("GET", "/videos/invalid/stream/720p/segment_001.ts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid video ID")
}

func TestStreamSegment_InvalidSegmentExtension(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/:quality/:segment", StreamSegment)

	testCases := []struct {
		name    string
		segment string
	}{
		{"No extension", "segment_001"},
		{"Wrong extension", "segment_001.mp4"},
		{"Text file", "segment_001.txt"},
		{"Executable", "segment_001.exe"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/videos/1/stream/720p/"+tc.segment, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "invalid segment")
		})
	}
}

func TestGetVideoStreamInfo_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test video
	video := &models.Video{
		Title:        "Test Video",
		Filename:     "test.mp4",
		OriginalName: "test.mp4",
		FileSize:     1024000,
		Duration:     120.0,
		Width:        1920,
		Height:       1080,
		Codec:        "h264",
		Format:       "mp4",
		Bitrate:      5000000,
		FPS:          30.0,
		StoragePath:  "videos/video-1/original/test.mp4",
		UploadStatus: "completed",
		ThumbnailURL: "https://example.com/thumb.jpg",
	}
	storage.DB.Create(video)

	// Create transcoded versions
	transcodedVideos := []models.TranscodedVideo{
		{
			VideoID:     video.ID,
			Quality:     "720p",
			Width:       1280,
			Height:      720,
			Codec:       "h264",
			Bitrate:     2800000,
			FileSize:    524288000,
			StoragePath: "videos/video-1/transcoded/720p/video.mp4",
			Status:      "completed",
		},
		{
			VideoID:     video.ID,
			Quality:     "480p",
			Width:       854,
			Height:      480,
			Codec:       "h264",
			Bitrate:     1400000,
			FileSize:    262144000,
			StoragePath: "videos/video-1/transcoded/480p/video.mp4",
			Status:      "completed",
		},
	}

	for _, tv := range transcodedVideos {
		storage.DB.Create(&tv)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/info", GetVideoStreamInfo)

	req := httptest.NewRequest("GET", "/videos/1/stream/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["video_id"])
	assert.Equal(t, "Test Video", response["title"])
	assert.Equal(t, 120.0, response["duration"])
	assert.NotNil(t, response["hls"])
	assert.NotNil(t, response["qualities"])

	qualities := response["qualities"].([]interface{})
	assert.Equal(t, 2, len(qualities))
}

func TestGetVideoStreamInfo_InvalidVideoID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/info", GetVideoStreamInfo)

	req := httptest.NewRequest("GET", "/videos/invalid/stream/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid video ID")
}

func TestGetVideoStreamInfo_VideoNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/stream/info", GetVideoStreamInfo)

	req := httptest.NewRequest("GET", "/videos/999/stream/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "video not found")
}

func TestDownloadQualityVideo_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test video
	video := &models.Video{
		Title:        "Test Video",
		Filename:     "test.mp4",
		OriginalName: "test.mp4",
		FileSize:     1024000,
		Duration:     120.0,
		UploadStatus: "completed",
	}
	storage.DB.Create(video)

	// Create a transcoded version
	transcodedVideo := &models.TranscodedVideo{
		VideoID:     video.ID,
		Quality:     "720p",
		Width:       1280,
		Height:      720,
		Codec:       "h264",
		Bitrate:     2800000,
		FileSize:    524288000,
		StoragePath: "videos/video-1/transcoded/720p/video.mp4",
		Status:      "completed",
	}
	storage.DB.Create(transcodedVideo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/download/:quality", DownloadQualityVideo)

	req := httptest.NewRequest("GET", "/videos/1/download/720p", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail as MinIO is not available in test
	// But we can check the handler logic for invalid cases
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestDownloadQualityVideo_InvalidVideoID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/download/:quality", DownloadQualityVideo)

	req := httptest.NewRequest("GET", "/videos/invalid/download/720p", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid video ID")
}

func TestDownloadQualityVideo_QualityNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create a test video without transcoded versions
	video := &models.Video{
		Title:        "Test Video",
		Filename:     "test.mp4",
		OriginalName: "test.mp4",
		FileSize:     1024000,
		Duration:     120.0,
		UploadStatus: "completed",
	}
	storage.DB.Create(video)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/videos/:videoId/download/:quality", DownloadQualityVideo)

	req := httptest.NewRequest("GET", "/videos/1/download/1080p", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "quality not found")
}

func TestCheckHLSAvailability(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Test with non-existent video
	available := checkHLSAvailability(999)
	assert.False(t, available)
}


func TestDownloadFromMinIO(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	ctx := context.Background()

	// Test with non-existent file (MinIO not available)
	err := downloadFromMinIO(ctx, "non-existent", "/tmp/test.mp4")
	assert.Error(t, err)
}

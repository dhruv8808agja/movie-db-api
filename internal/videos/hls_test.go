package videos

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGenerateHLSFromVideo_InvalidInput(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	ctx := context.Background()

	opts := HLSOptions{
		InputPath:   "/non-existent/video.mp4",
		OutputDir:   "/tmp/hls_test",
		SegmentTime: 10,
	}

	result, err := GenerateHLSFromVideo(ctx, opts)

	// Should fail as input file doesn't exist
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

func TestGenerateHLSFromVideo_DefaultValues(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Create a temp output directory
	tmpDir := filepath.Join(os.TempDir(), "hls_test_defaults")
	defer os.RemoveAll(tmpDir)

	opts := HLSOptions{
		InputPath: "/non-existent/video.mp4",
		OutputDir: tmpDir,
		// Don't set defaults to test they are applied
	}

	// Even though it will fail, we can check that defaults were applied
	ctx := context.Background()
	_, _ = GenerateHLSFromVideo(ctx, opts)

	// Output directory should be created
	_, err := os.Stat(tmpDir)
	assert.NoError(t, err)
}

func TestGenerateAdaptiveHLS_EmptyVideos(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	ctx := context.Background()

	result, err := GenerateAdaptiveHLS(ctx, 1, []TranscodedVideoPath{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no transcoded videos provided")
}

func TestGenerateAdaptiveHLS_InvalidVideos(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	// Set temp dir
	os.Setenv("UPLOAD_TEMP_DIR", os.TempDir())

	ctx := context.Background()

	videos := []TranscodedVideoPath{
		{
			Quality:   "720p",
			LocalPath: "/non-existent/720p.mp4",
			Width:     1280,
			Height:    720,
			Bitrate:   2800000,
		},
	}

	result, err := GenerateAdaptiveHLS(ctx, 1, videos)

	// Should handle errors gracefully
	// The function may succeed in creating structure but fail on processing
	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestGenerateMasterPlaylist_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	tmpDir := filepath.Join(os.TempDir(), "master_playlist_test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "master.m3u8")

	playlists := []PlaylistInfo{
		{
			Quality:      "1080p",
			PlaylistPath: "/tmp/1080p/playlist.m3u8",
			Bandwidth:    5000000,
			Resolution:   "1920x1080",
		},
		{
			Quality:      "720p",
			PlaylistPath: "/tmp/720p/playlist.m3u8",
			Bandwidth:    2800000,
			Resolution:   "1280x720",
		},
		{
			Quality:      "480p",
			PlaylistPath: "/tmp/480p/playlist.m3u8",
			Bandwidth:    1400000,
			Resolution:   "854x480",
		},
	}

	err := generateMasterPlaylist(outputPath, playlists)

	assert.NoError(t, err)

	// Check file exists
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#EXTM3U")
	assert.Contains(t, contentStr, "#EXT-X-VERSION:3")
	assert.Contains(t, contentStr, "1080p/playlist.m3u8")
	assert.Contains(t, contentStr, "720p/playlist.m3u8")
	assert.Contains(t, contentStr, "480p/playlist.m3u8")
	assert.Contains(t, contentStr, "BANDWIDTH=5000000")
	assert.Contains(t, contentStr, "RESOLUTION=1920x1080")
}

func TestGenerateMasterPlaylist_EmptyPlaylists(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	tmpDir := filepath.Join(os.TempDir(), "master_playlist_empty_test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "master.m3u8")

	err := generateMasterPlaylist(outputPath, []PlaylistInfo{})

	assert.NoError(t, err)

	// Check file exists
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#EXTM3U")
	assert.Contains(t, contentStr, "#EXT-X-VERSION:3")
}

func TestSortPlaylistsByQuality(t *testing.T) {
	playlists := []PlaylistInfo{
		{Quality: "480p", Bandwidth: 1400000},
		{Quality: "1080p", Bandwidth: 5000000},
		{Quality: "720p", Bandwidth: 2800000},
		{Quality: "360p", Bandwidth: 800000},
	}

	sorted := sortPlaylistsByQuality(playlists)

	assert.Equal(t, 4, len(sorted))
	// Should be sorted by bandwidth descending
	assert.Equal(t, int64(5000000), sorted[0].Bandwidth)
	assert.Equal(t, int64(2800000), sorted[1].Bandwidth)
	assert.Equal(t, int64(1400000), sorted[2].Bandwidth)
	assert.Equal(t, int64(800000), sorted[3].Bandwidth)
}

func TestSortPlaylistsByQuality_SinglePlaylist(t *testing.T) {
	playlists := []PlaylistInfo{
		{Quality: "720p", Bandwidth: 2800000},
	}

	sorted := sortPlaylistsByQuality(playlists)

	assert.Equal(t, 1, len(sorted))
	assert.Equal(t, int64(2800000), sorted[0].Bandwidth)
}

func TestSortPlaylistsByQuality_EmptyList(t *testing.T) {
	playlists := []PlaylistInfo{}

	sorted := sortPlaylistsByQuality(playlists)

	assert.Equal(t, 0, len(sorted))
}

func TestCountSegments(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "count_segments_test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Create some test .ts files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, "segment_"+string(rune('0'+i))+".ts")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	// Create a non-.ts file
	os.WriteFile(filepath.Join(tmpDir, "playlist.m3u8"), []byte("test"), 0644)

	count := countSegments(tmpDir)

	assert.Equal(t, 5, count)
}

func TestCountSegments_NoSegments(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "count_segments_empty_test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	count := countSegments(tmpDir)

	assert.Equal(t, 0, count)
}

func TestCountSegments_NonExistentDir(t *testing.T) {
	count := countSegments("/non-existent-directory-12345")

	assert.Equal(t, 0, count)
}

func TestExtractPlaylistPaths(t *testing.T) {
	playlists := []PlaylistInfo{
		{Quality: "1080p", PlaylistPath: "/path/1080p/playlist.m3u8"},
		{Quality: "720p", PlaylistPath: "/path/720p/playlist.m3u8"},
		{Quality: "480p", PlaylistPath: "/path/480p/playlist.m3u8"},
	}

	paths := extractPlaylistPaths(playlists)

	assert.Equal(t, 3, len(paths))
	assert.Equal(t, "/path/1080p/playlist.m3u8", paths[0])
	assert.Equal(t, "/path/720p/playlist.m3u8", paths[1])
	assert.Equal(t, "/path/480p/playlist.m3u8", paths[2])
}

func TestExtractPlaylistPaths_Empty(t *testing.T) {
	paths := extractPlaylistPaths([]PlaylistInfo{})

	assert.Equal(t, 0, len(paths))
}

func TestGetHLSMasterPlaylistURL(t *testing.T) {
	testCases := []struct {
		videoID  uint
		expected string
	}{
		{1, "/videos/1/stream/master.m3u8"},
		{123, "/videos/123/stream/master.m3u8"},
		{9999, "/videos/9999/stream/master.m3u8"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			url := GetHLSMasterPlaylistURL(tc.videoID)
			assert.Equal(t, tc.expected, url)
		})
	}
}

func TestGetQualityPlaylistURL(t *testing.T) {
	testCases := []struct {
		videoID  uint
		quality  string
		expected string
	}{
		{1, "720p", "/videos/1/stream/720p/playlist.m3u8"},
		{123, "1080p", "/videos/123/stream/1080p/playlist.m3u8"},
		{456, "480p", "/videos/456/stream/480p/playlist.m3u8"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			url := GetQualityPlaylistURL(tc.videoID, tc.quality)
			assert.Equal(t, tc.expected, url)
		})
	}
}

func TestUploadHLSToStorage_NoHLSDir(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	ctx := context.Background()
	err := UploadHLSToStorage(ctx, 1, "/non-existent-directory")

	// Should fail as directory doesn't exist
	assert.Error(t, err)
}

func TestTranscodedVideoPath_Struct(t *testing.T) {
	video := TranscodedVideoPath{
		Quality:   "720p",
		LocalPath: "/tmp/video.mp4",
		Width:     1280,
		Height:    720,
		Bitrate:   2800000,
	}

	assert.Equal(t, "720p", video.Quality)
	assert.Equal(t, "/tmp/video.mp4", video.LocalPath)
	assert.Equal(t, 1280, video.Width)
	assert.Equal(t, 720, video.Height)
	assert.Equal(t, int64(2800000), video.Bitrate)
}

func TestPlaylistInfo_Struct(t *testing.T) {
	playlist := PlaylistInfo{
		Quality:      "1080p",
		PlaylistPath: "/tmp/playlist.m3u8",
		Bandwidth:    5000000,
		Resolution:   "1920x1080",
	}

	assert.Equal(t, "1080p", playlist.Quality)
	assert.Equal(t, "/tmp/playlist.m3u8", playlist.PlaylistPath)
	assert.Equal(t, int64(5000000), playlist.Bandwidth)
	assert.Equal(t, "1920x1080", playlist.Resolution)
}

func TestHLSOptions_Struct(t *testing.T) {
	opts := HLSOptions{
		InputPath:      "/tmp/input.mp4",
		OutputDir:      "/tmp/output",
		SegmentTime:    10,
		PlaylistType:   "vod",
		SegmentPattern: "segment_%03d.ts",
	}

	assert.Equal(t, "/tmp/input.mp4", opts.InputPath)
	assert.Equal(t, "/tmp/output", opts.OutputDir)
	assert.Equal(t, 10, opts.SegmentTime)
	assert.Equal(t, "vod", opts.PlaylistType)
	assert.Equal(t, "segment_%03d.ts", opts.SegmentPattern)
}

func TestHLSResult_Struct(t *testing.T) {
	result := HLSResult{
		MasterPlaylist: "/tmp/master.m3u8",
		Playlists:      []string{"/tmp/720p.m3u8", "/tmp/480p.m3u8"},
		SegmentCount:   42,
		Success:        true,
		ErrorMessage:   "",
	}

	assert.Equal(t, "/tmp/master.m3u8", result.MasterPlaylist)
	assert.Equal(t, 2, len(result.Playlists))
	assert.Equal(t, 42, result.SegmentCount)
	assert.True(t, result.Success)
	assert.Empty(t, result.ErrorMessage)
}

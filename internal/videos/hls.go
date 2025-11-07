package videos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"go.uber.org/zap"
)

// HLSOptions contains options for HLS generation
type HLSOptions struct {
	InputPath      string
	OutputDir      string
	SegmentTime    int    // Segment duration in seconds (default: 10)
	PlaylistType   string // "vod" or "event"
	SegmentPattern string // Pattern for segment filenames
}

// HLSResult contains the result of HLS generation
type HLSResult struct {
	MasterPlaylist string   // Path to master playlist (m3u8)
	Playlists      []string // Paths to quality-specific playlists
	SegmentCount   int      // Total number of segments
	Success        bool
	ErrorMessage   string
}

// GenerateHLSFromVideo generates HLS segments and playlists from a video file
func GenerateHLSFromVideo(ctx context.Context, opts HLSOptions) (*HLSResult, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set defaults
	if opts.SegmentTime == 0 {
		opts.SegmentTime = 10 // 10 second segments
	}
	if opts.PlaylistType == "" {
		opts.PlaylistType = "vod"
	}
	if opts.SegmentPattern == "" {
		opts.SegmentPattern = "segment_%03d.ts"
	}

	playlistPath := filepath.Join(opts.OutputDir, "playlist.m3u8")

	// Build FFmpeg command for HLS
	args := []string{
		"-i", opts.InputPath,
		"-c:v", "copy", // Copy video codec (already encoded)
		"-c:a", "aac",  // AAC audio
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", opts.SegmentTime),
		"-hls_playlist_type", opts.PlaylistType,
		"-hls_segment_filename", filepath.Join(opts.OutputDir, opts.SegmentPattern),
		playlistPath,
	}

	logger.Log.Info("generating HLS segments",
		zap.String("input", opts.InputPath),
		zap.String("output_dir", opts.OutputDir),
		zap.Int("segment_time", opts.SegmentTime))

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Log.Error("HLS generation failed",
			zap.Error(err),
			zap.String("ffmpeg_output", string(output)))
		return &HLSResult{
			Success:      false,
			ErrorMessage: string(output),
		}, fmt.Errorf("HLS generation failed: %w", err)
	}

	// Count segments
	segmentCount := countSegments(opts.OutputDir)

	result := &HLSResult{
		MasterPlaylist: playlistPath,
		Playlists:      []string{playlistPath},
		SegmentCount:   segmentCount,
		Success:        true,
	}

	logger.Log.Info("HLS generation completed",
		zap.String("playlist", playlistPath),
		zap.Int("segments", segmentCount))

	return result, nil
}

// GenerateAdaptiveHLS generates HLS with multiple quality levels
func GenerateAdaptiveHLS(ctx context.Context, videoID uint, transcodedVideos []TranscodedVideoPath) (*HLSResult, error) {
	if len(transcodedVideos) == 0 {
		return nil, fmt.Errorf("no transcoded videos provided")
	}

	// Create HLS directory structure
	baseDir := filepath.Join(os.Getenv("UPLOAD_TEMP_DIR"), fmt.Sprintf("hls_%d", videoID))
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create HLS directory: %w", err)
	}

	var playlists []PlaylistInfo
	var allSegments []string

	// Generate HLS for each quality
	for _, video := range transcodedVideos {
		qualityDir := filepath.Join(baseDir, video.Quality)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			logger.Log.Error("failed to create quality directory",
				zap.String("quality", video.Quality),
				zap.Error(err))
			continue
		}

		opts := HLSOptions{
			InputPath:      video.LocalPath,
			OutputDir:      qualityDir,
			SegmentTime:    10,
			PlaylistType:   "vod",
			SegmentPattern: fmt.Sprintf("%s_%%03d.ts", video.Quality),
		}

		result, err := GenerateHLSFromVideo(ctx, opts)
		if err != nil {
			logger.Log.Error("failed to generate HLS for quality",
				zap.String("quality", video.Quality),
				zap.Error(err))
			continue
		}

		// Collect segments for upload
		segments, _ := filepath.Glob(filepath.Join(qualityDir, "*.ts"))
		allSegments = append(allSegments, segments...)

		playlists = append(playlists, PlaylistInfo{
			Quality:      video.Quality,
			PlaylistPath: result.MasterPlaylist,
			Bandwidth:    video.Bitrate,
			Resolution:   fmt.Sprintf("%dx%d", video.Width, video.Height),
		})
	}

	// Generate master playlist
	masterPlaylistPath := filepath.Join(baseDir, "master.m3u8")
	if err := generateMasterPlaylist(masterPlaylistPath, playlists); err != nil {
		return nil, fmt.Errorf("failed to generate master playlist: %w", err)
	}

	result := &HLSResult{
		MasterPlaylist: masterPlaylistPath,
		Playlists:      extractPlaylistPaths(playlists),
		SegmentCount:   len(allSegments),
		Success:        true,
	}

	logger.Log.Info("adaptive HLS generated",
		zap.String("master_playlist", masterPlaylistPath),
		zap.Int("qualities", len(playlists)),
		zap.Int("total_segments", len(allSegments)))

	return result, nil
}

// TranscodedVideoPath represents a transcoded video with local path
type TranscodedVideoPath struct {
	Quality   string
	LocalPath string
	Width     int
	Height    int
	Bitrate   int64
}

// PlaylistInfo contains information about a quality-specific playlist
type PlaylistInfo struct {
	Quality      string
	PlaylistPath string
	Bandwidth    int64
	Resolution   string
}

// generateMasterPlaylist creates a master m3u8 playlist for adaptive streaming
func generateMasterPlaylist(outputPath string, playlists []PlaylistInfo) error {
	var content strings.Builder
	content.WriteString("#EXTM3U\n")
	content.WriteString("#EXT-X-VERSION:3\n\n")

	// Sort playlists by quality (highest first)
	sortedPlaylists := sortPlaylistsByQuality(playlists)

	for _, p := range sortedPlaylists {
		// Get relative path from master playlist to quality playlist
		relPath := filepath.Join(p.Quality, "playlist.m3u8")

		content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n",
			p.Bandwidth, p.Resolution))
		content.WriteString(relPath + "\n")
	}

	if err := os.WriteFile(outputPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}

	logger.Log.Info("master playlist generated",
		zap.String("path", outputPath),
		zap.Int("variants", len(playlists)))

	return nil
}

// sortPlaylistsByQuality sorts playlists by bandwidth (highest first)
func sortPlaylistsByQuality(playlists []PlaylistInfo) []PlaylistInfo {
	sorted := make([]PlaylistInfo, len(playlists))
	copy(sorted, playlists)

	// Simple bubble sort by bandwidth (descending)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Bandwidth < sorted[j+1].Bandwidth {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// UploadHLSToStorage uploads HLS files to MinIO
func UploadHLSToStorage(ctx context.Context, videoID uint, hlsDir string) error {
	// Upload master playlist
	masterPlaylist := filepath.Join(hlsDir, "master.m3u8")
	masterStoragePath := fmt.Sprintf("videos/video-%d/hls/master.m3u8", videoID)

	if err := storage.UploadFile(ctx, masterStoragePath, masterPlaylist, "application/vnd.apple.mpegurl"); err != nil {
		return fmt.Errorf("failed to upload master playlist: %w", err)
	}

	logger.Log.Info("uploaded master playlist", zap.String("path", masterStoragePath))

	// Upload quality-specific playlists and segments
	qualities := []string{"1080p", "720p", "480p", "360p"}
	for _, quality := range qualities {
		qualityDir := filepath.Join(hlsDir, quality)
		if _, err := os.Stat(qualityDir); os.IsNotExist(err) {
			continue
		}

		// Upload playlist
		playlistPath := filepath.Join(qualityDir, "playlist.m3u8")
		playlistStoragePath := fmt.Sprintf("videos/video-%d/hls/%s/playlist.m3u8", videoID, quality)

		if err := storage.UploadFile(ctx, playlistStoragePath, playlistPath, "application/vnd.apple.mpegurl"); err != nil {
			logger.Log.Error("failed to upload playlist",
				zap.String("quality", quality),
				zap.Error(err))
			continue
		}

		// Upload segments
		segments, err := filepath.Glob(filepath.Join(qualityDir, "*.ts"))
		if err != nil {
			logger.Log.Error("failed to list segments",
				zap.String("quality", quality),
				zap.Error(err))
			continue
		}

		for _, segment := range segments {
			segmentName := filepath.Base(segment)
			segmentStoragePath := fmt.Sprintf("videos/video-%d/hls/%s/%s", videoID, quality, segmentName)

			if err := storage.UploadFile(ctx, segmentStoragePath, segment, "video/mp2t"); err != nil {
				logger.Log.Error("failed to upload segment",
					zap.String("segment", segmentName),
					zap.Error(err))
				continue
			}
		}

		logger.Log.Info("uploaded HLS quality",
			zap.String("quality", quality),
			zap.Int("segments", len(segments)))
	}

	return nil
}

// countSegments counts .ts files in a directory
func countSegments(dir string) int {
	segments, err := filepath.Glob(filepath.Join(dir, "*.ts"))
	if err != nil {
		return 0
	}
	return len(segments)
}

// extractPlaylistPaths extracts playlist paths from PlaylistInfo slice
func extractPlaylistPaths(playlists []PlaylistInfo) []string {
	paths := make([]string, len(playlists))
	for i, p := range playlists {
		paths[i] = p.PlaylistPath
	}
	return paths
}

// GetHLSMasterPlaylistURL returns the public URL for HLS master playlist
func GetHLSMasterPlaylistURL(videoID uint) string {
	return fmt.Sprintf("/videos/%d/stream/master.m3u8", videoID)
}

// GetQualityPlaylistURL returns the URL for a quality-specific playlist
func GetQualityPlaylistURL(videoID uint, quality string) string {
	return fmt.Sprintf("/videos/%d/stream/%s/playlist.m3u8", videoID, quality)
}

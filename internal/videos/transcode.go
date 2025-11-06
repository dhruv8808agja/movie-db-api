package videos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"go.uber.org/zap"
)

// QualityPreset defines a transcoding quality profile
type QualityPreset struct {
	Name       string // e.g., "1080p", "720p", "480p", "360p"
	Width      int
	Height     int
	VideoBitrate string // e.g., "5000k", "2500k"
	AudioBitrate string // e.g., "128k", "96k"
	MaxRate    string // Maximum bitrate for rate control
	BufSize    string // Buffer size for rate control
}

// Common quality presets for video transcoding
var QualityPresets = map[string]QualityPreset{
	"1080p": {
		Name:         "1080p",
		Width:        1920,
		Height:       1080,
		VideoBitrate: "5000k",
		AudioBitrate: "128k",
		MaxRate:      "5350k",
		BufSize:      "7500k",
	},
	"720p": {
		Name:         "720p",
		Width:        1280,
		Height:       720,
		VideoBitrate: "2800k",
		AudioBitrate: "128k",
		MaxRate:      "2996k",
		BufSize:      "4200k",
	},
	"480p": {
		Name:         "480p",
		Width:        854,
		Height:       480,
		VideoBitrate: "1400k",
		AudioBitrate: "96k",
		MaxRate:      "1498k",
		BufSize:      "2100k",
	},
	"360p": {
		Name:         "360p",
		Width:        640,
		Height:       360,
		VideoBitrate: "800k",
		AudioBitrate: "96k",
		MaxRate:      "856k",
		BufSize:      "1200k",
	},
}

// TranscodeOptions contains options for video transcoding
type TranscodeOptions struct {
	InputPath    string
	OutputPath   string
	Quality      QualityPreset
	Codec        string // "h264" or "h265"
	Preset       string // FFmpeg preset: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
	TwoPass      bool   // Enable two-pass encoding for better quality
	ProgressFunc func(percent float64) // Callback for progress updates
}

// TranscodeResult contains the result of transcoding
type TranscodeResult struct {
	OutputPath   string
	Quality      string
	FileSize     int64
	Duration     float64
	Bitrate      int64
	Width        int
	Height       int
	Success      bool
	ErrorMessage string
}

// TranscodeVideo transcodes a video to a specific quality preset
func TranscodeVideo(ctx context.Context, opts TranscodeOptions) (*TranscodeResult, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set defaults
	if opts.Codec == "" {
		opts.Codec = "h264"
	}
	if opts.Preset == "" {
		opts.Preset = "medium"
	}

	result := &TranscodeResult{
		OutputPath: opts.OutputPath,
		Quality:    opts.Quality.Name,
	}

	// Build FFmpeg command
	args := buildTranscodeArgs(opts)

	logger.Log.Info("starting transcoding",
		zap.String("input", opts.InputPath),
		zap.String("output", opts.OutputPath),
		zap.String("quality", opts.Quality.Name),
		zap.String("codec", opts.Codec),
		zap.String("preset", opts.Preset))

	// Execute FFmpeg
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Success = false
		result.ErrorMessage = string(output)
		logger.Log.Error("transcoding failed",
			zap.Error(err),
			zap.String("quality", opts.Quality.Name),
			zap.String("ffmpeg_output", string(output)))
		return result, fmt.Errorf("transcoding failed: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(opts.OutputPath)
	if err == nil {
		result.FileSize = fileInfo.Size()
	}

	// Extract metadata from transcoded file
	metadata, err := ExtractMetadata(opts.OutputPath)
	if err == nil {
		result.Duration = metadata.Duration
		result.Bitrate = metadata.Bitrate
		result.Width = metadata.Width
		result.Height = metadata.Height
	}

	result.Success = true

	logger.Log.Info("transcoding completed",
		zap.String("quality", opts.Quality.Name),
		zap.Int64("file_size", result.FileSize),
		zap.Float64("duration", result.Duration),
		zap.Int("width", result.Width),
		zap.Int("height", result.Height))

	return result, nil
}

// buildTranscodeArgs builds FFmpeg arguments for transcoding
func buildTranscodeArgs(opts TranscodeOptions) []string {
	args := []string{
		"-i", opts.InputPath,
		"-c:v", getVideoCodec(opts.Codec),
		"-preset", opts.Preset,
	}

	// Video settings
	args = append(args,
		"-b:v", opts.Quality.VideoBitrate,
		"-maxrate", opts.Quality.MaxRate,
		"-bufsize", opts.Quality.BufSize,
	)

	// Scale to target resolution
	scaleFilter := fmt.Sprintf("scale=%d:%d", opts.Quality.Width, opts.Quality.Height)
	args = append(args, "-vf", scaleFilter)

	// Audio settings
	args = append(args,
		"-c:a", "aac",
		"-b:a", opts.Quality.AudioBitrate,
		"-ar", "44100", // Sample rate
	)

	// Output settings
	args = append(args,
		"-movflags", "+faststart", // Enable streaming
		"-y", // Overwrite output
		opts.OutputPath,
	)

	return args
}

// getVideoCodec returns the FFmpeg codec name
func getVideoCodec(codec string) string {
	switch codec {
	case "h265", "hevc":
		return "libx265"
	case "h264", "avc":
		return "libx264"
	case "vp9":
		return "libvpx-vp9"
	case "av1":
		return "libaom-av1"
	default:
		return "libx264"
	}
}

// TranscodeToMultipleQualities transcodes a video to multiple quality levels
func TranscodeToMultipleQualities(ctx context.Context, inputPath string, outputDir string, qualities []string) ([]*TranscodeResult, error) {
	if len(qualities) == 0 {
		qualities = []string{"1080p", "720p", "480p", "360p"}
	}

	// Get source video metadata to determine which qualities to generate
	metadata, err := ExtractMetadata(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract source metadata: %w", err)
	}

	var results []*TranscodeResult

	for _, qualityName := range qualities {
		preset, exists := QualityPresets[qualityName]
		if !exists {
			logger.Log.Warn("unknown quality preset, skipping",
				zap.String("quality", qualityName))
			continue
		}

		// Skip if target resolution is higher than source
		if preset.Height > metadata.Height {
			logger.Log.Info("skipping quality higher than source",
				zap.String("quality", qualityName),
				zap.Int("source_height", metadata.Height),
				zap.Int("target_height", preset.Height))
			continue
		}

		// Build output path
		outputFilename := fmt.Sprintf("%s_%s.mp4",
			strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)),
			qualityName)
		outputPath := filepath.Join(outputDir, outputFilename)

		opts := TranscodeOptions{
			InputPath:  inputPath,
			OutputPath: outputPath,
			Quality:    preset,
			Codec:      "h264",
			Preset:     "medium",
			TwoPass:    false,
		}

		result, err := TranscodeVideo(ctx, opts)
		if err != nil {
			logger.Log.Error("failed to transcode quality",
				zap.String("quality", qualityName),
				zap.Error(err))
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// UploadTranscodedVideo uploads a transcoded video to MinIO
func UploadTranscodedVideo(ctx context.Context, localPath string, storagePath string) error {
	contentType := "video/mp4"
	if err := storage.UploadFile(ctx, storagePath, localPath, contentType); err != nil {
		return fmt.Errorf("failed to upload transcoded video: %w", err)
	}

	logger.Log.Info("transcoded video uploaded",
		zap.String("storage_path", storagePath))

	return nil
}

// GenerateTranscodedStoragePath generates the storage path for transcoded videos
func GenerateTranscodedStoragePath(videoID uint, quality string, filename string) string {
	return fmt.Sprintf("videos/video-%d/transcoded/%s/%s", videoID, quality, filename)
}

// EstimateTranscodingTime estimates transcoding time based on duration and quality
func EstimateTranscodingTime(duration float64, quality string) float64 {
	// Rough estimates: transcoding is typically 0.5x-2x realtime depending on quality and preset
	multipliers := map[string]float64{
		"360p": 0.3,
		"480p": 0.5,
		"720p": 1.0,
		"1080p": 2.0,
	}

	multiplier, exists := multipliers[quality]
	if !exists {
		multiplier = 1.0
	}

	return duration * multiplier
}

// GetOptimalQualities determines optimal quality levels based on source resolution
func GetOptimalQualities(sourceWidth, sourceHeight int) []string {
	var qualities []string

	// Only generate qualities equal to or lower than source
	if sourceHeight >= 1080 {
		qualities = append(qualities, "1080p")
	}
	if sourceHeight >= 720 {
		qualities = append(qualities, "720p")
	}
	if sourceHeight >= 480 {
		qualities = append(qualities, "480p")
	}
	if sourceHeight >= 360 {
		qualities = append(qualities, "360p")
	}

	// Always include at least one quality
	if len(qualities) == 0 {
		qualities = []string{"360p"}
	}

	return qualities
}

// ParseQualityFromFilename extracts quality level from filename
func ParseQualityFromFilename(filename string) string {
	for quality := range QualityPresets {
		if strings.Contains(filename, quality) {
			return quality
		}
	}
	return ""
}

// GetTranscodingProgress parses FFmpeg output to extract progress
func GetTranscodingProgress(output string, totalDuration float64) float64 {
	// Parse FFmpeg output for time=XX:XX:XX.XX
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.Contains(line, "time=") {
			parts := strings.Split(line, "time=")
			if len(parts) > 1 {
				timeStr := strings.Fields(parts[1])[0]
				currentTime := parseFFmpegTime(timeStr)
				if totalDuration > 0 {
					return (currentTime / totalDuration) * 100.0
				}
			}
		}
	}
	return 0.0
}

// parseFFmpegTime parses FFmpeg time format (HH:MM:SS.MS) to seconds
func parseFFmpegTime(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0.0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

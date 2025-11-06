package videos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"go.uber.org/zap"
)

// ThumbnailOptions contains options for thumbnail generation
type ThumbnailOptions struct {
	Width      int     // Thumbnail width (0 = auto based on height)
	Height     int     // Thumbnail height (0 = auto based on width)
	TimeOffset float64 // Time offset in seconds to extract frame (default: 1 second)
	Quality    int     // JPEG quality 1-31 (lower is better, default: 2)
}

// DefaultThumbnailOptions returns sensible defaults
func DefaultThumbnailOptions() ThumbnailOptions {
	return ThumbnailOptions{
		Width:      320,  // 320px width
		Height:     0,    // Auto height to maintain aspect ratio
		TimeOffset: 1.0,  // Extract frame at 1 second
		Quality:    2,    // High quality
	}
}

// GenerateThumbnail creates a thumbnail image from a video file
// Returns the path to the generated thumbnail file
func GenerateThumbnail(videoPath string, outputPath string, opts ThumbnailOptions) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		logger.Log.Warn("ffmpeg not found, skipping thumbnail generation",
			zap.String("video_path", videoPath))
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Build scale filter for aspect ratio preservation
	scaleFilter := fmt.Sprintf("scale=%d:%d", opts.Width, opts.Height)
	if opts.Height == 0 {
		scaleFilter = fmt.Sprintf("scale=%d:-1", opts.Width) // Auto height
	} else if opts.Width == 0 {
		scaleFilter = fmt.Sprintf("scale=-1:%d", opts.Height) // Auto width
	}

	// Build ffmpeg command
	// -ss: seek to time offset
	// -i: input file
	// -vf: video filter (scale)
	// -vframes 1: extract only 1 frame
	// -q:v: quality (2 is high quality)
	args := []string{
		"-ss", fmt.Sprintf("%.2f", opts.TimeOffset),
		"-i", videoPath,
		"-vf", scaleFilter,
		"-vframes", "1",
		"-q:v", fmt.Sprintf("%d", opts.Quality),
		"-y", // Overwrite output file
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Log.Error("failed to generate thumbnail",
			zap.Error(err),
			zap.String("video_path", videoPath),
			zap.String("output_path", outputPath),
			zap.String("ffmpeg_output", string(output)))
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	logger.Log.Info("thumbnail generated",
		zap.String("video_path", videoPath),
		zap.String("thumbnail_path", outputPath),
		zap.Int("width", opts.Width),
		zap.Int("height", opts.Height))

	return nil
}

// GenerateThumbnailWithDefault generates a thumbnail with default options
func GenerateThumbnailWithDefault(videoPath string, outputPath string) error {
	return GenerateThumbnail(videoPath, outputPath, DefaultThumbnailOptions())
}

// GenerateMultipleThumbnails creates multiple thumbnails at different time offsets
// Useful for preview strips or selecting the best thumbnail
func GenerateMultipleThumbnails(videoPath string, outputDir string, count int, duration float64) ([]string, error) {
	if count <= 0 {
		count = 3 // Default to 3 thumbnails
	}

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	opts := DefaultThumbnailOptions()
	var thumbnails []string

	// Generate thumbnails at evenly spaced intervals
	for i := 0; i < count; i++ {
		// Calculate time offset (avoid very start and end)
		timeOffset := 1.0 // Start at 1 second
		if duration > 5 && count > 1 {
			// Distribute thumbnails across the video duration
			// Skip first and last 10% of video
			usableDuration := duration * 0.8
			timeOffset = (duration * 0.1) + (usableDuration * float64(i) / float64(count-1))
		}

		outputPath := filepath.Join(outputDir, fmt.Sprintf("thumb_%d.jpg", i))
		opts.TimeOffset = timeOffset

		if err := GenerateThumbnail(videoPath, outputPath, opts); err != nil {
			logger.Log.Warn("failed to generate thumbnail",
				zap.Int("index", i),
				zap.Error(err))
			continue
		}

		thumbnails = append(thumbnails, outputPath)
	}

	if len(thumbnails) == 0 {
		return nil, fmt.Errorf("failed to generate any thumbnails")
	}

	logger.Log.Info("multiple thumbnails generated",
		zap.Int("count", len(thumbnails)),
		zap.String("video_path", videoPath))

	return thumbnails, nil
}

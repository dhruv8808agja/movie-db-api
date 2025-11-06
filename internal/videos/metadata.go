package videos

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"go.uber.org/zap"
)

// VideoMetadata contains extracted video information
type VideoMetadata struct {
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Codec    string  `json:"codec"`
	Format   string  `json:"format"`
	Bitrate  int64   `json:"bitrate"`
	FPS      float64 `json:"fps"`
}

// FFProbeOutput represents the JSON output from ffprobe
type FFProbeOutput struct {
	Streams []struct {
		CodecName      string `json:"codec_name"`
		CodecType      string `json:"codec_type"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		RFrameRate     string `json:"r_frame_rate"`
		AvgFrameRate   string `json:"avg_frame_rate"`
		BitRate        string `json:"bit_rate"`
		Duration       string `json:"duration"`
		DurationTs     int64  `json:"duration_ts"`
		TimeBase       string `json:"time_base"`
		NbFrames       string `json:"nb_frames"`
	} `json:"streams"`
	Format struct {
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		Duration       string `json:"duration"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
	} `json:"format"`
}

// ExtractMetadata uses ffprobe to extract video metadata
func ExtractMetadata(filePath string) (*VideoMetadata, error) {
	// Check if ffprobe is available
	if _, err := exec.LookPath("ffprobe"); err != nil {
		logger.Log.Warn("ffprobe not found, skipping metadata extraction",
			zap.String("path", filePath))
		return &VideoMetadata{}, nil // Return empty metadata instead of error
	}

	// Run ffprobe to get video information as JSON
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		logger.Log.Error("failed to run ffprobe", zap.Error(err), zap.String("path", filePath))
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	var probeData FFProbeOutput
	if err := json.Unmarshal(output, &probeData); err != nil {
		logger.Log.Error("failed to parse ffprobe output", zap.Error(err))
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata := &VideoMetadata{}

	// Extract duration from format (in seconds)
	if probeData.Format.Duration != "" {
		duration, err := strconv.ParseFloat(probeData.Format.Duration, 64)
		if err == nil {
			metadata.Duration = duration
		}
	}

	// Extract format name
	metadata.Format = probeData.Format.FormatName

	// Extract bitrate from format
	if probeData.Format.BitRate != "" {
		bitrate, err := strconv.ParseInt(probeData.Format.BitRate, 10, 64)
		if err == nil {
			metadata.Bitrate = bitrate
		}
	}

	// Find the video stream (first video stream)
	for _, stream := range probeData.Streams {
		if stream.CodecType == "video" {
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.Codec = stream.CodecName

			// Parse frame rate (e.g., "30/1" or "25/1")
			if stream.RFrameRate != "" {
				fps := parseFrameRate(stream.RFrameRate)
				if fps > 0 {
					metadata.FPS = fps
				}
			} else if stream.AvgFrameRate != "" {
				fps := parseFrameRate(stream.AvgFrameRate)
				if fps > 0 {
					metadata.FPS = fps
				}
			}

			// Use stream bitrate if format bitrate is not available
			if metadata.Bitrate == 0 && stream.BitRate != "" {
				bitrate, err := strconv.ParseInt(stream.BitRate, 10, 64)
				if err == nil {
					metadata.Bitrate = bitrate
				}
			}

			// Use stream duration if format duration is not available
			if metadata.Duration == 0 && stream.Duration != "" {
				duration, err := strconv.ParseFloat(stream.Duration, 64)
				if err == nil {
					metadata.Duration = duration
				}
			}

			break // Only process first video stream
		}
	}

	logger.Log.Info("metadata extracted",
		zap.String("path", filePath),
		zap.Float64("duration", metadata.Duration),
		zap.Int("width", metadata.Width),
		zap.Int("height", metadata.Height),
		zap.String("codec", metadata.Codec),
		zap.Float64("fps", metadata.FPS),
	)

	return metadata, nil
}

// parseFrameRate parses frame rate strings like "30/1" or "24000/1001"
func parseFrameRate(frameRate string) float64 {
	parts := strings.Split(frameRate, "/")
	if len(parts) != 2 {
		return 0
	}

	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)

	if err1 != nil || err2 != nil || denominator == 0 {
		return 0
	}

	return numerator / denominator
}

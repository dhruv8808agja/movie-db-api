package videos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

// ValidateVideoFile validates if a file is a valid video format
func ValidateVideoFile(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("file has no extension")
	}

	// Remove the dot from extension
	ext = strings.TrimPrefix(ext, ".")

	// Get allowed formats from environment or use defaults
	allowedFormats := os.Getenv("ALLOWED_VIDEO_FORMATS")
	if allowedFormats == "" {
		allowedFormats = "mp4,avi,mov,mkv,webm,flv"
	}

	formats := strings.Split(allowedFormats, ",")
	for _, format := range formats {
		if strings.TrimSpace(format) == ext {
			return nil
		}
	}

	return fmt.Errorf("file format '%s' not allowed. Allowed formats: %s", ext, allowedFormats)
}

// ValidateFileSize validates if file size is within limits
func ValidateFileSize(size int64, maxSize int64) error {
	if size <= 0 {
		return fmt.Errorf("invalid file size")
	}

	if size > maxSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", size, maxSize)
	}

	return nil
}

// DetectVideoMimeType detects the MIME type of a video file
func DetectVideoMimeType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read first 261 bytes for file type detection
	head := make([]byte, 261)
	_, err = file.Read(head)
	if err != nil {
		return "", err
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return "", err
	}

	if kind == filetype.Unknown {
		return "application/octet-stream", nil
	}

	return kind.MIME.Value, nil
}

// ValidateChunkIndex validates chunk index and total chunks
func ValidateChunkIndex(chunkIndex, totalChunks int) error {
	if chunkIndex < 0 {
		return fmt.Errorf("chunk index cannot be negative")
	}

	if totalChunks <= 0 {
		return fmt.Errorf("total chunks must be greater than 0")
	}

	if chunkIndex >= totalChunks {
		return fmt.Errorf("chunk index %d exceeds total chunks %d", chunkIndex, totalChunks)
	}

	return nil
}

// CalculateTotalChunks calculates the number of chunks for a file
func CalculateTotalChunks(fileSize, chunkSize int64) int {
	if chunkSize <= 0 {
		return 0
	}

	totalChunks := fileSize / chunkSize
	if fileSize%chunkSize != 0 {
		totalChunks++
	}

	return int(totalChunks)
}

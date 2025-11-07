package videos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateVideoFile_ValidFormats(t *testing.T) {
	validFiles := []string{
		"video.mp4",
		"movie.avi",
		"film.mov",
		"content.mkv",
		"clip.webm",
		"video.flv",
	}

	for _, filename := range validFiles {
		t.Run(filename, func(t *testing.T) {
			err := ValidateVideoFile(filename)
			assert.NoError(t, err)
		})
	}
}

func TestValidateVideoFile_InvalidFormats(t *testing.T) {
	invalidFiles := []string{
		"document.pdf",
		"image.jpg",
		"audio.mp3",
		"archive.zip",
		"text.txt",
		"script.exe",
	}

	for _, filename := range invalidFiles {
		t.Run(filename, func(t *testing.T) {
			err := ValidateVideoFile(filename)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not allowed")
		})
	}
}

func TestValidateVideoFile_NoExtension(t *testing.T) {
	err := ValidateVideoFile("videofile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no extension")
}

func TestValidateVideoFile_CaseInsensitive(t *testing.T) {
	testCases := []string{
		"video.MP4",
		"video.Mp4",
		"video.mP4",
		"VIDEO.MP4",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			err := ValidateVideoFile(filename)
			assert.NoError(t, err)
		})
	}
}

func TestValidateVideoFile_CustomFormats(t *testing.T) {
	// Set custom allowed formats
	os.Setenv("ALLOWED_VIDEO_FORMATS", "mp4,mov")
	defer os.Unsetenv("ALLOWED_VIDEO_FORMATS")

	// Should pass
	err := ValidateVideoFile("video.mp4")
	assert.NoError(t, err)

	err = ValidateVideoFile("video.mov")
	assert.NoError(t, err)

	// Should fail (not in custom list)
	err = ValidateVideoFile("video.avi")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateFileSize_ValidSize(t *testing.T) {
	testCases := []struct {
		name    string
		size    int64
		maxSize int64
	}{
		{"Small file", 1024, 1024 * 1024},
		{"Medium file", 1024 * 1024, 10 * 1024 * 1024},
		{"Large file", 100 * 1024 * 1024, 1024 * 1024 * 1024},
		{"Exact max size", 1024, 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFileSize(tc.size, tc.maxSize)
			assert.NoError(t, err)
		})
	}
}

func TestValidateFileSize_InvalidSize(t *testing.T) {
	testCases := []struct {
		name        string
		size        int64
		maxSize     int64
		expectedErr string
	}{
		{"Zero size", 0, 1024, "invalid file size"},
		{"Negative size", -1, 1024, "invalid file size"},
		{"Exceeds max", 2048, 1024, "exceeds maximum allowed size"},
		{"Much larger", 1024 * 1024 * 1024, 100 * 1024 * 1024, "exceeds maximum allowed size"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFileSize(tc.size, tc.maxSize)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestDetectVideoMimeType_NonExistentFile(t *testing.T) {
	_, err := DetectVideoMimeType("/non-existent-file.mp4")
	assert.Error(t, err)
}

func TestDetectVideoMimeType_EmptyFile(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mime_test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Create an empty file
	emptyFile := filepath.Join(tmpDir, "empty.mp4")
	os.WriteFile(emptyFile, []byte{}, 0644)

	mimeType, err := DetectVideoMimeType(emptyFile)
	// Should handle empty file gracefully
	if err == nil {
		assert.NotEmpty(t, mimeType)
	}
}

func TestValidateChunkIndex_Valid(t *testing.T) {
	testCases := []struct {
		chunkIndex  int
		totalChunks int
	}{
		{0, 10},
		{5, 10},
		{9, 10},
		{0, 1},
		{50, 100},
	}

	for _, tc := range testCases {
		t.Run("valid chunk", func(t *testing.T) {
			err := ValidateChunkIndex(tc.chunkIndex, tc.totalChunks)
			assert.NoError(t, err)
		})
	}
}

func TestValidateChunkIndex_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		chunkIndex  int
		totalChunks int
		expectedErr string
	}{
		{"Negative index", -1, 10, "cannot be negative"},
		{"Zero total chunks", 0, 0, "must be greater than 0"},
		{"Negative total chunks", 0, -1, "must be greater than 0"},
		{"Index equals total", 10, 10, "exceeds total chunks"},
		{"Index greater than total", 15, 10, "exceeds total chunks"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateChunkIndex(tc.chunkIndex, tc.totalChunks)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestCalculateTotalChunks(t *testing.T) {
	testCases := []struct {
		name        string
		fileSize    int64
		chunkSize   int64
		expected    int
	}{
		{"Exact division", 1024, 512, 2},
		{"With remainder", 1000, 512, 2},
		{"Single chunk", 512, 1024, 1},
		{"Large file", 10 * 1024 * 1024, 5 * 1024 * 1024, 2},
		{"Many chunks", 100 * 1024 * 1024, 1024 * 1024, 100},
		{"Small remainder", 1025, 1024, 2},
		{"Zero chunk size", 1024, 0, 0},
		{"Negative chunk size", 1024, -1, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateTotalChunks(tc.fileSize, tc.chunkSize)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateTotalChunks_EdgeCases(t *testing.T) {
	// Zero file size
	result := CalculateTotalChunks(0, 1024)
	assert.Equal(t, 0, result)

	// Very large file
	result = CalculateTotalChunks(10*1024*1024*1024, 5*1024*1024)
	expected := int(10 * 1024 * 1024 * 1024 / (5 * 1024 * 1024))
	assert.Equal(t, expected, result)
}

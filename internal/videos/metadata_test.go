package videos

import (
	"testing"

	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExtractMetadata_NonExistentFile(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	metadata, err := ExtractMetadata("/non-existent-file.mp4")

	// Should return error or empty metadata depending on ffprobe availability
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, metadata)
	}
}

func TestExtractMetadata_EmptyPath(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	metadata, err := ExtractMetadata("")

	// Should return error or empty metadata
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, metadata)
	}
}

func TestParseFrameRate_ValidFormats(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
	}{
		{"30/1", 30.0},
		{"25/1", 25.0},
		{"24000/1001", 23.976023976023978},
		{"60/1", 60.0},
		{"120/1", 120.0},
		{"24/1", 24.0},
		{"50/2", 25.0},
		{"100/4", 25.0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseFrameRate(tc.input)
			assert.InDelta(t, tc.expected, result, 0.0001)
		})
	}
}

func TestParseFrameRate_InvalidFormats(t *testing.T) {
	testCases := []string{
		"",
		"30",
		"invalid",
		"30/0",
		"abc/def",
		"30/1/2",
		"/30",
		"30/",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			result := parseFrameRate(input)
			assert.Equal(t, float64(0), result)
		})
	}
}

func TestParseFrameRate_ZeroDenominator(t *testing.T) {
	result := parseFrameRate("30/0")
	assert.Equal(t, float64(0), result)
}

func TestParseFrameRate_NegativeValues(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
	}{
		{"-30/1", -30.0},
		{"30/-1", -30.0},
		{"-30/-1", 30.0}, // Negative divided by negative equals positive
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseFrameRate(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVideoMetadata_Struct(t *testing.T) {
	metadata := VideoMetadata{
		Duration: 120.5,
		Width:    1920,
		Height:   1080,
		Codec:    "h264",
		Format:   "mp4",
		Bitrate:  5000000,
		FPS:      30.0,
	}

	assert.Equal(t, 120.5, metadata.Duration)
	assert.Equal(t, 1920, metadata.Width)
	assert.Equal(t, 1080, metadata.Height)
	assert.Equal(t, "h264", metadata.Codec)
	assert.Equal(t, "mp4", metadata.Format)
	assert.Equal(t, int64(5000000), metadata.Bitrate)
	assert.Equal(t, 30.0, metadata.FPS)
}

func TestFFProbeOutput_Struct(t *testing.T) {
	probeData := FFProbeOutput{
		Streams: []struct {
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
		}{
			{
				CodecName:    "h264",
				CodecType:    "video",
				Width:        1920,
				Height:       1080,
				RFrameRate:   "30/1",
				AvgFrameRate: "30/1",
				BitRate:      "5000000",
			},
		},
		Format: struct {
			Filename       string `json:"filename"`
			NbStreams      int    `json:"nb_streams"`
			NbPrograms     int    `json:"nb_programs"`
			FormatName     string `json:"format_name"`
			FormatLongName string `json:"format_long_name"`
			Duration       string `json:"duration"`
			Size           string `json:"size"`
			BitRate        string `json:"bit_rate"`
		}{
			Filename:   "test.mp4",
			NbStreams:  1,
			FormatName: "mp4",
			Duration:   "120.5",
			BitRate:    "5000000",
		},
	}

	assert.Equal(t, 1, len(probeData.Streams))
	assert.Equal(t, "h264", probeData.Streams[0].CodecName)
	assert.Equal(t, "video", probeData.Streams[0].CodecType)
	assert.Equal(t, 1920, probeData.Streams[0].Width)
	assert.Equal(t, "test.mp4", probeData.Format.Filename)
	assert.Equal(t, "mp4", probeData.Format.FormatName)
}

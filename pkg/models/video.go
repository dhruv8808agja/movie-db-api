package models

import (
	"time"
)

// Video represents a video file associated with a movie
type Video struct {
	ID           uint      `gorm:"primaryKey" json:"id" example:"1"`
	MovieID      *uint     `json:"movie_id" example:"1"` // Foreign key to movies table (nullable for standalone videos)
	Title        string    `json:"title" example:"The Matrix - Full Movie"`
	Filename     string    `json:"filename" example:"matrix_1999.mp4"`
	OriginalName string    `json:"original_name" example:"The Matrix (1999).mp4"`
	FileSize     int64     `json:"file_size" example:"2147483648"` // in bytes
	Duration     float64   `json:"duration" example:"136.5"`       // in seconds
	Width        int       `json:"width" example:"1920"`
	Height       int       `json:"height" example:"1080"`
	Codec        string    `json:"codec" example:"h264"`
	Format       string    `json:"format" example:"mp4"`
	Bitrate      int64     `json:"bitrate" example:"5000000"` // bits per second
	FPS          float64   `json:"fps" example:"23.976"`
	StoragePath  string    `json:"storage_path" example:"videos/movie-1/original/video-1.mp4"`
	ThumbnailURL string    `json:"thumbnail_url" example:"https://cdn.example.com/thumbnails/video-1-thumb.jpg"`
	UploadStatus string    `json:"upload_status" example:"completed"` // pending, uploading, processing, completed, failed
	CreatedAt    time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// UploadSession tracks chunk upload progress
type UploadSession struct {
	ID             string    `gorm:"primaryKey" json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // UUID
	VideoID        *uint     `json:"video_id" example:"1"`
	MovieID        *uint     `json:"movie_id" example:"1"`
	Filename       string    `json:"filename" example:"movie.mp4"`
	FileSize       int64     `json:"file_size" example:"1073741824"`
	ChunkSize      int64     `json:"chunk_size" example:"5242880"`
	TotalChunks    int       `json:"total_chunks" example:"205"`
	UploadedChunks int       `json:"uploaded_chunks" example:"102"`
	Status         string    `json:"status" example:"in_progress"` // in_progress, completed, failed, cancelled
	TempDir        string    `json:"-"`                            // Don't expose in JSON
	CreatedAt      time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	ExpiresAt      time.Time `json:"expires_at" example:"2024-01-01T01:00:00Z"`
}

// UploadProgress represents upload progress information
type UploadProgress struct {
	SessionID      string  `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TotalChunks    int     `json:"total_chunks" example:"205"`
	UploadedChunks int     `json:"uploaded_chunks" example:"102"`
	PercentComplete float64 `json:"percent_complete" example:"49.76"`
	Status         string  `json:"status" example:"in_progress"`
}

// VideoMetadata represents extracted video metadata
type VideoMetadata struct {
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Codec    string  `json:"codec"`
	Format   string  `json:"format"`
	Bitrate  int64   `json:"bitrate"`
	FPS      float64 `json:"fps"`
}

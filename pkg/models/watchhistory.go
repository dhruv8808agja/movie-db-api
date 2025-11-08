package models

import "time"

// WatchHistory represents a user's watch history for a movie
type WatchHistory struct {
	ID              uint      `gorm:"primaryKey" json:"id" example:"1"`
	UserID          uint      `gorm:"index;not null" json:"user_id" example:"1"`
	MovieID         uint      `gorm:"index;not null" json:"movie_id" example:"1"`
	WatchedAt       time.Time `gorm:"index;not null" json:"watched_at" example:"2024-01-01T12:00:00Z"`
	WatchDuration   int       `json:"watch_duration" example:"3600"` // In seconds
	Progress        float32   `json:"progress" example:"75.5"` // Percentage (0-100)
	Completed       bool      `json:"completed" example:"false"`
	LastPosition    int       `json:"last_position" example:"2700"` // Last playback position in seconds
	WatchCount      int       `gorm:"default:1" json:"watch_count" example:"1"` // Number of times watched
	Quality         string    `json:"quality" example:"1080p"` // Video quality watched
	DeviceType      string    `json:"device_type" example:"web"` // web, mobile, tv, etc.

	// Relationships
	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Movie Movie `gorm:"foreignKey:MovieID" json:"-"`
}

// WatchHistoryWithMovie includes movie details in the response
type WatchHistoryWithMovie struct {
	ID              uint      `json:"id" example:"1"`
	UserID          uint      `json:"user_id" example:"1"`
	MovieID         uint      `json:"movie_id" example:"1"`
	WatchedAt       time.Time `json:"watched_at" example:"2024-01-01T12:00:00Z"`
	WatchDuration   int       `json:"watch_duration" example:"3600"`
	Progress        float32   `json:"progress" example:"75.5"`
	Completed       bool      `json:"completed" example:"false"`
	LastPosition    int       `json:"last_position" example:"2700"`
	WatchCount      int       `json:"watch_count" example:"1"`
	Quality         string    `json:"quality" example:"1080p"`
	DeviceType      string    `json:"device_type" example:"web"`
	Movie           Movie     `json:"movie"`
}

// WatchStats represents aggregated watch statistics for a user
type WatchStats struct {
	TotalMoviesWatched  int       `json:"total_movies_watched" example:"42"`
	TotalWatchTime      int       `json:"total_watch_time" example:"151200"` // Total seconds
	CompletedMovies     int       `json:"completed_movies" example:"35"`
	InProgressMovies    int       `json:"in_progress_movies" example:"7"`
	FavoriteGenres      []string  `json:"favorite_genres" example:"Action,Drama"`
	LastWatchedAt       time.Time `json:"last_watched_at" example:"2024-01-01T12:00:00Z"`
	AverageProgress     float32   `json:"average_progress" example:"82.5"`
}

// ContinueWatching represents a movie that the user is currently watching
type ContinueWatching struct {
	WatchHistoryID  uint      `json:"watch_history_id" example:"1"`
	MovieID         uint      `json:"movie_id" example:"1"`
	Movie           Movie     `json:"movie"`
	LastPosition    int       `json:"last_position" example:"2700"`
	Progress        float32   `json:"progress" example:"75.5"`
	WatchedAt       time.Time `json:"watched_at" example:"2024-01-01T12:00:00Z"`
}

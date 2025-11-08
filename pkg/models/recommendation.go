package models

import "time"

// UserInteraction represents a user's interaction with a movie
type UserInteraction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	MovieID       uint      `gorm:"index" json:"movie_id"`
	InteractionType string  `json:"interaction_type"` // view, like, watch, rate
	Rating        *float32  `json:"rating,omitempty"` // Optional rating (1-10)
	WatchDuration *int      `json:"watch_duration,omitempty"` // In seconds
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Movie Movie `gorm:"foreignKey:MovieID" json:"-"`
}

// UserMovieRating represents a user's rating for a movie
type UserMovieRating struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	MovieID   uint      `gorm:"primaryKey;autoIncrement:false" json:"movie_id"`
	Rating    float32   `json:"rating" minimum:"0" maximum:"10"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MovieRecommendation represents a recommended movie for a user
type MovieRecommendation struct {
	Movie           Movie   `json:"movie"`
	Score           float64 `json:"score"`
	ReasonType      string  `json:"reason_type"` // collaborative, content-based, trending
	Reason          string  `json:"reason"`
}

// UserPreferences stores aggregated user preferences
type UserPreferences struct {
	UserID          uint      `gorm:"primaryKey" json:"user_id"`
	FavoriteGenres  Genres    `json:"favorite_genres"`
	FavoriteDirectors []string `gorm:"type:json" json:"favorite_directors"`
	AverageRating   float32   `json:"average_rating"`
	LastUpdated     time.Time `json:"last_updated"`
}

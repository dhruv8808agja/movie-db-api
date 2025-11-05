package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Genres is a custom type to store string slices as JSON in DB
type Genres []string

// Value converts the slice to JSON for storing in DB
func (g Genres) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan converts JSON from DB back into a slice
func (g *Genres) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Genres")
	}
	return json.Unmarshal(bytes, g)
}

// Movie is the main entity
type Movie struct {
	ID          uint      `gorm:"primaryKey" json:"id" example:"1"`
	Title       string    `json:"title" example:"The Matrix"`
	Description string    `json:"description" example:"A computer hacker learns from mysterious rebels about the true nature of his reality"`
	Director    string    `json:"director" example:"The Wachowskis"`
	ReleaseDate time.Time `json:"release_date" example:"1999-03-31T00:00:00Z"`
	Genres      Genres    `json:"genres" swaggertype:"array,string" example:"Action,Sci-Fi"` // stored as JSON
	Rating      float32   `json:"rating" example:"8.7" minimum:"0" maximum:"10"`
}

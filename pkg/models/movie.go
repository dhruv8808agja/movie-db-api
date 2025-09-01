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
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Director    string    `json:"director"`
	ReleaseDate time.Time `json:"release_date"`
	Genres      Genres    `json:"genres"` // stored as JSON
	Rating      float32   `json:"rating"`
}

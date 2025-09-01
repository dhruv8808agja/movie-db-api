package movie

import "time"

type Movie struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Director    string    `json:"director"`
	ReleaseDate time.Time `json:"release_date"`
	Genres      []string  `json:"genres"`
	Rating      float32   `json:"rating"`
}

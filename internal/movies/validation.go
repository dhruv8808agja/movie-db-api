package movies

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dhruv8808agja/movie-db-api/pkg/models"
)

// ValidateMovie validates movie fields for create and update operations
func ValidateMovie(movie *models.Movie) error {
	var validationErrors []string

	// Title validation - required, min 1 char, max 200 chars
	title := strings.TrimSpace(movie.Title)
	if title == "" {
		validationErrors = append(validationErrors, "title is required")
	} else if len(title) > 200 {
		validationErrors = append(validationErrors, "title must not exceed 200 characters")
	}

	// Description validation - optional, max 1000 chars
	if len(movie.Description) > 1000 {
		validationErrors = append(validationErrors, "description must not exceed 1000 characters")
	}

	// Director validation - optional but if provided, min 1 char after trimming
	director := strings.TrimSpace(movie.Director)
	if director == "" && movie.Director != "" {
		validationErrors = append(validationErrors, "director cannot be only whitespace")
	} else if len(director) > 100 {
		validationErrors = append(validationErrors, "director name must not exceed 100 characters")
	}

	// ReleaseDate validation - should be between 1800 and 5 years in the future
	minYear := 1800
	maxYear := time.Now().Year() + 5
	releaseYear := movie.ReleaseDate.Year()

	if !movie.ReleaseDate.IsZero() {
		if releaseYear < minYear || releaseYear > maxYear {
			validationErrors = append(validationErrors,
				fmt.Sprintf("release year must be between %d and %d", minYear, maxYear))
		}
	}

	// Rating validation - should be between 0 and 10
	if movie.Rating < 0 || movie.Rating > 10 {
		validationErrors = append(validationErrors, "rating must be between 0 and 10")
	}

	// Genres validation - each genre should be non-empty
	if len(movie.Genres) > 0 {
		for i, genre := range movie.Genres {
			if strings.TrimSpace(genre) == "" {
				validationErrors = append(validationErrors,
					fmt.Sprintf("genre at index %d cannot be empty", i))
			}
		}

		if len(movie.Genres) > 10 {
			validationErrors = append(validationErrors, "maximum 10 genres allowed")
		}
	}

	// Return combined error if any validations failed
	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

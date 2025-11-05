package movies

import (
	"strings"
	"testing"
	"time"

	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateMovie_ValidMovie(t *testing.T) {
	movie := &models.Movie{
		Title:       "The Matrix",
		Description: "A computer hacker learns about the true nature of reality",
		Director:    "The Wachowskis",
		ReleaseDate: time.Date(1999, 3, 31, 0, 0, 0, 0, time.UTC),
		Genres:      models.Genres{"Action", "Sci-Fi"},
		Rating:      8.7,
	}

	err := ValidateMovie(movie)
	assert.NoError(t, err)
}

func TestValidateMovie_EmptyTitle(t *testing.T) {
	movie := &models.Movie{
		Title:       "",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestValidateMovie_WhitespaceTitle(t *testing.T) {
	movie := &models.Movie{
		Title:       "   ",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestValidateMovie_TitleTooLong(t *testing.T) {
	movie := &models.Movie{
		Title:       strings.Repeat("a", 201),
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title must not exceed 200 characters")
}

func TestValidateMovie_DescriptionTooLong(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: strings.Repeat("a", 1001),
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description must not exceed 1000 characters")
}

func TestValidateMovie_DirectorWhitespace(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "   ",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "director cannot be only whitespace")
}

func TestValidateMovie_DirectorTooLong(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    strings.Repeat("a", 101),
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "director name must not exceed 100 characters")
}

func TestValidateMovie_ReleaseYearTooOld(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(1799, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "release year must be between")
}

func TestValidateMovie_ReleaseYearTooFuture(t *testing.T) {
	futureYear := time.Now().Year() + 10
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(futureYear, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "release year must be between")
}

func TestValidateMovie_RatingTooLow(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      -1.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rating must be between 0 and 10")
}

func TestValidateMovie_RatingTooHigh(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      11.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rating must be between 0 and 10")
}

func TestValidateMovie_EmptyGenre(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Genres:      models.Genres{"Action", "", "Drama"},
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "genre at index 1 cannot be empty")
}

func TestValidateMovie_TooManyGenres(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Genres:      models.Genres{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
		Rating:      5.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum 10 genres allowed")
}

func TestValidateMovie_MultipleErrors(t *testing.T) {
	movie := &models.Movie{
		Title:       "",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      15.0,
	}

	err := ValidateMovie(movie)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
	assert.Contains(t, err.Error(), "rating must be between 0 and 10")
}

func TestValidateMovie_EdgeCaseRatingZero(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      0.0,
	}

	err := ValidateMovie(movie)
	assert.NoError(t, err)
}

func TestValidateMovie_EdgeCaseRatingTen(t *testing.T) {
	movie := &models.Movie{
		Title:       "Valid Title",
		Description: "Description",
		Director:    "Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Rating:      10.0,
	}

	err := ValidateMovie(movie)
	assert.NoError(t, err)
}

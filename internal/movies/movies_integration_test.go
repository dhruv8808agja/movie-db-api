package movies

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func getAuthToken() string {
	testutil.SetTestEnv()
	token, _ := auth.GenerateToken("testuser")
	return token
}

func TestCreateMovie_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.POST("/movies", CreateMovie)

	movie := testutil.CreateTestMovie()
	jsonData, _ := json.Marshal(movie)

	req := httptest.NewRequest("POST", "/movies", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Movie
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, movie.Title, response.Title)
	assert.NotZero(t, response.ID)
}

func TestCreateMovie_ValidationError(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.POST("/movies", CreateMovie)

	// Invalid movie with empty title
	movie := &models.Movie{
		Title:  "",
		Rating: 5.0,
	}
	jsonData, _ := json.Marshal(movie)

	req := httptest.NewRequest("POST", "/movies", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "title is required")
}

func TestCreateMovie_InvalidJSON(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.POST("/movies", CreateMovie)

	req := httptest.NewRequest("POST", "/movies", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMovies_BulkSuccess(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.POST("/movies/bulk", CreateMovies)

	movies := []models.Movie{
		*testutil.CreateTestMovie(),
		{
			Title:       "Another Movie",
			Description: "Another description",
			Director:    "Another Director",
			ReleaseDate: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Comedy"},
			Rating:      7.5,
		},
	}
	jsonData, _ := json.Marshal(movies)

	req := httptest.NewRequest("POST", "/movies/bulk", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response []models.Movie
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestCreateMovies_BulkValidationError(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.POST("/movies/bulk", CreateMovies)

	movies := []models.Movie{
		*testutil.CreateTestMovie(),
		{
			Title:  "", // Invalid
			Rating: 5.0,
		},
	}
	jsonData, _ := json.Marshal(movies)

	req := httptest.NewRequest("POST", "/movies/bulk", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed for movie at index 1")
}

func TestGetMovieByID_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	testutil.InitTestStorage()
	defer testutil.CleanupTestRedis(storage.RedisClient)

	// Create a movie first
	movie := testutil.CreateTestMovie()
	storage.DB.Create(movie)

	router := setupRouter()
	router.GET("/movies/:id", GetMovieByID)

	req := httptest.NewRequest("GET", "/movies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Movie
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, movie.Title, response.Title)
}

func TestGetMovieByID_NotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.GET("/movies/:id", GetMovieByID)

	req := httptest.NewRequest("GET", "/movies/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMovieByID_InvalidID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.GET("/movies/:id", GetMovieByID)

	req := httptest.NewRequest("GET", "/movies/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateMovie_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	// Create a movie first
	movie := testutil.CreateTestMovie()
	storage.DB.Create(movie)

	router := setupRouter()
	router.PUT("/movies/:id", UpdateMovie)

	// Update the movie
	updatedMovie := &models.Movie{
		Title:       "Updated Title",
		Description: "Updated description",
		Director:    "Updated Director",
		ReleaseDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Genres:      models.Genres{"Drama"},
		Rating:      9.0,
	}
	jsonData, _ := json.Marshal(updatedMovie)

	req := httptest.NewRequest("PUT", "/movies/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Movie
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", response.Title)
	assert.Equal(t, uint(1), response.ID)
}

func TestUpdateMovie_NotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.PUT("/movies/:id", UpdateMovie)

	updatedMovie := testutil.CreateTestMovie()
	jsonData, _ := json.Marshal(updatedMovie)

	req := httptest.NewRequest("PUT", "/movies/999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMovie_ValidationError(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	// Create a movie first
	movie := testutil.CreateTestMovie()
	storage.DB.Create(movie)

	router := setupRouter()
	router.PUT("/movies/:id", UpdateMovie)

	// Invalid update
	updatedMovie := &models.Movie{
		Title:  "", // Invalid
		Rating: 5.0,
	}
	jsonData, _ := json.Marshal(updatedMovie)

	req := httptest.NewRequest("PUT", "/movies/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMovie_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	// Create a movie first
	movie := testutil.CreateTestMovie()
	storage.DB.Create(movie)

	router := setupRouter()
	router.DELETE("/movies/:id", DeleteMovie)

	req := httptest.NewRequest("DELETE", "/movies/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify movie is deleted
	var count int64
	storage.DB.Model(&models.Movie{}).Where("id = ?", 1).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteMovie_NotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	router := setupRouter()
	router.DELETE("/movies/:id", DeleteMovie)

	req := httptest.NewRequest("DELETE", "/movies/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListMoviesWithFilter_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()
	storage.DB = testutil.SetupTestDB()

	// Create test movies
	movies := testutil.CreateTestMovies(3)
	for i := range movies {
		storage.DB.Create(&movies[i])
	}

	router := setupRouter()
	router.GET("/movies", ListMoviesWithFilter)

	req := httptest.NewRequest("GET", "/movies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

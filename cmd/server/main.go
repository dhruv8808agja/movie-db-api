package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/dhruv8808agja/movie-db-api/docs" // Import generated docs
	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/middleware"
	"github.com/dhruv8808agja/movie-db-api/internal/monitor"
	"github.com/dhruv8808agja/movie-db-api/internal/movies"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/users"
	"github.com/dhruv8808agja/movie-db-api/internal/videos"
)

// @title           Movie Database API
// @version         1.0
// @description     A movie database API with authentication, rate limiting, and caching
// @description     This API allows you to manage movies with CRUD operations
//
// @contact.name    API Support
// @contact.email   support@moviedb.com
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and JWT token
//
// @tag.name         auth
// @tag.description  Authentication endpoints
//
// @tag.name         users
// @tag.description  User management and profile endpoints
//
// @tag.name         movies
// @tag.description  Movie management endpoints
func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize DB, Redis, and MinIO
	storage.InitDB()
	storage.InitRedis()
	storage.InitMinIO()

	// Setup Gin router
	r := gin.New()
	r.Use(logger.GinLogger(), gin.Recovery()) // logging + recovery

	// Setup rate limiting
	rateLimitConfig := middleware.GetRateLimiterConfig()
	r.Use(middleware.RateLimiter(rateLimitConfig))

	// Public routes
	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Video player demo
	r.StaticFile("/player", "./public/player.html")

	// Authentication & User Registration
	r.POST("/login", auth.Login)
	r.POST("/register", users.Register)

	// Public user profiles
	r.GET("/users/:id", users.GetProfileByID)

	// Read
	r.GET("/movies", movies.ListMoviesWithFilter)

	// Secured routes
	secured := r.Group("/")
	secured.Use(auth.JWTMiddleware())

	// User profile routes (authenticated users)
	secured.GET("/profile", users.GetProfile)
	secured.PUT("/profile", users.UpdateProfile)
	secured.DELETE("/profile", users.DeleteProfile)

	// Admin routes (admin only)
	admin := secured.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	admin.GET("/users", users.ListUsers)
	admin.PUT("/users/:id/role", users.UpdateUserRole)
	admin.POST("/users/:id/deactivate", users.DeactivateUser)
	admin.POST("/users/:id/activate", users.ActivateUser)

	// Movie routes
	// Create
	secured.POST("/movies", movies.CreateMovie)
	secured.POST("/movies/bulk", movies.CreateMovies)

	// Read
	secured.GET("/movies/:id", movies.GetMovieByID)

	// Update
	secured.PUT("/movies/:id", movies.UpdateMovie)
	// Delete
	secured.DELETE("/movies/:id", movies.DeleteMovie)
	secured.DELETE("/movies", movies.DeleteMovies)

	// Video upload routes
	secured.POST("/videos/upload/initiate", videos.InitiateUpload)
	secured.POST("/videos/upload/chunk", videos.UploadChunk)
	secured.POST("/videos/upload/complete", videos.CompleteUpload)
	secured.GET("/videos/upload/status/:sessionId", videos.GetUploadStatus)
	secured.DELETE("/videos/upload/cancel/:sessionId", videos.CancelUpload)

	// Video transcoding routes
	secured.POST("/videos/transcode", videos.StartTranscoding)
	secured.GET("/videos/transcode/job/:jobId", videos.GetTranscodingJobStatus)
	secured.GET("/videos/:videoId/transcoded", videos.ListTranscodedVideos)

	// Video streaming routes
	secured.POST("/videos/hls/generate", videos.GenerateHLS)
	secured.GET("/videos/:videoId/info", videos.GetVideoStreamInfo)
	secured.GET("/videos/:videoId/download/:quality", videos.DownloadQualityVideo)

	// Public streaming routes (no auth required for playback)
	r.GET("/videos/:videoId/stream/master.m3u8", videos.StreamMasterPlaylist)
	r.GET("/videos/:videoId/stream/:quality/playlist.m3u8", videos.StreamQualityPlaylist)
	r.GET("/videos/:videoId/stream/:quality/:segment", videos.StreamSegment)

	// Prometheus metrics
	monitor.RegisterMetrics(r)

	r.Run(":8080")
}

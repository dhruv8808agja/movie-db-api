package users

import (
	"net/http"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email     string `json:"email" binding:"required,email" example:"john@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"securepassword123"`
	FirstName string `json:"first_name" binding:"max=50" example:"John"`
	LastName  string `json:"last_name" binding:"max=50" example:"Doe"`
	Bio       string `json:"bio" binding:"max=500" example:"Movie enthusiast"`
}

// RegisterResponse represents the response after successful registration
type RegisterResponse struct {
	Message string              `json:"message" example:"User registered successfully"`
	User    models.UserProfile  `json:"user"`
	Token   string              `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Register creates a new user account
// @Summary      Register new user
// @Description  Create a new user account with username, email, and password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest   true  "User registration data"
// @Success      201   {object}  RegisterResponse  "User registered successfully"
// @Failure      400   {object}  gin.H             "Invalid request or validation error"
// @Failure      409   {object}  gin.H             "Username or email already exists"
// @Failure      500   {object}  gin.H             "Internal server error"
// @Router       /register [post]
func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize username and email
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Check if username already exists
	var existingUser models.User
	if err := storage.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// Check if email already exists
	if err := storage.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      models.RoleUser, // Default role
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Bio:       req.Bio,
		IsActive:  true,
	}

	if err := storage.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate JWT token with user claims
	token, err := auth.GenerateTokenWithClaims(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, RegisterResponse{
		Message: "User registered successfully",
		User:    user.ToProfile(),
		Token:   token,
	})
}

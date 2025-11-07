package users

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"max=50" example:"John"`
	LastName  string `json:"last_name" binding:"max=50" example:"Doe"`
	Bio       string `json:"bio" binding:"max=500" example:"Updated bio"`
	Avatar    string `json:"avatar" binding:"max=500" example:"https://example.com/new-avatar.jpg"`
}

// GetProfile retrieves the authenticated user's profile
// @Summary      Get user profile
// @Description  Retrieve the profile of the currently authenticated user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.UserProfile  "User profile"
// @Failure      401  {object}  gin.H               "Unauthorized"
// @Failure      404  {object}  gin.H               "User not found"
// @Router       /profile [get]
func GetProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := storage.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToProfile())
}

// GetProfileByID retrieves a user's profile by user ID
// @Summary      Get user profile by ID
// @Description  Retrieve any user's public profile by their ID
// @Tags         users
// @Produce      json
// @Param        id   path      int                 true  "User ID"
// @Success      200  {object}  models.UserProfile  "User profile"
// @Failure      404  {object}  gin.H               "User not found"
// @Router       /users/{id} [get]
func GetProfileByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := storage.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToProfile())
}

// UpdateProfile updates the authenticated user's profile
// @Summary      Update user profile
// @Description  Update the profile of the currently authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        profile  body      UpdateProfileRequest  true  "Profile update data"
// @Success      200      {object}  models.UserProfile    "Updated profile"
// @Failure      400      {object}  gin.H                 "Invalid request"
// @Failure      401      {object}  gin.H                 "Unauthorized"
// @Failure      404      {object}  gin.H                 "User not found"
// @Failure      500      {object}  gin.H                 "Internal server error"
// @Router       /profile [put]
func UpdateProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := storage.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Bio = req.Bio
	user.Avatar = req.Avatar

	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user.ToProfile())
}

// DeleteProfile deletes the authenticated user's account (soft delete)
// @Summary      Delete user account
// @Description  Soft delete the currently authenticated user's account
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  gin.H  "Account deleted successfully"
// @Failure      401  {object}  gin.H  "Unauthorized"
// @Failure      404  {object}  gin.H  "User not found"
// @Failure      500  {object}  gin.H  "Internal server error"
// @Router       /profile [delete]
func DeleteProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := storage.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Soft delete
	if err := storage.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

package users

import (
	"fmt"
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// ListUsersResponse represents paginated user list response
type ListUsersResponse struct {
	Users []models.UserProfile `json:"users"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	Role models.UserRole `json:"role" binding:"required" example:"moderator"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"success message"`
}

// ListUsers retrieves all users (admin only)
// @Summary      List all users
// @Description  Retrieve a paginated list of all users (admin only)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int  false  "Page number"      default(1)
// @Param        limit  query     int  false  "Items per page"   default(10)
// @Success      200    {object}  ListUsersResponse  "List of users"
// @Failure      401    {object}  ErrorResponse      "Unauthorized"
// @Failure      403    {object}  ErrorResponse      "Forbidden - admin only"
// @Failure      500    {object}  ErrorResponse      "Internal server error"
// @Router       /admin/users [get]
func ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10
	if p, ok := c.GetQuery("page"); ok {
		if val, err := parseIntDefault(p, 1); err == nil {
			page = val
		}
	}
	if l, ok := c.GetQuery("limit"); ok {
		if val, err := parseIntDefault(l, 10); err == nil {
			limit = val
		}
	}

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	// Get total count
	if err := storage.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
		return
	}

	// Get users with pagination
	if err := storage.DB.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	// Convert to profiles
	profiles := make([]models.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = user.ToProfile()
	}

	c.JSON(http.StatusOK, ListUsersResponse{
		Users: profiles,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// UpdateUserRole updates a user's role (admin only)
// @Summary      Update user role
// @Description  Update a user's role (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                   true  "User ID"
// @Param        role  body      UpdateRoleRequest     true  "Role update data"
// @Success      200   {object}  models.UserProfile    "Updated user profile"
// @Failure      400   {object}  ErrorResponse         "Invalid request"
// @Failure      401   {object}  ErrorResponse         "Unauthorized"
// @Failure      403   {object}  ErrorResponse         "Forbidden - admin only"
// @Failure      404   {object}  ErrorResponse         "User not found"
// @Failure      500   {object}  ErrorResponse         "Internal server error"
// @Router       /admin/users/{id}/role [put]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if req.Role != models.RoleAdmin && req.Role != models.RoleModerator && req.Role != models.RoleUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	var user models.User
	if err := storage.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update role
	user.Role = req.Role
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	c.JSON(http.StatusOK, user.ToProfile())
}

// DeactivateUser deactivates a user account (admin only)
// @Summary      Deactivate user
// @Description  Deactivate a user account (admin only)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int               true  "User ID"
// @Success      200 {object}  MessageResponse   "User deactivated successfully"
// @Failure      401 {object}  ErrorResponse     "Unauthorized"
// @Failure      403 {object}  ErrorResponse     "Forbidden - admin only"
// @Failure      404 {object}  ErrorResponse     "User not found"
// @Failure      500 {object}  ErrorResponse     "Internal server error"
// @Router       /admin/users/{id}/deactivate [post]
func DeactivateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := storage.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.IsActive = false
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}

// ActivateUser activates a user account (admin only)
// @Summary      Activate user
// @Description  Activate a user account (admin only)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int               true  "User ID"
// @Success      200 {object}  MessageResponse   "User activated successfully"
// @Failure      401 {object}  ErrorResponse     "Unauthorized"
// @Failure      403 {object}  ErrorResponse     "Forbidden - admin only"
// @Failure      404 {object}  ErrorResponse     "User not found"
// @Failure      500 {object}  ErrorResponse     "Internal server error"
// @Router       /admin/users/{id}/activate [post]
func ActivateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := storage.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.IsActive = true
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user activated successfully"})
}

// Helper function to parse integer with default value
func parseIntDefault(s string, defaultVal int) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return defaultVal, err
	}
	return val, nil
}

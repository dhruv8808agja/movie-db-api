package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"
	RoleUser      UserRole = "user"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username" example:"johndoe"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email" example:"john@example.com"`
	Password  string         `gorm:"not null" json:"-"` // Never expose password in JSON
	Role      UserRole       `gorm:"type:varchar(20);default:'user'" json:"role" example:"user"`
	FirstName string         `json:"first_name" example:"John"`
	LastName  string         `json:"last_name" example:"Doe"`
	Bio       string         `json:"bio" example:"Movie enthusiast and film critic"`
	Avatar    string         `json:"avatar" example:"https://example.com/avatar.jpg"`
	IsActive  bool           `gorm:"default:true" json:"is_active" example:"true"`
	CreatedAt time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// UserProfile represents public user profile information
type UserProfile struct {
	ID        uint      `json:"id" example:"1"`
	Username  string    `json:"username" example:"johndoe"`
	Email     string    `json:"email" example:"john@example.com"`
	Role      UserRole  `json:"role" example:"user"`
	FirstName string    `json:"first_name" example:"John"`
	LastName  string    `json:"last_name" example:"Doe"`
	Bio       string    `json:"bio" example:"Movie enthusiast and film critic"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ToProfile converts User to UserProfile (public representation)
func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Bio:       u.Bio,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
	}
}

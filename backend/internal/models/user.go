package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	Name         string     `json:"name"`
	AvatarURL    string     `json:"avatar_url"`
	Role         Role       `gorm:"default:user" json:"role"`
	GitHubToken  string     `gorm:"column:github_token" json:"-"`
	GitHubLogin  string     `gorm:"column:github_login" json:"github_login"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Repositories []Repository  `gorm:"foreignKey:UserID" json:"-"`
	Notifications []Notification `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        Role      `json:"role"`
	GitHubLogin string    `json:"github_login"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		AvatarURL:   u.AvatarURL,
		Role:        u.Role,
		GitHubLogin: u.GitHubLogin,
		CreatedAt:   u.CreatedAt,
	}
}

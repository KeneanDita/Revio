package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	GitHubID    int64      `gorm:"column:github_id;not null" json:"github_id"`
	Name        string     `gorm:"not null" json:"name"`
	FullName    string     `gorm:"column:full_name;not null;uniqueIndex:idx_user_repo" json:"full_name"`
	Owner       string     `gorm:"not null" json:"owner"`
	Description string     `json:"description"`
	Private     bool       `gorm:"default:false" json:"private"`
	HTMLURL     string     `gorm:"column:html_url" json:"html_url"`
	LastSyncAt  *time.Time `gorm:"column:last_sync_at" json:"last_sync_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User         User          `gorm:"foreignKey:UserID" json:"-"` // intentionally hidden — don't expose tokens
	PullRequests []PullRequest `gorm:"foreignKey:RepoID" json:"-"`
}

func (r *Repository) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

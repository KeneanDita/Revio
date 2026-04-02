package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewState string

const (
	ReviewStateApproved         ReviewState = "approved"
	ReviewStateChangesRequested ReviewState = "changes_requested"
	ReviewStateCommented        ReviewState = "commented"
	ReviewStateDismissed        ReviewState = "dismissed"
)

type Review struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	PRID            uuid.UUID   `gorm:"type:uuid;not null;index" json:"pr_id"`
	GitHubID        int64       `gorm:"column:github_id;not null" json:"github_id"`
	Reviewer        string      `gorm:"not null" json:"reviewer"`
	ReviewerAvatar  string      `gorm:"column:reviewer_avatar" json:"reviewer_avatar"`
	State           ReviewState `gorm:"not null;index" json:"state"`
	Body            string      `gorm:"type:text" json:"body"`
	HTMLURL         string      `gorm:"column:html_url" json:"html_url"`
	SubmittedAt     time.Time   `gorm:"column:submitted_at;not null;index" json:"submitted_at"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	PullRequest PullRequest `gorm:"foreignKey:PRID" json:"-"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string    `gorm:"not null" json:"type"`
	Title     string    `gorm:"not null" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	Link      string    `json:"link"`
	Read      bool      `gorm:"default:false;index" json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type SyncJob struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RepoID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"repo_id"`
	Status     string     `gorm:"not null;default:pending" json:"status"`
	StartedAt  *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finished_at"`
	ErrorMsg   string     `gorm:"column:error_msg;type:text" json:"error_msg,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (s *SyncJob) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

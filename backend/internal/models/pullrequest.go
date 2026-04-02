package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "open"
	PRStatusClosed PRStatus = "closed"
	PRStatusMerged PRStatus = "merged"
)

type PullRequest struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RepoID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"repo_id"`
	GitHubID         int64      `gorm:"column:github_id;not null" json:"github_id"`
	Number           int        `gorm:"not null" json:"number"`
	Title            string     `gorm:"not null" json:"title"`
	Body             string     `gorm:"type:text" json:"body"`
	Author           string     `gorm:"not null" json:"author"`
	AuthorAvatarURL  string     `gorm:"column:author_avatar_url" json:"author_avatar_url"`
	Status           PRStatus   `gorm:"not null;default:open;index" json:"status"`
	BaseBranch       string     `gorm:"column:base_branch" json:"base_branch"`
	HeadBranch       string     `gorm:"column:head_branch" json:"head_branch"`
	Additions        int        `json:"additions"`
	Deletions        int        `json:"deletions"`
	ChangedFiles     int        `gorm:"column:changed_files" json:"changed_files"`
	CommitCount      int        `gorm:"column:commit_count" json:"commit_count"`
	CommentCount     int        `gorm:"column:comment_count" json:"comment_count"`
	ReviewCount      int        `gorm:"column:review_count" json:"review_count"`
	HTMLURL          string     `gorm:"column:html_url" json:"html_url"`
	MergedAt         *time.Time `gorm:"column:merged_at;index" json:"merged_at"`
	ClosedAt         *time.Time `gorm:"column:closed_at" json:"closed_at"`
	FirstReviewAt    *time.Time `gorm:"column:first_review_at" json:"first_review_at"`
	GithubCreatedAt  time.Time  `gorm:"column:github_created_at;not null;index" json:"github_created_at"`
	GithubUpdatedAt  time.Time  `gorm:"column:github_updated_at;not null" json:"github_updated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	Repo    Repository `gorm:"foreignKey:RepoID" json:"repo,omitempty"`
	Reviews []Review   `gorm:"foreignKey:PRID" json:"reviews,omitempty"`
}

func (p *PullRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (p *PullRequest) TimeToMerge() *float64 {
	if p.MergedAt == nil {
		return nil
	}
	hours := p.MergedAt.Sub(p.GithubCreatedAt).Hours()
	return &hours
}

func (p *PullRequest) TimeToFirstReview() *float64 {
	if p.FirstReviewAt == nil {
		return nil
	}
	hours := p.FirstReviewAt.Sub(p.GithubCreatedAt).Hours()
	return &hours
}

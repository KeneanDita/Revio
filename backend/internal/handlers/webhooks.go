package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/revio/backend/internal/models"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	db     *gorm.DB
	secret string
}

func NewWebhookHandler(db *gorm.DB, secret string) *WebhookHandler {
	return &WebhookHandler{db: db, secret: secret}
}

type githubWebhookPayload struct {
	Action      string              `json:"action"`
	Number      int                 `json:"number"`
	PullRequest githubPRPayload     `json:"pull_request"`
	Repository  githubRepoPayload   `json:"repository"`
	Review      *githubReviewPayload `json:"review,omitempty"`
	Sender      githubSenderPayload `json:"sender"`
}

type githubPRPayload struct {
	ID        int64      `json:"id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	MergedAt  *time.Time `json:"merged_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	Base struct{ Ref string `json:"ref"` } `json:"base"`
	Head struct{ Ref string `json:"ref"` } `json:"head"`
}

type githubRepoPayload struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
}

type githubReviewPayload struct {
	ID          int64     `json:"id"`
	State       string    `json:"state"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	SubmittedAt time.Time `json:"submitted_at"`
	User        struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

type githubSenderPayload struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (h *WebhookHandler) GitHub(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	if h.secret != "" && !h.verifySignature(body, c.GetHeader("X-Hub-Signature-256")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")

	var payload githubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	switch event {
	case "pull_request":
		h.handlePREvent(payload)
	case "pull_request_review":
		h.handleReviewEvent(payload)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func (h *WebhookHandler) handlePREvent(payload githubWebhookPayload) {
	var repo models.Repository
	if err := h.db.Where("github_id = ?", payload.Repository.ID).First(&repo).Error; err != nil {
		return // Repository not connected
	}

	status := models.PRStatus(payload.PullRequest.State)
	if payload.PullRequest.MergedAt != nil {
		status = models.PRStatusMerged
	}

	prData := map[string]interface{}{
		"title":             payload.PullRequest.Title,
		"body":              payload.PullRequest.Body,
		"status":            status,
		"merged_at":         payload.PullRequest.MergedAt,
		"closed_at":         payload.PullRequest.ClosedAt,
		"github_updated_at": payload.PullRequest.UpdatedAt,
	}

	result := h.db.Model(&models.PullRequest{}).
		Where("repo_id = ? AND github_id = ?", repo.ID, payload.PullRequest.ID).
		Updates(prData)

	if result.RowsAffected == 0 {
		// Create new PR
		pr := models.PullRequest{
			RepoID:          repo.ID,
			GitHubID:        payload.PullRequest.ID,
			Number:          payload.PullRequest.Number,
			Title:           payload.PullRequest.Title,
			Body:            payload.PullRequest.Body,
			Author:          payload.PullRequest.User.Login,
			AuthorAvatarURL: payload.PullRequest.User.AvatarURL,
			Status:          status,
			BaseBranch:      payload.PullRequest.Base.Ref,
			HeadBranch:      payload.PullRequest.Head.Ref,
			HTMLURL:         payload.PullRequest.HTMLURL,
			MergedAt:        payload.PullRequest.MergedAt,
			ClosedAt:        payload.PullRequest.ClosedAt,
			GithubCreatedAt: payload.PullRequest.CreatedAt,
			GithubUpdatedAt: payload.PullRequest.UpdatedAt,
		}
		if err := h.db.Create(&pr).Error; err != nil {
			log.Printf("webhook: failed to create PR: %v", err)
		}
	}

	// Create in-app notification
	h.createNotification(repo.UserID.String(), payload.Action, payload.PullRequest.Title, payload.PullRequest.HTMLURL)
}

func (h *WebhookHandler) handleReviewEvent(payload githubWebhookPayload) {
	if payload.Review == nil {
		return
	}

	var pr models.PullRequest
	if err := h.db.Where("github_id = ?", payload.PullRequest.ID).First(&pr).Error; err != nil {
		return
	}

	review := models.Review{
		PRID:           pr.ID,
		GitHubID:       payload.Review.ID,
		Reviewer:       payload.Review.User.Login,
		ReviewerAvatar: payload.Review.User.AvatarURL,
		State:          models.ReviewState(strings.ToLower(payload.Review.State)),
		Body:           payload.Review.Body,
		HTMLURL:        payload.Review.HTMLURL,
		SubmittedAt:    payload.Review.SubmittedAt,
	}

	h.db.Where(models.Review{GitHubID: review.GitHubID}).FirstOrCreate(&review)

	// Update first review time if not set
	h.db.Model(&pr).Where("first_review_at IS NULL").
		UpdateColumn("first_review_at", payload.Review.SubmittedAt)
}

func (h *WebhookHandler) createNotification(userID, action, prTitle, link string) {
	title := "Pull request " + action
	notification := models.Notification{
		Type:  "pr_" + action,
		Title: title,
		Body:  prTitle,
		Link:  link,
	}
	_ = h.db.Raw("SELECT id FROM users WHERE id = ?", userID).Scan(&notification.UserID)
	h.db.Create(&notification)
}

func (h *WebhookHandler) verifySignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

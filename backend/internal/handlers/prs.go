package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/models"
	"github.com/revio/backend/internal/services"
	"gorm.io/gorm"
)

type PRHandler struct {
	db            *gorm.DB
	githubService *services.GitHubService
}

func NewPRHandler(db *gorm.DB, githubService *services.GitHubService) *PRHandler {
	return &PRHandler{db: db, githubService: githubService}
}

func (h *PRHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	query := h.db.
		Joins("JOIN repositories ON repositories.id = pull_requests.repo_id").
		Where("repositories.user_id = ?", userID)

	if status := c.Query("status"); status != "" {
		query = query.Where("pull_requests.status = ?", status)
	}
	if repoID := c.Query("repo_id"); repoID != "" {
		query = query.Where("pull_requests.repo_id = ?", repoID)
	}
	if author := c.Query("author"); author != "" {
		query = query.Where("pull_requests.author ILIKE ?", "%"+author+"%")
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("pull_requests.github_created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("pull_requests.github_created_at <= ?", to)
	}

	var total int64
	query.Model(&models.PullRequest{}).Count(&total)

	var prs []models.PullRequest
	if err := query.
		Preload("Repo").
		Order("pull_requests.github_created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&prs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pull requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pull_requests": prs,
		"total":         total,
		"page":          page,
		"per_page":      perPage,
		"total_pages":   (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *PRHandler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pull request ID"})
		return
	}

	var pr models.PullRequest
	if err := h.db.
		Joins("JOIN repositories ON repositories.id = pull_requests.repo_id").
		Where("pull_requests.id = ? AND repositories.user_id = ?", prID, userID).
		Preload("Repo").
		Preload("Reviews").
		First(&pr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pull request not found"})
		return
	}

	timeToMerge := pr.TimeToMerge()
	timeToFirstReview := pr.TimeToFirstReview()

	c.JSON(http.StatusOK, gin.H{
		"pull_request":         pr,
		"time_to_merge_hours":  timeToMerge,
		"time_to_review_hours": timeToFirstReview,
	})
}

type commentRequest struct {
	Body string `json:"body" binding:"required"`
}

func (h *PRHandler) Comment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pull request ID"})
		return
	}

	var req commentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pr models.PullRequest
	if err := h.db.
		Joins("JOIN repositories ON repositories.id = pull_requests.repo_id").
		Where("pull_requests.id = ? AND repositories.user_id = ?", prID, userID).
		Preload("Repo").
		First(&pr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pull request not found"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub account not connected"})
		return
	}

	comment, err := h.githubService.PostComment(
		c.Request.Context(),
		user.GitHubToken,
		pr.Repo.Owner,
		pr.Repo.Name,
		pr.Number,
		req.Body,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to post comment"})
		return
	}

	// Increment comment count
	h.db.Model(&pr).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

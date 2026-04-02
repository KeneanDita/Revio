package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/models"
	"github.com/revio/backend/internal/services"
	"gorm.io/gorm"
)

type RepoHandler struct {
	db            *gorm.DB
	githubService *services.GitHubService
}

func NewRepoHandler(db *gorm.DB, githubService *services.GitHubService) *RepoHandler {
	return &RepoHandler{db: db, githubService: githubService}
}

func (h *RepoHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var repos []models.Repository
	if err := h.db.Where("user_id = ?", userID).Order("full_name ASC").Find(&repos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch repositories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"total":        len(repos),
	})
}

type connectRepoRequest struct {
	FullName string `json:"full_name" binding:"required"`
}

func (h *RepoHandler) Connect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req connectRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub account not connected. Please login via GitHub OAuth."})
		return
	}

	// Check if already connected
	var existing models.Repository
	if err := h.db.Where("user_id = ? AND full_name = ?", userID, req.FullName).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "repository already connected"})
		return
	}

	ghRepo, err := h.githubService.GetRepository(c.Request.Context(), user.GitHubToken, req.FullName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository not found or access denied"})
		return
	}

	repo := models.Repository{
		UserID:      userID,
		GitHubID:    ghRepo.ID,
		Name:        ghRepo.Name,
		FullName:    ghRepo.FullName,
		Owner:       ghRepo.Owner.Login,
		Description: ghRepo.Description,
		Private:     ghRepo.Private,
		HTMLURL:     ghRepo.HTMLURL,
	}

	if err := h.db.Create(&repo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect repository"})
		return
	}

	// Trigger background sync
	go h.githubService.SyncRepository(repo.ID, user.GitHubToken)

	c.JSON(http.StatusCreated, repo)
}

func (h *RepoHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	repoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository ID"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", repoID, userID).Delete(&models.Repository{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disconnect repository"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "repository disconnected"})
}

func (h *RepoHandler) Sync(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	repoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository ID"})
		return
	}

	var repo models.Repository
	if err := h.db.Where("id = ? AND user_id = ?", repoID, userID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	go h.githubService.SyncRepository(repo.ID, user.GitHubToken)

	c.JSON(http.StatusAccepted, gin.H{"message": "sync started"})
}

func (h *RepoHandler) ListGitHub(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub account not connected"})
		return
	}

	repos, err := h.githubService.ListUserRepositories(c.Request.Context(), user.GitHubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch GitHub repositories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"total":        len(repos),
	})
}

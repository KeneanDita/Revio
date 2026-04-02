package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/revio/backend/internal/auth"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/models"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db          *gorm.DB
	jwtManager  *auth.Manager
	githubOAuth *auth.GitHubOAuth
	frontendURL string
}

func NewAuthHandler(db *gorm.DB, jwtManager *auth.Manager, githubOAuth *auth.GitHubOAuth, frontendURL string) *AuthHandler {
	return &AuthHandler{
		db:          db,
		jwtManager:  jwtManager,
		githubOAuth: githubOAuth,
		frontendURL: frontendURL,
	}
}

type signupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         models.RoleUser,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.setTokenCookie(c, tokens.AccessToken)

	c.JSON(http.StatusCreated, gin.H{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.setTokenCookie(c, tokens.AccessToken)

	c.JSON(http.StatusOK, gin.H{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	state := generateState()
	c.SetCookie("oauth_state", state, 600, "/", "", true, true)
	c.Redirect(http.StatusTemporaryRedirect, h.githubOAuth.AuthCodeURL(state))
}

func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth code"})
		return
	}

	token, err := h.githubOAuth.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange oauth code"})
		return
	}

	ghUser, err := h.githubOAuth.GetUser(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get GitHub user"})
		return
	}

	var user models.User
	result := h.db.Where("email = ?", ghUser.Email).First(&user)

	if result.Error != nil {
		// Create new user
		user = models.User{
			Email:       ghUser.Email,
			Name:        ghUser.Name,
			AvatarURL:   ghUser.AvatarURL,
			GitHubLogin: ghUser.Login,
			GitHubToken: token.AccessToken,
			Role:        models.RoleUser,
		}
		if user.Name == "" {
			user.Name = ghUser.Login
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else {
		// Update existing user
		h.db.Model(&user).Updates(map[string]interface{}{
			"github_token": token.AccessToken,
			"github_login": ghUser.Login,
			"avatar_url":   ghUser.AvatarURL,
		})
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.setTokenCookie(c, tokens.AccessToken)

	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/dashboard?token="+tokens.AccessToken)
}

func (h *AuthHandler) setTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		token,
		int(24*time.Hour/time.Second),
		"/",
		"",
		false, // secure: set true in production
		true,  // httpOnly
	)
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

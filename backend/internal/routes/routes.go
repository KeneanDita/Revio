package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/revio/backend/internal/auth"
	"github.com/revio/backend/internal/handlers"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/services"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB            *gorm.DB
	JWTManager    *auth.Manager
	GitHubOAuth   *auth.GitHubOAuth
	GitHubService *services.GitHubService
	FrontendURL   string
	WebhookSecret string
}

func Register(r *gin.Engine, deps Dependencies) {
	authHandler := handlers.NewAuthHandler(deps.DB, deps.JWTManager, deps.GitHubOAuth, deps.FrontendURL)
	repoHandler := handlers.NewRepoHandler(deps.DB, deps.GitHubService)
	prHandler := handlers.NewPRHandler(deps.DB, deps.GitHubService)
	analyticsHandler := handlers.NewAnalyticsHandler(deps.DB)
	webhookHandler := handlers.NewWebhookHandler(deps.DB, deps.WebhookSecret)
	notifHandler := handlers.NewNotificationHandler(deps.DB)

	api := r.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (rate limited)
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/signup", middleware.RateLimitAuth(), authHandler.Signup)
		authRoutes.POST("/login", middleware.RateLimitAuth(), authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.GET("/github", authHandler.GitHubLogin)
		authRoutes.GET("/github/callback", authHandler.GitHubCallback)
		authRoutes.GET("/me", middleware.AuthRequired(deps.JWTManager), authHandler.Me)
	}

	// Webhook routes (no auth, verified by signature)
	webhookRoutes := api.Group("/webhooks")
	{
		webhookRoutes.POST("/github", webhookHandler.GitHub)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(deps.JWTManager))
	protected.Use(middleware.RateLimitNormal())
	{
		// Repositories
		repoRoutes := protected.Group("/repos")
		{
			repoRoutes.GET("", repoHandler.List)
			repoRoutes.GET("/github", repoHandler.ListGitHub)
			repoRoutes.POST("/connect", repoHandler.Connect)
			repoRoutes.DELETE("/:id", repoHandler.Delete)
			repoRoutes.POST("/:id/sync", repoHandler.Sync)
		}

		// Pull Requests
		prRoutes := protected.Group("/prs")
		{
			prRoutes.GET("", prHandler.List)
			prRoutes.GET("/:id", prHandler.Get)
			prRoutes.POST("/:id/comment", prHandler.Comment)
		}

		// Analytics
		protected.GET("/analytics", analyticsHandler.Get)

		// Notifications
		notifRoutes := protected.Group("/notifications")
		{
			notifRoutes.GET("", notifHandler.List)
			notifRoutes.PATCH("/:id/read", notifHandler.MarkRead)
			notifRoutes.PATCH("/read-all", notifHandler.MarkAllRead)
		}
	}
}

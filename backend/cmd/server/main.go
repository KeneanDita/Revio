package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revio/backend/internal/auth"
	"github.com/revio/backend/internal/config"
	"github.com/revio/backend/internal/database"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/routes"
	"github.com/revio/backend/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.VerifySchema(db); err != nil {
		log.Fatalf("schema verification failed: %v", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.Recover())
	r.Use(middleware.CORS(cfg.FrontendURL))

	jwtManager := auth.NewManager(cfg.JWTSecret)
	githubOAuth := auth.NewGitHubOAuth(cfg)
	githubService := services.NewGitHubService(db)

	routes.Register(r, routes.Dependencies{
		DB:            db,
		JWTManager:    jwtManager,
		GitHubOAuth:   githubOAuth,
		GitHubService: githubService,
		FrontendURL:   cfg.FrontendURL,
		WebhookSecret: cfg.GitHubClientSecret,
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	addr := ":" + cfg.Port
	log.Printf("Revio API server starting on %s (env: %s)", addr, cfg.Env)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

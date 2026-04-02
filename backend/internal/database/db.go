package database

import (
	"log"
	"time"

	"github.com/revio/backend/internal/config"
	"github.com/revio/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if !cfg.IsProduction() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	DB = db
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database auto-migration...")
	return db.AutoMigrate(
		&models.User{},
		&models.Repository{},
		&models.PullRequest{},
		&models.Review{},
		&models.Notification{},
		&models.SyncJob{},
	)
}

package database

import (
	"fmt"
	"time"

	"github.com/revio/backend/internal/config"
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

// VerifySchema checks that the required tables exist.
// Schema is managed via backend/migrations/001_initial.sql — not GORM AutoMigrate.
func VerifySchema(db *gorm.DB) error {
	required := []string{"users", "repositories", "pull_requests", "reviews", "notifications", "sync_jobs"}
	for _, table := range required {
		var count int64
		row := db.Raw(
			"SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?",
			table,
		).Row()
		if err := row.Scan(&count); err != nil {
			return fmt.Errorf("schema check failed for table %q: %w", table, err)
		}
		if count == 0 {
			return fmt.Errorf("required table %q not found — run: psql -h localhost -U postgres -d revio -f backend/migrations/001_initial.sql", table)
		}
	}
	return nil
}

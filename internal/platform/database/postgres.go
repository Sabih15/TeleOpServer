package database

import (
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgres opens a connection and runs AutoMigrate for all registered models.
// Add new models to the AutoMigrate call as the domain grows.
func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Enable TimescaleDB extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE").Error; err != nil {
		return nil, err
	}

	return db, nil
}

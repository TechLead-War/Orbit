package database

import (
	"github.com/ayush/ORBIT/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Initialize sets up the database and registers all models
func Initialize(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Initializing database models...")

	// Register all models here
	err := db.AutoMigrate(
		&models.Student{},
		&models.Rating{},
		&models.ContestHistory{},
	)
	if err != nil {
		logger.Error("Failed to auto-migrate models", zap.Error(err))
		return err
	}

	logger.Info("Database models initialized successfully")
	return nil
}

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ayush/ORBIT/internal/config"
	"github.com/ayush/ORBIT/internal/database"
	"github.com/ayush/ORBIT/internal/handlers"
	"github.com/ayush/ORBIT/internal/jobs"
	"github.com/ayush/ORBIT/internal/repository"
	"github.com/ayush/ORBIT/internal/server"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	studentRepo := repository.NewStudentRepository(db)
	dbService := database.NewDatabase(db)
	handler := handlers.NewHandler(dbService)

	// Initialize rating updater with weekly interval
	ratingUpdater := jobs.NewRatingUpdater(studentRepo, logger, 7*24*time.Hour) // 7 days
	ratingUpdater.Start()
	defer ratingUpdater.Stop()

	// Initialize contest history updater with weekly interval
	contestHistoryUpdater := jobs.NewContestHistoryUpdater(studentRepo, logger, 7*24*time.Hour) // 7 days
	contestHistoryUpdater.Start()
	defer contestHistoryUpdater.Stop()

	router := server.SetupRouter(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

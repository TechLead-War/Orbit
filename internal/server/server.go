package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/ORBIT/handlers"
	"github.com/ayush/ORBIT/internal/config"
	"github.com/ayush/ORBIT/internal/database"
	"github.com/ayush/ORBIT/internal/worker"
	"github.com/ayush/ORBIT/pkg/leetcode"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	config      *config.Config
	logger      *zap.Logger
	db          *gorm.DB
	httpServer  *http.Server
	router      *gin.Engine
	statsWorker *worker.WeeklyStatsWorker
}

// New creates a new server instance
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize database handlers
	studentDB := database.NewStudentDB(db)
	weeklyStatsDB := database.NewWeeklyStatsDB(db)
	leetcodeClient := leetcode.NewClient()

	// Initialize handlers
	studentHandler := handlers.NewHandler(studentDB)
	weeklyStatsHandler := handlers.NewWeeklyStatsHandler(studentDB, leetcodeClient)

	// Initialize weekly stats worker
	statsWorker := worker.NewWeeklyStatsWorker(studentDB, weeklyStatsDB, leetcodeClient)

	// Setup router
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		students := v1.Group("/students")
		{
			// Basic student operations
			students.GET("", studentHandler.GetAllStudents)
			students.POST("", studentHandler.CreateStudent)
			students.POST("/bulk", studentHandler.BulkCreateStudents)
			students.GET("/:id", studentHandler.GetStudentDetails)

			// Weekly stats routes
			students.GET("/:id/weekly-stats", weeklyStatsHandler.GetStudentWeeklyStats)
			students.GET("/:id/current-week", weeklyStatsHandler.GetCurrentWeekStats)
			students.PUT("/:id/weekly-stats", weeklyStatsHandler.UpdateWeeklyStats)
			students.POST("/:id/weekly-stats/leetcode", weeklyStatsHandler.UpdateWeeklyStatsFromLeetCode)
			students.PUT("/weekly-stats/update-all", weeklyStatsHandler.UpdateAllWeeklyStats)

			// Contest history routes
			students.GET("/:id/contest-history", studentHandler.GetContestHistory)
			students.PUT("/:id/contest-history", studentHandler.UpdateContestHistory)
			students.PUT("/contest-history/update-all", studentHandler.UpdateAllContestHistories)
		}
	}

	server := &Server{
		config:      cfg,
		logger:      logger,
		db:          db,
		router:      r,
		statsWorker: statsWorker,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return server, nil
}

// Start starts the HTTP server and background workers
func (s *Server) Start() error {
	// Start weekly stats worker
	s.statsWorker.Start()

	s.logger.Info("Starting server...", zap.String("port", s.config.ServerPort))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}

// DB returns the database connection
func (s *Server) DB() *gorm.DB {
	return s.db
}

// Logger returns the logger instance
func (s *Server) Logger() *zap.Logger {
	return s.logger
}

// Router returns the gin router instance
func (s *Server) Router() *gin.Engine {
	return s.router
}

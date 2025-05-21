package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/ORBIT/internal/api"
	"github.com/ayush/ORBIT/internal/config"
	"github.com/ayush/ORBIT/internal/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	db         *gorm.DB
	httpServer *http.Server
	router     *gin.Engine
}

// New creates a new server instance
func New(cfg *config.Config, logger *zap.Logger, db *gorm.DB) (*Server, error) {
	// Initialize database models
	if err := database.Initialize(db, logger); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	router := gin.Default()

	// Initialize routes
	api.InitializeRoutes(router, cfg, logger, db)

	server := &Server{
		config: cfg,
		logger: logger,
		db:     db,
		router: router,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
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

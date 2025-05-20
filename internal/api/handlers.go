package api

import (
	"net/http"
	"strconv"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"github.com/ayush/ORBIT/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler contains all the HTTP handlers
type Handler struct {
	studentService *service.StudentService
	logger         *zap.Logger
}

// NewHandler creates a new handler instance
func NewHandler(studentService *service.StudentService, logger *zap.Logger) *Handler {
	return &Handler{
		studentService: studentService,
		logger:         logger,
	}
}

// InitializeRoutes sets up all the routes for the application
func InitializeRoutes(router *gin.Engine, cfg interface{}, logger *zap.Logger, db *gorm.DB) {
	// Middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	studentRepo := repository.NewStudentRepository(db)
	studentService := service.NewStudentService(studentRepo, logger)
	handler := NewHandler(studentService, logger)

	// Routes
	v1 := router.Group("/api/v1")
	{
		// Student routes
		students := v1.Group("/students")
		{
			students.POST("", handler.CreateStudent)
			students.GET("", handler.ListStudents)
			students.GET("/:id", handler.GetStudent)
			students.PUT("/:id/rating", handler.UpdateStudentRating)
			students.GET("/:id/stats", handler.GetStudentStats)
		}
	}
}

// CreateStudent handles student registration
func (h *Handler) CreateStudent(c *gin.Context) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.studentService.RegisterStudent(c.Request.Context(), &student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, student)
}

// ListStudents handles retrieving a list of students
func (h *Handler) ListStudents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	students, err := h.studentService.ListStudents(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// GetStudent handles retrieving a single student
func (h *Handler) GetStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	student, err := h.studentService.GetStudentStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, student)
}

// UpdateStudentRating handles updating a student's LeetCode rating
func (h *Handler) UpdateStudentRating(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	var rating models.Rating
	if err := c.ShouldBindJSON(&rating); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.studentService.UpdateStudentRating(c.Request.Context(), uint(id), &rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rating)
}

// GetStudentStats handles retrieving a student's statistics
func (h *Handler) GetStudentStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	stats, err := h.studentService.GetStudentStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

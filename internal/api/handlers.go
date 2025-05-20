package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"github.com/ayush/ORBIT/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	studentService *service.StudentService
	logger         *zap.Logger
}

func NewHandler(studentService *service.StudentService, logger *zap.Logger) *Handler {
	return &Handler{
		studentService: studentService,
		logger:         logger,
	}
}

func InitializeRoutes(router *gin.Engine, cfg interface{}, logger *zap.Logger, db *gorm.DB) {
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	studentRepo := repository.NewStudentRepository(db)
	studentService := service.NewStudentService(studentRepo, logger)
	handler := NewHandler(studentService, logger)

	v1 := router.Group("/api/v1")
	{
		students := v1.Group("/students")
		{
			students.POST("", handler.CreateStudent)
			students.GET("", handler.ListStudents)
			students.GET("/:id", handler.GetStudent)
			students.PUT("/:id/rating", handler.UpdateStudentRating)
			students.GET("/:id/stats", handler.GetStudentStats)
			students.GET("/:id/leetcode", handler.GetLeetCodeStats)
			students.GET("/:id/contest-rankings", handler.GetContestRankings)
		}
	}
}

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

func (h *Handler) GetStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	student, err := h.studentService.GetStudentStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, student)
}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rating)
}

func (h *Handler) GetStudentStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	stats, err := h.studentService.GetStudentStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student not found or no ratings available",
				"details": "The student either does not exist or has no LeetCode ratings recorded yet. " +
					"Try updating their rating first using PUT /api/v1/students/:id/rating",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetLeetCodeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	student, err := h.studentService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.studentService.GetLeetCodeStats(c.Request.Context(), student.LeetcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch LeetCode stats",
			"details": "Could not retrieve data from LeetCode API. This could be due to:" +
				"\n- Invalid LeetCode username" +
				"\n- LeetCode API rate limiting" +
				"\n- LeetCode API being unavailable",
			"technical_error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetContestRankings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	student, err := h.studentService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rankings, err := h.studentService.GetContestRankings(c.Request.Context(), student.LeetcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch LeetCode contest rankings",
			"details": "Could not retrieve contest data from LeetCode API. This could be due to:" +
				"\n- Invalid LeetCode username" +
				"\n- LeetCode API rate limiting" +
				"\n- LeetCode API being unavailable",
			"technical_error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, rankings)
}

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

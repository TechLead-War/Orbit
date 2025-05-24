package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/pkg/leetcode"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WeeklyStatsDB defines the database interface for weekly stats
type WeeklyStatsDB interface {
	GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error)
	GetWeeklyStats(studentID string, start, end time.Time) (*models.WeeklyStats, error)
	UpdateWeeklyStats(stats *models.WeeklyStats) error
	GetAllStudents() ([]*models.Student, error)
	GetStudentByID(studentID string) (*models.Student, error)
}

// WeeklyStatsHandler handles weekly stats operations
type WeeklyStatsHandler struct {
	db     WeeklyStatsDB
	lc     *leetcode.Client
	logger *zap.Logger
}

// NewWeeklyStatsHandler creates a new weekly stats handler
func NewWeeklyStatsHandler(db WeeklyStatsDB, lc *leetcode.Client, logger *zap.Logger) *WeeklyStatsHandler {
	return &WeeklyStatsHandler{
		db:     db,
		lc:     lc,
		logger: logger,
	}
}

// GetStudentWeeklyStats retrieves all weekly stats for a student
func (h *WeeklyStatsHandler) GetStudentWeeklyStats(c *gin.Context) {
	start := time.Now()
	studentID := c.Param("id")
	logger := h.logger.With(
		zap.String("handler", "GetStudentWeeklyStats"),
		zap.String("request_id", c.GetString("request_id")),
		zap.String("student_id", studentID),
	)

	logger.Info("Starting weekly stats retrieval")

	if studentID == "" {
		logger.Warn("Missing student ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	stats, err := h.db.GetStudentWeeklyStats(studentID)
	if err != nil {
		logger.Error("Failed to get weekly stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get weekly stats: %v", err)})
		return
	}

	duration := time.Since(start)
	logger.Info("Weekly stats retrieved successfully",
		zap.Duration("duration", duration),
		zap.Int("stats_count", len(stats)),
	)

	c.JSON(http.StatusOK, stats)
}

// GetCurrentWeekStats retrieves stats for the current week
func (h *WeeklyStatsHandler) GetCurrentWeekStats(c *gin.Context) {
	start := time.Now()
	studentID := c.Param("id")
	logger := h.logger.With(
		zap.String("handler", "GetCurrentWeekStats"),
		zap.String("request_id", c.GetString("request_id")),
		zap.String("student_id", studentID),
	)

	logger.Info("Starting current week stats retrieval")

	if studentID == "" {
		logger.Warn("Missing student ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	logger = logger.With(
		zap.Time("week_start", weekStart),
		zap.Time("week_end", weekEnd),
	)

	stats, err := h.db.GetWeeklyStats(studentID, weekStart, weekEnd)
	if err != nil {
		if err.Error() == "record not found" {
			logger.Info("No stats found for current week")
			c.JSON(http.StatusNotFound, gin.H{"error": "no stats found for current week"})
			return
		}
		logger.Error("Failed to get current week stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get current week stats: %v", err)})
		return
	}

	duration := time.Since(start)
	logger.Info("Current week stats retrieved successfully",
		zap.Duration("duration", duration),
	)

	c.JSON(http.StatusOK, stats)
}

// UpdateWeeklyStats updates weekly stats for a student
func (h *WeeklyStatsHandler) UpdateWeeklyStats(c *gin.Context) {
	start := time.Now()
	studentID := c.Param("id")
	logger := h.logger.With(
		zap.String("handler", "UpdateWeeklyStats"),
		zap.String("request_id", c.GetString("request_id")),
		zap.String("student_id", studentID),
	)

	logger.Info("Starting weekly stats update")

	if studentID == "" {
		logger.Warn("Missing student ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	var stats models.WeeklyStats
	if err := c.ShouldBindJSON(&stats); err != nil {
		logger.Error("Invalid request body",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedID, err := strconv.ParseUint(studentID, 10, 64)
	if err != nil {
		logger.Error("Invalid student ID format",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID format"})
		return
	}
	stats.StudentID = uint(parsedID)

	if err := h.db.UpdateWeeklyStats(&stats); err != nil {
		logger.Error("Failed to update weekly stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update weekly stats: %v", err)})
		return
	}

	duration := time.Since(start)
	logger.Info("Weekly stats updated successfully",
		zap.Duration("duration", duration),
		zap.Time("week_start", stats.WeekStart),
		zap.Time("week_end", stats.WeekEnd),
	)

	c.JSON(http.StatusOK, stats)
}

// UpdateWeeklyStatsFromLeetCode updates weekly stats from LeetCode data
func (h *WeeklyStatsHandler) UpdateWeeklyStatsFromLeetCode(c *gin.Context) {
	start := time.Now()
	studentID := c.Param("id")
	logger := h.logger.With(
		zap.String("handler", "UpdateWeeklyStatsFromLeetCode"),
		zap.String("request_id", c.GetString("request_id")),
		zap.String("student_id", studentID),
	)

	logger.Info("Starting LeetCode stats update")

	student, err := h.db.GetStudentByID(studentID)
	if err != nil {
		logger.Error("Failed to find student",
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	logger = logger.With(
		zap.String("leetcode_id", student.LeetcodeID),
	)

	// Get LeetCode stats
	logger.Info("Fetching LeetCode stats")
	stats, err := h.lc.GetUserStats(student.LeetcodeID)
	if err != nil {
		logger.Error("Failed to fetch LeetCode stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch LeetCode stats: %v", err)})
		return
	}

	// Create weekly stats
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	parsedID, err := strconv.ParseUint(studentID, 10, 64)
	if err != nil {
		logger.Error("Invalid student ID format",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID format"})
		return
	}

	weeklyStats := &models.WeeklyStats{
		StudentID:      uint(parsedID),
		WeekStart:      weekStart,
		WeekEnd:        weekEnd,
		ProblemsSolved: stats.TotalSolved,
		EasySolved:     stats.EasyCount,
		MediumSolved:   stats.MediumCount,
		HardSolved:     stats.HardCount,
		ContestRating:  float64(stats.ContestRating),
		GlobalRanking:  stats.ContestRanking,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.db.UpdateWeeklyStats(weeklyStats); err != nil {
		logger.Error("Failed to update weekly stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update weekly stats: %v", err)})
		return
	}

	duration := time.Since(start)
	logger.Info("Weekly stats updated from LeetCode successfully",
		zap.Duration("duration", duration),
		zap.Int("problems_count", stats.TotalSolved),
		zap.Int("contest_rating", stats.ContestRating),
	)

	c.JSON(http.StatusOK, weeklyStats)
}

// UpdateAllWeeklyStats updates weekly stats for all students
func (h *WeeklyStatsHandler) UpdateAllWeeklyStats(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "UpdateAllWeeklyStats"),
		zap.String("request_id", c.GetString("request_id")),
	)

	logger.Info("Starting weekly stats update for all students")

	// Get all students (paginated)
	page := 1
	pageSize := 10
	var processedCount int
	var failedCount int
	var totalProcessed int

	for {
		logger.Info("Processing student batch",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Int("processed_so_far", totalProcessed),
		)

		students, err := h.db.GetAllStudents()
		if err != nil {
			logger.Error("Failed to fetch students",
				zap.Error(err),
				zap.Int("page", page),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch students"})
			return
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			studentLogger := logger.With(
				zap.String("student_id", student.StudentID),
				zap.String("leetcode_id", student.LeetcodeID),
			)

			studentLogger.Info("Processing student")

			// Get LeetCode stats
			stats, err := h.lc.GetUserStats(student.LeetcodeID)
			if err != nil {
				studentLogger.Error("Failed to fetch LeetCode stats",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			// Create weekly stats
			now := time.Now()
			weekStart := now.AddDate(0, 0, -int(now.Weekday()))
			weekEnd := weekStart.AddDate(0, 0, 7)

			parsedID, err := strconv.ParseUint(student.StudentID, 10, 64)
			if err != nil {
				studentLogger.Error("Invalid student ID format",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			weeklyStats := &models.WeeklyStats{
				StudentID:      uint(parsedID),
				WeekStart:      weekStart,
				WeekEnd:        weekEnd,
				ProblemsSolved: stats.TotalSolved,
				EasySolved:     stats.EasyCount,
				MediumSolved:   stats.MediumCount,
				HardSolved:     stats.HardCount,
				ContestRating:  float64(stats.ContestRating),
				GlobalRanking:  stats.ContestRanking,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := h.db.UpdateWeeklyStats(weeklyStats); err != nil {
				studentLogger.Error("Failed to update weekly stats",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			studentLogger.Info("Successfully updated weekly stats",
				zap.Int("problems_count", stats.TotalSolved),
				zap.Int("contest_rating", stats.ContestRating),
			)
			processedCount++
			totalProcessed++

			// Add a small delay between students to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}

		page++
	}

	duration := time.Since(start)
	logger.Info("Weekly stats update completed for all students",
		zap.Duration("duration", duration),
		zap.Int("total_processed", processedCount),
		zap.Int("failed_count", failedCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Weekly stats update completed",
		"processed_count": processedCount,
		"failed_count":    failedCount,
		"duration":        duration.String(),
	})
}

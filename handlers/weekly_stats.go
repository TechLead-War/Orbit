package handlers

import (
	"net/http"
	"time"

	"log"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/pkg/leetcode"
	"github.com/gin-gonic/gin"
)

// WeeklyStatsDB interface for weekly stats operations
type WeeklyStatsDB interface {
	CreateWeeklyStats(stats *models.WeeklyStats) error
	GetWeeklyStats(studentID string, startDate, endDate time.Time) (*models.WeeklyStats, error)
	UpdateWeeklyStats(stats *models.WeeklyStats) error
	GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error)
	GetAllStudentsLatestStats() ([]models.WeeklyStats, error)
	GetAllStudents() ([]models.Student, error)
	GetStudentByID(studentID string) (*models.Student, error)
}

// WeeklyStatsHandler handles weekly statistics operations
type WeeklyStatsHandler struct {
	db WeeklyStatsDB
	lc *leetcode.Client
}

// NewWeeklyStatsHandler creates a new weekly stats handler
func NewWeeklyStatsHandler(db WeeklyStatsDB, lc *leetcode.Client) *WeeklyStatsHandler {
	return &WeeklyStatsHandler{db: db, lc: lc}
}

// GetStudentWeeklyStats gets all weekly statistics for a student
func (h *WeeklyStatsHandler) GetStudentWeeklyStats(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	stats, err := h.db.GetStudentWeeklyStats(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCurrentWeekStats gets the current week's statistics for a student
func (h *WeeklyStatsHandler) GetCurrentWeekStats(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	stats, err := h.db.GetWeeklyStats(studentID, weekStart, weekEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateWeeklyStats updates the weekly statistics for a student
func (h *WeeklyStatsHandler) UpdateWeeklyStats(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	// First verify student exists
	_, err := h.db.GetStudentByID(studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	// Parse request body
	var manualStats models.CreateWeeklyStatsRequest
	err = c.ShouldBindJSON(&manualStats)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// For manual updates, validate the request
	if manualStats.WeekStartDate.IsZero() || manualStats.WeekEndDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start_date and week_end_date are required"})
		return
	}

	stats := &models.WeeklyStats{
		StudentID:     studentID,
		WeekStartDate: manualStats.WeekStartDate,
		WeekEndDate:   manualStats.WeekEndDate,
		EasyCount:     manualStats.EasyCount,
		MediumCount:   manualStats.MediumCount,
		HardCount:     manualStats.HardCount,
		TotalCount:    manualStats.TotalCount,
	}

	// Check if stats already exist for this week
	existingStats, err := h.db.GetWeeklyStats(studentID, manualStats.WeekStartDate, manualStats.WeekEndDate)
	if err != nil && err.Error() != "record not found" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingStats != nil {
		stats.ID = existingStats.ID
		if err := h.db.UpdateWeeklyStats(stats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := h.db.CreateWeeklyStats(stats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateWeeklyStatsFromLeetCode updates the weekly statistics for a student from LeetCode
func (h *WeeklyStatsHandler) UpdateWeeklyStatsFromLeetCode(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	// Get student with LeetCode ID
	student, err := h.db.GetStudentByID(studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	if student.LeetcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student has no LeetCode ID"})
		return
	}

	// Get current LeetCode stats
	leetcodeStats, err := h.lc.GetUserStats(student.LeetcodeID)
	if err != nil {
		if leetcodeStats != nil && leetcodeStats.Error != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": leetcodeStats.Error})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch LeetCode stats: " + err.Error()})
		}
		return
	}

	// Calculate week boundaries
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Get previous week's stats if they exist
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := weekStart
	prevStats, err := h.db.GetWeeklyStats(studentID, prevWeekStart, prevWeekEnd)

	// Calculate the differences
	var stats *models.WeeklyStats
	if prevStats != nil {
		// Calculate progress since last week
		stats = &models.WeeklyStats{
			StudentID:     studentID,
			WeekStartDate: weekStart,
			WeekEndDate:   weekEnd,
			EasyCount:     leetcodeStats.EasyCount - prevStats.EasyCount,
			MediumCount:   leetcodeStats.MediumCount - prevStats.MediumCount,
			HardCount:     leetcodeStats.HardCount - prevStats.HardCount,
			TotalCount:    leetcodeStats.TotalSolved - prevStats.TotalCount,
		}
	} else {
		// No previous stats, use current totals
		stats = &models.WeeklyStats{
			StudentID:     studentID,
			WeekStartDate: weekStart,
			WeekEndDate:   weekEnd,
			EasyCount:     leetcodeStats.EasyCount,
			MediumCount:   leetcodeStats.MediumCount,
			HardCount:     leetcodeStats.HardCount,
			TotalCount:    leetcodeStats.TotalSolved,
		}
	}

	// Ensure no negative values (in case LeetCode stats were reset)
	if stats.EasyCount < 0 {
		stats.EasyCount = 0
	}
	if stats.MediumCount < 0 {
		stats.MediumCount = 0
	}
	if stats.HardCount < 0 {
		stats.HardCount = 0
	}
	if stats.TotalCount < 0 {
		stats.TotalCount = 0
	}

	// Check if stats already exist for this week
	existingStats, err := h.db.GetWeeklyStats(studentID, weekStart, weekEnd)
	if err != nil && err.Error() != "record not found" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingStats != nil {
		stats.ID = existingStats.ID
		if err := h.db.UpdateWeeklyStats(stats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := h.db.CreateWeeklyStats(stats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateAllWeeklyStats updates weekly stats for all students
func (h *WeeklyStatsHandler) UpdateAllWeeklyStats(c *gin.Context) {
	// Start processing in background
	go h.processAllStudents()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Weekly stats update started. Check individual student stats for progress.",
	})
}

// processAllStudents handles the background processing of all students' stats
func (h *WeeklyStatsHandler) processAllStudents() {
	log.Printf("[WeeklyStats] Starting weekly stats update for all students")

	students, err := h.db.GetAllStudents()
	if err != nil {
		log.Printf("[WeeklyStats] Error fetching students: %v", err)
		return
	}
	log.Printf("[WeeklyStats] Found %d students to process", len(students))

	// Calculate week boundaries
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	log.Printf("[WeeklyStats] Processing stats for week: %s to %s",
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"))

	// Get previous week's boundaries
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := weekStart

	// Process students in smaller batches to avoid overwhelming the LeetCode API
	batchSize := 5
	totalProcessed := 0
	totalSuccess := 0
	totalSkipped := 0
	totalErrors := 0

	for i := 0; i < len(students); i += batchSize {
		end := i + batchSize
		if end > len(students) {
			end = len(students)
		}

		batch := students[i:end]
		log.Printf("[WeeklyStats] Processing batch %d-%d of %d students", i+1, end, len(students))

		for _, student := range batch {
			if student.LeetcodeID == "" {
				log.Printf("[WeeklyStats] Skipping student %s (ID: %s) - No LeetCode ID",
					student.Name, student.StudentID)
				totalSkipped++
				continue
			}

			log.Printf("[WeeklyStats] Processing student: %s (LeetCode: %s)",
				student.Name, student.LeetcodeID)

			// Get current LeetCode stats
			leetcodeStats, err := h.lc.GetUserStats(student.LeetcodeID)
			if err != nil || leetcodeStats.Error != "" {
				log.Printf("[WeeklyStats] Error fetching LeetCode stats for %s: %v",
					student.LeetcodeID, err)
				totalErrors++
				continue
			}

			log.Printf("[WeeklyStats] Retrieved LeetCode stats for %s - Easy: %d, Medium: %d, Hard: %d, Total: %d",
				student.LeetcodeID,
				leetcodeStats.EasyCount,
				leetcodeStats.MediumCount,
				leetcodeStats.HardCount,
				leetcodeStats.TotalSolved)

			// Get previous week's stats if they exist
			prevStats, err := h.db.GetWeeklyStats(student.StudentID, prevWeekStart, prevWeekEnd)
			if err != nil && err.Error() != "record not found" {
				log.Printf("[WeeklyStats] Error fetching previous week stats for %s: %v",
					student.StudentID, err)
			}

			// Calculate the differences
			var stats *models.WeeklyStats
			if prevStats != nil {
				log.Printf("[WeeklyStats] Found previous week stats for %s", student.StudentID)
				// Calculate progress since last week
				stats = &models.WeeklyStats{
					StudentID:     student.StudentID,
					WeekStartDate: weekStart,
					WeekEndDate:   weekEnd,
					EasyCount:     leetcodeStats.EasyCount - prevStats.EasyCount,
					MediumCount:   leetcodeStats.MediumCount - prevStats.MediumCount,
					HardCount:     leetcodeStats.HardCount - prevStats.HardCount,
					TotalCount:    leetcodeStats.TotalSolved - prevStats.TotalCount,
				}
			} else {
				log.Printf("[WeeklyStats] No previous week stats found for %s, using current totals",
					student.StudentID)
				// No previous stats, use current totals
				stats = &models.WeeklyStats{
					StudentID:     student.StudentID,
					WeekStartDate: weekStart,
					WeekEndDate:   weekEnd,
					EasyCount:     leetcodeStats.EasyCount,
					MediumCount:   leetcodeStats.MediumCount,
					HardCount:     leetcodeStats.HardCount,
					TotalCount:    leetcodeStats.TotalSolved,
				}
			}

			// Ensure no negative values (in case LeetCode stats were reset)
			if stats.EasyCount < 0 {
				stats.EasyCount = 0
				log.Printf("[WeeklyStats] Corrected negative easy count for %s", student.StudentID)
			}
			if stats.MediumCount < 0 {
				stats.MediumCount = 0
				log.Printf("[WeeklyStats] Corrected negative medium count for %s", student.StudentID)
			}
			if stats.HardCount < 0 {
				stats.HardCount = 0
				log.Printf("[WeeklyStats] Corrected negative hard count for %s", student.StudentID)
			}
			if stats.TotalCount < 0 {
				stats.TotalCount = 0
				log.Printf("[WeeklyStats] Corrected negative total count for %s", student.StudentID)
			}

			// Check if stats already exist for this week
			existingStats, err := h.db.GetWeeklyStats(student.StudentID, weekStart, weekEnd)
			if err != nil && err.Error() != "record not found" {
				log.Printf("[WeeklyStats] Error checking existing stats for %s: %v",
					student.StudentID, err)
				totalErrors++
				continue
			}

			// Update or create stats in the weekly_stats table
			if existingStats != nil {
				stats.ID = existingStats.ID
				if err := h.db.UpdateWeeklyStats(stats); err != nil {
					log.Printf("[WeeklyStats] Error updating weekly_stats for %s: %v",
						student.StudentID, err)
					totalErrors++
					continue
				}
				log.Printf("[WeeklyStats] Updated existing weekly_stats record (ID: %d) for %s",
					stats.ID, student.StudentID)
			} else {
				if err := h.db.CreateWeeklyStats(stats); err != nil {
					log.Printf("[WeeklyStats] Error creating weekly_stats for %s: %v",
						student.StudentID, err)
					totalErrors++
					continue
				}
				log.Printf("[WeeklyStats] Created new weekly_stats record for %s", student.StudentID)
			}

			totalSuccess++
			totalProcessed++

			// Add delay to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("[WeeklyStats] Weekly stats update completed:")
	log.Printf("- Total students processed: %d", totalProcessed)
	log.Printf("- Successfully updated: %d", totalSuccess)
	log.Printf("- Skipped (no LeetCode ID): %d", totalSkipped)
	log.Printf("- Errors encountered: %d", totalErrors)
	log.Printf("- Database table updated: weekly_stats")
}

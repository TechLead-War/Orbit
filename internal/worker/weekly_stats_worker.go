package worker

import (
	"log"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/pkg/leetcode"
)

type StudentDB interface {
	GetAllStudents() ([]*models.Student, error)
}

type WeeklyStatsDB interface {
	CreateWeeklyStats(stats *models.WeeklyStats) error
	GetWeeklyStats(studentID string, startDate, endDate time.Time) (*models.WeeklyStats, error)
}

// WeeklyStatsWorker handles periodic updates of student LeetCode statistics
type WeeklyStatsWorker struct {
	studentDB StudentDB
	statsDB   WeeklyStatsDB
	lc        *leetcode.Client
}

// NewWeeklyStatsWorker creates a new WeeklyStatsWorker
func NewWeeklyStatsWorker(studentDB StudentDB, statsDB WeeklyStatsDB, lc *leetcode.Client) *WeeklyStatsWorker {
	return &WeeklyStatsWorker{
		studentDB: studentDB,
		statsDB:   statsDB,
		lc:        lc,
	}
}

// Start begins the weekly stats update process
func (w *WeeklyStatsWorker) Start() {
	ticker := time.NewTicker(7 * 24 * time.Hour) // Run every week
	go func() {
		for {
			<-ticker.C
			w.UpdateAllStudentsStats()
		}
	}()
}

// UpdateAllStudentsStats updates the weekly statistics for all students
func (w *WeeklyStatsWorker) UpdateAllStudentsStats() {
	students, err := w.studentDB.GetAllStudents()
	if err != nil {
		log.Printf("Failed to get students: %v", err)
		return
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)

	for _, student := range students {
		if student.LeetcodeID == "" {
			continue
		}

		stats, err := w.lc.GetUserStats(student.LeetcodeID)
		if err != nil {
			log.Printf("Failed to get stats for student %s: %v", student.StudentID, err)
			continue
		}

		weeklyStats := &models.WeeklyStats{
			StudentID:     student.StudentID,
			WeekStartDate: weekStart,
			WeekEndDate:   now,
			EasyCount:     stats.EasyCount,
			MediumCount:   stats.MediumCount,
			HardCount:     stats.HardCount,
			ProblemsCount: stats.TotalSolved,
			ContestRating: stats.ContestRating,
			GlobalRanking: stats.ContestRanking,
		}

		// Check if stats already exist for this week
		existing, err := w.statsDB.GetWeeklyStats(student.StudentID, weekStart, now)
		if err == nil && existing != nil {
			continue // Skip if stats already exist for this week
		}

		if err := w.statsDB.CreateWeeklyStats(weeklyStats); err != nil {
			log.Printf("Failed to save stats for student %s: %v", student.StudentID, err)
		}

		// Add delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}
}

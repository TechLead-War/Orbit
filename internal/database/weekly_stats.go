package database

import (
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/gorm"
)

// WeeklyStatsDB handles database operations for weekly statistics
type WeeklyStatsDB struct {
	db *gorm.DB
}

// NewWeeklyStatsDB creates a new WeeklyStatsDB instance
func NewWeeklyStatsDB(db *gorm.DB) *WeeklyStatsDB {
	return &WeeklyStatsDB{db: db}
}

// CreateWeeklyStats creates a new weekly stats record
func (d *WeeklyStatsDB) CreateWeeklyStats(stats *models.WeeklyStats) error {
	return d.db.Create(stats).Error
}

// GetWeeklyStats retrieves weekly stats for a student within a date range
func (d *WeeklyStatsDB) GetWeeklyStats(studentID string, startDate, endDate time.Time) (*models.WeeklyStats, error) {
	var stats models.WeeklyStats
	err := d.db.Where("student_id = ? AND week_start_date = ? AND week_end_date = ?",
		studentID, startDate, endDate).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateWeeklyStats updates an existing weekly stats record
func (d *WeeklyStatsDB) UpdateWeeklyStats(stats *models.WeeklyStats) error {
	return d.db.Save(stats).Error
}

// GetStudentWeeklyStats retrieves all weekly stats for a student
func (d *WeeklyStatsDB) GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error) {
	var stats []models.WeeklyStats
	err := d.db.Where("student_id = ?", studentID).Order("week_start_date desc").Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetAllStudentsLatestStats retrieves the latest weekly stats for all students
func (d *WeeklyStatsDB) GetAllStudentsLatestStats() ([]models.WeeklyStats, error) {
	var stats []models.WeeklyStats
	err := d.db.Raw(`
		SELECT DISTINCT ON (student_id) *
		FROM weekly_stats
		ORDER BY student_id, week_start_date DESC
	`).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

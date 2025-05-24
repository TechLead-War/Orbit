package database

import (
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/gorm"
)

// WeeklyStatsDB defines the database interface for weekly stats
type WeeklyStatsDB interface {
	GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error)
	GetWeeklyStats(studentID string, start, end time.Time) (*models.WeeklyStats, error)
	UpdateWeeklyStats(stats *models.WeeklyStats) error
	CreateWeeklyStats(stats *models.WeeklyStats) error
	GetAllStudents() ([]*models.Student, error)
	GetStudentByID(studentID string) (*models.Student, error)
}

// weeklyStatsDB implements WeeklyStatsDB
type weeklyStatsDB struct {
	db *gorm.DB
}

// NewWeeklyStatsDB creates a new weekly stats database handler
func NewWeeklyStatsDB(db *gorm.DB) WeeklyStatsDB {
	return &weeklyStatsDB{db: db}
}

// GetStudentWeeklyStats retrieves all weekly stats for a student
func (db *weeklyStatsDB) GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error) {
	var stats []models.WeeklyStats
	err := db.db.Where("student_id = ?", studentID).
		Order("week_start_date DESC").
		Find(&stats).Error
	return stats, err
}

// GetWeeklyStats retrieves weekly stats for a specific week
func (db *weeklyStatsDB) GetWeeklyStats(studentID string, start, end time.Time) (*models.WeeklyStats, error) {
	var stats models.WeeklyStats
	err := db.db.Where("student_id = ? AND week_start_date = ?", studentID, start).
		First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateWeeklyStats updates or creates weekly stats
func (db *weeklyStatsDB) UpdateWeeklyStats(stats *models.WeeklyStats) error {
	return db.db.Save(stats).Error
}

// CreateWeeklyStats creates a new weekly stats record
func (db *weeklyStatsDB) CreateWeeklyStats(stats *models.WeeklyStats) error {
	return db.db.Create(stats).Error
}

// GetAllStudents retrieves all students
func (db *weeklyStatsDB) GetAllStudents() ([]*models.Student, error) {
	var students []*models.Student
	err := db.db.Find(&students).Error
	return students, err
}

// GetStudentByID retrieves a student by their ID
func (db *weeklyStatsDB) GetStudentByID(studentID string) (*models.Student, error) {
	var student models.Student
	err := db.db.Where("student_id = ?", studentID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

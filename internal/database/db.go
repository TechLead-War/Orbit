package database

import (
	"fmt"
	"os"
	"time"

	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "Orbit"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

// StudentDB handles database operations for students
type StudentDB struct {
	db *gorm.DB
}

// NewStudentDB creates a new StudentDB instance
func NewStudentDB(db *gorm.DB) *StudentDB {
	return &StudentDB{db: db}
}

// CreateStudent creates a new student record
func (d *StudentDB) CreateStudent(student *models.Student) error {
	return d.db.Create(student).Error
}

// CreateFileUpload creates a new file upload record
func (d *StudentDB) CreateFileUpload(upload *models.FileUpload) error {
	return d.db.Create(upload).Error
}

// UpdateFileUpload updates an existing file upload record
func (d *StudentDB) UpdateFileUpload(upload *models.FileUpload) error {
	return d.db.Save(upload).Error
}

// GetAllStudents retrieves all students from the database
func (d *StudentDB) GetAllStudents() ([]models.Student, error) {
	var students []models.Student
	if err := d.db.Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}

// GetStudentByID retrieves a student by their ID
func (d *StudentDB) GetStudentByID(studentID string) (*models.Student, error) {
	var student models.Student
	if err := d.db.Where("student_id = ?", studentID).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetStudentWeeklyStats retrieves all weekly stats for a student
func (d *StudentDB) GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error) {
	var stats []models.WeeklyStats
	if err := d.db.Where("student_id = ?", studentID).Order("week_start_date DESC").Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// GetWeeklyStats retrieves weekly stats for a student within a date range
func (d *StudentDB) GetWeeklyStats(studentID string, start, end time.Time) (*models.WeeklyStats, error) {
	var stats models.WeeklyStats
	if err := d.db.Where("student_id = ? AND week_start_date = ?", studentID, start).First(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateWeeklyStats updates an existing weekly stats record
func (d *StudentDB) UpdateWeeklyStats(stats *models.WeeklyStats) error {
	return d.db.Save(stats).Error
}

// CreateWeeklyStats creates a new weekly stats record
func (d *StudentDB) CreateWeeklyStats(stats *models.WeeklyStats) error {
	return d.db.Create(stats).Error
}

// GetAllStudentsLatestStats retrieves the latest weekly stats for all students
func (d *StudentDB) GetAllStudentsLatestStats() ([]models.WeeklyStats, error) {
	var stats []models.WeeklyStats
	subQuery := d.db.Model(&models.WeeklyStats{}).
		Select("student_id, MAX(week_start_date) as max_date").
		Group("student_id")

	if err := d.db.Table("weekly_stats").
		Joins("JOIN (?) as latest ON weekly_stats.student_id = latest.student_id AND weekly_stats.week_start_date = latest.max_date",
			subQuery).
		Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// ListStudents retrieves a paginated list of students with their ratings
func (d *StudentDB) ListStudents(page, pageSize int) ([]*models.Student, error) {
	var students []*models.Student
	offset := (page - 1) * pageSize

	err := d.db.Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC").Limit(1)
	}).Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}

// AddContestHistories adds multiple contest history records for a student
func (d *StudentDB) AddContestHistories(studentID uint, histories []*models.ContestHistory) error {
	// Use a transaction to ensure all histories are added or none
	return d.db.Transaction(func(tx *gorm.DB) error {
		for _, history := range histories {
			history.StudentID = studentID
			if err := tx.Create(history).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteContestHistory deletes all contest history records for a student
func (d *StudentDB) DeleteContestHistory(studentID uint) error {
	return d.db.Where("student_id = ?", studentID).Delete(&models.ContestHistory{}).Error
}

// GetContestRankings retrieves contest history records for a student
func (d *StudentDB) GetContestRankings(studentID uint) ([]*models.ContestHistory, error) {
	var rankings []*models.ContestHistory
	err := d.db.Where("student_id = ?", studentID).
		Order("contest_date desc").
		Find(&rankings).Error
	if err != nil {
		return nil, err
	}
	return rankings, nil
}

// GetLeetCodeStats retrieves LeetCode statistics for a student
func (d *StudentDB) GetLeetCodeStats(leetcodeID string) (*models.LeetCodeStats, error) {
	// Create a LeetCode service instance
	leetcodeService := leetcode.NewService()

	// Get user profile from LeetCode API
	profile, err := leetcodeService.GetUserProfile(leetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LeetCode profile: %w", err)
	}

	// Get contest ranking
	contestInfo, err := leetcodeService.GetContestRanking(leetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}

	// Map the data to our stats model
	stats := &models.LeetCodeStats{
		TotalSolved:    0,
		EasySolved:     0,
		MediumSolved:   0,
		HardSolved:     0,
		ContestRating:  contestInfo.Data.UserContestRanking.Rating,
		ContestRanking: contestInfo.Data.UserContestRanking.GlobalRanking,
	}

	// Map submission stats
	for _, submission := range profile.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		count := submission.Count
		switch submission.Difficulty {
		case "Easy":
			stats.EasySolved = count
		case "Medium":
			stats.MediumSolved = count
		case "Hard":
			stats.HardSolved = count
		}
	}
	stats.TotalSolved = stats.EasySolved + stats.MediumSolved + stats.HardSolved

	return stats, nil
}

// GetStudent retrieves a student by their numeric ID
func (d *StudentDB) GetStudent(id uint) (*models.Student, error) {
	var student models.Student
	err := d.db.First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// GetStudentWithStats retrieves a student with their latest rating and contest stats
func (d *StudentDB) GetStudentWithStats(id uint) (*models.Student, error) {
	var student models.Student

	// Get student with latest rating
	err := d.db.Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC").Limit(1)
	}).First(&student, id).Error
	if err != nil {
		return nil, err
	}

	// Get contest stats
	var contestStats models.ContestStats
	err = d.db.Model(&models.ContestHistory{}).
		Select("COUNT(*) as contests_participated, "+
			"AVG(rating) as average_rating, "+
			"MIN(ranking) as best_ranking, "+
			"SUM(problems_solved) as total_problems_solved").
		Where("student_id = ?", id).
		Scan(&contestStats).Error
	if err == nil {
		student.ContestStats = &contestStats
	}

	// Get contest history
	var contestHistory []*models.ContestHistory
	err = d.db.Where("student_id = ?", id).
		Order("contest_date desc").
		Find(&contestHistory).Error
	if err == nil {
		student.ContestHistory = contestHistory
	}

	return &student, nil
}

// UpdateStudentRating updates or creates a rating record for a student
func (d *StudentDB) UpdateStudentRating(studentID uint, rating *models.Rating) error {
	rating.StudentID = studentID
	rating.CreatedAt = time.Now()
	rating.RecordedAt = time.Now()
	return d.db.Create(rating).Error
}

package repository

import (
	"context"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/gorm"
)

// StudentRepository handles all database operations for students
type StudentRepository struct {
	db *gorm.DB
}

// NewStudentRepository creates a new student repository
func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create adds a new student to the database
func (r *StudentRepository) Create(ctx context.Context, student *models.Student) error {
	return r.db.WithContext(ctx).Create(student).Error
}

// GetByID retrieves a student by their ID
func (r *StudentRepository) GetByID(ctx context.Context, id uint) (*models.Student, error) {
	var student models.Student
	err := r.db.WithContext(ctx).First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// GetByLeetcodeID retrieves a student by their LeetCode ID
func (r *StudentRepository) GetByLeetcodeID(ctx context.Context, leetcodeID string) (*models.Student, error) {
	var student models.Student
	err := r.db.WithContext(ctx).Where("leetcode_id = ?", leetcodeID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// Update updates a student's information
func (r *StudentRepository) Update(ctx context.Context, student *models.Student) error {
	return r.db.WithContext(ctx).Save(student).Error
}

// Delete removes a student from the database
func (r *StudentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Student{}, id).Error
}

// AddRating adds a new rating record for a student
func (r *StudentRepository) AddRating(ctx context.Context, rating *models.Rating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

// GetStudentStats retrieves statistics for a student
func (r *StudentRepository) GetStudentStats(ctx context.Context, studentID uint) (*models.StudentStats, error) {
	var stats models.StudentStats
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Get current rating
	var currentRating models.Rating
	err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("recorded_at desc").
		First(&currentRating).Error
	if err != nil {
		return nil, err
	}

	// Get rating from 30 days ago
	var oldRating models.Rating
	err = r.db.WithContext(ctx).
		Where("student_id = ? AND recorded_at <= ?", studentID, thirtyDaysAgo).
		Order("recorded_at desc").
		First(&oldRating).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Calculate statistics
	stats.StudentID = studentID
	stats.CurrentRank = currentRating.GlobalRank
	stats.TotalProblems = currentRating.ProblemsCount
	stats.RatingChange30Days = currentRating.Rating - oldRating.Rating

	// Calculate average rating
	var avgRating struct {
		Average float64
	}
	err = r.db.WithContext(ctx).
		Table("ratings").
		Select("AVG(rating) as average").
		Where("student_id = ?", studentID).
		Scan(&avgRating).Error
	if err != nil {
		return nil, err
	}
	stats.AverageRating = avgRating.Average

	return &stats, nil
}

// ListStudents retrieves a paginated list of students
func (r *StudentRepository) ListStudents(ctx context.Context, offset, limit int) ([]models.Student, error) {
	var students []models.Student
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&students).Error
	return students, err
}

func (r *StudentRepository) GetByIDWithRatings(ctx context.Context, id uint) (*models.Student, error) {
	var student models.Student
	err := r.db.WithContext(ctx).Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC")
	}).First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepository) List(ctx context.Context, page, pageSize int) ([]models.Student, error) {
	var students []models.Student
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

// AddContestHistory adds contest history records for a student
func (r *StudentRepository) AddContestHistory(ctx context.Context, history *models.ContestHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetContestHistory retrieves contest history for a student
func (r *StudentRepository) GetContestHistory(ctx context.Context, studentID uint) ([]models.ContestHistory, error) {
	var history []models.ContestHistory
	err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("contest_date DESC").
		Find(&history).Error
	return history, err
}

// DeleteContestHistory deletes all contest history for a student
func (r *StudentRepository) DeleteContestHistory(ctx context.Context, studentID uint) error {
	return r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Delete(&models.ContestHistory{}).Error
}

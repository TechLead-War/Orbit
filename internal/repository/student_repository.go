package repository

import (
	"context"

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
	stats.StudentID = studentID

	// Get current rating and problems count
	var currentRating models.Rating
	err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("recorded_at desc").
		First(&currentRating).Error
	if err != nil {
		return nil, err
	}

	stats.CurrentRating = currentRating.Rating
	stats.ProblemsCount = currentRating.ProblemsCount
	stats.GlobalRank = currentRating.GlobalRank

	// Get contest statistics
	var contestCount int64
	err = r.db.WithContext(ctx).
		Model(&models.ContestHistory{}).
		Where("student_id = ?", studentID).
		Count(&contestCount).Error
	if err != nil {
		return nil, err
	}
	stats.ContestsParticipated = int(contestCount)

	// Get latest contest rating
	var latestContest models.ContestHistory
	err = r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("contest_date desc").
		First(&latestContest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		stats.ContestRating = latestContest.Rating
	}

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

// GetContestStats retrieves aggregated contest statistics for a student
func (r *StudentRepository) GetContestStats(ctx context.Context, studentID uint) (*models.ContestStats, error) {
	var stats models.ContestStats
	var count int64

	// Get total contests participated
	err := r.db.WithContext(ctx).
		Model(&models.ContestHistory{}).
		Where("student_id = ?", studentID).
		Count(&count).Error
	if err != nil {
		return nil, err
	}
	stats.ContestsParticipated = int(count)

	if stats.ContestsParticipated == 0 {
		return &stats, nil
	}

	// Get average rating and total problems solved
	err = r.db.WithContext(ctx).
		Model(&models.ContestHistory{}).
		Select("COALESCE(AVG(rating), 0) as average_rating, COALESCE(SUM(problems_solved), 0) as total_problems_solved").
		Where("student_id = ?", studentID).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Get best ranking
	var bestContest models.ContestHistory
	err = r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("ranking ASC").
		First(&bestContest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		stats.BestRanking = bestContest.Ranking
	}

	return &stats, nil
}

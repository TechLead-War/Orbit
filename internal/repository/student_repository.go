package repository

import (
	"context"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/gorm"
)

type StudentRepository struct {
	DB *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{
		DB: db,
	}
}

func (r *StudentRepository) Create(ctx context.Context, student *models.Student) error {
	return r.DB.WithContext(ctx).Create(student).Error
}

func (r *StudentRepository) GetByID(ctx context.Context, id uint) (*models.Student, error) {
	var student models.Student
	if err := r.DB.WithContext(ctx).First(&student, id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepository) GetByLeetcodeID(ctx context.Context, leetcodeID string) (*models.Student, error) {
	var student models.Student
	if err := r.DB.WithContext(ctx).Where("leetcode_id = ?", leetcodeID).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepository) Update(ctx context.Context, student *models.Student) error {
	return r.DB.WithContext(ctx).Save(student).Error
}

func (r *StudentRepository) Delete(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&models.Student{}, id).Error
}

func (r *StudentRepository) List(ctx context.Context, page, pageSize int) ([]models.Student, error) {
	var students []models.Student
	offset := (page - 1) * pageSize
	if err := r.DB.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}

func (r *StudentRepository) GetByIDWithRatings(ctx context.Context, id uint) (*models.Student, error) {
	var student models.Student
	if err := r.DB.WithContext(ctx).Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC")
	}).First(&student, id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *StudentRepository) AddRating(ctx context.Context, rating *models.Rating) error {
	return r.DB.WithContext(ctx).Create(rating).Error
}

func (r *StudentRepository) UpdateStudentRating(ctx context.Context, studentID uint, rating *models.Rating) error {
	rating.StudentID = studentID
	return r.DB.WithContext(ctx).Create(rating).Error
}

func (r *StudentRepository) GetStudentStats(ctx context.Context, studentID uint) (*models.StudentStats, error) {
	var stats models.StudentStats
	// Implement stats collection logic here
	return &stats, nil
}

func (r *StudentRepository) ListStudents(ctx context.Context, offset, limit int, department, batch string) ([]*models.Student, error) {
	var students []*models.Student
	query := r.DB.WithContext(ctx)

	if department != "" {
		query = query.Where("department = ?", department)
	}
	if batch != "" {
		query = query.Where("batch = ?", batch)
	}

	err := query.Offset(offset).Limit(limit).Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (r *StudentRepository) GetContestHistory(ctx context.Context, studentID uint) ([]models.ContestHistory, error) {
	var history []models.ContestHistory
	if err := r.DB.WithContext(ctx).Where("student_id = ?", studentID).Order("contest_date DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func (r *StudentRepository) DeleteContestHistory(ctx context.Context, studentID uint) error {
	return r.DB.WithContext(ctx).Where("student_id = ?", studentID).Delete(&models.ContestHistory{}).Error
}

func (r *StudentRepository) AddContestHistories(ctx context.Context, studentID uint, histories []*models.ContestHistory) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, history := range histories {
			history.StudentID = studentID
			if err := tx.WithContext(ctx).Create(history).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *StudentRepository) GetLeetCodeStats(ctx context.Context, leetcodeID string) (*models.LeetCodeStats, error) {
	var stats models.LeetCodeStats
	if err := r.DB.WithContext(ctx).Where("leetcode_id = ?", leetcodeID).First(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *StudentRepository) UpdateLeetCodeStats(ctx context.Context, id uint, stats *models.LeetCodeStats) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Updates(stats).Error
}

func (r *StudentRepository) GetDailyProgress(ctx context.Context, studentID uint, start, end time.Time) ([]*models.DailyProgress, error) {
	var progress []*models.DailyProgress
	if err := r.DB.WithContext(ctx).
		Where("student_id = ? AND date BETWEEN ? AND ?", studentID, start, end).
		Order("date ASC").
		Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (r *StudentRepository) GetWeeklyStats(ctx context.Context, studentID uint, start, end time.Time) (*models.WeeklyStats, error) {
	var stats models.WeeklyStats
	if err := r.DB.WithContext(ctx).
		Where("student_id = ? AND week_start_date BETWEEN ? AND ?", studentID, start, end).
		First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &stats, nil
}

func (r *StudentRepository) GetLeaderboard(ctx context.Context, start time.Time, department, batch string) ([]*models.Student, error) {
	var students []*models.Student
	query := r.DB.WithContext(ctx).
		Joins("LEFT JOIN weekly_stats ON students.id = weekly_stats.student_id").
		Where("weekly_stats.week_start_date >= ?", start)

	if department != "" {
		query = query.Where("students.department = ?", department)
	}
	if batch != "" {
		query = query.Where("students.batch = ?", batch)
	}

	err := query.
		Order("weekly_stats.total_problems_solved DESC").
		Group("students.id").
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (r *StudentRepository) GetTrendingStudents(ctx context.Context, start time.Time, limit int) ([]*models.Student, error) {
	var students []*models.Student
	err := r.DB.WithContext(ctx).
		Joins("LEFT JOIN weekly_stats ON students.id = weekly_stats.student_id").
		Where("weekly_stats.week_start_date >= ?", start).
		Order("weekly_stats.problems_solved_this_week DESC").
		Limit(limit).
		Group("students.id").
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

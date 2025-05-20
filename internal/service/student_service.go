package service

import (
	"context"
	"errors"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"go.uber.org/zap"
)

// StudentService handles business logic for students
type StudentService struct {
	repo   *repository.StudentRepository
	logger *zap.Logger
}

// NewStudentService creates a new student service
func NewStudentService(repo *repository.StudentRepository, logger *zap.Logger) *StudentService {
	return &StudentService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterStudent registers a new student
func (s *StudentService) RegisterStudent(ctx context.Context, student *models.Student) error {
	// Validate student data
	if err := s.validateStudent(student); err != nil {
		s.logger.Error("Invalid student data", zap.Error(err))
		return err
	}

	// Set creation time
	now := time.Now()
	student.CreatedAt = now
	student.UpdatedAt = now

	// Create student
	if err := s.repo.Create(ctx, student); err != nil {
		s.logger.Error("Failed to create student", zap.Error(err))
		return err
	}

	s.logger.Info("Student registered successfully", zap.String("leetcode_id", student.LeetcodeID))
	return nil
}

// UpdateStudentRating updates a student's LeetCode rating
func (s *StudentService) UpdateStudentRating(ctx context.Context, studentID uint, rating *models.Rating) error {
	// Validate rating data
	if rating.Rating < 0 {
		return errors.New("rating cannot be negative")
	}

	// Set recording time
	rating.RecordedAt = time.Now()
	rating.StudentID = studentID

	// Add rating
	if err := s.repo.AddRating(ctx, rating); err != nil {
		s.logger.Error("Failed to update student rating",
			zap.Uint("student_id", studentID),
			zap.Error(err))
		return err
	}

	s.logger.Info("Student rating updated",
		zap.Uint("student_id", studentID),
		zap.Int("rating", rating.Rating))
	return nil
}

// GetStudentStats retrieves a student's statistics
func (s *StudentService) GetStudentStats(ctx context.Context, studentID uint) (*models.StudentStats, error) {
	stats, err := s.repo.GetStudentStats(ctx, studentID)
	if err != nil {
		s.logger.Error("Failed to get student stats",
			zap.Uint("student_id", studentID),
			zap.Error(err))
		return nil, err
	}

	return stats, nil
}

// ListStudents retrieves a paginated list of students
func (s *StudentService) ListStudents(ctx context.Context, page, pageSize int) ([]models.Student, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	students, err := s.repo.ListStudents(ctx, offset, pageSize)
	if err != nil {
		s.logger.Error("Failed to list students", zap.Error(err))
		return nil, err
	}

	return students, nil
}

// validateStudent validates student data
func (s *StudentService) validateStudent(student *models.Student) error {
	if student.Name == "" {
		return errors.New("student name is required")
	}
	if student.Email == "" {
		return errors.New("student email is required")
	}
	if student.LeetcodeID == "" {
		return errors.New("LeetCode ID is required")
	}
	return nil
}

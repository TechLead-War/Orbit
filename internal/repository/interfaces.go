package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ayush/ORBIT/internal/models"
)

// Error definitions
var (
	ErrNotFound = errors.New("record not found")
)

type WeeklyStatsDB interface {
	GetStudentWeeklyStats(ctx context.Context, studentID uint) (*models.WeeklyStats, error)
	UpdateWeeklyStats(ctx context.Context, stats *models.WeeklyStats) error
	UpdateAllWeeklyStats(ctx context.Context) error
}

type StudentDB interface {
	Create(ctx context.Context, student *models.Student) error
	GetByID(ctx context.Context, id uint) (*models.Student, error)
	GetByLeetcodeID(ctx context.Context, leetcodeID string) (*models.Student, error)
	Update(ctx context.Context, student *models.Student) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int) ([]models.Student, error)
	GetByIDWithRatings(ctx context.Context, id uint) (*models.Student, error)
	AddRating(ctx context.Context, rating *models.Rating) error
	GetStudentStats(ctx context.Context, studentID uint) (*models.StudentStats, error)
	ListStudents(ctx context.Context, offset, limit int, department, batch string) ([]*models.Student, error)
	GetContestHistory(ctx context.Context, studentID uint) ([]models.ContestHistory, error)
	DeleteContestHistory(ctx context.Context, studentID uint) error
	AddContestHistories(ctx context.Context, studentID uint, histories []*models.ContestHistory) error
	GetLeetCodeStats(ctx context.Context, leetcodeID string) (*models.LeetCodeStats, error)
	UpdateLeetCodeStats(ctx context.Context, id uint, stats *models.LeetCodeStats) error
	GetDailyProgress(ctx context.Context, studentID uint, start, end time.Time) ([]*models.DailyProgress, error)
	GetWeeklyStats(ctx context.Context, studentID uint, start, end time.Time) (*models.WeeklyStats, error)
	GetLeaderboard(ctx context.Context, start time.Time, department, batch string) ([]*models.Student, error)
	GetTrendingStudents(ctx context.Context, start time.Time, limit int) ([]*models.Student, error)
}

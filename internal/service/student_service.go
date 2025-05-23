package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"go.uber.org/zap"
)

type StudentService struct {
	repo            *repository.StudentRepository
	logger          *zap.Logger
	leetcodeService *LeetCodeService
}

func NewStudentService(repo *repository.StudentRepository, logger *zap.Logger) *StudentService {
	return &StudentService{
		repo:            repo,
		logger:          logger,
		leetcodeService: NewLeetCodeService(),
	}
}

func (s *StudentService) RegisterStudent(ctx context.Context, student *models.Student) error {
	stats, err := s.leetcodeService.GetUserStats(student.LeetcodeID)
	if err != nil {
		s.logger.Error("failed to fetch leetcode stats", zap.Error(err))
		return fmt.Errorf("failed to fetch leetcode stats: %w", err)
	}

	// Get contest stats
	contestStats, err := s.leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		s.logger.Error("failed to fetch contest stats", zap.Error(err))
		return fmt.Errorf("failed to fetch contest stats: %w", err)
	}

	now := time.Now()

	rating := models.Rating{
		StudentID:     uint(student.ID),
		Rating:        0,
		ProblemsCount: stats["All"],
		EasyCount:     stats["Easy"],
		MediumCount:   stats["Medium"],
		HardCount:     stats["Hard"],
		GlobalRank:    stats["ranking"],
		RecordedAt:    now,
		CreatedAt:     now,
	}

	// Calculate overall rating (problems + contest performance)
	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(contestStats.Data.UserContestRanking.Rating * 0.2) // Contest rating contributes 20% to overall rating
	rating.Rating = problemRating + contestBonus

	student.Ratings = []models.Rating{rating}

	// Create student with initial rating
	if err := s.repo.Create(ctx, student); err != nil {
		return err
	}

	// Store contest history
	for _, contest := range contestStats.Data.UserContestRankingHistory {
		history := &models.ContestHistory{
			StudentID:         uint(student.ID),
			ContestTitle:      contest.Contest.Title,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: contest.FinishTimeInSeconds,
			ContestDate:       now, // We'll use current time as contest date since it's not provided in the API
			CreatedAt:         now,
		}
		if err := s.repo.AddContestHistory(ctx, history); err != nil {
			s.logger.Error("failed to store contest history", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *StudentService) UpdateStudentRating(ctx context.Context, studentID uint, rating *models.Rating) error {
	student, err := s.repo.GetByID(ctx, studentID)
	if err != nil {
		return err
	}

	// Get problem stats
	stats, err := s.leetcodeService.GetUserStats(student.LeetcodeID)
	if err != nil {
		s.logger.Error("failed to fetch leetcode stats", zap.Error(err))
		return fmt.Errorf("failed to fetch leetcode stats: %w", err)
	}

	contestStats, err := s.leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		s.logger.Error("failed to fetch contest stats", zap.Error(err))
		return fmt.Errorf("failed to fetch contest stats: %w", err)
	}

	now := time.Now()

	rating.ProblemsCount = stats["All"]
	rating.EasyCount = stats["Easy"]
	rating.MediumCount = stats["Medium"]
	rating.HardCount = stats["Hard"]
	rating.GlobalRank = stats["ranking"]
	rating.RecordedAt = now
	rating.StudentID = studentID

	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(contestStats.Data.UserContestRanking.Rating * 0.2) // Contest rating contributes 20% to overall rating
	rating.Rating = problemRating + contestBonus

	if err := s.repo.AddRating(ctx, rating); err != nil {
		return err
	}

	if err := s.repo.DeleteContestHistory(ctx, studentID); err != nil {
		s.logger.Error("failed to delete old contest history", zap.Error(err))
		return err
	}

	for _, contest := range contestStats.Data.UserContestRankingHistory {
		history := &models.ContestHistory{
			StudentID:         uint(studentID),
			ContestTitle:      contest.Contest.Title,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: contest.FinishTimeInSeconds,
			ContestDate:       now,
			CreatedAt:         now,
		}
		if err := s.repo.AddContestHistory(ctx, history); err != nil {
			s.logger.Error("failed to store contest history", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *StudentService) GetStudentStats(ctx context.Context, studentID uint) (*models.Student, error) {
	student, err := s.repo.GetByIDWithRatings(ctx, studentID)
	if err != nil {
		return nil, err
	}

	contestStats, err := s.repo.GetContestStats(ctx, studentID)
	if err != nil {
		s.logger.Error("failed to fetch contest stats", zap.Error(err))
	} else {
		student.ContestStats = contestStats
	}

	return student, nil
}

func (s *StudentService) ListStudents(ctx context.Context, page, pageSize int) ([]models.Student, error) {
	students, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list students", zap.Error(err))
		return nil, fmt.Errorf("failed to list students: %w", err)
	}

	// Get contest stats for each student
	for i := range students {
		if contestStats, err := s.repo.GetContestStats(ctx, students[i].ID); err == nil {
			students[i].ContestStats = contestStats
		} else {
			s.logger.Error("failed to get contest stats for student",
				zap.Uint("student_id", students[i].ID),
				zap.Error(err))
		}
	}

	return students, nil
}

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

func (s *StudentService) GetByID(ctx context.Context, studentID uint) (*models.Student, error) {
	return s.repo.GetByID(ctx, studentID)
}

func (s *StudentService) GetLeetCodeStats(ctx context.Context, leetcodeID string) (*leetcode.UserProfile, error) {
	return s.leetcodeService.GetUserProfile(leetcodeID)
}

func (s *StudentService) GetContestRankings(ctx context.Context, leetcodeID string) (*leetcode.ContestRankingInfo, error) {
	return s.leetcodeService.GetContestRanking(leetcodeID)
}

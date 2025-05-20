package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	// Get problem stats
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
		Rating:        0,
		ProblemsCount: stats["All"],
		EasyCount:     stats["Easy"],
		MediumCount:   stats["Medium"],
		HardCount:     stats["Hard"],
		GlobalRank:    stats["ranking"],
		ContestRating: contestStats.Data.UserContestRanking.Rating,
		ContestCount:  contestStats.Data.UserContestRanking.AttendedContestsCount,
		TopPercentage: contestStats.Data.UserContestRanking.TopPercentage,
		RecordedAt:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Calculate overall rating (problems + contest performance)
	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(rating.ContestRating * 0.2) // Contest rating contributes 20% to overall rating
	rating.Rating = problemRating + contestBonus

	student.Ratings = []models.Rating{rating}
	return s.repo.Create(ctx, student)
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

	// Get contest stats
	contestStats, err := s.leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		s.logger.Error("failed to fetch contest stats", zap.Error(err))
		return fmt.Errorf("failed to fetch contest stats: %w", err)
	}

	now := time.Now()

	// Update rating with both problem and contest data
	rating.ProblemsCount = stats["All"]
	rating.EasyCount = stats["Easy"]
	rating.MediumCount = stats["Medium"]
	rating.HardCount = stats["Hard"]
	rating.GlobalRank = stats["ranking"]
	rating.RecordedAt = now
	rating.StudentID = studentID
	rating.ContestRating = contestStats.Data.UserContestRanking.Rating
	rating.ContestCount = contestStats.Data.UserContestRanking.AttendedContestsCount
	rating.TopPercentage = contestStats.Data.UserContestRanking.TopPercentage
	rating.UpdatedAt = now

	// Calculate overall rating (problems + contest performance)
	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(rating.ContestRating * 0.2) // Contest rating contributes 20% to overall rating
	rating.Rating = problemRating + contestBonus

	// Store the rating
	if err := s.repo.AddRating(ctx, rating); err != nil {
		return err
	}

	// Delete existing contest history for this update
	if err := s.repo.DeleteContestHistory(ctx, studentID); err != nil {
		s.logger.Error("failed to delete old contest history", zap.Error(err))
		return err
	}

	// Store contest history
	for _, contest := range contestStats.Data.UserContestRankingHistory {
		history := &models.ContestHistory{
			StudentID:         studentID,
			ContestTitle:      contest.Contest.Title,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			Attended:          contest.Attended,
			TrendDirection:    contest.TrendDirection,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: contest.FinishTimeInSeconds,
			ContestDate:       now, // We'll use current time as contest date since it's not provided in the API
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := s.repo.AddContestHistory(ctx, history); err != nil {
			s.logger.Error("failed to store contest history", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *StudentService) GetStudentStats(ctx context.Context, studentID uint) (*models.Student, error) {
	return s.repo.GetByIDWithRatings(ctx, studentID)
}

func (s *StudentService) ListStudents(ctx context.Context, page, pageSize int) ([]models.Student, error) {
	return s.repo.List(ctx, page, pageSize)
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

func (s *StudentService) GetLeetCodeStats(ctx context.Context, leetcodeID string) (*LeetCodeUserProfile, error) {
	return s.leetcodeService.GetUserProfile(leetcodeID)
}

func (s *StudentService) GetContestRankings(ctx context.Context, leetcodeID string) (*ContestRankingInfo, error) {
	return s.leetcodeService.GetContestRanking(leetcodeID)
}

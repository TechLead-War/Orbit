package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ayush/ORBIT/internal/database"
	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrNotFound = errors.New("not found")
)

type StudentService struct {
	repo            *repository.StudentRepository
	logger          *zap.Logger
	leetcodeService *leetcode.Service
}

func NewStudentService(repo *repository.StudentRepository, logger *zap.Logger) *StudentService {
	return &StudentService{
		repo:            repo,
		logger:          logger,
		leetcodeService: leetcode.NewService(),
	}
}

// ListStudents retrieves a paginated list of students
func (s *StudentService) ListStudents(ctx context.Context, page, pageSize int, department, batch string) ([]*models.Student, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStudents(ctx, offset, pageSize, department, batch)
}

// GetStudent retrieves a student by ID
func (s *StudentService) GetStudent(ctx context.Context, id uint) (*models.Student, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return student, nil
}

// GetLeetCodeStats retrieves a student's LeetCode statistics using their LeetCode ID
func (s *StudentService) GetLeetCodeStats(ctx context.Context, id interface{}) (*models.LeetCodeStats, error) {
	var student *models.Student
	var err error

	switch v := id.(type) {
	case uint:
		student, err = s.repo.GetByID(ctx, v)
	case string:
		student, err = s.repo.GetByLeetcodeID(ctx, v)
	default:
		return nil, fmt.Errorf("invalid id type: %T", id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	profile, err := s.leetcodeService.GetUserProfile(student.LeetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LeetCode stats: %w", err)
	}

	contestInfo, err := s.leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}

	// Calculate total submissions from individual counts
	totalSolved := 0
	for _, submission := range profile.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		totalSolved += submission.Count
	}

	// Map the data to our stats model
	stats := &models.LeetCodeStats{
		StudentID:            student.ID,
		Rating:               profile.Data.MatchedUser.Profile.Ranking,
		GlobalRanking:        profile.Data.MatchedUser.Profile.Ranking,
		TotalSolved:          totalSolved,
		ContestsParticipated: len(contestInfo.Data.UserContestRankingHistory),
		ContestRating:        float64(contestInfo.Data.UserContestRanking.Rating),
		ContestGlobalRanking: contestInfo.Data.UserContestRanking.GlobalRanking,
		LastSolvedAt:         time.Now(),
		InitialRating:        profile.Data.MatchedUser.Profile.Ranking,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Calculate solved problems by difficulty
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

	return stats, nil
}

// GetDailyProgress retrieves a student's daily progress
func (s *StudentService) GetDailyProgress(ctx context.Context, id uint, startDate, endDate string) ([]*models.DailyProgress, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	return s.repo.GetDailyProgress(ctx, student.ID, start, end)
}

// GetWeeklyStats retrieves a student's weekly statistics
func (s *StudentService) GetWeeklyStats(ctx context.Context, id uint) (*models.WeeklyStats, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	return s.repo.GetWeeklyStats(ctx, student.ID, weekStart, weekEnd)
}

// GetContestHistory retrieves a student's contest history
func (s *StudentService) GetContestHistory(ctx context.Context, id uint) ([]models.ContestHistory, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.repo.GetContestHistory(ctx, student.ID)
}

// UpdateContestHistory updates a student's contest history
func (s *StudentService) UpdateContestHistory(ctx context.Context, id uint) ([]models.ContestHistory, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	contestInfo, err := s.leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}

	// Delete existing contest history
	if err := s.repo.DeleteContestHistory(ctx, student.ID); err != nil {
		return nil, fmt.Errorf("failed to delete existing contest history: %w", err)
	}

	// Create new contest history entries
	var histories []models.ContestHistory
	var historyPtrs []*models.ContestHistory
	now := time.Now()

	for _, contest := range contestInfo.Data.UserContestRankingHistory {
		history := models.ContestHistory{
			StudentID:         student.ID,
			ContestTitle:      contest.Contest.Title,
			ContestDate:       now,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: int64(contest.FinishTimeInSeconds),
			CreatedAt:         now,
		}
		histories = append(histories, history)
		historyPtrs = append(historyPtrs, &histories[len(histories)-1])
	}

	// Add new contest histories using pointer slice
	if err := s.repo.AddContestHistories(ctx, student.ID, historyPtrs); err != nil {
		return nil, fmt.Errorf("failed to add contest histories: %w", err)
	}

	return histories, nil
}

// GetLeaderboard retrieves the student leaderboard
func (s *StudentService) GetLeaderboard(ctx context.Context, timeframe, department, batch string) ([]*models.Student, error) {
	var start time.Time
	now := time.Now()

	switch timeframe {
	case "week":
		start = now.AddDate(0, 0, -7)
	case "month":
		start = now.AddDate(0, -1, 0)
	case "year":
		start = now.AddDate(-1, 0, 0)
	default:
		start = time.Time{} // All time
	}

	return s.repo.GetLeaderboard(ctx, start, department, batch)
}

// GetTrendingStudents retrieves trending students based on recent activity
func (s *StudentService) GetTrendingStudents(ctx context.Context, days, limit int) ([]*models.Student, error) {
	start := time.Now().AddDate(0, 0, -days)
	return s.repo.GetTrendingStudents(ctx, start, limit)
}

// UpdateAllStudentStats updates LeetCode stats for all students
func (s *StudentService) UpdateAllStudentStats(ctx context.Context) error {
	students, err := s.repo.List(ctx, 1, 1000) // Get first 1000 students
	if err != nil {
		return fmt.Errorf("failed to list students: %w", err)
	}

	for _, student := range students {
		stats, err := s.GetLeetCodeStats(ctx, student.ID)
		if err != nil {
			s.logger.Error("failed to get LeetCode stats",
				zap.Error(err),
				zap.String("leetcode_id", student.LeetcodeID),
			)
			continue
		}

		if err := s.repo.UpdateLeetCodeStats(ctx, student.ID, stats); err != nil {
			s.logger.Error("failed to update LeetCode stats",
				zap.Error(err),
				zap.String("leetcode_id", student.LeetcodeID),
			)
			continue
		}

		// Add a small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (s *StudentService) RegisterStudent(ctx context.Context, student *models.Student) error {
	// Validate student data
	if err := s.validateStudent(student); err != nil {
		return err
	}

	// Get initial LeetCode stats
	stats, err := s.GetLeetCodeStats(ctx, student.ID)
	if err != nil {
		s.logger.Error("failed to fetch leetcode stats", zap.Error(err))
		return fmt.Errorf("failed to fetch leetcode stats: %w", err)
	}

	// Create student record
	if err := s.repo.Create(ctx, student); err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}

	// Update LeetCode stats
	if err := s.repo.UpdateLeetCodeStats(ctx, student.ID, stats); err != nil {
		return fmt.Errorf("failed to update leetcode stats: %w", err)
	}

	return nil
}

func (s *StudentService) validateStudent(student *models.Student) error {
	if student.Name == "" {
		return errors.New("name is required")
	}
	if student.Email == "" {
		return errors.New("email is required")
	}
	if student.LeetcodeID == "" {
		return errors.New("leetcode ID is required")
	}
	if student.PassingYear < time.Now().Year() {
		return errors.New("invalid passing year")
	}
	if student.Batch == "" {
		return errors.New("batch is required")
	}
	if student.Department == "" {
		return errors.New("department is required")
	}
	return nil
}

func (s *StudentService) GetByID(ctx context.Context, studentID uint) (*models.Student, error) {
	return s.repo.GetByID(ctx, studentID)
}

func (s *StudentService) GetContestRankings(ctx context.Context, leetcodeID string) (*leetcode.ContestRankingInfo, error) {
	return s.leetcodeService.GetContestRanking(leetcodeID)
}

// CreateStudent creates a new student record
func (s *StudentService) CreateStudent(student *models.Student) error {
	return s.repo.Create(context.Background(), student)
}

// UpdateLeetCodeStats updates a student's LeetCode statistics
func (s *StudentService) UpdateLeetCodeStats(ctx context.Context, id uint) (*models.LeetCodeStats, error) {
	return s.GetLeetCodeStats(ctx, id)
}

// GetStudentWithStats retrieves a student with all their statistics
func (s *StudentService) GetStudentWithStats(ctx context.Context, id uint) (*models.Student, error) {
	student, err := s.repo.GetByIDWithRatings(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get LeetCode stats
	stats, err := s.GetLeetCodeStats(ctx, id)
	if err == nil {
		student.LeetCodeStats = stats
	}

	return student, nil
}

// CreateFileUpload creates a new file upload record
func (s *StudentService) CreateFileUpload(upload *models.FileUpload) error {
	// Create a new FileUploadDB instance
	fileUploadDB := database.NewFileUploadDB(s.repo.DB)
	return fileUploadDB.Create(context.Background(), upload)
}

// UpdateFileUpload updates a file upload record
func (s *StudentService) UpdateFileUpload(upload *models.FileUpload) error {
	// Create a new FileUploadDB instance
	fileUploadDB := database.NewFileUploadDB(s.repo.DB)
	return fileUploadDB.Update(context.Background(), upload)
}

// DeleteContestHistory deletes all contest history for a student
func (s *StudentService) DeleteContestHistory(ctx context.Context, studentID uint) error {
	return s.repo.DeleteContestHistory(ctx, studentID)
}

// AddContestHistories adds multiple contest histories for a student
func (s *StudentService) AddContestHistories(ctx context.Context, studentID uint, histories []models.ContestHistory) error {
	historyPtrs := make([]*models.ContestHistory, len(histories))
	for i := range histories {
		historyPtrs[i] = &histories[i]
	}
	return s.repo.AddContestHistories(ctx, studentID, historyPtrs)
}

// UpdateStudentRating updates a student's rating
func (s *StudentService) UpdateStudentRating(studentID uint, rating *models.Rating) error {
	return s.repo.UpdateStudentRating(context.Background(), studentID, rating)
}

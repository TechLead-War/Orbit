package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"go.uber.org/zap"
)

type ContestHistoryUpdater struct {
	repo           *repository.StudentRepository
	logger         *zap.Logger
	ticker         *time.Ticker
	updateInterval time.Duration
	done           chan bool
	batchSize      int
	wg             sync.WaitGroup
}

func NewContestHistoryUpdater(repo *repository.StudentRepository, logger *zap.Logger, updateInterval time.Duration) *ContestHistoryUpdater {
	return &ContestHistoryUpdater{
		repo:           repo,
		logger:         logger,
		updateInterval: updateInterval,
		done:           make(chan bool),
		batchSize:      10, // Process 10 students at a time
	}
}

// calculateNextWeeklyUpdate returns the duration until the next weekly update
// Updates are scheduled for Sunday at 00:00 UTC
func (u *ContestHistoryUpdater) calculateNextWeeklyUpdate() time.Duration {
	now := time.Now().UTC()
	nextSunday := now

	// Find the next Sunday
	daysUntilSunday := time.Sunday - now.Weekday()
	if daysUntilSunday <= 0 {
		daysUntilSunday += 7 // If today is Sunday, go to next Sunday
	}

	nextSunday = time.Date(
		now.Year(),
		now.Month(),
		now.Day()+int(daysUntilSunday),
		0, // Hour
		0, // Minute
		0, // Second
		0, // Nanosecond
		time.UTC,
	)

	return nextSunday.Sub(now)
}

func (u *ContestHistoryUpdater) Start() {
	// Calculate the initial delay to start at the next weekly interval
	initialDelay := u.calculateNextWeeklyUpdate()
	u.logger.Info("Contest history updater scheduled",
		zap.Duration("initial_delay", initialDelay),
		zap.Duration("interval", u.updateInterval))

	// Create a timer for the initial delay
	initialTimer := time.NewTimer(initialDelay)

	u.wg.Add(1)
	go func() {
		defer u.wg.Done()

		// Wait for the initial delay
		select {
		case <-u.done:
			initialTimer.Stop()
			u.logger.Info("Contest history updater stopped before first update")
			return
		case <-initialTimer.C:
			// Initial delay completed, start the regular updates
			u.ticker = time.NewTicker(u.updateInterval)

			// Perform the first update immediately after the initial delay
			if err := u.updateAllStudents(); err != nil {
				u.logger.Error("Failed to update contest histories", zap.Error(err))
			}

			// Continue with regular updates
			for {
				select {
				case <-u.done:
					u.ticker.Stop()
					u.logger.Info("Contest history updater stopped")
					return
				case <-u.ticker.C:
					if err := u.updateAllStudents(); err != nil {
						u.logger.Error("Failed to update contest histories", zap.Error(err))
					}
				}
			}
		}
	}()
}

func (u *ContestHistoryUpdater) Stop() {
	u.ticker.Stop()
	u.done <- true
	u.wg.Wait()
}

func (u *ContestHistoryUpdater) updateAllStudents() error {
	page := 1
	leetcodeService := leetcode.NewService()

	for {
		students, err := u.repo.List(context.Background(), page, u.batchSize)
		if err != nil {
			return fmt.Errorf("failed to fetch students: %w", err)
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			if err := u.updateStudentContestHistory(&student, leetcodeService); err != nil {
				u.logger.Error("Failed to update student contest history",
					zap.String("student_id", student.StudentID),
					zap.Error(err))
				continue // Continue with next student even if one fails
			}
			// Add a small delay between students to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}

		page++
	}

	return nil
}

func (u *ContestHistoryUpdater) updateStudentContestHistory(student *models.Student, leetcodeService *leetcode.Service) error {
	ctx := context.Background()

	// Get fresh contest stats
	contestStats, err := leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		return fmt.Errorf("failed to get contest stats: %w", err)
	}

	// Delete existing contest history
	if err := u.repo.DeleteContestHistory(ctx, student.ID); err != nil {
		return fmt.Errorf("failed to delete existing contest history: %w", err)
	}

	now := time.Now()
	var histories []*models.ContestHistory

	// Create new contest history entries
	for _, contest := range contestStats.Data.UserContestRankingHistory {
		history := &models.ContestHistory{
			StudentID:         student.ID,
			ContestTitle:      contest.Contest.Title,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: contest.FinishTimeInSeconds,
			ContestDate:       now,
			CreatedAt:         now,
		}
		histories = append(histories, history)
	}

	// Add new contest histories
	if err := u.repo.AddContestHistories(ctx, student.ID, histories); err != nil {
		return fmt.Errorf("failed to add contest histories: %w", err)
	}

	u.logger.Info("Updated student contest history",
		zap.String("student_id", student.StudentID),
		zap.Int("contests_count", len(histories)))

	return nil
}

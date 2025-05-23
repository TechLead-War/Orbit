package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/repository"
	"go.uber.org/zap"
)

type RatingUpdater struct {
	repo           *repository.StudentRepository
	logger         *zap.Logger
	ticker         *time.Ticker
	updateInterval time.Duration
	done           chan bool
	batchSize      int
	wg             sync.WaitGroup
}

func NewRatingUpdater(repo *repository.StudentRepository, logger *zap.Logger, updateInterval time.Duration) *RatingUpdater {
	return &RatingUpdater{
		repo:           repo,
		logger:         logger,
		updateInterval: updateInterval,
		done:           make(chan bool),
		batchSize:      10, // Process 10 students at a time to avoid overwhelming LeetCode API
	}
}

// calculateNextWeeklyUpdate returns the duration until the next weekly update
// Updates are scheduled for Sunday at 00:00 UTC
func (r *RatingUpdater) calculateNextWeeklyUpdate() time.Duration {
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

func (r *RatingUpdater) Start() {
	// Calculate the initial delay to start at the next weekly interval
	initialDelay := r.calculateNextWeeklyUpdate()
	r.logger.Info("Rating updater scheduled",
		zap.Duration("initial_delay", initialDelay),
		zap.Duration("interval", r.updateInterval))

	// Create a timer for the initial delay
	initialTimer := time.NewTimer(initialDelay)

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// Wait for the initial delay
		select {
		case <-r.done:
			initialTimer.Stop()
			r.logger.Info("Rating updater stopped before first update")
			return
		case <-initialTimer.C:
			// Initial delay completed, start the regular updates
			r.ticker = time.NewTicker(r.updateInterval)

			// Perform the first update immediately after the initial delay
			if err := r.updateAllStudents(); err != nil {
				r.logger.Error("Failed to update student ratings", zap.Error(err))
			}

			// Continue with regular updates
			for {
				select {
				case <-r.done:
					r.ticker.Stop()
					r.logger.Info("Rating updater stopped")
					return
				case <-r.ticker.C:
					if err := r.updateAllStudents(); err != nil {
						r.logger.Error("Failed to update student ratings", zap.Error(err))
					}
				}
			}
		}
	}()
}

func (r *RatingUpdater) Stop() {
	r.ticker.Stop()
	r.done <- true
	r.wg.Wait()
}

func (r *RatingUpdater) updateAllStudents() error {
	page := 1
	for {
		students, err := r.repo.List(context.Background(), page, r.batchSize)
		if err != nil {
			return fmt.Errorf("failed to fetch students: %w", err)
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			if err := r.updateStudentRating(&student); err != nil {
				r.logger.Error("Failed to update student rating",
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

func (r *RatingUpdater) updateStudentRating(student *models.Student) error {
	ctx := context.Background()

	// Get fresh LeetCode stats
	leetcodeStats, err := r.repo.GetLeetCodeStats(ctx, student.LeetcodeID)
	if err != nil {
		return fmt.Errorf("failed to get LeetCode stats: %w", err)
	}

	now := time.Now()
	rating := &models.Rating{
		StudentID:     student.ID,
		ProblemsCount: leetcodeStats.TotalSolved,
		EasyCount:     leetcodeStats.EasySolved,
		MediumCount:   leetcodeStats.MediumSolved,
		HardCount:     leetcodeStats.HardSolved,
		GlobalRank:    leetcodeStats.ContestRanking,
		RecordedAt:    now,
		CreatedAt:     now,
	}

	// Calculate rating using the formula: Easy(x1) + Medium(x3) + Hard(x5) + 20% of contest rating
	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(leetcodeStats.ContestRating * 0.2)
	rating.Rating = problemRating + contestBonus

	if err := r.repo.AddRating(ctx, rating); err != nil {
		return fmt.Errorf("failed to add rating: %w", err)
	}

	r.logger.Info("Updated student rating",
		zap.String("student_id", student.StudentID),
		zap.Int("rating", rating.Rating),
		zap.Int("problems", rating.ProblemsCount))

	return nil
}

package database

import (
	"fmt"

	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{db: db}
}

func (d *Database) CreateStudent(student *models.Student) error {
	return d.db.Create(student).Error
}

func (d *Database) CreateFileUpload(upload *models.FileUpload) error {
	// Ensure uploaded_by is nil if it's 0
	if upload.UploadedBy != nil && *upload.UploadedBy == 0 {
		upload.UploadedBy = nil
	}
	return d.db.Create(upload).Error
}

func (d *Database) UpdateFileUpload(upload *models.FileUpload) error {
	return d.db.Save(upload).Error
}

func (d *Database) GetStudent(id uint) (*models.Student, error) {
	var student models.Student
	err := d.db.First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (d *Database) ListStudents(page, pageSize int) ([]*models.Student, error) {
	var students []*models.Student
	offset := (page - 1) * pageSize

	// Query with preloaded ratings and order by creation date
	err := d.db.Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC")
	}).Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}

func (d *Database) UpdateStudentRating(studentID uint, rating *models.Rating) error {
	// Ensure the rating is associated with the correct student
	rating.StudentID = studentID

	// Create a new rating record
	result := d.db.Create(rating)
	if result.Error != nil {
		return fmt.Errorf("failed to create rating: %w", result.Error)
	}

	// Verify the rating was created
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rating record was created")
	}

	return nil
}

func (d *Database) GetStudentStats(studentID uint) (*models.StudentStats, error) {
	var stats models.StudentStats
	stats.StudentID = studentID

	// Get latest rating
	var rating models.Rating
	err := d.db.Where("student_id = ?", studentID).Order("recorded_at desc").First(&rating).Error
	if err != nil {
		return nil, err
	}

	stats.CurrentRating = rating.Rating
	stats.ProblemsCount = rating.ProblemsCount
	stats.GlobalRank = rating.GlobalRank

	// Get contest stats
	var contestStats models.ContestStats
	err = d.db.Model(&models.ContestHistory{}).
		Select("COUNT(*) as contests_participated, "+
			"AVG(rating) as average_rating, "+
			"MIN(ranking) as best_ranking, "+
			"SUM(problems_solved) as total_problems_solved").
		Where("student_id = ?", studentID).
		Scan(&contestStats).Error
	if err != nil {
		return nil, err
	}

	stats.ContestRating = contestStats.AverageRating
	stats.ContestsParticipated = contestStats.ContestsParticipated

	return &stats, nil
}

func (d *Database) GetLeetCodeStats(leetcodeID string) (*models.LeetCodeStats, error) {
	// Create a LeetCode service instance
	leetcodeService := leetcode.NewService()

	// Get user profile from LeetCode API
	profile, err := leetcodeService.GetUserProfile(leetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LeetCode profile: %w", err)
	}

	// Get contest ranking
	contestInfo, err := leetcodeService.GetContestRanking(leetcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}

	// Map the data to our stats model
	stats := &models.LeetCodeStats{
		TotalSolved:    0,
		EasySolved:     0,
		MediumSolved:   0,
		HardSolved:     0,
		ContestRating:  contestInfo.Data.UserContestRanking.Rating,
		ContestRanking: contestInfo.Data.UserContestRanking.GlobalRanking,
	}

	// Map submission stats
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
	stats.TotalSolved = stats.EasySolved + stats.MediumSolved + stats.HardSolved

	return stats, nil
}

func (d *Database) GetContestRankings(studentID uint) ([]*models.ContestHistory, error) {
	var rankings []*models.ContestHistory
	err := d.db.Where("student_id = ?", studentID).
		Order("contest_date desc").
		Find(&rankings).Error
	if err != nil {
		return nil, err
	}
	return rankings, nil
}

func (d *Database) GetStudentWithStats(id uint) (*models.Student, error) {
	var student models.Student

	// Get student with latest rating
	err := d.db.Preload("Ratings", func(db *gorm.DB) *gorm.DB {
		return db.Order("recorded_at DESC").Limit(1)
	}).First(&student, id).Error
	if err != nil {
		return nil, err
	}

	// Get contest stats
	var contestStats models.ContestStats
	err = d.db.Model(&models.ContestHistory{}).
		Select("COUNT(*) as contests_participated, "+
			"AVG(rating) as average_rating, "+
			"MIN(ranking) as best_ranking, "+
			"SUM(problems_solved) as total_problems_solved").
		Where("student_id = ?", id).
		Scan(&contestStats).Error
	if err == nil {
		student.ContestStats = &contestStats
	}

	// Get contest history
	var contestHistory []*models.ContestHistory
	err = d.db.Where("student_id = ?", id).
		Order("contest_date desc").
		Find(&contestHistory).Error
	if err == nil {
		student.ContestHistory = contestHistory
	}

	return &student, nil
}

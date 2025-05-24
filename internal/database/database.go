package database

import (
	"context"
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
		// Convert []*models.ContestHistory to []models.ContestHistory
		var ch []models.ContestHistory
		for _, v := range contestHistory {
			ch = append(ch, *v)
		}
		student.ContestHistory = ch
	}

	return &student, nil
}

func (d *Database) DeleteContestHistory(studentID uint) error {
	return d.db.Where("student_id = ?", studentID).Delete(&models.ContestHistory{}).Error
}

func (d *Database) AddContestHistories(studentID uint, histories []*models.ContestHistory) error {
	// Use a transaction to ensure all histories are added or none
	return d.db.Transaction(func(tx *gorm.DB) error {
		for _, history := range histories {
			history.StudentID = studentID
			if err := tx.Create(history).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type FileUploadDB struct {
	db *gorm.DB
}

func NewFileUploadDB(db *gorm.DB) *FileUploadDB {
	return &FileUploadDB{db: db}
}

func (d *FileUploadDB) Create(ctx context.Context, upload *models.FileUpload) error {
	return d.db.WithContext(ctx).Create(upload).Error
}

func (d *FileUploadDB) GetByID(ctx context.Context, id uint) (*models.FileUpload, error) {
	var upload models.FileUpload
	if err := d.db.WithContext(ctx).First(&upload, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get file upload: %w", err)
	}
	return &upload, nil
}

func (d *FileUploadDB) Update(ctx context.Context, upload *models.FileUpload) error {
	return d.db.WithContext(ctx).Save(upload).Error
}

func (d *FileUploadDB) Delete(ctx context.Context, id uint) error {
	return d.db.WithContext(ctx).Delete(&models.FileUpload{}, id).Error
}

type BatchStatsDB struct {
	db *gorm.DB
}

func NewBatchStatsDB(db *gorm.DB) *BatchStatsDB {
	return &BatchStatsDB{db: db}
}

func (d *BatchStatsDB) Create(ctx context.Context, stats *models.BatchStats) error {
	return d.db.WithContext(ctx).Create(stats).Error
}

func (d *BatchStatsDB) GetByBatch(ctx context.Context, batch string) (*models.BatchStats, error) {
	var stats models.BatchStats
	if err := d.db.WithContext(ctx).Where("batch = ?", batch).First(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get batch stats: %w", err)
	}
	return &stats, nil
}

func (d *BatchStatsDB) Update(ctx context.Context, stats *models.BatchStats) error {
	return d.db.WithContext(ctx).Save(stats).Error
}

func (d *BatchStatsDB) Delete(ctx context.Context, batch string) error {
	return d.db.WithContext(ctx).Where("batch = ?", batch).Delete(&models.BatchStats{}).Error
}

type DepartmentStatsDB struct {
	db *gorm.DB
}

func NewDepartmentStatsDB(db *gorm.DB) *DepartmentStatsDB {
	return &DepartmentStatsDB{db: db}
}

func (d *DepartmentStatsDB) Create(ctx context.Context, stats *models.DepartmentStats) error {
	return d.db.WithContext(ctx).Create(stats).Error
}

func (d *DepartmentStatsDB) GetByDepartment(ctx context.Context, department string) (*models.DepartmentStats, error) {
	var stats models.DepartmentStats
	if err := d.db.WithContext(ctx).Where("department = ?", department).First(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get department stats: %w", err)
	}
	return &stats, nil
}

func (d *DepartmentStatsDB) Update(ctx context.Context, stats *models.DepartmentStats) error {
	return d.db.WithContext(ctx).Save(stats).Error
}

func (d *DepartmentStatsDB) Delete(ctx context.Context, department string) error {
	return d.db.WithContext(ctx).Where("department = ?", department).Delete(&models.DepartmentStats{}).Error
}

type SystemStatsDB struct {
	db *gorm.DB
}

func NewSystemStatsDB(db *gorm.DB) *SystemStatsDB {
	return &SystemStatsDB{db: db}
}

func (d *SystemStatsDB) Create(ctx context.Context, stats *models.SystemStats) error {
	return d.db.WithContext(ctx).Create(stats).Error
}

func (d *SystemStatsDB) GetLatest(ctx context.Context) (*models.SystemStats, error) {
	var stats models.SystemStats
	if err := d.db.WithContext(ctx).Order("created_at DESC").First(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get system stats: %w", err)
	}
	return &stats, nil
}

func (d *SystemStatsDB) Update(ctx context.Context, stats *models.SystemStats) error {
	return d.db.WithContext(ctx).Save(stats).Error
}

func (d *SystemStatsDB) Delete(ctx context.Context, id uint) error {
	return d.db.WithContext(ctx).Delete(&models.SystemStats{}, id).Error
}

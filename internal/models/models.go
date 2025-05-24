package models

import (
	"time"
)

// Student represents a student in the system
type Student struct {
	ID             uint             `json:"id" gorm:"primaryKey"`
	StudentID      string           `json:"student_id" gorm:"uniqueIndex"`
	Name           string           `json:"name"`
	Email          string           `json:"email" gorm:"uniqueIndex"`
	LeetcodeID     string           `json:"leetcode_id" gorm:"uniqueIndex"`
	PassingYear    int              `json:"passing_year"`
	Batch          string           `json:"batch"`
	Department     string           `json:"department"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	LeetCodeStats  *LeetCodeStats   `json:"leetcode_stats,omitempty" gorm:"foreignKey:StudentID"`
	Ratings        []Rating         `json:"ratings,omitempty" gorm:"foreignKey:StudentID"`
	ContestHistory []ContestHistory `json:"contest_history,omitempty" gorm:"foreignKey:StudentID"`
	ContestStats   *ContestStats    `json:"contest_stats,omitempty" gorm:"-"`
}

// LeetCodeStats represents a student's LeetCode statistics
type LeetCodeStats struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	StudentID              uint      `json:"student_id" gorm:"uniqueIndex"`
	Rating                 int       `json:"rating"`
	GlobalRanking          int       `json:"global_ranking"`
	TotalSolved            int       `json:"total_solved"`
	EasySolved             int       `json:"easy_solved"`
	MediumSolved           int       `json:"medium_solved"`
	HardSolved             int       `json:"hard_solved"`
	ContestsParticipated   int       `json:"contests_participated"`
	ContestRating          float64   `json:"contest_rating"`
	ContestGlobalRanking   int       `json:"contest_global_ranking"`
	ContestRanking         int       `json:"contest_ranking"`
	Reputation             int       `json:"reputation"`
	ContributionPoints     int       `json:"contribution_points"`
	LastSolvedAt           time.Time `json:"last_solved_at"`
	InitialRating          int       `json:"initial_rating"`
	ProblemsSolvedLastWeek int       `json:"problems_solved_last_week"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Rating represents a student's rating at a point in time
type Rating struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	StudentID     uint      `json:"student_id"`
	Rating        int       `json:"rating"`
	ProblemsCount int       `json:"problems_count"`
	EasyCount     int       `json:"easy_count"`
	MediumCount   int       `json:"medium_count"`
	HardCount     int       `json:"hard_count"`
	GlobalRank    int       `json:"global_rank"`
	RecordedAt    time.Time `json:"recorded_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// ContestHistory represents a student's contest participation history
type ContestHistory struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	StudentID         uint      `json:"student_id"`
	ContestTitle      string    `json:"contest_title"`
	Rating            float64   `json:"rating"`
	Ranking           int       `json:"ranking"`
	ProblemsSolved    int       `json:"problems_solved"`
	FinishTimeSeconds int64     `json:"finish_time_seconds"`
	ContestDate       time.Time `json:"contest_date"`
	CreatedAt         time.Time `json:"created_at"`
}

// ContestStats represents aggregated contest statistics for a student
// Used for analytics and reporting
// This struct is referenced in database logic
type ContestStats struct {
	ContestsParticipated int     `json:"contests_participated"`
	AverageRating        float64 `json:"average_rating"`
	BestRanking          int     `json:"best_ranking"`
	TotalProblemsSolved  int     `json:"total_problems_solved"`
}

// DailyProgress represents a student's daily problem-solving progress
type DailyProgress struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	StudentID      uint      `json:"student_id"`
	Date           time.Time `json:"date"`
	ProblemsSolved int       `json:"problems_solved"`
	EasySolved     int       `json:"easy_solved"`
	MediumSolved   int       `json:"medium_solved"`
	HardSolved     int       `json:"hard_solved"`
	TimeSpent      int       `json:"time_spent"` // in minutes
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WeeklyStats represents a student's weekly statistics
type WeeklyStats struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	StudentID        uint      `json:"student_id"`
	WeekStart        time.Time `json:"week_start"`
	WeekEnd          time.Time `json:"week_end"`
	ProblemsSolved   int       `json:"problems_solved"`
	EasySolved       int       `json:"easy_solved"`
	MediumSolved     int       `json:"medium_solved"`
	HardSolved       int       `json:"hard_solved"`
	TimeSpent        int       `json:"time_spent"` // in minutes
	ContestsAttended int       `json:"contests_attended"`
	AverageRating    float64   `json:"average_rating"`
	ContestRating    float64   `json:"contest_rating"`
	GlobalRanking    int       `json:"global_ranking"`
	ProblemsCount    int       `json:"problems_count"`
	EasyCount        int       `json:"easy_count"`
	MediumCount      int       `json:"medium_count"`
	HardCount        int       `json:"hard_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StudentStats represents aggregated statistics for a student
type StudentStats struct {
	StudentID            uint             `json:"student_id"`
	TotalProblemsSolved  int              `json:"total_problems_solved"`
	EasySolved           int              `json:"easy_solved"`
	MediumSolved         int              `json:"medium_solved"`
	HardSolved           int              `json:"hard_solved"`
	ContestsParticipated int              `json:"contests_participated"`
	AverageRating        float64          `json:"average_rating"`
	HighestRating        int              `json:"highest_rating"`
	CurrentRating        int              `json:"current_rating"`
	ProblemsCount        int              `json:"problems_count"`
	GlobalRank           int              `json:"global_rank"`
	ContestRating        float64          `json:"contest_rating"`
	LastActive           time.Time        `json:"last_active"`
	WeeklyProgress       []WeeklyStats    `json:"weekly_progress"`
	ContestHistory       []ContestHistory `json:"contest_history"`
}

// FileUpload represents a file upload record
type FileUpload struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	FileName          string    `json:"file_name"`
	OriginalName      string    `json:"original_name"`
	FileType          string    `json:"file_type"`
	FileSize          int64     `json:"file_size"`
	StoragePath       string    `json:"storage_path"`
	Status            string    `json:"status"`
	ProcessedAt       time.Time `json:"processed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	UploadedBy        *uint     `json:"uploaded_by"`
	TotalRecords      int       `json:"total_records"`
	SuccessRecords    int       `json:"success_records"`
	SuccessfulRecords int       `json:"successful_records"` // Adding this field
	FailedRecords     int       `json:"failed_records"`
	ErrorDetails      *string   `json:"error_details"`
}

// BatchStats represents statistics for a batch of students
type BatchStats struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	Batch               string    `json:"batch"`
	TotalStudents       int       `json:"total_students"`
	ActiveStudents      int       `json:"active_students"`
	AverageRating       float64   `json:"average_rating"`
	HighestRating       int       `json:"highest_rating"`
	TotalProblemsSolved int       `json:"total_problems_solved"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// DepartmentStats represents statistics for a department
type DepartmentStats struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	Department          string    `json:"department"`
	TotalStudents       int       `json:"total_students"`
	ActiveStudents      int       `json:"active_students"`
	AverageRating       float64   `json:"average_rating"`
	HighestRating       int       `json:"highest_rating"`
	TotalProblemsSolved int       `json:"total_problems_solved"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	TotalStudents        int       `json:"total_students"`
	ActiveStudents       int       `json:"active_students"`
	TotalProblemsSolved  int       `json:"total_problems_solved"`
	TotalContests        int       `json:"total_contests"`
	AverageStudentRating float64   `json:"average_student_rating"`
	HighestStudentRating int       `json:"highest_student_rating"`
	LastUpdated          time.Time `json:"last_updated"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateStudentRequest represents the request body for creating a new student
type CreateStudentRequest struct {
	StudentID   string `json:"student_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	LeetcodeID  string `json:"leetcode_id" binding:"required"`
	PassingYear int    `json:"passing_year" binding:"required"`
	Batch       string `json:"batch" binding:"required"`
	Department  string `json:"department" binding:"required"`
}

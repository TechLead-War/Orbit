package models

import (
	"time"
)

type Student struct {
	ID             uint              `json:"id" db:"id"`
	StudentID      string            `json:"student_id" db:"student_id"`
	Name           string            `json:"name" db:"name"`
	Email          string            `json:"email" db:"email"`
	LeetcodeID     string            `json:"leetcode_id" db:"leetcode_id"`
	PassingYear    int               `json:"passing_year" db:"passing_year"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	Ratings        []Rating          `json:"ratings,omitempty" gorm:"foreignKey:StudentID"`
	ContestStats   *ContestStats     `json:"contest_stats,omitempty" gorm:"-"`
	LeetCodeStats  *LeetCodeStats    `json:"leetcode_stats,omitempty" gorm:"-"`
	ContestHistory []*ContestHistory `json:"contest_history,omitempty" gorm:"foreignKey:StudentID"`
}

type StudentStats struct {
	StudentID            uint    `json:"student_id"`
	CurrentRating        int     `json:"current_rating"`
	ProblemsCount        int     `json:"problems_count"`
	ContestRating        float64 `json:"contest_rating"`
	GlobalRank           int     `json:"global_rank"`
	ContestsParticipated int     `json:"contests_participated"`
}

type CreateStudentRequest struct {
	StudentID   string `json:"student_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	LeetcodeID  string `json:"leetcode_id" binding:"required"`
	PassingYear int    `json:"passing_year" binding:"required"`
}

type FileUpload struct {
	ID                uint      `json:"id" db:"id"`
	FileName          string    `json:"file_name" db:"file_name"`
	OriginalName      string    `json:"original_name" db:"original_name"`
	FileType          string    `json:"file_type" db:"file_type"`
	FileSize          int64     `json:"file_size" db:"file_size"`
	StoragePath       string    `json:"storage_path" db:"storage_path"`
	TotalRecords      int       `json:"total_records" db:"total_records"`
	SuccessfulRecords int       `json:"successful_records" db:"successful_records"`
	FailedRecords     int       `json:"failed_records" db:"failed_records"`
	Status            string    `json:"status" db:"status"`
	ErrorDetails      *string   `json:"error_details,omitempty" db:"error_details"`
	UploadedBy        *uint     `json:"uploaded_by,omitempty" db:"uploaded_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type FileUploadResponse struct {
	ID                uint      `json:"id"`
	FileName          string    `json:"file_name"`
	TotalRecords      int       `json:"total_records"`
	SuccessfulRecords int       `json:"successful_records"`
	FailedRecords     int       `json:"failed_records"`
	Status            string    `json:"status"`
	ErrorDetails      *string   `json:"error_details,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

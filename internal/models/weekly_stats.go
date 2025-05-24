package models

import "time"

// WeeklyStats represents a student's LeetCode statistics for a specific week
type WeeklyStats struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	StudentID     string    `json:"student_id" gorm:"index"`
	WeekStartDate time.Time `json:"week_start_date"`
	WeekEndDate   time.Time `json:"week_end_date"`
	EasyCount     int       `json:"easy_count"`
	MediumCount   int       `json:"medium_count"`
	HardCount     int       `json:"hard_count"`
	TotalCount    int       `json:"total_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Student *Student `json:"-" gorm:"foreignKey:StudentID;references:StudentID"`
}

// CreateWeeklyStatsRequest represents the request body for creating weekly stats
type CreateWeeklyStatsRequest struct {
	StudentID     string    `json:"student_id,omitempty"`
	WeekStartDate time.Time `json:"week_start_date" binding:"required"`
	WeekEndDate   time.Time `json:"week_end_date" binding:"required"`
	EasyCount     int       `json:"easy_count"`
	MediumCount   int       `json:"medium_count"`
	HardCount     int       `json:"hard_count"`
	TotalCount    int       `json:"total_count"`
}

package models

import (
	"time"
)

// Student represents a student in the system
type Student struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"`
	Email      string    `json:"email" gorm:"unique"`
	LeetcodeID string    `json:"leetcode_id" gorm:"unique;column:leetcode_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Ratings    []Rating  `json:"ratings,omitempty" gorm:"foreignKey:StudentID"`
}

// Rating represents a student's LeetCode rating at a point in time
type Rating struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	StudentID     uint      `json:"student_id"`
	Rating        int       `json:"rating"`
	ProblemsCount int       `json:"problems_count"`
	EasyCount     int       `json:"easy_count"`
	MediumCount   int       `json:"medium_count"`
	HardCount     int       `json:"hard_count"`
	ContestRating int       `json:"contest_rating"`
	GlobalRank    int       `json:"global_rank"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// StudentStats represents aggregated statistics for a student
type StudentStats struct {
	StudentID          uint    `json:"student_id"`
	AverageRating      float64 `json:"average_rating"`
	RatingChange30Days int     `json:"rating_change_30_days"`
	ProblemsLast30Days int     `json:"problems_last_30_days"`
	TotalProblems      int     `json:"total_problems"`
	CurrentRank        int     `json:"current_rank"`
}

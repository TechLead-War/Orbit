package models

import (
	"time"
)

// Student represents a student in the system
type Student struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	Name         string        `json:"name"`
	Email        string        `json:"email" gorm:"unique"`
	LeetcodeID   string        `json:"leetcode_id" gorm:"unique"`
	CreatedAt    time.Time     `json:"created_at"`
	Ratings      []Rating      `json:"ratings,omitempty" gorm:"foreignKey:StudentID"`
	ContestStats *ContestStats `json:"contest_stats,omitempty" gorm:"-"`
}

// StudentStats represents aggregated statistics for a student
type StudentStats struct {
	StudentID            uint    `json:"student_id"`
	CurrentRating        int     `json:"current_rating"`
	ProblemsCount        int     `json:"problems_count"`
	ContestRating        float64 `json:"contest_rating"`
	GlobalRank           int     `json:"global_rank"`
	ContestsParticipated int     `json:"contests_participated"`
}

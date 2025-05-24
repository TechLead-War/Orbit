package models

import (
	"time"
)

type Student struct {
	ID             uint              `json:"id" gorm:"primaryKey"`
	StudentID      string            `json:"student_id" gorm:"uniqueIndex"`
	Name           string            `json:"name" db:"name"`
	Email          string            `json:"email" db:"email"`
	LeetcodeID     string            `json:"leetcode_id" db:"leetcode_id"`
	PassingYear    int               `json:"passing_year" db:"passing_year"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	Ratings        []Rating          `json:"ratings,omitempty" gorm:"foreignKey:StudentID;references:ID"`
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

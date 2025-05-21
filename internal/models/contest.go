package models

import "time"

// ContestHistory represents a student's performance in a specific LeetCode contest
type ContestHistory struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	StudentID         uint      `json:"student_id"`
	ContestTitle      string    `json:"contest_title"`
	Rating            float64   `json:"rating"`
	Ranking           int       `json:"ranking"`
	ProblemsSolved    int       `json:"problems_solved"`
	FinishTimeSeconds int       `json:"finish_time_seconds"`
	ContestDate       time.Time `json:"contest_date"`
	CreatedAt         time.Time `json:"created_at"`
}

// ContestStats represents aggregated contest statistics for a student
type ContestStats struct {
	ContestsParticipated int     `json:"contests_participated"`
	AverageRating        float64 `json:"average_rating"`
	BestRanking          int     `json:"best_ranking"`
	TotalProblemsSolved  int     `json:"total_problems_solved"`
}

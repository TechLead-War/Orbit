package models

// LeetCodeStats represents a user's LeetCode statistics
type LeetCodeStats struct {
	TotalSolved        int     `json:"total_solved"`
	EasySolved         int     `json:"easy_solved"`
	MediumSolved       int     `json:"medium_solved"`
	HardSolved         int     `json:"hard_solved"`
	AcceptanceRate     float64 `json:"acceptance_rate"`
	ContestRating      float64 `json:"contest_rating"`
	ContestRanking     int     `json:"contest_ranking"`
	ContributionPoints int     `json:"contribution_points"`
}

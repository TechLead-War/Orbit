package models

import "time"

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

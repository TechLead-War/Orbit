package models

import "time"

// Rating represents a student's LeetCode performance snapshot
type Rating struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	StudentID     uint      `json:"student_id" gorm:"index:idx_student_recorded,unique:student_recorded"`
	Rating        int       `json:"rating" gorm:"comment:Composite score (problem_rating + contest_bonus). Problem rating: Easy(x1) + Medium(x3) + Hard(x5). Contest bonus: 20% of LeetCode contest rating"`
	ProblemsCount int       `json:"problems_count" gorm:"comment:Total number of problems solved"`
	EasyCount     int       `json:"easy_count" gorm:"comment:Number of easy problems solved"`
	MediumCount   int       `json:"medium_count" gorm:"comment:Number of medium problems solved"`
	HardCount     int       `json:"hard_count" gorm:"comment:Number of hard problems solved"`
	GlobalRank    int       `json:"global_rank" gorm:"comment:Student's position in global LeetCode rankings (lower is better)"`
	RecordedAt    time.Time `json:"recorded_at" gorm:"index:idx_student_recorded,unique:student_recorded;not null;comment:When this rating snapshot was taken"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CalculateRating computes the composite rating score
func (r *Rating) CalculateRating(contestRating float64) {
	// Problem solving rating
	problemRating := (r.EasyCount * 1) + (r.MediumCount * 3) + (r.HardCount * 5)

	// Contest performance bonus (20% of contest rating)
	contestBonus := int(contestRating * 0.2)

	// Final rating
	r.Rating = problemRating + contestBonus
}

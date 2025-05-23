package models

import "time"

type Student struct {
	ID          int       `json:"id" db:"id"`
	StudentID   string    `json:"student_id" db:"student_id"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email" db:"email"`
	LeetcodeID  string    `json:"leetcode_id" db:"leetcode_id"`
	PassingYear int       `json:"passing_year" db:"passing_year"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateStudentRequest struct {
	StudentID   string `json:"student_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	LeetcodeID  string `json:"leetcode_id" binding:"required"`
	PassingYear int    `json:"passing_year" binding:"required"`
}

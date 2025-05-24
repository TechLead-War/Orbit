package models

import (
	"time"
)

// FileUpload represents a file upload record
type FileUpload struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	StoragePath  string    `json:"storage_path"`
	Status       string    `json:"status"`
	ErrorDetails *string   `json:"error_details,omitempty"`
	UploadedBy   *uint     `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	TotalRecords      int `json:"total_records"`
	SuccessfulRecords int `json:"successful_records"`
	FailedRecords     int `json:"failed_records"`
}

// FileUploadResponse represents the response for file upload operations
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

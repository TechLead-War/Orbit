package models

import "time"

type FileUpload struct {
	ID                int       `json:"id" db:"id"`
	FileName          string    `json:"file_name" db:"file_name"`
	OriginalName      string    `json:"original_name" db:"original_name"`
	FileType          string    `json:"file_type" db:"file_type"`
	FileSize          int64     `json:"file_size" db:"file_size"`
	StoragePath       string    `json:"storage_path" db:"storage_path"`
	TotalRecords      int       `json:"total_records" db:"total_records"`
	SuccessfulRecords int       `json:"successful_records" db:"successful_records"`
	FailedRecords     int       `json:"failed_records" db:"failed_records"`
	Status            string    `json:"status" db:"status"`
	ErrorDetails      *string   `json:"error_details" db:"error_details"`
	UploadedBy        int       `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type FileUploadResponse struct {
	ID                int       `json:"id"`
	FileName          string    `json:"file_name"`
	TotalRecords      int       `json:"total_records"`
	SuccessfulRecords int       `json:"successful_records"`
	FailedRecords     int       `json:"failed_records"`
	Status            string    `json:"status"`
	ErrorDetails      *string   `json:"error_details,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

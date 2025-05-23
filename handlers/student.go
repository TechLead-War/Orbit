package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"orbit/models"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type Handler struct {
	db Database
}

type Database interface {
	CreateStudent(student *models.Student) error
	CreateFileUpload(upload *models.FileUpload) error
	UpdateFileUpload(upload *models.FileUpload) error
}

const (
	uploadDir   = "./uploads"
	maxFileSize = 10 << 20 // 10MB
)

func NewHandler(db Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) CreateStudent(c *gin.Context) {
	var req models.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student := models.Student{
		StudentID:   req.StudentID,
		Name:        req.Name,
		Email:       req.Email,
		LeetcodeID:  req.LeetcodeID,
		PassingYear: req.PassingYear,
	}

	if err := h.db.CreateStudent(&student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, student)
}

func (h *Handler) BulkCreateStudents(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".xlsx" && ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Only .xlsx and .csv files are allowed"})
		return
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s%s", timestamp, strings.TrimSuffix(file.Filename, ext), ext)
	filepath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	fileUpload := &models.FileUpload{
		FileName:     filename,
		OriginalName: file.Filename,
		FileType:     ext[1:],
		FileSize:     file.Size,
		StoragePath:  filepath,
		Status:       "processing",
		UploadedBy:   1, // TODO: Get from authenticated user
	}

	if err := h.db.CreateFileUpload(fileUpload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record file upload"})
		return
	}

	go h.processUploadedFile(fileUpload)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "File uploaded successfully. Processing started.",
		"file_id": fileUpload.ID,
	})
}

func (h *Handler) processUploadedFile(fileUpload *models.FileUpload) {
	var records [][]string
	var err error

	if strings.HasSuffix(fileUpload.FileName, ".xlsx") {
		records, err = readExcelFile(fileUpload.StoragePath)
	} else {
		records, err = readCSVFile(fileUpload.StoragePath)
	}

	if err != nil {
		updateFileUploadStatus(h, fileUpload, "failed", err.Error(), 0, 0, 0)
		return
	}

	if len(records) < 2 {
		updateFileUploadStatus(h, fileUpload, "failed", "File is empty or has no data rows", 0, 0, 0)
		return
	}

	successful := 0
	failed := 0
	errors := make([]string, 0)

	for i, record := range records[1:] {
		if len(record) != 5 {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: Invalid number of columns", i+2))
			continue
		}

		student := models.Student{
			StudentID:   record[0],
			Name:        record[1],
			Email:       record[2],
			LeetcodeID:  record[3],
			PassingYear: parseInt(record[4]),
		}

		if err := h.db.CreateStudent(&student); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+2, err.Error()))
			continue
		}

		successful++
	}

	status := "completed"
	if failed > 0 {
		status = "completed_with_errors"
	}

	errorDetails := strings.Join(errors, "\n")
	updateFileUploadStatus(h, fileUpload, status, errorDetails, len(records)-1, successful, failed)
}

func readExcelFile(filepath string) ([][]string, error) {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in excel file")
	}

	return f.GetRows(sheets[0])
}

func readCSVFile(filepath string) ([][]string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func updateFileUploadStatus(h *Handler, fileUpload *models.FileUpload, status, errorDetails string, total, successful, failed int) {
	fileUpload.Status = status
	fileUpload.TotalRecords = total
	fileUpload.SuccessfulRecords = successful
	fileUpload.FailedRecords = failed
	if errorDetails != "" {
		fileUpload.ErrorDetails = &errorDetails
	}
	fileUpload.UpdatedAt = time.Now()

	if err := h.db.UpdateFileUpload(fileUpload); err != nil {
		log.Printf("Failed to update file upload status: %v", err)
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

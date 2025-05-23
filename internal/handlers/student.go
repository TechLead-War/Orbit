package handlers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ayush/ORBIT/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db Database
}

type Database interface {
	CreateStudent(student *models.Student) error
	CreateFileUpload(upload *models.FileUpload) error
	UpdateFileUpload(upload *models.FileUpload) error
	GetStudent(id uint) (*models.Student, error)
	GetStudentWithStats(id uint) (*models.Student, error)
	ListStudents(page, pageSize int) ([]*models.Student, error)
	UpdateStudentRating(studentID uint, rating *models.Rating) error
	GetLeetCodeStats(leetcodeID string) (*models.LeetCodeStats, error)
	GetContestRankings(studentID uint) ([]*models.ContestHistory, error)
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

	student := &models.Student{
		StudentID:   req.StudentID,
		Name:        req.Name,
		Email:       req.Email,
		LeetcodeID:  req.LeetcodeID,
		PassingYear: req.PassingYear,
	}

	if err := h.db.CreateStudent(student); err != nil {
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

		student := &models.Student{
			StudentID:   record[0],
			Name:        record[1],
			Email:       record[2],
			LeetcodeID:  record[3],
			PassingYear: parseInt(record[4]),
		}

		if err := h.db.CreateStudent(student); err != nil {
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

type StudentResponse struct {
	ID             uint      `json:"id"`
	StudentID      string    `json:"student_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	LeetcodeID     string    `json:"leetcode_id"`
	PassingYear    int       `json:"passing_year"`
	CreatedAt      time.Time `json:"created_at"`
	Rating         int       `json:"rating"`
	TotalProblems  int       `json:"total_problems"`
	EasyProblems   int       `json:"easy_problems"`
	MediumProblems int       `json:"medium_problems"`
	HardProblems   int       `json:"hard_problems"`
	GlobalRank     int       `json:"global_rank"`
}

func (h *Handler) ListStudents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	students, err := h.db.ListStudents(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform the response to include rating details
	response := make([]StudentResponse, 0, len(students))
	for _, student := range students {
		studentResp := StudentResponse{
			ID:          student.ID,
			StudentID:   student.StudentID,
			Name:        student.Name,
			Email:       student.Email,
			LeetcodeID:  student.LeetcodeID,
			PassingYear: student.PassingYear,
			CreatedAt:   student.CreatedAt,
		}

		// Add rating details if available
		if len(student.Ratings) > 0 {
			latestRating := student.Ratings[len(student.Ratings)-1] // Get the latest rating
			studentResp.Rating = latestRating.Rating
			studentResp.TotalProblems = latestRating.ProblemsCount
			studentResp.EasyProblems = latestRating.EasyCount
			studentResp.MediumProblems = latestRating.MediumCount
			studentResp.HardProblems = latestRating.HardCount
			studentResp.GlobalRank = latestRating.GlobalRank
		}

		response = append(response, studentResp)
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"students":  response,
	})
}

func (h *Handler) GetStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Get student with all stats
	student, err := h.db.GetStudentWithStats(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student not found or no ratings available",
				"details": "The student either does not exist or has no LeetCode ratings recorded yet. " +
					"Try updating their rating first using PUT /api/v1/students/:id/rating",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get LeetCode stats
	leetcodeStats, err := h.db.GetLeetCodeStats(student.LeetcodeID)
	if err == nil && leetcodeStats != nil {
		student.LeetCodeStats = leetcodeStats
	}

	// Get contest rankings
	contestRankings, err := h.db.GetContestRankings(uint(id))
	if err == nil && contestRankings != nil {
		student.ContestHistory = contestRankings
	}

	c.JSON(http.StatusOK, student)
}

func (h *Handler) UpdateStudentRating(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Get student to get their LeetCode ID
	student, err := h.db.GetStudent(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get fresh LeetCode stats
	leetcodeStats, err := h.db.GetLeetCodeStats(student.LeetcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch LeetCode stats",
			"details": "Could not retrieve data from LeetCode API. This could be due to:" +
				"\n- Invalid LeetCode username" +
				"\n- LeetCode API rate limiting" +
				"\n- LeetCode API being unavailable",
			"technical_error": err.Error(),
		})
		return
	}

	now := time.Now()
	rating := &models.Rating{
		StudentID:     uint(id),
		ProblemsCount: leetcodeStats.TotalSolved,
		EasyCount:     leetcodeStats.EasySolved,
		MediumCount:   leetcodeStats.MediumSolved,
		HardCount:     leetcodeStats.HardSolved,
		GlobalRank:    leetcodeStats.ContestRanking,
		RecordedAt:    now,
		CreatedAt:     now,
	}

	// Calculate rating using the formula: Easy(x1) + Medium(x3) + Hard(x5) + 20% of contest rating
	problemRating := (rating.EasyCount * 1) + (rating.MediumCount * 3) + (rating.HardCount * 5)
	contestBonus := int(leetcodeStats.ContestRating * 0.2)
	rating.Rating = problemRating + contestBonus

	if err := h.db.UpdateStudentRating(uint(id), rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rating)
}

func (h *Handler) GetLeetCodeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	student, err := h.db.GetStudent(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.db.GetLeetCodeStats(student.LeetcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetContestRankings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	rankings, err := h.db.GetContestRankings(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rankings)
}

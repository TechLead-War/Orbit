package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"

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
	GetStudent(id uint) (*models.Student, error)
	GetStudentWithStats(id uint) (*models.Student, error)
	ListStudents(page, pageSize int) ([]*models.Student, error)
	UpdateStudentRating(studentID uint, rating *models.Rating) error
	GetLeetCodeStats(leetcodeID string) (*models.LeetCodeStats, error)
	GetContestRankings(studentID uint) ([]*models.ContestHistory, error)
	DeleteContestHistory(studentID uint) error
	AddContestHistories(studentID uint, histories []*models.ContestHistory) error
	GetStudentByID(studentID string) (*models.Student, error)
	GetStudentWeeklyStats(studentID string) ([]models.WeeklyStats, error)
	GetWeeklyStats(studentID string, start, end time.Time) (*models.WeeklyStats, error)
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
		UploadedBy:   nil,
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

// GetStudentDetails retrieves student details including weekly stats
func (h *Handler) GetStudentDetails(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	student, err := h.db.GetStudentByID(studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	// Get all weekly stats for the student
	weeklyStats, err := h.db.GetStudentWeeklyStats(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get weekly stats: %v", err)})
		return
	}

	// Get current week's stats
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)
	currentWeekStats, err := h.db.GetWeeklyStats(studentID, weekStart, weekEnd)
	if err != nil && err.Error() != "record not found" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get current week stats: %v", err)})
		return
	}

	response := gin.H{
		"student":      student,
		"weekly_stats": weeklyStats,
		"current_week": currentWeekStats,
	}

	c.JSON(http.StatusOK, response)
}

// GetAllStudents retrieves all students with optional pagination
func (h *Handler) GetAllStudents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "60"))

	students, err := h.db.ListStudents(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get students: %v", err)})
		return
	}

	// Transform the response to include rating details
	response := make([]gin.H, 0, len(students))
	for _, student := range students {
		studentData := gin.H{
			"student_id":   student.StudentID,
			"name":         student.Name,
			"email":        student.Email,
			"leetcode_id":  student.LeetcodeID,
			"passing_year": student.PassingYear,
			"created_at":   student.CreatedAt,
		}

		// Add rating details if available
		if len(student.Ratings) > 0 {
			latestRating := student.Ratings[0] // Get the latest rating (already ordered by recorded_at DESC)
			studentData["rating"] = latestRating.Rating
			studentData["total_problems"] = latestRating.ProblemsCount
			studentData["easy_problems"] = latestRating.EasyCount
			studentData["medium_problems"] = latestRating.MediumCount
			studentData["hard_problems"] = latestRating.HardCount
			studentData["global_rank"] = latestRating.GlobalRank
		} else {
			// Add default values if no ratings are available
			studentData["rating"] = 0
			studentData["total_problems"] = 0
			studentData["easy_problems"] = 0
			studentData["medium_problems"] = 0
			studentData["hard_problems"] = 0
			studentData["global_rank"] = 0
		}

		response = append(response, studentData)
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"students":  response,
	})
}

// UpdateContestHistory updates the contest history for a specific student
func (h *Handler) UpdateContestHistory(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	student, err := h.db.GetStudentByID(studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	// Create LeetCode service
	leetcodeService := leetcode.NewService()

	// Get fresh contest stats
	contestStats, err := leetcodeService.GetContestRanking(student.LeetcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch LeetCode contest stats",
			"details": "Could not retrieve contest data from LeetCode API. This could be due to:" +
				"\n- Invalid LeetCode username" +
				"\n- LeetCode API rate limiting" +
				"\n- LeetCode API being unavailable",
			"technical_error": err.Error(),
		})
		return
	}

	// Delete existing contest history
	if err := h.db.DeleteContestHistory(student.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing contest history"})
		return
	}

	now := time.Now()
	var histories []*models.ContestHistory

	// Create new contest history entries
	for _, contest := range contestStats.Data.UserContestRankingHistory {
		history := &models.ContestHistory{
			StudentID:         student.ID,
			ContestTitle:      contest.Contest.Title,
			Rating:            contest.Rating,
			Ranking:           contest.Ranking,
			ProblemsSolved:    contest.ProblemsSolved,
			FinishTimeSeconds: contest.FinishTimeInSeconds,
			ContestDate:       now,
			CreatedAt:         now,
		}
		histories = append(histories, history)
	}

	// Add new contest histories
	if err := h.db.AddContestHistories(student.ID, histories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add contest histories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contest histories updated successfully",
		"count":   len(histories),
	})
}

// UpdateAllContestHistories updates contest histories for all students
func (h *Handler) UpdateAllContestHistories(c *gin.Context) {
	// Create LeetCode service
	leetcodeService := leetcode.NewService()

	// Get all students (paginated)
	page := 1
	pageSize := 10
	var processedCount int
	var failedCount int

	for {
		students, err := h.db.ListStudents(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch students"})
			return
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			// Get fresh contest stats
			contestStats, err := leetcodeService.GetContestRanking(student.LeetcodeID)
			if err != nil {
				failedCount++
				continue
			}

			// Delete existing contest history
			if err := h.db.DeleteContestHistory(student.ID); err != nil {
				failedCount++
				continue
			}

			now := time.Now()
			var histories []*models.ContestHistory

			// Create new contest history entries
			for _, contest := range contestStats.Data.UserContestRankingHistory {
				history := &models.ContestHistory{
					StudentID:         student.ID,
					ContestTitle:      contest.Contest.Title,
					Rating:            contest.Rating,
					Ranking:           contest.Ranking,
					ProblemsSolved:    contest.ProblemsSolved,
					FinishTimeSeconds: contest.FinishTimeInSeconds,
					ContestDate:       now,
					CreatedAt:         now,
				}
				histories = append(histories, history)
			}

			// Add new contest histories
			if err := h.db.AddContestHistories(student.ID, histories); err != nil {
				failedCount++
				continue
			}

			processedCount++
			// Add a small delay between students to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}

		page++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Contest histories update completed",
		"processed_count": processedCount,
		"failed_count":    failedCount,
	})
}

// GetContestHistory retrieves contest history for a specific student
func (h *Handler) GetContestHistory(c *gin.Context) {
	studentID := c.Param("id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required"})
		return
	}

	student, err := h.db.GetStudentByID(studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	// Get contest history
	rankings, err := h.db.GetContestRankings(student.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch contest history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"student_id": student.StudentID,
		"history":    rankings,
	})
}

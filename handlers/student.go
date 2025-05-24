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

	"github.com/ayush/ORBIT/internal/cache"
	"github.com/ayush/ORBIT/internal/leetcode"
	"github.com/ayush/ORBIT/internal/models"
	"github.com/ayush/ORBIT/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	service *service.StudentService
	cache   *cache.RedisCache
	logger  *zap.Logger
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

func NewHandler(service *service.StudentService, cache *cache.RedisCache, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		cache:   cache,
		logger:  logger,
	}
}

// ListStudents retrieves a paginated list of students
func (h *Handler) ListStudents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	department := c.Query("department")
	batch := c.Query("batch")

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:list:page:%d:size:%d:dept:%s:batch:%s", page, pageSize, department, batch)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	students, err := h.service.ListStudents(c.Request.Context(), page, pageSize, department, batch)
	if err != nil {
		h.logger.Error("failed to list students", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list students"})
		return
	}

	// Cache the result for 5 minutes
	h.cache.Set(c, cacheKey, students, 5*time.Minute)
	c.JSON(http.StatusOK, students)
}

// GetStudent retrieves a single student by ID
func (h *Handler) GetStudent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:detail:%d", id)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	student, err := h.service.GetStudent(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to get student", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get student"})
		return
	}

	// Cache the result for 5 minutes
	h.cache.Set(c, cacheKey, student, 5*time.Minute)
	c.JSON(http.StatusOK, student)
}

// GetLeetCodeStats retrieves a student's LeetCode statistics
func (h *Handler) GetLeetCodeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:leetcode:%d", id)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	stats, err := h.service.GetLeetCodeStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to get LeetCode stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get LeetCode stats"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, stats, time.Hour)
	c.JSON(http.StatusOK, stats)
}

// UpdateLeetCodeStats updates a student's LeetCode statistics
func (h *Handler) UpdateLeetCodeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	stats, err := h.service.UpdateLeetCodeStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to update LeetCode stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update LeetCode stats"})
		return
	}

	// Invalidate cache
	h.cache.Del(c, fmt.Sprintf("students:leetcode:%d", id))
	c.JSON(http.StatusOK, stats)
}

// GetDailyProgress retrieves a student's daily progress
func (h *Handler) GetDailyProgress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:daily:%d:%s:%s", id, startDate, endDate)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	progress, err := h.service.GetDailyProgress(c.Request.Context(), uint(id), startDate, endDate)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to get daily progress", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get daily progress"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, progress, time.Hour)
	c.JSON(http.StatusOK, progress)
}

// GetWeeklyStats retrieves a student's weekly statistics
func (h *Handler) GetWeeklyStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:weekly:%d", id)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	stats, err := h.service.GetWeeklyStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to get weekly stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weekly stats"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, stats, time.Hour)
	c.JSON(http.StatusOK, stats)
}

// GetContestHistory retrieves a student's contest history
func (h *Handler) GetContestHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("students:contests:%d", id)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	history, err := h.service.GetContestHistory(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to get contest history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contest history"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, history, time.Hour)
	c.JSON(http.StatusOK, history)
}

// UpdateContestHistory updates a student's contest history
func (h *Handler) UpdateContestHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("invalid student ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	history, err := h.service.GetContestHistory(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		h.logger.Error("failed to update contest history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update contest history"})
		return
	}

	// Invalidate cache
	h.cache.Del(c, fmt.Sprintf("students:contests:%d", id))
	c.JSON(http.StatusOK, history)
}

// GetLeaderboard retrieves the student leaderboard
func (h *Handler) GetLeaderboard(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "all")
	department := c.Query("department")
	batch := c.Query("batch")

	// Try to get from cache first
	cacheKey := fmt.Sprintf("leaderboard:%s:dept:%s:batch:%s", timeframe, department, batch)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), timeframe, department, batch)
	if err != nil {
		h.logger.Error("failed to get leaderboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get leaderboard"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, leaderboard, time.Hour)
	c.JSON(http.StatusOK, leaderboard)
}

// GetTrendingStudents retrieves trending students based on recent activity
func (h *Handler) GetTrendingStudents(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Try to get from cache first
	cacheKey := fmt.Sprintf("trending:days:%d:limit:%d", days, limit)
	if cached, err := h.cache.Get(c, cacheKey); err == nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	trending, err := h.service.GetTrendingStudents(c.Request.Context(), days, limit)
	if err != nil {
		h.logger.Error("failed to get trending students", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trending students"})
		return
	}

	// Cache the result for 1 hour
	h.cache.Set(c, cacheKey, trending, time.Hour)
	c.JSON(http.StatusOK, trending)
}

// CreateStudent creates a new student record
func (h *Handler) CreateStudent(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "CreateStudent"),
		zap.String("request_id", c.GetString("request_id")),
	)

	logger.Info("Starting student creation")

	var req models.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger = logger.With(
		zap.String("student_id", req.StudentID),
		zap.String("leetcode_id", req.LeetcodeID),
		zap.String("email", req.Email),
	)

	student := models.Student{
		StudentID:   req.StudentID,
		Name:        req.Name,
		Email:       req.Email,
		LeetcodeID:  req.LeetcodeID,
		PassingYear: req.PassingYear,
	}

	if err := h.service.CreateStudent(&student); err != nil {
		logger.Error("Failed to create student",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	duration := time.Since(start)
	logger.Info("Student created successfully",
		zap.Duration("duration", duration),
		zap.Uint("internal_id", student.ID),
	)

	c.JSON(http.StatusCreated, student)
}

// GetAllStudents retrieves all students with optional pagination
func (h *Handler) GetAllStudents(c *gin.Context) {
	start := time.Now()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "60"))
	department := c.Query("department")
	batch := c.Query("batch")

	logger := h.logger.With(
		zap.String("handler", "GetAllStudents"),
		zap.String("request_id", c.GetString("request_id")),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	logger.Info("Starting student list retrieval")

	students, err := h.service.ListStudents(c.Request.Context(), page, pageSize, department, batch)
	if err != nil {
		logger.Error("Failed to get students",
			zap.Error(err),
		)
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

	duration := time.Since(start)
	logger.Info("Student list retrieved successfully",
		zap.Duration("duration", duration),
		zap.Int("student_count", len(students)),
	)

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"students":  response,
	})
}

// GetStudentDetails retrieves student details including weekly stats
func (h *Handler) GetStudentDetails(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "GetStudentDetails"),
		zap.String("request_id", c.GetString("request_id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid student ID",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	// Get student with all stats
	student, err := h.service.GetStudentWithStats(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Student not found",
				zap.Uint("student_id", uint(id)),
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student not found or no ratings available",
				"details": "The student either does not exist or has no LeetCode ratings recorded yet. " +
					"Try updating their rating first using PUT /api/v1/students/:id/rating",
			})
			return
		}
		logger.Error("Failed to get student",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get LeetCode stats
	leetcodeStats, err := h.service.GetLeetCodeStats(c.Request.Context(), uint(id))
	if err == nil && leetcodeStats != nil {
		student.LeetCodeStats = leetcodeStats
	}

	// Get contest rankings
	contestRankings, err := h.service.GetContestHistory(c.Request.Context(), uint(id))
	if err == nil && len(contestRankings) > 0 {
		student.ContestHistory = contestRankings
	}

	duration := time.Since(start)
	logger.Info("Student details retrieved successfully",
		zap.Duration("duration", duration),
		zap.String("leetcode_id", student.LeetcodeID),
	)

	c.JSON(http.StatusOK, student)
}

// BulkCreateStudents handles bulk student creation from file upload
func (h *Handler) BulkCreateStudents(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "BulkCreateStudents"),
		zap.String("request_id", c.GetString("request_id")),
	)

	logger.Info("Starting bulk student creation")

	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("No file uploaded",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	logger = logger.With(
		zap.String("filename", file.Filename),
		zap.Int64("file_size", file.Size),
	)

	if file.Size > maxFileSize {
		logger.Warn("File too large",
			zap.Int64("max_size", maxFileSize),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".xlsx" && ext != ".csv" {
		logger.Warn("Invalid file format",
			zap.String("extension", ext),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Only .xlsx and .csv files are allowed"})
		return
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory",
			zap.Error(err),
			zap.String("directory", uploadDir),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s%s", timestamp, strings.TrimSuffix(file.Filename, ext), ext)
	filepath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		logger.Error("Failed to save file",
			zap.Error(err),
			zap.String("filepath", filepath),
		)
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

	if err := h.service.CreateFileUpload(fileUpload); err != nil {
		logger.Error("Failed to record file upload",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record file upload"})
		return
	}

	duration := time.Since(start)
	logger.Info("File upload processed successfully",
		zap.Duration("duration", duration),
		zap.String("status", "processing"),
		zap.Uint("file_upload_id", fileUpload.ID),
	)

	go h.processUploadedFile(fileUpload)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "File uploaded successfully. Processing started.",
		"file_id": fileUpload.ID,
	})
}

func (h *Handler) processUploadedFile(fileUpload *models.FileUpload) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "processUploadedFile"),
		zap.Uint("file_upload_id", fileUpload.ID),
		zap.String("filename", fileUpload.FileName),
	)

	logger.Info("Starting file processing")

	var records [][]string
	var err error

	if strings.HasSuffix(fileUpload.FileName, ".xlsx") {
		records, err = readExcelFile(fileUpload.StoragePath)
	} else {
		records, err = readCSVFile(fileUpload.StoragePath)
	}

	if err != nil {
		logger.Error("Failed to read file",
			zap.Error(err),
		)
		updateFileUploadStatus(h, fileUpload, "failed", err.Error(), 0, 0, 0)
		return
	}

	if len(records) < 2 {
		logger.Warn("Empty file or no data rows")
		updateFileUploadStatus(h, fileUpload, "failed", "File is empty or has no data rows", 0, 0, 0)
		return
	}

	successful := 0
	failed := 0
	errors := make([]string, 0)

	logger.Info("Processing records",
		zap.Int("total_records", len(records)-1), // Subtract header row
	)

	for i, record := range records[1:] {
		recordLogger := logger.With(
			zap.Int("row_number", i+2),
		)

		if len(record) != 5 {
			failed++
			msg := fmt.Sprintf("Row %d: Invalid number of columns", i+2)
			errors = append(errors, msg)
			recordLogger.Warn("Invalid record format",
				zap.Int("columns_found", len(record)),
				zap.Int("columns_expected", 5),
			)
			continue
		}

		student := models.Student{
			StudentID:   record[0],
			Name:        record[1],
			Email:       record[2],
			LeetcodeID:  record[3],
			PassingYear: parseInt(record[4]),
		}

		recordLogger = recordLogger.With(
			zap.String("student_id", student.StudentID),
			zap.String("leetcode_id", student.LeetcodeID),
		)

		if err := h.service.CreateStudent(&student); err != nil {
			failed++
			msg := fmt.Sprintf("Row %d: %s", i+2, err.Error())
			errors = append(errors, msg)
			recordLogger.Error("Failed to create student",
				zap.Error(err),
			)
			continue
		}

		successful++
		recordLogger.Info("Student created successfully")
	}

	status := "completed"
	if failed > 0 {
		status = "completed_with_errors"
	}

	duration := time.Since(start)
	logger.Info("File processing completed",
		zap.Duration("duration", duration),
		zap.String("status", status),
		zap.Int("successful", successful),
		zap.Int("failed", failed),
		zap.Int("total_processed", len(records)-1),
	)

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

	if err := h.service.UpdateFileUpload(fileUpload); err != nil {
		log.Printf("Failed to update file upload status: %v", err)
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

// UpdateAllContestHistories updates contest histories for all students
func (h *Handler) UpdateAllContestHistories(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "UpdateAllContestHistories"),
		zap.String("request_id", c.GetString("request_id")),
	)

	logger.Info("Starting contest history update for all students")

	// Create LeetCode service
	leetcodeService := leetcode.NewService()

	// Get all students (paginated)
	page := 1
	pageSize := 10
	var processedCount int
	var failedCount int
	var totalProcessed int

	for {
		logger.Info("Processing student batch",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Int("processed_so_far", totalProcessed),
		)

		students, err := h.service.ListStudents(c.Request.Context(), page, pageSize, "", "")
		if err != nil {
			logger.Error("Failed to fetch students",
				zap.Error(err),
				zap.Int("page", page),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch students"})
			return
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			studentLogger := logger.With(
				zap.String("student_id", student.StudentID),
				zap.String("leetcode_id", student.LeetcodeID),
				zap.Uint("internal_id", student.ID),
			)

			studentLogger.Info("Processing student")

			// Get fresh contest stats
			contestStats, err := leetcodeService.GetContestRanking(student.LeetcodeID)
			if err != nil {
				studentLogger.Error("Failed to fetch contest stats",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			// Delete existing contest history
			if err := h.service.DeleteContestHistory(c.Request.Context(), student.ID); err != nil {
				studentLogger.Error("Failed to delete existing contest history",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			now := time.Now()
			var histories []models.ContestHistory

			// Create new contest history entries
			for _, contest := range contestStats.Data.UserContestRankingHistory {
				history := models.ContestHistory{
					StudentID:         student.ID,
					ContestTitle:      contest.Contest.Title,
					Rating:            contest.Rating,
					Ranking:           contest.Ranking,
					ProblemsSolved:    contest.ProblemsSolved,
					FinishTimeSeconds: int64(contest.FinishTimeInSeconds),
					ContestDate:       now,
					CreatedAt:         now,
				}
				histories = append(histories, history)
			}

			// Add new contest histories
			if err := h.service.AddContestHistories(c.Request.Context(), student.ID, histories); err != nil {
				studentLogger.Error("Failed to add contest histories",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			studentLogger.Info("Successfully processed student",
				zap.Int("histories_added", len(histories)),
			)
			processedCount++
			totalProcessed++

			// Add a small delay between students to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}

		page++
	}

	duration := time.Since(start)
	logger.Info("Contest history update completed for all students",
		zap.Duration("duration", duration),
		zap.Int("total_processed", processedCount),
		zap.Int("failed_count", failedCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Contest histories update completed",
		"processed_count": processedCount,
		"failed_count":    failedCount,
		"duration":        duration.String(),
	})
}

// UpdateStudentRating updates a student's LeetCode rating
func (h *Handler) UpdateStudentRating(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "UpdateStudentRating"),
		zap.String("request_id", c.GetString("request_id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid student ID",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student ID"})
		return
	}

	logger = logger.With(zap.Uint("student_id", uint(id)))
	logger.Info("Starting rating update")

	// Get student to get their LeetCode ID
	student, err := h.service.GetStudent(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Student not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		logger.Error("Failed to get student",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if student.LeetcodeID == "" {
		logger.Error("Student has no LeetCode ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "student has no LeetCode ID configured"})
		return
	}

	logger = logger.With(
		zap.String("leetcode_id", student.LeetcodeID),
	)
	logger.Info("Found student, fetching LeetCode stats")

	// Get fresh LeetCode stats with retries
	var leetcodeStats *models.LeetCodeStats
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		leetcodeStats, err = h.service.GetLeetCodeStats(c.Request.Context(), uint(id))
		if err == nil {
			break
		}
		logger.Warn("Failed to fetch LeetCode stats, retrying",
			zap.Error(err),
			zap.Int("attempt", i+1),
		)
		time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
	}

	if err != nil {
		logger.Error("Failed to fetch LeetCode stats after retries",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch LeetCode stats",
			"details": "Could not retrieve data from LeetCode API after multiple attempts. This could be due to:" +
				"\n- Invalid LeetCode username" +
				"\n- LeetCode API rate limiting" +
				"\n- LeetCode API being unavailable",
			"technical_error": err.Error(),
		})
		return
	}

	if leetcodeStats.TotalSolved == 0 && leetcodeStats.EasySolved == 0 && leetcodeStats.MediumSolved == 0 && leetcodeStats.HardSolved == 0 {
		logger.Error("Got zero values from LeetCode API",
			zap.Any("stats", leetcodeStats),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid LeetCode stats",
			"details": "Received all zero values from LeetCode API. This could mean:" +
				"\n- The LeetCode username is incorrect" +
				"\n- The user has not solved any problems" +
				"\n- There was an error fetching the data",
			"stats": leetcodeStats,
		})
		return
	}

	logger.Info("Got LeetCode stats",
		zap.Int("total_solved", leetcodeStats.TotalSolved),
		zap.Int("easy_solved", leetcodeStats.EasySolved),
		zap.Int("medium_solved", leetcodeStats.MediumSolved),
		zap.Int("hard_solved", leetcodeStats.HardSolved),
		zap.Float64("contest_rating", leetcodeStats.ContestRating),
		zap.Int("contest_ranking", leetcodeStats.ContestRanking),
	)

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

	logger.Info("Calculated rating",
		zap.Int("problem_rating", problemRating),
		zap.Int("contest_bonus", contestBonus),
		zap.Int("final_rating", rating.Rating),
	)

	if err := h.service.UpdateStudentRating(uint(id), rating); err != nil {
		logger.Error("Failed to update student rating",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	duration := time.Since(start)
	logger.Info("Student rating updated successfully",
		zap.Duration("duration", duration),
		zap.Int("rating", rating.Rating),
		zap.Int("problems_count", rating.ProblemsCount),
	)

	c.JSON(http.StatusOK, rating)
}

// UpdateAllStudentRatings updates LeetCode ratings for all students
func (h *Handler) UpdateAllStudentRatings(c *gin.Context) {
	start := time.Now()
	logger := h.logger.With(
		zap.String("handler", "UpdateAllStudentRatings"),
		zap.String("request_id", c.GetString("request_id")),
	)

	logger.Info("Starting rating update for all students")

	// Get all students (paginated)
	page := 1
	pageSize := 10
	var processedCount int
	var failedCount int
	var totalProcessed int

	for {
		logger.Info("Processing student batch",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Int("processed_so_far", totalProcessed),
		)

		students, err := h.service.ListStudents(c.Request.Context(), page, pageSize, "", "")
		if err != nil {
			logger.Error("Failed to fetch students",
				zap.Error(err),
				zap.Int("page", page),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch students"})
			return
		}

		if len(students) == 0 {
			break // No more students to process
		}

		// Process each student in the batch
		for _, student := range students {
			studentLogger := logger.With(
				zap.String("student_id", student.StudentID),
				zap.String("leetcode_id", student.LeetcodeID),
				zap.Uint("internal_id", student.ID),
			)

			if student.LeetcodeID == "" {
				studentLogger.Warn("Student has no LeetCode ID, skipping")
				failedCount++
				continue
			}

			studentLogger.Info("Processing student")

			// Get fresh LeetCode stats with retries
			var leetcodeStats *models.LeetCodeStats
			maxRetries := 3
			var lastErr error
			for i := 0; i < maxRetries; i++ {
				leetcodeStats, lastErr = h.service.GetLeetCodeStats(c.Request.Context(), student.LeetcodeID)
				if lastErr == nil {
					break
				}
				studentLogger.Warn("Failed to fetch LeetCode stats, retrying",
					zap.Error(lastErr),
					zap.Int("attempt", i+1),
				)
				time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
			}

			if lastErr != nil {
				studentLogger.Error("Failed to fetch LeetCode stats after retries",
					zap.Error(lastErr),
				)
				failedCount++
				continue
			}

			if leetcodeStats.TotalSolved == 0 && leetcodeStats.EasySolved == 0 && leetcodeStats.MediumSolved == 0 && leetcodeStats.HardSolved == 0 {
				studentLogger.Error("Got zero values from LeetCode API",
					zap.Any("stats", leetcodeStats),
				)
				failedCount++
				continue
			}

			now := time.Now()
			rating := &models.Rating{
				StudentID:     student.ID,
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

			if err := h.service.UpdateStudentRating(student.ID, rating); err != nil {
				studentLogger.Error("Failed to update student rating",
					zap.Error(err),
				)
				failedCount++
				continue
			}

			studentLogger.Info("Successfully processed student",
				zap.Int("rating", rating.Rating),
				zap.Int("problems_count", rating.ProblemsCount),
			)
			processedCount++
			totalProcessed++

			// Add a small delay between students to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}

		page++
	}

	duration := time.Since(start)
	logger.Info("Rating update completed for all students",
		zap.Duration("duration", duration),
		zap.Int("total_processed", processedCount),
		zap.Int("failed_count", failedCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Rating updates completed",
		"processed_count": processedCount,
		"failed_count":    failedCount,
		"duration":        duration.String(),
	})
}

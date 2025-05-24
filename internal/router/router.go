package router

import (
	"internal/middleware"
	"time"

	"github.com/ayush/ORBIT/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	// Add global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Rate limit the API to 100 requests per minute per IP
		v1.Use(middleware.RateLimit(100, 1*time.Minute))

		// Student routes
		students := v1.Group("/students")
		{
			// Basic student operations
			students.GET("", h.ListStudents)
			students.POST("", h.CreateStudent)
			students.GET("/:id", h.GetStudent)
			students.PUT("/:id", h.UpdateStudent)
			students.DELETE("/:id", h.DeleteStudent)

			// Bulk operations
			bulk := students.Group("/bulk")
			{
				bulk.POST("/import", h.BulkCreateStudents)
				bulk.PUT("/update-stats", h.UpdateAllStudentStats)
			}

			// LeetCode stats
			stats := students.Group("/:id/stats")
			{
				stats.GET("/leetcode", h.GetLeetCodeStats)
				stats.PUT("/leetcode", h.UpdateLeetCodeStats)
				stats.GET("/daily", h.GetDailyProgress)
				stats.GET("/weekly", h.GetWeeklyStats)
			}

			// Contest history
			contests := students.Group("/:id/contests")
			{
				contests.GET("", h.GetContestHistory)
				contests.PUT("", h.UpdateContestHistory)
				contests.GET("/latest", h.GetLatestContestResult)
				contests.GET("/stats", h.GetContestStats)
			}

			// Achievements
			achievements := students.Group("/:id/achievements")
			{
				achievements.GET("", h.GetAchievements)
				achievements.POST("", h.UnlockAchievement)
			}
		}

		// Achievement management
		achievements := v1.Group("/achievements")
		{
			achievements.GET("", h.ListAchievements)
			achievements.POST("", h.CreateAchievement)
			achievements.PUT("/:id", h.UpdateAchievement)
			achievements.DELETE("/:id", h.DeleteAchievement)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/department/:dept", h.GetDepartmentStats)
			analytics.GET("/batch/:batch", h.GetBatchStats)
			analytics.GET("/leaderboard", h.GetLeaderboard)
			analytics.GET("/trending", h.GetTrendingStudents)
		}
	}

	return r
}

package router

import (
	"github.com/ayush/ORBIT/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handlers.Handler, wh *handlers.WeeklyStatsHandler) *gin.Engine {
	r := gin.Default()

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Student routes
		students := v1.Group("/students")
		{
			// Basic student operations
			students.GET("", h.GetAllStudents)
			students.POST("", h.CreateStudent)
			students.POST("/bulk", h.BulkCreateStudents)
			students.GET("/:id", h.GetStudentDetails)

			// Weekly stats routes
			students.GET("/:id/weekly-stats", wh.GetStudentWeeklyStats)
			students.GET("/:id/current-week", wh.GetCurrentWeekStats)
			students.PUT("/:id/weekly-stats", wh.UpdateWeeklyStats)
			students.POST("/:id/weekly-stats/leetcode", wh.UpdateWeeklyStatsFromLeetCode)
			students.PUT("/weekly-stats/update-all", wh.UpdateAllWeeklyStats)

			// Contest history routes
			students.GET("/:id/contest-history", h.GetContestHistory)
			students.PUT("/:id/contest-history", h.UpdateContestHistory)
			students.PUT("/contest-history/update-all", h.UpdateAllContestHistories)
		}
	}

	return r
}

package server

import (
	"github.com/ayush/ORBIT/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		students := v1.Group("/students")
		{
			students.POST("", h.CreateStudent)
			students.POST("/bulk", h.BulkCreateStudents)
			students.GET("", h.ListStudents)
			students.GET("/:id", h.GetStudent)
			students.PUT("/:id/rating", h.UpdateStudentRating)
			students.GET("/:id/leetcode", h.GetLeetCodeStats)
			students.GET("/:id/contest-rankings", h.GetContestRankings)
		}
	}

	return r
}

package routes

import (
	"github.com/ayush/ORBIT/handlers"
	"github.com/ayush/ORBIT/internal/database"
	"github.com/ayush/ORBIT/pkg/leetcode"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *database.StudentDB) {
	studentHandler := handlers.NewHandler(db)
	leetcodeClient := leetcode.NewClient()
	weeklyStatsHandler := handlers.NewWeeklyStatsHandler(db, leetcodeClient)

	api := r.Group("/api/v1")
	{
		// Student routes
		api.GET("/students", studentHandler.GetAllStudents)
		api.POST("/students", studentHandler.CreateStudent)
		api.POST("/students/bulk", studentHandler.BulkCreateStudents)
		api.GET("/students/:id", studentHandler.GetStudentDetails)

		// Weekly stats routes
		api.GET("/students/:id/weekly-stats", weeklyStatsHandler.GetStudentWeeklyStats)
		api.PUT("/students/:id/weekly-stats", weeklyStatsHandler.UpdateWeeklyStats)
		api.PUT("/students/weekly-stats/update-all", weeklyStatsHandler.UpdateAllWeeklyStats)
	}
}

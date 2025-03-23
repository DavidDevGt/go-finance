package routes

import (
	"github.com/DavidDevGt/go-finance/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Expenses
		api.GET("/expenses", controllers.GetExpenses)
		api.GET("/expenses/:id", controllers.GetExpenseByID)
		api.GET("/expenses/week/:week", controllers.GetExpensesByWeek)
		api.GET("/expenses/week/:week/export", controllers.ExportExpensesByWeekToCSV)
		api.POST("/expenses", controllers.CreateExpense)
		api.PUT("/expenses/:id", controllers.UpdateExpense)
		api.DELETE("/expenses/:id", controllers.DeleteExpense)

		// Budget
		api.POST("/budget", controllers.SetWeeklyBudget)
		api.GET("/budget/:year/:week", controllers.GetWeeklySummary)
	}
}

package main

import (
	"github.com/DavidDevGt/go-finance/database"
	"github.com/DavidDevGt/go-finance/routes"

	"github.com/gin-gonic/gin"

	// Swagger
	_ "github.com/DavidDevGt/go-finance/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title API de Finanzas Personales Semanales
// @version 1.0
// @description Esta API permite gestionar gastos y presupuestos semanales
// @termsOfService https://github.com/DavidDevGt/go-finance

// @contact.name David Dev GT
// @contact.url https://github.com/DavidDevGt
// @contact.email davidgt@davidwebgt.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8055
// @BasePath /api
func main() {
	database.ConnectDatabase()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	routes.SetupRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8055")
}

package database

import (
	"log"
	"os"
	"time"

	"github.com/DavidDevGt/go-finance/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	database, err := gorm.Open(sqlite.Open("expenses.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("❌ No se pudo conectar a la base de datos: ", err)
	}
	if err := database.AutoMigrate(&models.Expense{}, &models.WeeklyBudget{}); err != nil {
		log.Fatal("❌ Error al migrar los modelos: ", err)
	}

	DB = database
	log.Println("✅ Base de datos conectada.")
}

package main

import (
	"log"
	"os"
	"time"

	"segmentation-api/internal/api"
	lgr "segmentation-api/internal/logger"
	mysqlRepo "segmentation-api/internal/repository/mysql"
	"segmentation-api/internal/service"

	_ "segmentation-api/docs" // Swagger documentation

	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Initialize logger
	log_, file, err := lgr.New()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer file.Close()

	// GORM logger for database
	gormLog := gormLogger.New(
		log.New(log_.Writer(), log_.Prefix(), log_.Flags()),
		gormLogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormLogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Database connection using NewMySQL helper
	db, err := mysqlRepo.NewMySQL(gormLog)
	if err != nil {
		log_.Printf("Failed to connect to database: %v", err)
		panic("failed to connect to database")
	}

	// Run migrations
	if err := mysqlRepo.RunMigrations(db); err != nil {
		log_.Printf("Failed to run migrations: %v", err)
		panic("failed to run migrations")
	}

	// Initialize repository and service
	repo := mysqlRepo.NewSegmentationRepository(db)
	svc := service.NewSegmentationService(repo)

	// Setup router
	router := api.SetupRouter(svc)

	// Get port from environment or default to 8080
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log_.Printf("Starting API server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log_.Printf("Failed to start server: %v", err)
		panic(err)
	}
}

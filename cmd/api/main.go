package main

import (
	"os"

	"segmentation-api/internal/api"
	lgr "segmentation-api/internal/logger"
	"segmentation-api/internal/models"
	mysqlRepo "segmentation-api/internal/repository/mysql"
	"segmentation-api/internal/service"

	_ "segmentation-api/docs" // Swagger documentation

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	log_, file, err := lgr.New()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer file.Close()

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/segmentation?charset=utf8mb4&parseTime=true"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log_.Printf("Failed to connect to database: %v", err)
		panic("failed to connect to database")
	}

	// Run migrations
	if err := db.AutoMigrate(&models.Segmentation{}); err != nil {
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

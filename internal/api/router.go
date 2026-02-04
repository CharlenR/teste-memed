package api

import (
	"segmentation-api/internal/api/handler"
	"segmentation-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all API routes
func SetupRouter(svc *service.SegmentationService) *gin.Engine {
	router := gin.Default()

	// Initialize handler
	h := handler.NewSegmentationHandler(svc)

	// Health check endpoint
	router.GET("/health", h.Health)

	// Segmentation endpoints
	router.GET("/users/:user_id/segmentations", h.GetUserSegmentations)

	// Swagger documentation
	// Available at http://localhost:8080/swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

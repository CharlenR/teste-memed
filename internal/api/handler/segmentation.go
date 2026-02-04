package handler

import (
	"net/http"
	"strconv"

	"segmentation-api/internal/service"

	"github.com/gin-gonic/gin"
)

// SegmentationHandler handles segmentation-related HTTP requests
type SegmentationHandler struct {
	service *service.SegmentationService
}

// NewSegmentationHandler creates a new segmentation handler
func NewSegmentationHandler(s *service.SegmentationService) *SegmentationHandler {
	return &SegmentationHandler{service: s}
}

// GetUserSegmentations retrieves all segmentations for a user
// GET /users/:user_id/segmentations
func (h *SegmentationHandler) GetUserSegmentations(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	ctx := c.Request.Context()
	result, err := h.service.GetByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Health returns the health status of the API
// GET /health
func (h *SegmentationHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// MockRepository for testing
type MockRepository struct {
	findByUserIDFunc func(ctx context.Context, userID uint64) ([]models.Segmentation, error)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	return repository.UpsertInserted, nil
}

func TestGetUserSegmentations_Success(t *testing.T) {
	// Setup mock data
	mockData := []models.Segmentation{
		{
			ID:               1,
			UserID:           123,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{"experience_years": 5, "certification": "CRM123"}`),
		},
		{
			ID:               2,
			UserID:           123,
			SegmentationType: "drug",
			SegmentationName: "Alopáticos",
			Data:             datatypes.JSON(`{"quantity": "200"}`),
		},
	}

	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 123 {
				return mockData, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	// Create request
	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "123"}}

	// Call handler
	handler.GetUserSegmentations(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp service.SegmentationResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.UserID != 123 {
		t.Fatalf("expected user_id 123, got %d", resp.UserID)
	}

	if len(resp.Segmentations) != 2 {
		t.Fatalf("expected 2 segmentation types, got %d", len(resp.Segmentations))
	}

	// Verify specialties
	if len(resp.Segmentations["specialties"]) != 1 {
		t.Fatalf("expected 1 specialty, got %d", len(resp.Segmentations["specialties"]))
	}

	if resp.Segmentations["specialties"][0].Name != "Cardiologia" {
		t.Fatalf("expected specialty name 'Cardiologia', got %s", resp.Segmentations["specialties"][0].Name)
	}

	// Verify drugs
	if len(resp.Segmentations["drugs"]) != 1 {
		t.Fatalf("expected 1 drug, got %d", len(resp.Segmentations["drugs"]))
	}

	if resp.Segmentations["drugs"][0].Name != "Alopáticos" {
		t.Fatalf("expected drug name 'Alopáticos', got %s", resp.Segmentations["drugs"][0].Name)
	}
}

func TestGetUserSegmentations_InvalidUserID(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	// Create request with invalid user_id
	req := httptest.NewRequest("GET", "/users/invalid/segmentations", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "invalid"}}

	// Call handler
	handler.GetUserSegmentations(c)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid user_id format" {
		t.Fatalf("expected error message about invalid format, got %s", resp["error"])
	}
}

func TestGetUserSegmentations_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	// Create request
	req := httptest.NewRequest("GET", "/users/999/segmentations", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "999"}}

	// Call handler
	handler.GetUserSegmentations(c)

	// Verify response - should return empty segmentations, not 404
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.UserID != 999 {
		t.Fatalf("expected user_id 999, got %d", resp.UserID)
	}
}

func TestHealth(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Setup Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.Health(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Fatalf("expected status 'healthy', got %s", resp["status"])
	}
}

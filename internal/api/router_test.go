package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"
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

func TestSetupRouter_RoutesDefined(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)

	router := SetupRouter(svc)

	if router == nil {
		t.Fatal("expected router to be initialized")
	}

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected health endpoint to return 200, got %d", w.Code)
	}
}

func TestSetupRouter_SegmentationEndpoint(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return nil, nil
		},
	}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test segmentation endpoint
	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected segmentation endpoint to return 200, got %d", w.Code)
	}
}

func TestSetupRouter_InvalidRoute(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test non-existent endpoint
	req := httptest.NewRequest("GET", "/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected non-existent endpoint to return 404, got %d", w.Code)
	}
}

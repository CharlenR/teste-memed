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

func (m *MockRepository) BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error) {
	results := make([]repository.UpsertResult, len(*s))
	errors := make([]error, len(*s))
	for i := range results {
		results[i] = repository.UpsertInserted
		errors[i] = nil
	}
	return results, errors
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

func TestSetupRouter_SwaggerEndpoint(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test swagger endpoint
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected swagger endpoint to return 200, got %d", w.Code)
	}
}

func TestSetupRouter_HealthMultipleCalls(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Call health endpoint multiple times
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("call %d: expected health endpoint to return 200, got %d", i+1, w.Code)
		}
	}
}

func TestSetupRouter_SegmentationEndpointMultipleUsers(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return nil, nil
		},
	}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test with different user IDs
	userIDs := []string{"1", "100", "999", "123456789"}
	for _, userID := range userIDs {
		req := httptest.NewRequest("GET", "/users/"+userID+"/segmentations", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected segmentation endpoint for user %s to return 200, got %d", userID, w.Code)
		}
	}
}

func TestSetupRouter_InvalidSegmentationEndpoint(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test with invalid user_id
	req := httptest.NewRequest("GET", "/users/invalid/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid user_id to return 400, got %d", w.Code)
	}
}

func TestSetupRouter_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Test POST on GET-only endpoint - Gin returns 404 for undefined routes by default
	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin doesn't define a POST /health route, so it returns 404
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected POST /health to return 404 (route not found), got %d", w.Code)
	}
}

func TestSetupRouter_HealthResponseContentType(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Fatalf("expected content-type 'application/json; charset=utf-8', got %s", contentType)
	}
}

func TestSetupRouter_SegmentationResponseContentType(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return []models.Segmentation{}, nil
		},
	}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Fatalf("expected content-type 'application/json; charset=utf-8', got %s", contentType)
	}
}

func TestSetupRouter_SegmentationEndpointZeroUserID(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return []models.Segmentation{}, nil
		},
	}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/0/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected user_id 0 to return 200, got %d", w.Code)
	}
}

func TestSetupRouter_PathNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	paths := []string{
		"/users",
		"/users/123",
		"/segmentations",
		"/users/123/segmentations/456",
	}

	for _, path := range paths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected path %s to return 404, got %d", path, w.Code)
		}
	}
}

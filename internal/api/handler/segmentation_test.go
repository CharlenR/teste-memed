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

func (m *MockRepository) BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error) {
	return []repository.UpsertResult{repository.UpsertInserted}, nil
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

func TestGetUserSegmentations_EmptyUserID(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users//segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: ""}}

	handler.GetUserSegmentations(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty user_id, got %d", w.Code)
	}
}

func TestGetUserSegmentations_NegativeUserID(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/-1/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "-1"}}

	handler.GetUserSegmentations(c)

	// -1 can parse to uint64 but as a very large number due to two's complement
	// The handler should still process it
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestGetUserSegmentations_LargeUserID(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return []models.Segmentation{}, nil
		},
	}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/18446744073709551615/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "18446744073709551615"}}

	handler.GetUserSegmentations(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for large user_id, got %d", w.Code)
	}
}

func TestGetUserSegmentations_MultipleSegmentationTypes(t *testing.T) {
	mockData := []models.Segmentation{
		{
			ID:               1,
			UserID:           456,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               2,
			UserID:           456,
			SegmentationType: "specialty",
			SegmentationName: "Neurologia",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               3,
			UserID:           456,
			SegmentationType: "drug",
			SegmentationName: "Antibióticos",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               4,
			UserID:           456,
			SegmentationType: "drug",
			SegmentationName: "Analgésicos",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               5,
			UserID:           456,
			SegmentationType: "patient",
			SegmentationName: "Crônicos",
			Data:             datatypes.JSON(`{}`),
		},
	}

	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 456 {
				return mockData, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/456/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "456"}}

	handler.GetUserSegmentations(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Segmentations["specialties"]) != 2 {
		t.Fatalf("expected 2 specialties, got %d", len(resp.Segmentations["specialties"]))
	}

	if len(resp.Segmentations["drugs"]) != 2 {
		t.Fatalf("expected 2 drugs, got %d", len(resp.Segmentations["drugs"]))
	}

	if len(resp.Segmentations["patients"]) != 1 {
		t.Fatalf("expected 1 patient, got %d", len(resp.Segmentations["patients"]))
	}
}

func TestGetUserSegmentations_ResponseContentType(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return []models.Segmentation{}, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "123"}}

	handler.GetUserSegmentations(c)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Fatalf("expected content-type 'application/json; charset=utf-8', got %s", contentType)
	}
}

func TestHealth_ResponseFormat(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Health(c)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if _, exists := resp["status"]; !exists {
		t.Fatalf("expected 'status' field in response")
	}

	if resp["status"] != "healthy" {
		t.Fatalf("expected status 'healthy', got %v", resp["status"])
	}
}

func TestNewSegmentationHandler_NotNil(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	if handler == nil {
		t.Fatal("expected handler to be initialized")
	}
}

func TestGetUserSegmentations_ServiceError(t *testing.T) {
	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "123"}}

	handler.GetUserSegmentations(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Fatal("expected error field in response")
	}
}

func TestGetUserSegmentations_SpecificUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		want   uint64
	}{
		{
			name:   "user 1",
			userID: "1",
			want:   1,
		},
		{
			name:   "user 100",
			userID: "100",
			want:   100,
		},
		{
			name:   "user 999999",
			userID: "999999",
			want:   999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{
				findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
					if userID == tt.want {
						return []models.Segmentation{}, nil
					}
					return nil, nil
				},
			}

			svc := service.NewSegmentationService(mockRepo)
			handler := NewSegmentationHandler(svc)

			req := httptest.NewRequest("GET", "/users/"+tt.userID+"/segmentations", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{{Key: "user_id", Value: tt.userID}}

			handler.GetUserSegmentations(c)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestHealth_MultipleResponses(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Health(c)

		if w.Code != http.StatusOK {
			t.Fatalf("call %d: expected status 200, got %d", i+1, w.Code)
		}

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"] != "healthy" {
			t.Fatalf("call %d: expected status 'healthy'", i+1)
		}
	}
}

func TestGetUserSegmentations_GroupingByType(t *testing.T) {
	// Test that segmentations are properly grouped
	mockData := []models.Segmentation{
		{
			ID:               1,
			UserID:           500,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               2,
			UserID:           500,
			SegmentationType: "specialty",
			SegmentationName: "Neurologia",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               3,
			UserID:           500,
			SegmentationType: "specialty",
			SegmentationName: "Pediatria",
			Data:             datatypes.JSON(`{}`),
		},
	}

	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 500 {
				return mockData, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	handler := NewSegmentationHandler(svc)

	req := httptest.NewRequest("GET", "/users/500/segmentations", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "500"}}

	handler.GetUserSegmentations(c)

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Segmentations["specialties"]) != 3 {
		t.Fatalf("expected 3 specialties, got %d", len(resp.Segmentations["specialties"]))
	}

	// Verify all names are present
	names := map[string]bool{}
	for _, seg := range resp.Segmentations["specialties"] {
		names[seg.Name] = true
	}

	for _, name := range []string{"Cardiologia", "Neurologia", "Pediatria"} {
		if !names[name] {
			t.Fatalf("expected specialty %s to be present", name)
		}
	}
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"

	"gorm.io/datatypes"
)

// IntegrationMockRepository for integration tests
type IntegrationMockRepository struct {
	findByUserIDFunc func(ctx context.Context, userID uint64) ([]models.Segmentation, error)
	upsertFunc       func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error)
}

func (m *IntegrationMockRepository) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *IntegrationMockRepository) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, s)
	}
	return repository.UpsertInserted, nil
}

// TestIntegration_HealthEndpoint tests health check through full stack
func TestIntegration_HealthEndpoint(t *testing.T) {
	mockRepo := &IntegrationMockRepository{}
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Fatalf("expected status 'healthy', got %v", resp["status"])
	}
}

// TestIntegration_GetSegmentationsFlow tests complete request flow
func TestIntegration_GetSegmentationsFlow(t *testing.T) {
	mockData := []models.Segmentation{
		{
			ID:               1,
			UserID:           123,
			SegmentationType: "drug",
			SegmentationName: "Antibióticos",
			Data:             datatypes.JSON(`{"category": "antibiotic"}`),
		},
		{
			ID:               2,
			UserID:           123,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{"experience_years": 5}`),
		},
	}

	mockRepo := &IntegrationMockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 123 {
				return mockData, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/123/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.UserID != 123 {
		t.Fatalf("expected user_id 123, got %d", resp.UserID)
	}

	if len(resp.Segmentations) != 2 {
		t.Fatalf("expected 2 segmentation types, got %d", len(resp.Segmentations))
	}

	// Verify data is properly grouped
	if len(resp.Segmentations["drugs"]) != 1 {
		t.Fatalf("expected 1 drug, got %d", len(resp.Segmentations["drugs"]))
	}

	if len(resp.Segmentations["specialties"]) != 1 {
		t.Fatalf("expected 1 specialty, got %d", len(resp.Segmentations["specialties"]))
	}
}

// TestIntegration_RepositoryCalledCorrectly verifies service calls repository
func TestIntegration_RepositoryCalledCorrectly(t *testing.T) {
	var calledWithUserID uint64
	var called bool

	mockRepo := &IntegrationMockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			called = true
			calledWithUserID = userID
			return []models.Segmentation{}, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/456/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !called {
		t.Fatal("expected repository.FindByUserID to be called")
	}

	if calledWithUserID != 456 {
		t.Fatalf("expected repository to be called with user_id 456, got %d", calledWithUserID)
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

// TestIntegration_DataTransformationFlow verifies data transformation through layers
func TestIntegration_DataTransformationFlow(t *testing.T) {
	rawData := []models.Segmentation{
		{
			ID:               1,
			UserID:           789,
			SegmentationType: "DRUG",
			SegmentationName: "Antibióticos",
			Data:             datatypes.JSON(`{"info": "test"}`),
		},
		{
			ID:               2,
			UserID:           789,
			SegmentationType: "SPECIALTY",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{"info": "test"}`),
		},
		{
			ID:               3,
			UserID:           789,
			SegmentationType: "PATIENT",
			SegmentationName: "Crônicos",
			Data:             datatypes.JSON(`{"info": "test"}`),
		},
	}

	mockRepo := &IntegrationMockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 789 {
				return rawData, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/789/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify normalization and grouping worked
	if len(resp.Segmentations["drugs"]) != 1 {
		t.Error("expected normalized 'drugs' key")
	}
	if len(resp.Segmentations["specialties"]) != 1 {
		t.Error("expected normalized 'specialties' key")
	}
	if len(resp.Segmentations["patients"]) != 1 {
		t.Error("expected normalized 'patients' key")
	}
}

// TestIntegration_ErrorPropagation verifies errors bubble up correctly
func TestIntegration_ErrorPropagation(t *testing.T) {
	mockRepo := &IntegrationMockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return nil, context.DeadlineExceeded
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	req := httptest.NewRequest("GET", "/users/999/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Error should propagate to 500 response
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on repository error, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Fatal("expected error field in response")
	}
}

// TestIntegration_MultipleRequestsIndependent verifies requests don't interfere
func TestIntegration_MultipleRequestsIndependent(t *testing.T) {
	user1Data := []models.Segmentation{
		{
			ID:               1,
			UserID:           100,
			SegmentationType: "drug",
			SegmentationName: "Drug1",
			Data:             datatypes.JSON(`{}`),
		},
	}

	user2Data := []models.Segmentation{
		{
			ID:               2,
			UserID:           200,
			SegmentationType: "specialty",
			SegmentationName: "Specialty1",
			Data:             datatypes.JSON(`{}`),
		},
		{
			ID:               3,
			UserID:           200,
			SegmentationType: "specialty",
			SegmentationName: "Specialty2",
			Data:             datatypes.JSON(`{}`),
		},
	}

	mockRepo := &IntegrationMockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 100 {
				return user1Data, nil
			}
			if userID == 200 {
				return user2Data, nil
			}
			return nil, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Request 1
	req1 := httptest.NewRequest("GET", "/users/100/segmentations", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var resp1 service.SegmentationResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)

	// Request 2
	req2 := httptest.NewRequest("GET", "/users/200/segmentations", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var resp2 service.SegmentationResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	// Verify responses are independent
	if resp1.UserID != 100 || len(resp1.Segmentations) != 1 {
		t.Error("request 1 data corrupted")
	}

	if resp2.UserID != 200 || len(resp2.Segmentations) != 1 {
		t.Error("request 2 data corrupted")
	}

	if len(resp2.Segmentations["specialties"]) != 2 {
		t.Error("request 2 should have 2 specialties")
	}
}

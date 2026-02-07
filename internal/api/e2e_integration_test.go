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

// E2EMockRepository simulates a real database for end-to-end tests
type E2EMockRepository struct {
	database map[uint64][]models.Segmentation
	upserts  []models.Segmentation
}

func NewE2EMockRepository() *E2EMockRepository {
	return &E2EMockRepository{
		database: make(map[uint64][]models.Segmentation),
		upserts:  []models.Segmentation{},
	}
}

func (m *E2EMockRepository) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	if data, exists := m.database[userID]; exists {
		return data, nil
	}
	return []models.Segmentation{}, nil
}

func (m *E2EMockRepository) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	m.upserts = append(m.upserts, *s)

	if _, exists := m.database[s.UserID]; !exists {
		m.database[s.UserID] = []models.Segmentation{}
	}

	// Check if already exists (for update scenario)
	for i, existing := range m.database[s.UserID] {
		if existing.UserID == s.UserID &&
			existing.SegmentationType == s.SegmentationType &&
			existing.SegmentationName == s.SegmentationName {
			m.database[s.UserID][i] = *s
			return repository.UpsertUpdated, nil
		}
	}

	// Insert new
	m.database[s.UserID] = append(m.database[s.UserID], *s)
	return repository.UpsertInserted, nil
}

func (m *E2EMockRepository) BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error) {
	m.upserts = append(m.upserts, *s...)

	results := make([]repository.UpsertResult, len(*s))

	for idx, seg := range *s {
		if _, exists := m.database[seg.UserID]; !exists {
			m.database[seg.UserID] = []models.Segmentation{}
		}

		// Check if already exists (for update scenario)
		found := false
		for i, existing := range m.database[seg.UserID] {
			if existing.UserID == seg.UserID &&
				existing.SegmentationType == seg.SegmentationType &&
				existing.SegmentationName == seg.SegmentationName {
				m.database[seg.UserID][i] = seg
				results[idx] = repository.UpsertUpdated
				found = true
				break
			}
		}

		if !found {
			// Insert new
			m.database[seg.UserID] = append(m.database[seg.UserID], seg)
			results[idx] = repository.UpsertInserted
		}
	}

	return results, nil
}

// TestE2E_CompleteWorkflow tests full request-response cycle
func TestE2E_CompleteWorkflow(t *testing.T) {
	mockRepo := NewE2EMockRepository()

	// Seed database
	mockRepo.database[999] = []models.Segmentation{
		{
			ID:               1,
			UserID:           999,
			SegmentationType: "drug",
			SegmentationName: "Aspirin",
			Data:             datatypes.JSON(`{"dosage": "500mg"}`),
		},
		{
			ID:               2,
			UserID:           999,
			SegmentationType: "specialty",
			SegmentationName: "Orthopedia",
			Data:             datatypes.JSON(`{"experience": 10}`),
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Make request
	req := httptest.NewRequest("GET", "/users/999/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.UserID != 999 {
		t.Fatal("wrong user id")
	}

	if len(resp.Segmentations["drugs"]) != 1 {
		t.Fatal("missing drug")
	}

	if len(resp.Segmentations["specialties"]) != 1 {
		t.Fatal("missing specialty")
	}
}

// TestE2E_DataPersistence tests that inserts persist and can be queried
func TestE2E_DataPersistence(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	ctx := context.Background()

	// Insert data via service
	newSeg := &models.Segmentation{
		UserID:           1001,
		SegmentationType: "drug",
		SegmentationName: "Ibuprofen",
		Data:             datatypes.JSON(`{"strength": "200mg"}`),
	}
	svc.Create(ctx, newSeg)

	// Query via API
	req := httptest.NewRequest("GET", "/users/1001/segmentations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp service.SegmentationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Segmentations["drugs"]) != 1 {
		t.Fatal("inserted data not found in query")
	}

	if resp.Segmentations["drugs"][0].Name != "Ibuprofen" {
		t.Fatal("wrong drug name")
	}
}

// TestE2E_MultiUserIsolation verifies users don't see each other's data
func TestE2E_MultiUserIsolation(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	ctx := context.Background()

	// User 2001 data
	svc.Create(ctx, &models.Segmentation{
		UserID:           2001,
		SegmentationType: "drug",
		SegmentationName: "DrugA",
		Data:             datatypes.JSON(`{}`),
	})

	// User 2002 data
	svc.Create(ctx, &models.Segmentation{
		UserID:           2002,
		SegmentationType: "specialty",
		SegmentationName: "SpecB",
		Data:             datatypes.JSON(`{}`),
	})

	// Query user 2001
	req1 := httptest.NewRequest("GET", "/users/2001/segmentations", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var resp1 service.SegmentationResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)

	// Query user 2002
	req2 := httptest.NewRequest("GET", "/users/2002/segmentations", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var resp2 service.SegmentationResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	// Verify isolation
	if len(resp1.Segmentations["drugs"]) != 1 {
		t.Fatal("user 2001 should have drugs")
	}

	if len(resp1.Segmentations["specialties"]) != 0 {
		t.Fatal("user 2001 should not have specialties")
	}

	if len(resp2.Segmentations["drugs"]) != 0 {
		t.Fatal("user 2002 should not have drugs")
	}

	if len(resp2.Segmentations["specialties"]) != 1 {
		t.Fatal("user 2002 should have specialties")
	}
}

// TestE2E_GroupingAndNormalization tests complete data transformation
func TestE2E_GroupingAndNormalization(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)

	ctx := context.Background()

	// Insert mixed case data
	svc.Create(ctx, &models.Segmentation{
		UserID:           3001,
		SegmentationType: "DRUG",
		SegmentationName: "Drug1",
		Data:             datatypes.JSON(`{}`),
	})

	svc.Create(ctx, &models.Segmentation{
		UserID:           3001,
		SegmentationType: "drug",
		SegmentationName: "Drug2",
		Data:             datatypes.JSON(`{}`),
	})

	svc.Create(ctx, &models.Segmentation{
		UserID:           3001,
		SegmentationType: "SPECIALTY",
		SegmentationName: "Spec1",
		Data:             datatypes.JSON(`{}`),
	})

	svc.Create(ctx, &models.Segmentation{
		UserID:           3001,
		SegmentationType: "specialty",
		SegmentationName: "Spec2",
		Data:             datatypes.JSON(`{}`),
	})

	svc.Create(ctx, &models.Segmentation{
		UserID:           3001,
		SegmentationType: "PATIENT",
		SegmentationName: "Patient1",
		Data:             datatypes.JSON(`{}`),
	})

	// Query and verify grouping + normalization
	result, _ := svc.GetByUserID(ctx, 3001)

	if len(result.Segmentations) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(result.Segmentations))
	}

	if len(result.Segmentations["drugs"]) != 2 {
		t.Fatalf("expected 2 drugs, got %d", len(result.Segmentations["drugs"]))
	}

	if len(result.Segmentations["specialties"]) != 2 {
		t.Fatalf("expected 2 specialties, got %d", len(result.Segmentations["specialties"]))
	}

	if len(result.Segmentations["patients"]) != 1 {
		t.Fatalf("expected 1 patient, got %d", len(result.Segmentations["patients"]))
	}
}

// TestE2E_UpsertBehavior tests insert vs update behavior
func TestE2E_UpsertBehavior(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)

	ctx := context.Background()

	// First insert
	seg1 := &models.Segmentation{
		UserID:           4001,
		SegmentationType: "drug",
		SegmentationName: "Drug1",
		Data:             datatypes.JSON(`{"v": "1"}`),
	}

	result1, _ := svc.Create(ctx, seg1)

	// Second insert same key, different data (should update)
	seg2 := &models.Segmentation{
		UserID:           4001,
		SegmentationType: "drug",
		SegmentationName: "Drug1",
		Data:             datatypes.JSON(`{"v": "2"}`),
	}

	result2, _ := svc.Create(ctx, seg2)

	// Verify behavior
	if result1 != repository.UpsertInserted {
		t.Fatalf("first insert should return UpsertInserted, got %v", result1)
	}

	if result2 != repository.UpsertUpdated {
		t.Fatalf("second insert (duplicate key) should return UpsertUpdated, got %v", result2)
	}

	// Verify database has only one record for this user-type-name combo
	result, _ := svc.GetByUserID(ctx, 4001)
	if len(result.Segmentations["drugs"]) != 1 {
		t.Fatalf("expected 1 drug after upsert, got %d", len(result.Segmentations["drugs"]))
	}
}

// TestE2E_ErrorRecovery tests system handles errors gracefully
func TestE2E_ErrorRecovery(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)
	router := SetupRouter(svc)

	// Insert valid data
	ctx := context.Background()
	svc.Create(ctx, &models.Segmentation{
		UserID:           5001,
		SegmentationType: "drug",
		SegmentationName: "Drug1",
		Data:             datatypes.JSON(`{}`),
	})

	// Query valid user
	req1 := httptest.NewRequest("GET", "/users/5001/segmentations", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatal("valid query should succeed")
	}

	// Query invalid format
	req2 := httptest.NewRequest("GET", "/users/invalid/segmentations", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatal("invalid user_id should return 400")
	}

	// Query non-existent user (should still work, return empty)
	req3 := httptest.NewRequest("GET", "/users/9999/segmentations", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatal("non-existent user should return 200")
	}
}

// TestE2E_SequentialOperations verifies multiple operations work correctly
func TestE2E_SequentialOperations(t *testing.T) {
	mockRepo := NewE2EMockRepository()
	svc := service.NewSegmentationService(mockRepo)

	ctx := context.Background()

	// Operation 1: Create drug
	svc.Create(ctx, &models.Segmentation{
		UserID:           6001,
		SegmentationType: "drug",
		SegmentationName: "Drug1",
		Data:             datatypes.JSON(`{}`),
	})

	result1, _ := svc.GetByUserID(ctx, 6001)
	if len(result1.Segmentations) != 1 {
		t.Fatal("after first insert, should have 1 type")
	}

	// Operation 2: Create specialty
	svc.Create(ctx, &models.Segmentation{
		UserID:           6001,
		SegmentationType: "specialty",
		SegmentationName: "Spec1",
		Data:             datatypes.JSON(`{}`),
	})

	result2, _ := svc.GetByUserID(ctx, 6001)
	if len(result2.Segmentations) != 2 {
		t.Fatal("after second insert, should have 2 types")
	}

	// Operation 3: Create patient
	svc.Create(ctx, &models.Segmentation{
		UserID:           6001,
		SegmentationType: "patient",
		SegmentationName: "Patient1",
		Data:             datatypes.JSON(`{}`),
	})

	result3, _ := svc.GetByUserID(ctx, 6001)
	if len(result3.Segmentations) != 3 {
		t.Fatal("after third insert, should have 3 types")
	}

	// Verify all data is present
	if len(result3.Segmentations["drugs"]) != 1 {
		t.Fatal("should have 1 drug")
	}
	if len(result3.Segmentations["specialties"]) != 1 {
		t.Fatal("should have 1 specialty")
	}
	if len(result3.Segmentations["patients"]) != 1 {
		t.Fatal("should have 1 patient")
	}
}

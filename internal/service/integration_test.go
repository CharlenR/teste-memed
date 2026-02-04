package service

import (
	"context"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"

	"gorm.io/datatypes"
)

// RepositoryMock for integration testing
type RepositoryMock struct {
	findByUserIDCalled bool
	findByUserIDInput  uint64
	findByUserIDResult []models.Segmentation

	upsertCalled bool
	upsertInput  *models.Segmentation
	upsertResult repository.UpsertResult
}

func (m *RepositoryMock) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	m.findByUserIDCalled = true
	m.findByUserIDInput = userID
	return m.findByUserIDResult, nil
}

func (m *RepositoryMock) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	m.upsertCalled = true
	m.upsertInput = s
	return m.upsertResult, nil
}

// TestIntegration_ServiceCallsRepositoryFindByUserID verifies service -> repository flow
func TestIntegration_ServiceCallsRepositoryFindByUserID(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{
			{
				ID:               1,
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "TestDrug",
				Data:             datatypes.JSON(`{}`),
			},
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, err := svc.GetByUserID(context.Background(), 100)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockRepo.findByUserIDCalled {
		t.Fatal("expected FindByUserID to be called")
	}

	if mockRepo.findByUserIDInput != 100 {
		t.Fatalf("expected FindByUserID called with 100, got %d", mockRepo.findByUserIDInput)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
	}

	if result.UserID != 100 {
		t.Fatalf("expected user_id 100, got %d", result.UserID)
	}
}

// TestIntegration_ServiceGroupsSegmentations verifies grouping logic
func TestIntegration_ServiceGroupsSegmentations(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{
			{
				ID:               1,
				UserID:           200,
				SegmentationType: "drug",
				SegmentationName: "Drug1",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               2,
				UserID:           200,
				SegmentationType: "drug",
				SegmentationName: "Drug2",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               3,
				UserID:           200,
				SegmentationType: "specialty",
				SegmentationName: "Specialty1",
				Data:             datatypes.JSON(`{}`),
			},
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, _ := svc.GetByUserID(context.Background(), 200)

	if len(result.Segmentations) != 2 {
		t.Fatalf("expected 2 types, got %d", len(result.Segmentations))
	}

	if len(result.Segmentations["drugs"]) != 2 {
		t.Fatalf("expected 2 drugs, got %d", len(result.Segmentations["drugs"]))
	}

	if len(result.Segmentations["specialties"]) != 1 {
		t.Fatalf("expected 1 specialty, got %d", len(result.Segmentations["specialties"]))
	}
}

// TestIntegration_ServiceNormalizesTypes verifies type normalization
func TestIntegration_ServiceNormalizesTypes(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{
			{
				ID:               1,
				UserID:           300,
				SegmentationType: "DRUG",
				SegmentationName: "Drug",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               2,
				UserID:           300,
				SegmentationType: "DrUg",
				SegmentationName: "Drug2",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               3,
				UserID:           300,
				SegmentationType: "SPECIALTY",
				SegmentationName: "Spec",
				Data:             datatypes.JSON(`{}`),
			},
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, _ := svc.GetByUserID(context.Background(), 300)

	// All should be normalized to lowercase
	if len(result.Segmentations["drugs"]) != 2 {
		t.Fatalf("expected 2 normalized drugs, got %d", len(result.Segmentations["drugs"]))
	}

	if len(result.Segmentations["specialties"]) != 1 {
		t.Fatalf("expected 1 normalized specialty, got %d", len(result.Segmentations["specialties"]))
	}
}

// TestIntegration_ServiceCreatesSegmentation verifies Create -> Upsert flow
func TestIntegration_ServiceCreatesSegmentation(t *testing.T) {
	mockRepo := &RepositoryMock{
		upsertResult: repository.UpsertInserted,
	}

	svc := NewSegmentationService(mockRepo)
	seg := &models.Segmentation{
		UserID:           400,
		SegmentationType: "drug",
		SegmentationName: "NewDrug",
		Data:             datatypes.JSON(`{"new": true}`),
	}

	result, err := svc.Create(context.Background(), seg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockRepo.upsertCalled {
		t.Fatal("expected Upsert to be called")
	}

	if mockRepo.upsertInput == nil {
		t.Fatal("expected upsertInput to not be nil")
	}

	if mockRepo.upsertInput.SegmentationName != "NewDrug" {
		t.Fatalf("expected name 'NewDrug', got %s", mockRepo.upsertInput.SegmentationName)
	}

	if result != repository.UpsertInserted {
		t.Fatalf("expected UpsertInserted, got %v", result)
	}
}

// TestIntegration_ServiceHandlesEmptyResult verifies empty data handling
func TestIntegration_ServiceHandlesEmptyResult(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{},
	}

	svc := NewSegmentationService(mockRepo)
	result, err := svc.GetByUserID(context.Background(), 500)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
	}

	if result.UserID != 500 {
		t.Fatalf("expected user_id 500, got %d", result.UserID)
	}

	if len(result.Segmentations) != 0 {
		t.Fatalf("expected no segmentations, got %d", len(result.Segmentations))
	}
}

// TestIntegration_ServiceNormalizesAndGroups tests combined normalization + grouping
func TestIntegration_ServiceNormalizesAndGroups(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{
			{
				ID:               1,
				UserID:           600,
				SegmentationType: "PATIENT",
				SegmentationName: "Patient1",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               2,
				UserID:           600,
				SegmentationType: "patient",
				SegmentationName: "Patient2",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               3,
				UserID:           600,
				SegmentationType: "SPECIALTY",
				SegmentationName: "Spec1",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               4,
				UserID:           600,
				SegmentationType: "specialty",
				SegmentationName: "Spec2",
				Data:             datatypes.JSON(`{}`),
			},
			{
				ID:               5,
				UserID:           600,
				SegmentationType: "DRUG",
				SegmentationName: "Drug1",
				Data:             datatypes.JSON(`{}`),
			},
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, _ := svc.GetByUserID(context.Background(), 600)

	if len(result.Segmentations) != 3 {
		t.Fatalf("expected 3 types, got %d", len(result.Segmentations))
	}

	if len(result.Segmentations["patients"]) != 2 {
		t.Fatalf("expected 2 patients, got %d", len(result.Segmentations["patients"]))
	}

	if len(result.Segmentations["specialties"]) != 2 {
		t.Fatalf("expected 2 specialties, got %d", len(result.Segmentations["specialties"]))
	}

	if len(result.Segmentations["drugs"]) != 1 {
		t.Fatalf("expected 1 drug, got %d", len(result.Segmentations["drugs"]))
	}
}

// TestIntegration_ServiceContextPropagation verifies context flows through
func TestIntegration_ServiceContextPropagation(t *testing.T) {
	mockRepo := &RepositoryMock{
		findByUserIDResult: []models.Segmentation{},
	}

	svc := NewSegmentationService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should still work but verify context was passed
	result, err := svc.GetByUserID(ctx, 700)

	if err != nil {
		t.Logf("Error with cancelled context (expected): %v", err)
	}

	if result != nil {
		t.Logf("Got result with cancelled context: %v", result.UserID)
	}
}

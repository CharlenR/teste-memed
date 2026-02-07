package service

import (
	"context"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"

	"gorm.io/datatypes"
)

type MockRepository struct {
	findByUserIDFunc func(ctx context.Context, userID uint64) ([]models.Segmentation, error)
	upsertFunc       func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error)
	bulkUpsertFunc   func(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, s)
	}
	return repository.UpsertNoOp, nil
}

func (m *MockRepository) BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error) {
	if m.bulkUpsertFunc != nil {
		return m.bulkUpsertFunc(ctx, s)
	}
	return nil, nil
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "specialty", expected: "specialties"},
		{input: "SPECIALTY", expected: "specialties"},
		{input: "drug", expected: "drugs"},
		{input: "patient", expected: "patients"},
		{input: "custom", expected: "customs"},
		{input: "", expected: "s"},
	}

	for _, tt := range tests {
		t.Run("normalize_"+tt.input, func(t *testing.T) {
			result := normalizeType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSegmentationServiceGetByUserID(t *testing.T) {
	ctx := context.Background()

	mockRecords := []models.Segmentation{
		{
			ID:               1,
			UserID:           100,
			SegmentationType: "drug",
			SegmentationName: "Antibióticos",
			Data:             datatypes.JSON(`{"type": "antibiotic"}`),
		},
		{
			ID:               2,
			UserID:           100,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{"years": 15}`),
		},
	}

	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			if userID == 100 {
				return mockRecords, nil
			}
			return nil, nil
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, err := svc.GetByUserID(ctx, 100)

	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if result.UserID != 100 {
		t.Errorf("UserID = %d, want 100", result.UserID)
	}

	totalCount := 0
	for _, items := range result.Segmentations {
		totalCount += len(items)
	}
	if totalCount != 2 {
		t.Errorf("Expected 2 segmentations, got %d", totalCount)
	}
}

func TestSegmentationServiceGetByUserIDGrouping(t *testing.T) {
	ctx := context.Background()

	mockRecords := []models.Segmentation{
		{
			ID:               1,
			UserID:           100,
			SegmentationType: "drug",
			SegmentationName: "Antibióticos",
			Data:             datatypes.JSON(`{"type": "antibiotic"}`),
		},
		{
			ID:               2,
			UserID:           100,
			SegmentationType: "drug",
			SegmentationName: "Analgésicos",
			Data:             datatypes.JSON(`{"type": "analgesic"}`),
		},
		{
			ID:               3,
			UserID:           100,
			SegmentationType: "specialty",
			SegmentationName: "Cardiologia",
			Data:             datatypes.JSON(`{"years": 15}`),
		},
	}

	mockRepo := &MockRepository{
		findByUserIDFunc: func(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
			return mockRecords, nil
		},
	}

	svc := NewSegmentationService(mockRepo)
	result, err := svc.GetByUserID(ctx, 100)

	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if len(result.Segmentations) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(result.Segmentations))
	}

	if drugs, ok := result.Segmentations["drugs"]; ok {
		if len(drugs) != 2 {
			t.Errorf("Expected 2 drugs, got %d", len(drugs))
		}
	} else {
		t.Error("drugs group not found")
	}

	if specialties, ok := result.Segmentations["specialties"]; ok {
		if len(specialties) != 1 {
			t.Errorf("Expected 1 specialty, got %d", len(specialties))
		}
	} else {
		t.Error("specialties group not found")
	}
}

func TestSegmentationServiceCreate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		seg            *models.Segmentation
		mockResult     repository.UpsertResult
		expectedResult repository.UpsertResult
	}{
		{
			name: "create new segmentation",
			seg: &models.Segmentation{
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{"type": "antibiotic"}`),
			},
			mockResult:     repository.UpsertInserted,
			expectedResult: repository.UpsertInserted,
		},
		{
			name: "update existing segmentation",
			seg: &models.Segmentation{
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{"type": "antibiotic", "updated": true}`),
			},
			mockResult:     repository.UpsertUpdated,
			expectedResult: repository.UpsertUpdated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{
				upsertFunc: func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
					return tt.mockResult, nil
				},
			}

			svc := NewSegmentationService(mockRepo)
			result, err := svc.Create(ctx, tt.seg)

			if err != nil {
				t.Errorf("Create() error = %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("Create() result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

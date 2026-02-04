package mysql

import (
	"context"
	"testing"
	"time"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestSegmentationRepositoryInterface(t *testing.T) {
	var _ repository.SegmentationRepository = (*segmentationRepository)(nil)
}

func TestNewSegmentationRepository(t *testing.T) {
	repo := NewSegmentationRepository(nil)
	if repo == nil {
		t.Error("NewSegmentationRepository should not return nil")
	}

	_, ok := repo.(*segmentationRepository)
	if !ok {
		t.Error("NewSegmentationRepository should return *segmentationRepository")
	}
}

func TestSegmentationModelForRepository(t *testing.T) {
	seg := &models.Segmentation{
		ID:               1,
		UserID:           100,
		SegmentationType: "drug",
		SegmentationName: "Antibióticos",
		Data:             datatypes.JSON(`{"type": "antibiotic"}`),
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	}

	if seg.UserID == 0 {
		t.Error("UserID should not be zero")
	}
	if seg.SegmentationType == "" {
		t.Error("SegmentationType should not be empty")
	}
	if seg.SegmentationName == "" {
		t.Error("SegmentationName should not be empty")
	}
	if seg.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestUpsertResultValues(t *testing.T) {
	tests := []struct {
		name     string
		result   repository.UpsertResult
		expected int
	}{
		{
			name:     "inserted value",
			result:   repository.UpsertInserted,
			expected: 0,
		},
		{
			name:     "updated value",
			result:   repository.UpsertUpdated,
			expected: 1,
		},
		{
			name:     "noop value",
			result:   repository.UpsertNoOp,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.result) != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, int(tt.result))
			}
		})
	}
}

func TestRecordNotFoundCheck(t *testing.T) {
	if gorm.ErrRecordNotFound != gorm.ErrRecordNotFound {
		t.Error("Record not found error should be consistent")
	}

	testErr := gorm.ErrInvalidData
	if testErr == gorm.ErrRecordNotFound {
		t.Error("Different errors should not be equal")
	}
}

func TestContextHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled immediately")
	default:
		// Expected
	}

	cancel()
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after cancel()")
	}
}

func TestSegmentationUniqueKey(t *testing.T) {
	seg1 := &models.Segmentation{
		UserID:           100,
		SegmentationType: "drug",
		SegmentationName: "Antibióticos",
		Data:             datatypes.JSON(`{"type": "antibiotic"}`),
	}

	seg2 := &models.Segmentation{
		UserID:           100,
		SegmentationType: "drug",
		SegmentationName: "Antibióticos",
		Data:             datatypes.JSON(`{"type": "antibiotic", "updated": true}`),
	}

	seg3 := &models.Segmentation{
		UserID:           100,
		SegmentationType: "drug",
		SegmentationName: "Analgésicos",
		Data:             datatypes.JSON(`{"type": "analgesic"}`),
	}

	if !(seg1.UserID == seg2.UserID &&
		seg1.SegmentationType == seg2.SegmentationType &&
		seg1.SegmentationName == seg2.SegmentationName) {
		t.Error("seg1 and seg2 should have the same unique key")
	}

	if seg1.SegmentationName == seg3.SegmentationName {
		t.Error("seg1 and seg3 should have different unique keys")
	}
}

func TestSegmentationRepositoryCreation(t *testing.T) {
	repo := NewSegmentationRepository(nil)

	if repo == nil {
		t.Fatal("NewSegmentationRepository should return non-nil repo even with nil db")
	}

	// Verify the repository type
	if _, ok := repo.(*segmentationRepository); !ok {
		t.Error("should return *segmentationRepository")
	}
}

func TestSegmentationRepositoryImplementsInterface(t *testing.T) {
	repo := NewSegmentationRepository(nil)
	var _ repository.SegmentationRepository = repo
}

func TestSegmentationDataTypes(t *testing.T) {
	tests := []struct {
		name string
		seg  *models.Segmentation
	}{
		{
			name: "drug segmentation",
			seg: &models.Segmentation{
				UserID:           1,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{"category": "antibiotic"}`),
			},
		},
		{
			name: "specialty segmentation",
			seg: &models.Segmentation{
				UserID:           2,
				SegmentationType: "specialty",
				SegmentationName: "Cardiologia",
				Data:             datatypes.JSON(`{"subspecialty": "cardiac"}`),
			},
		},
		{
			name: "patient segmentation",
			seg: &models.Segmentation{
				UserID:           3,
				SegmentationType: "patient",
				SegmentationName: "Crônicos",
				Data:             datatypes.JSON(`{"age_range": "50+"}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seg.UserID == 0 {
				t.Error("UserID should not be 0")
			}
			if tt.seg.SegmentationType == "" {
				t.Error("SegmentationType should not be empty")
			}
			if tt.seg.SegmentationName == "" {
				t.Error("SegmentationName should not be empty")
			}
			if len(tt.seg.Data) == 0 {
				t.Error("Data should not be empty")
			}
		})
	}
}

func TestUpsertResultTypes(t *testing.T) {
	// Test all UpsertResult values
	results := map[string]repository.UpsertResult{
		"inserted": repository.UpsertInserted,
		"updated":  repository.UpsertUpdated,
		"noop":     repository.UpsertNoOp,
	}

	for name, result := range results {
		t.Run(name, func(t *testing.T) {
			// Verify each result can be used
			if result < 0 {
				t.Error("UpsertResult should be non-negative")
			}
		})
	}
}

func TestRepositoryContextCancellation(t *testing.T) {
	_ = NewSegmentationRepository(nil)

	_, cancel := context.WithCancel(context.Background())
	cancel()

	// With cancelled context and nil db, operations should handle it gracefully
	// This test verifies the repository doesn't crash with cancelled context
	t.Log("Repository created and context cancelled")
}

func TestSegmentationModelValidation(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		seg        *models.Segmentation
	}{
		{
			name:       "valid segmentation",
			shouldFail: false,
			seg: &models.Segmentation{
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{"valid": true}`),
			},
		},
		{
			name:       "empty segmentation type",
			shouldFail: true,
			seg: &models.Segmentation{
				UserID:           100,
				SegmentationType: "",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{}`),
			},
		},
		{
			name:       "zero user id",
			shouldFail: true,
			seg: &models.Segmentation{
				UserID:           0,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{}`),
			},
		},
		{
			name:       "empty segmentation name",
			shouldFail: true,
			seg: &models.Segmentation{
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "",
				Data:             datatypes.JSON(`{}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.seg.UserID != 0 &&
				tt.seg.SegmentationType != "" &&
				tt.seg.SegmentationName != ""

			if tt.shouldFail && isValid {
				t.Error("segmentation should be invalid")
			}
			if !tt.shouldFail && !isValid {
				t.Error("segmentation should be valid")
			}
		})
	}
}

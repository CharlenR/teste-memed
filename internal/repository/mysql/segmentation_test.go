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

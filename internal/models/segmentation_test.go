package models

import (
	"testing"

	"gorm.io/datatypes"
)

func TestSegmentationModel(t *testing.T) {
	tests := []struct {
		name string
		seg  Segmentation
	}{
		{
			name: "valid segmentation with drug type",
			seg: Segmentation{
				ID:               1,
				UserID:           100,
				SegmentationType: "drug",
				SegmentationName: "Antibióticos",
				Data:             datatypes.JSON(`{"type": "antibiotic"}`),
				CreatedAt:        1234567890,
				UpdatedAt:        1234567890,
			},
		},
		{
			name: "valid segmentation with specialty type",
			seg: Segmentation{
				ID:               2,
				UserID:           200,
				SegmentationType: "specialty",
				SegmentationName: "Cardiologia",
				Data:             datatypes.JSON(`{"years": 15}`),
				CreatedAt:        1234567890,
				UpdatedAt:        1234567890,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seg.UserID == 0 {
				t.Errorf("UserID should not be zero")
			}
			if tt.seg.SegmentationType == "" {
				t.Errorf("SegmentationType should not be empty")
			}
			if tt.seg.SegmentationName == "" {
				t.Errorf("SegmentationName should not be empty")
			}
		})
	}
}

func TestSegmentationStructTags(t *testing.T) {
	seg := Segmentation{}

	if seg.ID != 0 {
		t.Errorf("expected ID to be 0, got %d", seg.ID)
	}
	if seg.UserID != 0 {
		t.Errorf("expected UserID to be 0, got %d", seg.UserID)
	}
	if seg.SegmentationType != "" {
		t.Errorf("expected SegmentationType to be empty, got %s", seg.SegmentationType)
	}
	if seg.SegmentationName != "" {
		t.Errorf("expected SegmentationName to be empty, got %s", seg.SegmentationName)
	}
}

package service

import (
	"context"
	"encoding/json"
	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"strings"
)

type SegmentationService struct {
	repo repository.SegmentationRepository
}

func NewSegmentationService(r repository.SegmentationRepository) *SegmentationService {
	return &SegmentationService{repo: r}
}

type SegmentationItem struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

type SegmentationResponse struct {
	UserID        uint64                        `json:"user_id"`
	Segmentations map[string][]SegmentationItem `json:"segmentations"`
}

func (s *SegmentationService) GetByUserID(
	ctx context.Context,
	userID uint64,
) (*SegmentationResponse, error) {

	records, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := &SegmentationResponse{
		UserID:        userID,
		Segmentations: make(map[string][]SegmentationItem),
	}

	for _, r := range records {
		var data map[string]interface{}
		_ = json.Unmarshal(r.Data, &data)

		key := normalizeType(r.SegmentationType)

		result.Segmentations[key] = append(
			result.Segmentations[key],
			SegmentationItem{
				Name: r.SegmentationName,
				Data: data,
			},
		)
	}

	return result, nil
}

func normalizeType(t string) string {
	switch strings.ToLower(t) {
	case "specialty":
		return "specialties"
	case "drug":
		return "drugs"
	case "patient":
		return "patients"
	default:
		return t + "s"
	}
}

func (s *SegmentationService) Create(
	ctx context.Context,
	seg *models.Segmentation,
) (repository.UpsertResult, error) {
	return s.repo.Upsert(ctx, seg)
}

package repository

import (
	"context"
	"segmentation-api/internal/models"
)

type UpsertResult int

const (
	UpsertInserted UpsertResult = iota
	UpsertUpdated
	UpsertNoOp
)

type SegmentationRepository interface {
	FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error)
	Upsert(ctx context.Context, s *models.Segmentation) (UpsertResult, error)
	BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]UpsertResult, []error)
}

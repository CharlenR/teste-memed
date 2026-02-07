package mysql

import (
	"context"
	"log"

	// "log"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

type segmentationRepository struct {
	db *gorm.DB
}

func NewSegmentationRepository(db *gorm.DB) repository.SegmentationRepository {
	return &segmentationRepository{db: db}
}

func (r *segmentationRepository) FindByUserID(
	ctx context.Context,
	userID uint64,
) ([]models.Segmentation, error) {

	var segs []models.Segmentation

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("segmentation_type, segmentation_name").
		Find(&segs).Error

	return segs, err
}

func (r *segmentationRepository) Upsert(
	ctx context.Context,
	s *models.Segmentation,
) (repository.UpsertResult, error) {

	// tx := r.db.WithContext(ctx).
	// 	Clauses(clause.OnConflict{
	// 		Columns: []clause.Column{
	// 			{Name: "user_id"},
	// 			{Name: "segmentation_type"},
	// 			{Name: "segmentation_name"},
	// 		},
	// 		DoUpdates: clause.Assignments(map[string]interface{}{
	// 			"data":       s.Data,
	// 			"updated_at": time.Now().Unix(),
	// 		}),
	// 	}).
	// 	Create(s)

	tx := r.db.WithContext(ctx).Exec(`
	INSERT INTO segmentations
	(user_id, segmentation_type, segmentation_name, data, updated_at)
	VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
	data = VALUES(data),
	updated_at = VALUES(updated_at)
	`,
		s.UserID,
		s.SegmentationType,
		s.SegmentationName,
		s.Data,
		time.Now().Unix(),
	)

	if tx.Error != nil {
		log.Printf(
			"upsert_error user_id=%d seg_type=%s seg_name=%s error=%v",
			s.UserID, s.SegmentationType, s.SegmentationName, tx.Error,
		)
		return repository.UpsertNoOp, tx.Error
	}

	// MySQL trick (important)
	if tx.RowsAffected == 1 {
		return repository.UpsertInserted, nil
	}
	return repository.UpsertUpdated, nil
}

func (r *segmentationRepository) BulkUpsert(
	ctx context.Context,
	items *[]models.Segmentation,
) ([]repository.UpsertResult, []error) {

	results := make([]repository.UpsertResult, len(*items))
	errors := make([]error, len(*items))

	for i, item := range *items {
		result, err := r.Upsert(ctx, &item)
		results[i] = result
		errors[i] = err
	}

	return results, errors
}

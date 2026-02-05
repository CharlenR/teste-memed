package mysql

import (
	"context"
	"log"
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

	// Check if record exists
	var existing models.Segmentation
	existsQuery := r.db.WithContext(ctx).
		Where("user_id = ? AND segmentation_type = ? AND segmentation_name = ?",
			s.UserID, s.SegmentationType, s.SegmentationName).
		First(&existing)

	if existsQuery.Error == gorm.ErrRecordNotFound {
		// Record doesn't exist, insert it
		tx := r.db.WithContext(ctx).Create(s)
		if tx.Error != nil {
			log.Printf("upsert_error user_id=%d seg_type=%s seg_name=%s error=%v",
				s.UserID, s.SegmentationType, s.SegmentationName, tx.Error)
			return repository.UpsertNoOp, tx.Error
		}
		//log.Printf("upsert_debug user_id=%d seg_type=%s seg_name=%s action=inserted",
		//s.UserID, s.SegmentationType, s.SegmentationName)
		return repository.UpsertInserted, nil
	}

	if existsQuery.Error != nil {
		log.Printf("upsert_error user_id=%d seg_type=%s seg_name=%s error=%v",
			s.UserID, s.SegmentationType, s.SegmentationName, existsQuery.Error)
		return repository.UpsertNoOp, existsQuery.Error
	}

	// Record exists, update it
	s.ID = existing.ID
	tx := r.db.WithContext(ctx).Model(s).Updates(map[string]interface{}{
		"data":       s.Data,
		"updated_at": time.Now().Unix(),
	})

	if tx.Error != nil {
		log.Printf("upsert_error user_id=%d seg_type=%s seg_name=%s error=%v",
			s.UserID, s.SegmentationType, s.SegmentationName, tx.Error)
		return repository.UpsertNoOp, tx.Error
	}

	//log.Printf("upsert_debug user_id=%d seg_type=%s seg_name=%s action=updated",
	//s.UserID, s.SegmentationType, s.SegmentationName)
	return repository.UpsertUpdated, nil
}

package models

import "gorm.io/datatypes"

type Segmentation struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement"`
	UserID           uint64         `gorm:"not null;uniqueIndex:uniq_user_seg"`
	SegmentationType string         `gorm:"size:50;not null;uniqueIndex:uniq_user_seg"`
	SegmentationName string         `gorm:"size:100;not null;uniqueIndex:uniq_user_seg"`
	Data             datatypes.JSON `gorm:"type:json"`
	CreatedAt        int64
	UpdatedAt        int64
}

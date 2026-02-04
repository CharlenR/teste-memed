package mysql

import (
	"segmentation-api/internal/models"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Segmentation{},
	)
}

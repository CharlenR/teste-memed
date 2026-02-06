package mysql

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(gormLogger logger.Interface) (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")

	if host == "" || name == "" || user == "" {
		return nil, fmt.Errorf("database env vars not set")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		user,
		pass,
		host,
		port,
		name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:      gormLogger,
		PrepareStmt: true,
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// sqlDB.SetMaxOpenConns(50)
	// sqlDB.SetMaxIdleConns(10)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	// 2gb
	// sqlDB.SetMaxOpenConns(80)
	// sqlDB.SetMaxIdleConns(80)
	// sqlDB.SetConnMaxLifetime(2 * time.Minute)

	// 4gb
	// sqlDB.SetMaxOpenConns(32)
	// sqlDB.SetMaxIdleConns(32)
	// sqlDB.SetConnMaxLifetime(60 * time.Minute)

	sqlDB.SetMaxOpenConns(32)
	sqlDB.SetMaxIdleConns(32)
	sqlDB.SetConnMaxLifetime(30 * time.Second)

	// 👇 garante DB disponível antes de subir worker
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

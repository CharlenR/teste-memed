package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gormLogger "gorm.io/gorm/logger"

	"segmentation-api/internal/processor"
	"segmentation-api/internal/repository/mysql"
	"segmentation-api/internal/service"
)

func message() {
	fmt.Println("SEGMENTATION PROCESSOR")
}

func main() {
	message()

	// ─────────────────────────────────────────────
	// Logs
	// ─────────────────────────────────────────────
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatal(err)
	}

	logPath := fmt.Sprintf(
		"%s/%s-processor.log",
		logDir,
		time.Now().Format("2006-01-02T15-04-05"),
	)

	logFile, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	// Write to both stdout (for docker-compose logs) and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// logger base (both stdout and file)
	fileLogger := log.New(
		multiWriter,
		"",
		log.LstdFlags|log.Lmicroseconds,
	)

	// ─────────────────────────────────────────────
	// GORM logger (arquivo only, sem spam)
	// ─────────────────────────────────────────────
	gormLog := gormLogger.New(
		fileLogger,
		gormLogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormLogger.Warn, // 🔥 SEM INSERT OK
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// ─────────────────────────────────────────────
	// Context + signals
	// ─────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// ─────────────────────────────────────────────
	// Database
	// ─────────────────────────────────────────────
	db, err := mysql.NewMySQL(gormLog)
	if err != nil {
		fileLogger.Fatalf("db_init_error=%v", err)
	}

	if err := mysql.RunMigrations(db); err != nil {
		fileLogger.Fatalf("migration_error=%v", err)
	}

	// ─────────────────────────────────────────────
	// Service wiring
	// ─────────────────────────────────────────────
	repo := mysql.NewSegmentationRepository(db)
	svc := service.NewSegmentationService(repo)

	// ─────────────────────────────────────────────
	// Processor
	// ─────────────────────────────────────────────
	fileLogger.Println("processor_started")

	if err := processor.Run(ctx, svc, fileLogger); err != nil {
		fileLogger.Fatalf("processor_error=%v", err)
	}

	fileLogger.Println("processor_finished_successfully")
}

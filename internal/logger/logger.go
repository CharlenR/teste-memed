package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func New() (*log.Logger, *os.File, error) {
	logDir := os.Getenv("LOG_DIR")
	logOut := os.Getenv("PRINTLOG")
	if logDir == "" {
		logDir = "./logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, err
	}

	filename := time.Now().Format("2006-01-02T15-04-05") + "-processor.log"
	path := filepath.Join(logDir, filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}
	var multi io.Writer
	multi = file
	if logOut == "true" {
		multi = io.MultiWriter(os.Stdout, file)
	}

	logger := log.New(multi, "", log.LstdFlags|log.Lmicroseconds)

	return logger, file, nil
}

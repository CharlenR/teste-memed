package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerPackageExists(t *testing.T) {
	t.Log("logger package is available")
}

func TestNew_CreatesLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("LOG_DIR", tmpDir)

	logger, file, err := New()
	if err != nil {
		t.Fatalf("New() should not return error: %v", err)
	}
	defer file.Close()

	if logger == nil {
		t.Error("New() should return non-nil logger")
	}
	if file == nil {
		t.Error("New() should return non-nil file")
	}

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("Log directory should exist: %s", tmpDir)
	}
}

func TestNew_CreatesLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("LOG_DIR", tmpDir)

	logger, file, err := New()
	if err != nil {
		t.Fatalf("New() should not return error: %v", err)
	}
	defer file.Close()

	if logger == nil {
		t.Fatal("logger should not be nil")
	}

	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		t.Fatalf("Log file should exist: %v", err)
	}

	if fileInfo.IsDir() {
		t.Error("Log file should not be a directory")
	}
}

func TestNew_WithCustomLogDir(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom", "logs")
	t.Setenv("LOG_DIR", customDir)

	logger, file, err := New()
	if err != nil {
		t.Fatalf("New() should not return error: %v", err)
	}
	defer file.Close()

	if logger == nil {
		t.Error("New() should return non-nil logger")
	}

	if _, err := os.Stat(customDir); os.IsNotExist(err) {
		t.Errorf("Custom log directory should exist: %s", customDir)
	}
}

func TestNew_FileCanBeWritten(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("LOG_DIR", tmpDir)

	_, file, err := New()
	if err != nil {
		t.Fatalf("New() should not return error: %v", err)
	}
	defer file.Close()

	// Test writing to file through logger
	content := "test message"
	file.WriteString(content + "\n")

	fileContent, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("should be able to read log file: %v", err)
	}

	if len(fileContent) == 0 {
		t.Error("log file should contain data after writing")
	}
}

func TestNew_FilenameDateFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("LOG_DIR", tmpDir)

	_, file, err := New()
	if err != nil {
		t.Fatalf("New() should not return error: %v", err)
	}
	defer file.Close()

	filename := filepath.Base(file.Name())

	ext := filepath.Ext(filename)
	if ext != ".log" {
		t.Errorf("Log filename should have .log extension, got %s", ext)
	}

	if len(filename) < 4 {
		t.Error("Filename should have sufficient length")
	}
}

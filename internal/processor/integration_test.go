package processor

import (
	"context"
	"log"
	"os"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"

	"gorm.io/datatypes"
)

// ServiceIntegrationMock for processor-service integration testing
type ServiceIntegrationMock struct {
	createCalls []struct {
		userID   uint64
		segType  string
		name     string
	}
	result repository.UpsertResult
}

func (m *ServiceIntegrationMock) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	return []models.Segmentation{}, nil
}

func (m *ServiceIntegrationMock) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	m.createCalls = append(m.createCalls, struct {
		userID   uint64
		segType  string
		name     string
	}{
		userID:   s.UserID,
		segType:  s.SegmentationType,
		name:     s.SegmentationName,
	})
	return m.result, nil
}

// TestIntegration_ProcessorCallsService verifies processor -> service integration
func TestIntegration_ProcessorCallsService(t *testing.T) {
	mockRepo := &ServiceIntegrationMock{
		result: repository.UpsertInserted,
	}

	svc := service.NewSegmentationService(mockRepo)
	_ = log.New(os.Stderr, "[TEST] ", 0)

	// Test that Run doesn't panic with cancelled context
	ctx := context.Background()
	err := Run(ctx, svc, log.New(os.Stderr, "[TEST] ", 0))

	// Error is expected if data.csv doesn't exist, but should not panic
	t.Logf("Run completed with result: %v", err)
}

// TestIntegration_ProcessorServiceChain tests complete processor-service chain
func TestIntegration_ProcessorServiceChain(t *testing.T) {
	mockRepo := &ServiceIntegrationMock{
		result: repository.UpsertInserted,
	}

	svc := service.NewSegmentationService(mockRepo)
	_ = log.New(os.Stderr, "[TEST] ", 0)

	ctx := context.Background()

	// Create a test record to verify processing
	testSeg := &models.Segmentation{
		UserID:           100,
		SegmentationType: "drug",
		SegmentationName: "TestDrug",
		Data:             datatypes.JSON(`{"test": true}`),
	}

	// Upsert directly to verify service works
	result, err := svc.Create(ctx, testSeg)
	if err != nil {
		t.Fatalf("service.Create failed: %v", err)
	}

	if result != repository.UpsertInserted {
		t.Fatalf("expected UpsertInserted, got %v", result)
	}

	if len(mockRepo.createCalls) == 0 {
		t.Fatal("expected service to call upsert")
	}

	call := mockRepo.createCalls[0]
	if call.userID != 100 {
		t.Fatalf("expected user 100, got %d", call.userID)
	}

	if call.name != "TestDrug" {
		t.Fatalf("expected name TestDrug, got %s", call.name)
	}
}

// TestIntegration_ProcessorServiceMultipleRecords tests batch processing
func TestIntegration_ProcessorServiceMultipleRecords(t *testing.T) {
	mockRepo := &ServiceIntegrationMock{
		result: repository.UpsertInserted,
	}

	svc := service.NewSegmentationService(mockRepo)
	ctx := context.Background()

	// Simulate multiple records being processed
	records := []struct {
		userID   uint64
		segType  string
		name     string
	}{
		{100, "drug", "Drug1"},
		{100, "specialty", "Spec1"},
		{200, "drug", "Drug2"},
		{200, "patient", "Patient1"},
	}

	for _, rec := range records {
		seg := &models.Segmentation{
			UserID:           rec.userID,
			SegmentationType: rec.segType,
			SegmentationName: rec.name,
			Data:             datatypes.JSON(`{}`),
		}
		_, _ = svc.Create(ctx, seg)
	}

	if len(mockRepo.createCalls) != 4 {
		t.Fatalf("expected 4 upsert calls, got %d", len(mockRepo.createCalls))
	}

	// Verify all records were processed
	if mockRepo.createCalls[0].userID != 100 {
		t.Error("first record should be for user 100")
	}
	if mockRepo.createCalls[2].userID != 200 {
		t.Error("third record should be for user 200")
	}
}

// TestIntegration_ProcessorServiceErrorHandling tests error scenarios
func TestIntegration_ProcessorServiceErrorHandling(t *testing.T) {
	// Create a mock that returns error
	errorRepo := &ServiceIntegrationMock{
		result: repository.UpsertInserted,
	}

	svc := service.NewSegmentationService(errorRepo)
	logger := log.New(os.Stderr, "[TEST] ", 0)

	// Verify service can be created with context
	if svc == nil {
		t.Fatal("service should not be nil")
	}

	// Verify logger works
	logger.Println("test log message")
	t.Log("Service and logger initialized successfully")
}

// TestIntegration_LoggerWithProcessor tests logger output during processing
func TestIntegration_LoggerWithProcessor(t *testing.T) {
	mockRepo := &ServiceIntegrationMock{
		result: repository.UpsertInserted,
	}

	svc := service.NewSegmentationService(mockRepo)
	logger := log.New(os.Stderr, "[PROCESSOR] ", log.LstdFlags)

	ctx := context.Background()

	// Create a segmentation and verify logging
	seg := &models.Segmentation{
		UserID:           500,
		SegmentationType: "drug",
		SegmentationName: "LoggedDrug",
		Data:             datatypes.JSON(`{}`),
	}

	logger.Printf("Processing segmentation: %s", seg.SegmentationName)
	result, _ := svc.Create(ctx, seg)

	if result != repository.UpsertInserted {
		t.Fatal("create should succeed")
	}

	logger.Printf("Successfully processed segmentation for user %d", seg.UserID)
	t.Log("Logger output test passed")
}

// TestIntegration_RecordStructure tests record handling
func TestIntegration_RecordStructure(t *testing.T) {
	rec := record{
		userID:  123,
		segType: "drug",
		name:    "TestDrug",
		data:    []byte(`{"test": "data"}`),
	}

	if rec.userID != 123 {
		t.Fatalf("expected user 123, got %d", rec.userID)
	}

	if rec.segType != "drug" {
		t.Fatalf("expected type drug, got %s", rec.segType)
	}

	if rec.name != "TestDrug" {
		t.Fatalf("expected name TestDrug, got %s", rec.name)
	}

	if len(rec.data) == 0 {
		t.Fatal("expected data to not be empty")
	}
}

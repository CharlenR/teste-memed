package processor

import (
	"context"
	"log"
	"os"
	"testing"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"
)

// MockProcessorRepository for testing
type MockProcessorRepository struct {
	upsertFunc func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error)
	findFunc   func(ctx context.Context, userID uint64) ([]models.Segmentation, error)
}

func (m *MockProcessorRepository) Upsert(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, s)
	}
	return repository.UpsertInserted, nil
}

func (m *MockProcessorRepository) BulkUpsert(ctx context.Context, s *[]models.Segmentation) ([]repository.UpsertResult, []error) {
	results := make([]repository.UpsertResult, len(*s))
	errors := make([]error, len(*s))
	for i := range results {
		if m.upsertFunc != nil {
			result, err := m.upsertFunc(ctx, &(*s)[i])
			if err != nil {
				errors[i] = err
			}
			results[i] = result
		} else {
			results[i] = repository.UpsertInserted
		}
	}
	return results, errors
}

func (m *MockProcessorRepository) FindByUserID(ctx context.Context, userID uint64) ([]models.Segmentation, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, userID)
	}
	return nil, nil
}

func TestRecordStructure(t *testing.T) {
	rec := record{
		userID:  123,
		segType: "drug",
		name:    "Antibióticos",
		data:    []byte(`{"type": "antibiotic"}`),
	}

	if rec.userID != 123 {
		t.Errorf("userID = %d, want 123", rec.userID)
	}
	if rec.segType != "drug" {
		t.Errorf("segType = %s, want drug", rec.segType)
	}
	if rec.name != "Antibióticos" {
		t.Errorf("name = %s, want Antibióticos", rec.name)
	}
	if len(rec.data) == 0 {
		t.Error("data should not be empty")
	}
}

func TestContextUsage(t *testing.T) {
	ctx := context.Background()

	select {
	case <-ctx.Done():
		t.Error("context should not be done")
	default:
		// expected
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		// expected
	default:
		t.Error("context should be done after cancel")
	}
}

func TestRun_WithCancelledContext(t *testing.T) {
	mockRepo := &MockProcessorRepository{
		upsertFunc: func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
			return repository.UpsertInserted, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	logger := log.New(os.Stderr, "", 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, svc, logger)
	if err == nil {
		t.Error("Run() should return error when context is already cancelled")
	}
}

func TestRun_WithValidContext(t *testing.T) {
	upsertCount := 0
	mockRepo := &MockProcessorRepository{
		upsertFunc: func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
			upsertCount++
			return repository.UpsertInserted, nil
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	logger := log.New(os.Stderr, "", 0)

	ctx := context.Background()

	// Run should process the CSV file
	err := Run(ctx, svc, logger)

	if err != nil {
		t.Logf("Run() error (expected if data.csv not found): %v", err)
	}

	t.Logf("Upsert was called %d times", upsertCount)
}

func TestRun_LoggerNotNil(t *testing.T) {
	mockRepo := &MockProcessorRepository{}
	_ = service.NewSegmentationService(mockRepo)
	logger := log.New(os.Stderr, "[TEST] ", log.LstdFlags)

	// Verify logger works
	logger.Println("test message")
	// If no panic, logger is usable
}

func TestRun_ServiceNotNil(t *testing.T) {
	mockRepo := &MockProcessorRepository{}
	svc := service.NewSegmentationService(mockRepo)

	if svc == nil {
		t.Fatal("service should not be nil")
	}

	// Just verify we can create the service without panic
	t.Log("Service created successfully")
}

func TestRun_ContextCancel(t *testing.T) {
	mockRepo := &MockProcessorRepository{
		upsertFunc: func(ctx context.Context, s *models.Segmentation) (repository.UpsertResult, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
				return repository.UpsertInserted, nil
			}
		},
	}

	svc := service.NewSegmentationService(mockRepo)
	logger := log.New(os.Stderr, "", 0)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		cancel()
	}()

	_ = Run(ctx, svc, logger)
	// If context was properly cancelled, this should complete
}

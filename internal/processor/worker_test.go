package processor

import (
"context"
"testing"
)

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

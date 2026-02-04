package repository

import (
	"testing"
)

func TestUpsertResultConstants(t *testing.T) {
	tests := []struct {
		name     string
		result   UpsertResult
		expected int
	}{
		{
			name:     "UpsertInserted",
			result:   UpsertInserted,
			expected: 0,
		},
		{
			name:     "UpsertUpdated",
			result:   UpsertUpdated,
			expected: 1,
		},
		{
			name:     "UpsertNoOp",
			result:   UpsertNoOp,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.result) != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, int(tt.result))
			}
		})
	}
}

func TestUpsertResultOrder(t *testing.T) {
	if UpsertInserted >= UpsertUpdated {
		t.Error("UpsertInserted should be less than UpsertUpdated")
	}
	if UpsertUpdated >= UpsertNoOp {
		t.Error("UpsertUpdated should be less than UpsertNoOp")
	}
}

func TestUpsertResultDistinct(t *testing.T) {
	results := []UpsertResult{UpsertInserted, UpsertUpdated, UpsertNoOp}
	seen := make(map[int]bool)

	for _, r := range results {
		val := int(r)
		if seen[val] {
			t.Errorf("Duplicate UpsertResult value: %d", val)
		}
		seen[val] = true
	}
}

package exclusion

import (
	"path/filepath"
	"testing"
)

func TestExcluder(t *testing.T) {
	// Path to the test exclusion YAML file
	yamlPath := filepath.Join("..", "..", "testdata", "test-exclusions.yaml")

	excluder, err := NewExcluder(yamlPath)
	if err != nil {
		t.Fatalf("failed to create Excluder: %v", err)
	}

	tests := []struct {
		result   map[string]string
		excluded bool
	}{
		// Test 'equals' operator with 'and' conditions
		{
			result: map[string]string{
				"username":   "test",
				"ip_address": "192.168.1.1",
			},
			excluded: true,
		},
		// Test 'contains' operator with 'or' conditions
		{
			result: map[string]string{
				"email":    "user@example.com",
				"domain":   "external.com",
				"username": "user1",
			},
			excluded: true,
		},
		// Test 'equals' operator with 'or' conditions
		{
			result: map[string]string{
				"email":    "user@external.com",
				"domain":   "internal.local",
				"username": "user2",
			},
			excluded: true,
		},
		// Test 'equals' operator with 'and' conditions
		{
			result: map[string]string{
				"response_time": "fast",
				"status_code":   "200",
			},
			excluded: true,
		},
		// Test 'equals' operator with 'or' conditions
		{
			result: map[string]string{
				"user_role": "admin",
				"username":  "adminuser",
			},
			excluded: true,
		},
		// Test 'in' operator
		{
			result: map[string]string{
				"department": "sales",
			},
			excluded: true,
		},
		// Test 'not_equals' operator
		{
			result: map[string]string{
				"status": "inactive",
			},
			excluded: true,
		},
		// Test 'not_in' operator
		{
			result: map[string]string{
				"region": "eu-west-1",
			},
			excluded: true,
		},
		// Test non-excluded result
		{
			result: map[string]string{
				"user_role": "user",
				"username":  "regularuser",
			},
			excluded: false,
		},
		// Test partial match for 'and' conditions (should not exclude)
		{
			result: map[string]string{
				"username":      "test",
				"ip_address":    "10.0.0.1",
				"response_time": "slow",
			},
			excluded: false,
		},
		// Test 'regex' operator - matching URLs
		{
			result: map[string]string{
				"url": "https://www.example.com/path",
			},
			excluded: true,
		},
		{
			result: map[string]string{
				"url": "http://example.com/anotherpath",
			},
			excluded: true,
		},
		{
			result: map[string]string{
				"url": "https://sub.example.com/path",
			},
			excluded: true, // Now should pass with updated regex
		},
		// Test 'regex' operator with non-matching URL
		{
			result: map[string]string{
				"url": "https://www.test.com/path",
			},
			excluded: false,
		},
		{
			result: map[string]string{
				"url": "ftp://example.com/resource",
			},
			excluded: false,
		},
	}

	for i, tt := range tests {
		excluded := excluder.IsExcluded(tt.result)
		if excluded != tt.excluded {
			t.Errorf("Test case %d: expected excluded=%v, got %v", i+1, tt.excluded, excluded)
		}
	}
}

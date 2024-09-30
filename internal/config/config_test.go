package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nianticlabs/venator/internal/config"
)

const (
	existentConfigPath          = "../../testdata/test-config.yaml"
	existentGlobalConfigPath    = "../../testdata/test-global-config.yaml"
	invalidConfigPath           = "../../testdata/test-invalid-config.yaml"
	invalidGlobalConfigPath     = "../../testdata/test-invalid-global-config.yaml"
	nonExistentConfigPath       = "non_existent_file.yaml"
	nonExistentGlobalConfigPath = "non_existent_global_file.yaml"
)

func TestParseRuleConfig(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectedErr string
		wantCfg     *config.RuleConfig
	}{
		{
			// Test parsing of an existing, valid config file.
			name:        "ExistentConfig",
			filePath:    existentConfigPath,
			expectedErr: "",
			wantCfg:     MakeRuleConfig(),
		},
		{
			// Test parsing of a non-existent config file.
			name:        "NonExistentConfig",
			filePath:    nonExistentConfigPath,
			expectedErr: "failed to open config file",
			wantCfg:     nil,
		},
		{
			// Test parsing of a config file with invalid YAML (unexpected fields).
			name:        "InvalidYAML",
			filePath:    invalidConfigPath,
			expectedErr: "failed to decode config YAML",
			wantCfg:     nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel testing if applicable

			cfg, err := config.ParseRuleConfig(tt.filePath)

			if err != nil {
				if tt.expectedErr == "" {
					t.Fatalf("ParseConfig() unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("ParseConfig() error = %v, expected error containing %q", err, tt.expectedErr)
				}
			} else if tt.expectedErr != "" {
				t.Fatalf("ParseConfig() expected error containing %q, got nil", tt.expectedErr)
			}

			if tt.expectedErr == "" && cfg != nil {
				// Validate parsed fields if parsing was successful.
				if diff := cmp.Diff(tt.wantCfg, cfg); diff != "" {
					t.Errorf("ParseConfig() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// MakeRuleConfig constructs the expected Config struct based on testdata/test-config.yaml
func MakeRuleConfig() *config.RuleConfig {
	return &config.RuleConfig{
		Name:           "test-rule",
		UID:            "2001416b-bdd3-4a31-af52-3b1933c4f926",
		Status:         "test",
		Confidence:     config.ConfidenceLow,
		Enabled:        true,
		Schedule:       "0 */2 * * *",
		QueryEngine:    "opensearch",
		ExclusionsPath: "config/exclusions/example-exclusions.yaml",
		Publishers:     []string{"opensearch", "pubsub"},
		Language:       "SQL",
		Query:          "SELECT * FROM logs",
		Output: config.Output{
			Format: config.OutputFormatSignal,
			Fields: []config.OutputField{
				{
					Field:  "Field1",
					Source: "f1",
				},
				{
					Field:  "Field2",
					Source: "f2",
				},
			},
		},
		Description: "this is a test rule.",
		References:  []string{"https://ref1", "https://ref2"},
		Tags:        []string{"test"},
		Author:      "adelka",
		TTPs: []config.TTP{
			{
				Framework: "MITRE ATT&CK",
				Tactic:    "tactic1",
				Name:      "technique1",
				ID:        "T111",
				Reference: "https://example1.com",
			},
			{
				Framework: "MITRE ATT&CK",
				Tactic:    "tactic2",
				Name:      "technique2",
				ID:        "T222",
				Reference: "https://example2.com",
			},
		},
		LLM: &config.LLM{
			Enabled: true,
			Prompt:  "Prompt template",
		},
	}
}

func TestParseGlobalConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("OPENSEARCH_DEV_PASSWORD", "dev-secret-password")
	os.Setenv("OPENSEARCH_PROD_PASSWORD", "prod-secret-password")
	os.Setenv("LLM_API_KEY", "test-api-key")

	defer func() {
		// Clean up environment variables after the test
		os.Unsetenv("OPENSEARCH_DEV_PASSWORD")
		os.Unsetenv("OPENSEARCH_PROD_PASSWORD")
		os.Unsetenv("LLM_API_KEY")
	}()

	tests := []struct {
		name        string
		filePath    string
		expectedErr string
		wantCfg     *config.GlobalConfig
	}{
		{
			name:        "ExistentGlobalConfig",
			filePath:    existentGlobalConfigPath,
			expectedErr: "",
			wantCfg:     makeGlobalConfig(),
		},
		{
			name:        "NonExistentGlobalConfig",
			filePath:    nonExistentGlobalConfigPath,
			expectedErr: "failed to read global config file",
			wantCfg:     nil,
		},
		{
			name:        "InvalidGlobalYAML",
			filePath:    invalidGlobalConfigPath,
			expectedErr: "failed to decode global config YAML",
			wantCfg:     nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.ParseGlobalConfig(tt.filePath)

			if err != nil {
				if tt.expectedErr == "" {
					t.Fatalf("ParseGlobalConfig() unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("ParseGlobalConfig() error = %v, expected error containing %q", err, tt.expectedErr)
				}
			} else if tt.expectedErr != "" {
				t.Fatalf("ParseGlobalConfig() expected error containing %q, got nil", tt.expectedErr)
			}

			if tt.expectedErr == "" && cfg != nil {
				// Validate parsed fields if parsing was successful.
				if diff := cmp.Diff(tt.wantCfg, cfg); diff != "" {
					t.Errorf("ParseGlobalConfig() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func makeGlobalConfig() *config.GlobalConfig {
	return &config.GlobalConfig{
		OpenSearch: config.OpenSearchConnectors{
			Instances: map[string]config.OpenSearchConfig{
				"dev": {
					URL:                "https://opensearch-instance1.example.com:9200",
					Username:           "admin1",
					Password:           "dev-secret-password", // Expect the expanded value from the envVar
					InsecureSkipVerify: false,
				},
				"prod": {
					URL:                "https://opensearch-instance2.example.com:9200",
					Username:           "admin2",
					Password:           "prod-secret-password", // Expect the expanded value from the envVar
					InsecureSkipVerify: true,
				},
			},
		},
		LLM: config.LLMConfig{
			Provider:    "openai",
			APIKey:      "test-api-key", // Expect the expanded value from the envVar
			Model:       "gpt-4o",
			ServerURL:   "",
			Temperature: 0.7,
		},
	}
}

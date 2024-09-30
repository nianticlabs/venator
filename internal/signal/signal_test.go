package signal_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nianticlabs/venator/internal/config"
	"github.com/nianticlabs/venator/internal/signal"
)

func TestBuildSignal(t *testing.T) {
	cfg := &config.RuleConfig{
		Name: "test-rule",
		UID:  "test-uid",
		Output: config.Output{
			Fields: []config.OutputField{
				{Field: "Timestamp", Source: "timestamp"},
				{Field: "Message", Source: "message"},
				{Field: "ResourceName", Source: "hostname"},
				{Field: "ActorUserName", Source: "actor.user.name"},
				{Field: "SrcHostname", Source: "src.hostname"},
				{Field: "SrcIP", Source: "src.ip"},
				{Field: "RuleSpecificData", Source: "rule_specific_data"},
			},
		},
	}
	// config with unsupported output field
	cfgInvalidOutput := &config.RuleConfig{
		Name: "test-rule",
		UID:  "test-uid",
		Output: config.Output{
			Fields: []config.OutputField{
				{Field: "Timestamp", Source: "timestamp"},
				{Field: "Message", Source: "message"},
				{Field: "ResourceName", Source: "device.hostname"},
				{Field: "UnsupportedField", Source: "unsupported"},
			},
		},
	}

	tests := []struct {
		name       string
		result     map[string]string
		cfg        *config.RuleConfig
		errMessage string
		expected   *signal.Signal
	}{
		{
			name: "successful mapping",
			result: map[string]string{
				"timestamp":          "2023-05-14T10:00:00Z",
				"message":            "process xyz created/modified the file abc",
				"hostname":           "hostname123",
				"actor.user.name":    "user123",
				"src.hostname":       "src-hostname",
				"src.ip":             "src-ip",
				"rule_specific_data": `{"key1": "value1", "key2": "value2"}`,
			},
			cfg:        cfg,
			errMessage: "",
			expected: &signal.Signal{
				Timestamp: time.Date(2023, 5, 14, 10, 0, 0, 0, time.UTC),
				Rule_ID:   cfg.UID,
				Rule_Name: cfg.Name,
				TTPs:      []map[string]string{},
				Message:   "process xyz created/modified the file abc",
				Resource:  signal.Resource{Name: "hostname123", Type: "", UID: ""},
				Actor: signal.Actor{
					User: signal.User{Name: "user123", UID: ""},
				},
				SrcEndpoint: signal.Endpoint{Hostname: "src-hostname", IP: "src-ip"},
				DstEndpoint: signal.Endpoint{Hostname: "", IP: ""},
				RuleSpecificData: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "successful mapping with invalid RuleSpecificData format",
			result: map[string]string{
				"timestamp":          "2023-05-14T10:00:00Z",
				"message":            "process xyz created/modified the file abc",
				"hostname":           "hostname123",
				"actor.user.name":    "user123",
				"src.hostname":       "src-hostname",
				"src.ip":             "src-ip",
				"rule_specific_data": `key1: value1, key2: value2`, // invalid JSON
			},
			cfg:        cfg,
			errMessage: "",
			expected: &signal.Signal{
				Timestamp: time.Date(2023, 5, 14, 10, 0, 0, 0, time.UTC),
				Rule_ID:   cfg.UID,
				Rule_Name: cfg.Name,
				TTPs:      []map[string]string{},
				Message:   "process xyz created/modified the file abc",
				Resource:  signal.Resource{Name: "hostname123", Type: "", UID: ""},
				Actor: signal.Actor{
					User: signal.User{Name: "user123", UID: ""},
				},
				SrcEndpoint: signal.Endpoint{Hostname: "src-hostname", IP: "src-ip"},
				DstEndpoint: signal.Endpoint{Hostname: "", IP: ""},
				RuleSpecificData: map[string]string{
					"raw": "key1: value1, key2: value2",
				},
			},
		},
		{
			name: "missing field in result",
			result: map[string]string{
				"timestamp":          "2023-05-14T10:00:00Z",
				"raw":                "process xyz created/modified the file abc",
				"device.hostname":    "hostname123",
				"actor.user.name":    "user123",
				"src.hostname":       "src-hostname",
				"src.ip":             "src-ip",
				"rule_specific_data": "",
			},
			cfg:        cfg,
			errMessage: "source field message not found in query results",
			expected:   nil,
		},
		{
			name: "query result field count mismatch",
			result: map[string]string{
				"timestamp":       "2023-05-14T10:00:00Z",
				"message":         "process xyz created/modified the file abc",
				"device.hostname": "hostname123",
				"actor.user.name": "user123",
				"src.hostname":    "src-hostname",
			},
			cfg:        cfg,
			errMessage: "number of query result fields mismatches expected count",
			expected:   nil,
		},
		{
			name: "invalid timestamp format",
			result: map[string]string{
				"timestamp":          "invalid-timestamp",
				"message":            "process xyz created/modified the file abc",
				"hostname":           "hostname123",
				"actor.user.name":    "user123",
				"src.hostname":       "src-hostname",
				"src.ip":             "src-ip",
				"rule_specific_data": "",
			},
			cfg:        cfg,
			errMessage: "parsing time",
			expected:   nil,
		},
		{
			name: "unsupported output field",
			result: map[string]string{
				"timestamp":       "2023-05-14T10:00:00Z",
				"message":         "process xyz created/modified the file abc",
				"device.hostname": "hostname123",
				"unsupported":     "unsupported value",
			},
			cfg:        cfgInvalidOutput,
			errMessage: "unsupported output field UnsupportedField",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RuleConfig{}
			*cfg = *tt.cfg
			signal, err := signal.BuildSignal(tt.result, cfg)
			if err != nil {
				if tt.expected != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tt.errMessage) {
					t.Fatalf("error message does not contain expected substring: %s", tt.errMessage)
				}
				return
			}
			if diff := cmp.Diff(tt.expected, signal); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

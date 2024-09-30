package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/nianticlabs/venator/internal/config"
)

type mockLLMClient struct {
	response string
	err      error
}

func (m *mockLLMClient) Call(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestGeneratePrompt(t *testing.T) {
	tests := []struct {
		name           string
		prompt         string
		results        []map[string]string
		expectedPrompt string
		expectError    bool
	}{
		{
			name: "Valid prompt generation",
			prompt: `Analyze the following signals:
{{ .FormattedResults }}
`,
			results: []map[string]string{
				{"signal_name": "UserLogin", "user": "test-user", "timestamp": "2023-10-10 08:00:00"},
				{"signal_name": "PrivilegeEscalation", "user": "test-user", "timestamp": "2023-10-10 08:05:00"},
			},
			expectedPrompt: `Analyze the following signals:
signal_name: UserLogin, timestamp: 2023-10-10 08:00:00, user: test-user
signal_name: PrivilegeEscalation, timestamp: 2023-10-10 08:05:00, user: test-user
`,
			expectError: false,
		},
		{
			name: "Invalid template",
			prompt: `Analyze the following signals:
{{ .FormattedResults `,
			results:     []map[string]string{{"signal_name": "UserLogin"}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := generatePrompt(tt.prompt, tt.results)
			if (err != nil) != tt.expectError {
				t.Fatalf("generatePrompt() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil {
				return
			}

			// Remove any extra whitespace for consistent comparison
			prompt = strings.TrimSpace(prompt)
			expectedPrompt := strings.TrimSpace(tt.expectedPrompt)

			if prompt != expectedPrompt {
				t.Errorf("Prompt does not match expected.\nGot:\n%s\nExpected:\n%s", prompt, expectedPrompt)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		expectedResult []map[string]string
		expectError    bool
	}{
		{
			name: "Valid JSON response",
			response: `[
				{
					"Title": "Suspicious Privilege Escalation",
					"Type": "multi-stage",
					"Confidence": "high",
					"Message": "The user 'test-user' performed a privilege escalation.",
					"Next_Steps": "Investigate the user's activities.",
					"User": "test-user",
					"Resource": "MACHINE2",
					"RawData": ["{Signal Name: UserLogin, Timestamp: 2023-10-10 08:00:00}", "{Signal Name: PrivilegeEscalation, Timestamp: 2023-10-10 08:05:00}"]
				}
			]`,
			expectedResult: []map[string]string{
				{
					"Title":      "Suspicious Privilege Escalation",
					"Type":       "multi-stage",
					"Confidence": "high",
					"Message":    "The user 'test-user' performed a privilege escalation.",
					"Next_Steps": "Investigate the user's activities.",
					"User":       "test-user",
					"Resource":   "MACHINE2",
					"RawData":    `["{Signal Name: UserLogin, Timestamp: 2023-10-10 08:00:00}","{Signal Name: PrivilegeEscalation, Timestamp: 2023-10-10 08:05:00}"]`,
				},
			},
			expectError: false,
		},
		{
			name:        "Empty JSON array",
			response:    "[]",
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			response:    `{"Title": "Missing brackets"`,
			expectError: true,
		},
		{
			name:     "Response with code fences",
			response: "```json\n[\n    {\n        \"Title\": \"Suspicious Activity\",\n        \"Type\": \"single-stage\",\n        \"Confidence\": \"high\",\n        \"Message\": \"A test message.\",\n        \"Next_Steps\": \"Investigate further.\",\n        \"User\": \"test-user\",\n        \"Resource\": \"test-resource\",\n        \"RawData\": [\"{...}\"]\n    }\n]\n```",
			expectedResult: []map[string]string{
				{
					"Title":      "Suspicious Activity",
					"Type":       "single-stage",
					"Confidence": "high",
					"Message":    "A test message.",
					"Next_Steps": "Investigate further.",
					"User":       "test-user",
					"Resource":   "test-resource",
					"RawData":    `["{...}"]`,
				},
			},
			expectError: false,
		},
		{
			name:        "Response with extra text",
			response:    "Here is the analysis:\n[\n    {\n        \"Title\": \"Suspicious Activity\"\n    }\n]\nThank you.",
			expectError: true, // Expect error since extra text is not handled
		},
		{
			name:        "Empty response with code fences",
			response:    "```json\n[]\n```",
			expectError: false,
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			results, err := parseResponse(tt.response)
			if (err != nil) != tt.expectError {
				t.Fatalf("parseResponse() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil {
				return
			}

			if tt.expectedResult != nil && len(results) != len(tt.expectedResult) {
				t.Fatalf("Expected %d results, got %d", len(tt.expectedResult), len(results))
			}

			if tt.expectedResult != nil {
				for i, result := range results {
					for key, expectedValue := range tt.expectedResult[i] {
						if result[key] != expectedValue {
							t.Errorf("At index %d, key '%s': expected '%s', got '%s'", i, key, expectedValue, result[key])
						}
					}
				}
			} else if results != nil {
				t.Errorf("Expected results to be nil, got %v", results)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name           string
		results        []map[string]string
		prompt         string
		mockResponse   string
		mockError      error
		expectedResult []map[string]string
		expectError    bool
	}{
		{
			name: "Successful processing",
			results: []map[string]string{
				{"signal_name": "UserLogin", "user": "test-user", "timestamp": "2023-10-10 08:00:00"},
				{"signal_name": "PrivilegeEscalation", "user": "test-user", "timestamp": "2023-10-10 08:05:00"},
			},
			prompt: `Analyze the following signals:
{{ .FormattedResults }}
`,
			mockResponse: `[
				{
					"Title": "Suspicious Privilege Escalation",
					"Type": "multi-stage",
					"Confidence": "high",
					"Message": "The user 'test-user' performed a privilege escalation.",
					"Next_Steps": "Investigate the user's activities.",
					"User": "test-user",
					"Resource": "MACHINE2",
					"RawData": ["{Signal Name: UserLogin, Timestamp: 2023-10-10 08:00:00}", "{Signal Name: PrivilegeEscalation, Timestamp: 2023-10-10 08:05:00}"]
				}
			]`,
			expectedResult: []map[string]string{
				{
					"Title":      "Suspicious Privilege Escalation",
					"Type":       "multi-stage",
					"Confidence": "high",
					"Message":    "The user 'test-user' performed a privilege escalation.",
					"Next_Steps": "Investigate the user's activities.",
					"User":       "test-user",
					"Resource":   "MACHINE2",
					"RawData":    `["{Signal Name: UserLogin, Timestamp: 2023-10-10 08:00:00}","{Signal Name: PrivilegeEscalation, Timestamp: 2023-10-10 08:05:00}"]`,
				},
			},
			expectError: false,
		},
		{
			name: "LLM client returns error",
			results: []map[string]string{
				{"signal_name": "UserLogin", "user": "test-user", "timestamp": "2023-10-10 08:00:00"},
			},
			prompt: `Analyze the following signals:
{{ .FormattedResults }}
`,
			mockError:   fmt.Errorf("simulated LLM error"),
			expectError: true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			llmClient := &mockLLMClient{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			cfg := &config.RuleConfig{
				LLM: &config.LLM{
					Enabled: true,
					Prompt:  tt.prompt,
				},
			}

			newResults, err := Process(ctx, llmClient, tt.results, cfg)
			if (err != nil) != tt.expectError {
				t.Fatalf("Process() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil {
				return
			}

			if len(newResults) != len(tt.expectedResult) {
				t.Fatalf("Expected %d result(s), got %d", len(tt.expectedResult), len(newResults))
			}

			for i, result := range newResults {
				for key, expectedValue := range tt.expectedResult[i] {
					if result[key] != expectedValue {
						t.Errorf("Key '%s': expected '%s', got '%s'", key, expectedValue, result[key])
					}
				}
			}
		})
	}
}

func TestConvertToMapStringString(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedOutput map[string]string
		expectError    bool
	}{
		{
			name: "Valid conversion with various types",
			input: map[string]interface{}{
				"string": "value",
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"nil":    nil,
				"array":  []interface{}{"item1", "item2"},
				"map":    map[string]interface{}{"key": "value"},
			},
			expectedOutput: map[string]string{
				"string": "value",
				"number": "42",
				"float":  "3.14",
				"bool":   "true",
				"nil":    "",
				"array":  `["item1","item2"]`,
				"map":    `{"key":"value"}`,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			output, err := convertToMapStringString(tt.input)
			if (err != nil) != tt.expectError {
				t.Fatalf("convertToMapStringString() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil {
				return
			}

			for key, expectedValue := range tt.expectedOutput {
				if output[key] != expectedValue {
					t.Errorf("Key '%s': expected '%s', got '%s'", key, expectedValue, output[key])
				}
			}
		})
	}
}

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/nianticlabs/venator/internal/config"
	llmconfig "github.com/nianticlabs/venator/internal/llm/config"
	"github.com/nianticlabs/venator/internal/llm/model"
	"github.com/nianticlabs/venator/internal/llm/openai"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

var logger = logrus.StandardLogger()

// New creates a new LLM client based on the provided configuration
func New(llmConfig llmconfig.Config) (model.Client, error) {
	switch llmConfig.Provider {
	case llmconfig.ProviderOpenAI:
		return openai.New(llmConfig)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmConfig.Provider)
	}
}

// Process runs the LLM analysis on the query results.
func Process(ctx context.Context, client model.Client, results []map[string]string, cfg *config.RuleConfig) ([]map[string]string, error) {
	if len(results) == 0 {
		logger.Infof("No results to process with LLM")
		return nil, nil
	}

	if client == nil {
		return nil, fmt.Errorf("LLM client is not initialized")
	}

	prompt, err := generatePrompt(cfg.LLM.Prompt, results)
	if err != nil {
		return nil, fmt.Errorf("error generating prompt: %w", err)
	}

	response, err := client.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	newResults, err := parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	if len(newResults) == 0 {
		logger.Infof("LLM response is empty. No high-confidence findings.")
		return nil, nil
	}

	return newResults, nil
}

func generatePrompt(promptTemplate string, results []map[string]string) (string, error) {
	// Format the results to be readable in the prompt
	var formattedResults bytes.Buffer
	for _, result := range results {
		var fields []string

		// Collect and sort keys alphabetically
		keys := maps.Keys(result)
		sort.Strings(keys)

		for _, key := range keys {
			value := result[key]
			fields = append(fields, fmt.Sprintf("%s: %s", key, value))
		}
		line := strings.Join(fields, ", ")
		formattedResults.WriteString(line + "\n")
	}

	data := map[string]interface{}{
		"FormattedResults": formattedResults.String(),
	}

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing prompt template: %w", err)
	}

	return buf.String(), nil
}

func parseResponse(response string) ([]map[string]string, error) {
	// Remove any leading/trailing whitespace
	response = strings.TrimSpace(response)

	// Regular expression to match code fences
	codeFenceRegex := regexp.MustCompile("(?s)```(?:json)?\\n(.*)\\n```")

	matches := codeFenceRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		// Extract the content inside the code fences
		response = matches[1]
		logger.Debugf("Extracted JSON from code fences:\n%s", response)
	}

	// Remove any leading/trailing whitespace again
	response = strings.TrimSpace(response)

	if response == "[]" || response == "" {
		return nil, nil
	}

	logger.Debugf("Cleaned LLM response:\n%s", response)

	// Try to unmarshal into a []map[string]interface{}
	var rawResults []map[string]interface{}
	err := json.Unmarshal([]byte(response), &rawResults)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling LLM response: %w", err)
	}

	// Convert []map[string]interface{} to []map[string]string
	results := make([]map[string]string, len(rawResults))
	for i, rawResult := range rawResults {
		res, err := convertToMapStringString(rawResult)
		if err != nil {
			return nil, fmt.Errorf("error converting result at index %d: %w", i, err)
		}
		results[i] = res
	}

	return results, nil
}

func convertToMapStringString(m map[string]interface{}) (map[string]string, error) {
	res := make(map[string]string)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			res[k] = val
		case nil:
			res[k] = ""
		case []interface{}, map[string]interface{}:
			// Convert array or map to JSON string
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("error marshaling %T to JSON string: %w", val, err)
			}
			res[k] = string(jsonBytes)
		default:
			// For other types, convert to string
			res[k] = fmt.Sprintf("%v", val)
		}
	}
	return res, nil
}

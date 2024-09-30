package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds the entire global configuration.
type GlobalConfig struct {
	OpenSearch OpenSearchConnectors `yaml:"opensearch"`
	PubSub     PubSubConnectors     `yaml:"pubsub"`
	BigQuery   BigQueryConnectors   `yaml:"bigquery"`
	Slack      SlackConnectors      `yaml:"slack"`
	LLM        LLMConfig            `yaml:"llm"`
}

type OpenSearchConnectors struct {
	Instances map[string]OpenSearchConfig `yaml:"instances"`
}

type PubSubConnectors struct {
	Instances map[string]PubSubConfig `yaml:"instances"`
}

type BigQueryConnectors struct {
	Instances map[string]BigQueryConfig `yaml:"instances"`
}

type SlackConnectors struct {
	Instances map[string]SlackConfig `yaml:"instances"`
}

type OpenSearchConfig struct {
	URL                string `yaml:"url"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

type PubSubConfig struct {
	ProjectID string `yaml:"projectID"`
	TopicID   string `yaml:"topicID"`
}

type BigQueryConfig struct {
	ProjectID string `yaml:"projectID"`
	DatasetID string `yaml:"datasetID"`
	TableID   string `yaml:"tableID"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhookURL,omitempty"`
}

type LLMConfig struct {
	Provider    string  `yaml:"provider"`
	APIKey      string  `yaml:"apiKey"`
	Model       string  `yaml:"model"`
	ServerURL   string  `yaml:"serverURL,omitempty"`
	Temperature float64 `yaml:"temperature"`
}

// ParseGlobalConfig parses the global YAML configuration file.
func ParseGlobalConfig(path string) (*GlobalConfig, error) {
	var cfg GlobalConfig

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config file: %w", err)
	}

	// Expand environment variables in the configuration file
	expandedContent := os.ExpandEnv(string(fileContent))

	decoder := yaml.NewDecoder(strings.NewReader(expandedContent))
	decoder.KnownFields(true) // Enforce strict field matching

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode global config YAML: %w", err)
	}

	return &cfg, nil
}

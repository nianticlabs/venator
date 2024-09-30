package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuleConfig represents the entire config structure for a single rule.
type RuleConfig struct {
	Author         string          `yaml:"author"`
	Confidence     ConfidenceLevel `yaml:"confidence"`
	Description    string          `yaml:"description"`
	Enabled        bool            `yaml:"enabled"`
	ExclusionsPath string          `yaml:"exclusionsPath,omitempty"`
	Language       string          `yaml:"language"`
	LLM            *LLM            `yaml:"llm,omitempty"`
	Name           string          `yaml:"name"`
	Output         Output          `yaml:"output"`
	Publishers     []string        `yaml:"publishers"`
	Query          string          `yaml:"query"`
	QueryEngine    string          `yaml:"queryEngine"`
	References     []string        `yaml:"references"`
	Schedule       string          `yaml:"schedule"`
	Status         string          `yaml:"status"`
	Tags           []string        `yaml:"tags"`
	TTPs           []TTP           `yaml:"ttps"`
	UID            string          `yaml:"uid"`
}

type LLM struct {
	Enabled bool   `yaml:"enabled"`
	Prompt  string `yaml:"prompt"`
}

type Output struct {
	Format OutputFormat  `yaml:"format"`
	Fields []OutputField `yaml:"fields"`
}

type OutputField struct {
	Field  string `yaml:"field"`
	Source string `yaml:"source"`
}

type TTP struct {
	Framework string `yaml:"framework"`
	Tactic    string `yaml:"tactic"`
	Name      string `yaml:"name"`
	ID        string `yaml:"id"`
	Reference string `yaml:"reference"`
}

type ConfidenceLevel string

const (
	ConfidenceUnknown ConfidenceLevel = "unknown"
	ConfidenceLow     ConfidenceLevel = "low"
	ConfidenceMedium  ConfidenceLevel = "medium"
	ConfidenceHigh    ConfidenceLevel = "high"
)

type OutputFormat string

const (
	OutputFormatRaw    OutputFormat = "raw"
	OutputFormatSignal OutputFormat = "signal"
)

// ParseRuleConfig parses the rules YAML configuration file.
func ParseRuleConfig(path string) (*RuleConfig, error) {
	var cfg RuleConfig
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true) // Enforce strict field matching

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config YAML: %w", err)
	}

	return &cfg, nil
}

package config

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	// Add more providers as needed
)

type Config struct {
	Provider    Provider
	APIKey      string
	Model       string
	Temperature float64
	ServerURL   string // Optional
}

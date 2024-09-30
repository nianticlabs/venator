package model

import "context"

// Client defines the interface for LLM clients
type Client interface {
	Call(ctx context.Context, prompt string) (string, error)
}

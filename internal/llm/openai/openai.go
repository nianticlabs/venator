package openai

import (
	"context"
	"fmt"

	goopenai "github.com/sashabaranov/go-openai"

	llmconfig "github.com/nianticlabs/venator/internal/llm/config"
	"github.com/nianticlabs/venator/internal/llm/model"
)

type Client struct {
	client      *goopenai.Client
	model       string
	temperature float32
}

func New(llmConfig llmconfig.Config) (model.Client, error) {
	clientConfig := goopenai.DefaultConfig(llmConfig.APIKey)
	if llmConfig.ServerURL != "" {
		clientConfig.BaseURL = llmConfig.ServerURL
	}

	client := goopenai.NewClientWithConfig(clientConfig)
	return &Client{
		client:      client,
		model:       llmConfig.Model,
		temperature: float32(llmConfig.Temperature),
	}, nil
}

func (o *Client) Call(ctx context.Context, prompt string) (string, error) {
	req := goopenai.ChatCompletionRequest{
		Model:       o.model,
		Messages:    []goopenai.ChatCompletionMessage{{Role: goopenai.ChatMessageRoleUser, Content: prompt}},
		Temperature: o.temperature,
	}

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error calling OpenAI API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return resp.Choices[0].Message.Content, nil
}

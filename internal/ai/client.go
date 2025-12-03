package ai

import (
	"context"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// Client wraps the Anthropic SDK for runbook generation tasks.
type Client struct {
	client anthropic.Client
	model  anthropic.Model
}

// Available returns true if an API key is configured.
func Available() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}

// NewClient creates a new AI client.
// It reads ANTHROPIC_API_KEY from the environment.
// Uses Claude 3.5 Haiku for fast, cost-effective processing.
func NewClient() *Client {
	return &Client{
		client: anthropic.NewClient(),
		model:  anthropic.ModelClaude3_5HaikuLatest,
	}
}

// NewClientWithModel creates a client with a specific model.
func NewClientWithModel(model anthropic.Model) *Client {
	return &Client{
		client: anthropic.NewClient(),
		model:  model,
	}
}

// sendMessage sends a message to Claude and returns the text response.
func (c *Client) sendMessage(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	params := anthropic.MessageNewParams{
		MaxTokens: 4096,
		Model:     c.model,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	// Extract text from response
	for _, block := range message.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", nil
}

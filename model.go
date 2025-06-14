package aigit

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

type Model interface {
	Query(ctx context.Context, query string) (string, error)
}

func GetDefaultModel() Model {
	if _, exists := os.LookupEnv("ANTHROPIC_API_KEY"); exists {
		return NewAnthropicModel()
	}
	panic("No model configured.")
}

type AnthropicModel struct {
	client anthropic.Client
}

func NewAnthropicModel() *AnthropicModel {
	return &AnthropicModel{
		client: anthropic.NewClient(),
	}
}

func (m *AnthropicModel) Query(ctx context.Context, query string) (string, error) {
	message, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{{
			Content: []anthropic.ContentBlockParamUnion{{
				OfText: &anthropic.TextBlockParam{Text: query},
			}},
			Role: anthropic.MessageParamRoleUser,
		}},
		Model: anthropic.ModelClaude3_7SonnetLatest,
	})
	if err != nil {
		return "", fmt.Errorf("failed to query model: %w", err)
	}
	return message.Content[0].Text, nil
}

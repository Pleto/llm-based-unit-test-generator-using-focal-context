package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"pleto.dev/chattest/internal/integrations/llm"
)

type Config struct {
	APIKey string
}

type Client struct {
	openai *openai.Client
}

var _ llm.LLM = &Client{}

func NewClient(cfg Config) *Client {
	openAIConfig := openai.DefaultConfig(cfg.APIKey)
	return &Client{
		openai: openai.NewClientWithConfig(openAIConfig),
	}
}

func (c *Client) CreateChatCompletion(ctx context.Context, request *llm.CreateChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(request.Messages))
	for i, message := range request.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	resp, err := c.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("openai.CreateChatCompletion(): %w", err)
	}

	return &llm.ChatCompletionResponse{
		Content: resp.Choices[0].Message.Content,
	}, nil
}

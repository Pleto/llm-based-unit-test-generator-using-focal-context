package mockai

import (
	"context"

	"pleto.dev/chattest/internal/integrations/llm"
)

type Client struct {
}

var _ llm.LLM = &Client{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateChatCompletion(ctx context.Context, request *llm.CreateChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	return nil, nil
}

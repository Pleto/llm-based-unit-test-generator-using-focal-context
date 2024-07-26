package llm

import "context"

type LLM interface {
	CreateChatCompletion(context.Context, *CreateChatCompletionRequest) (*ChatCompletionResponse, error)
}

type CreateChatCompletionRequest struct {
	Messages []ChatCompletionMessage
}

type ChatCompletionMessage struct {
	Role    string
	Content string
}

type ChatCompletionResponse struct {
	Content string
}

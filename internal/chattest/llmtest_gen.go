package chattest

import (
	"context"
	"fmt"

	"pleto.dev/chattest/internal/integrations/llm"
)

// LLMTestGenerator uses the prompts from LLMTestContext to generate a test using the LLM API.
// It returns the parsed generated test.
type LLMTestGenerator struct {
	llm llm.LLM
}

func NewLLMTestGenerator(llm llm.LLM) *LLMTestGenerator {
	return &LLMTestGenerator{
		llm,
	}
}

func (t *LLMTestGenerator) Generate(ctx context.Context, context *LLMTestContext) (*LLMGeneratedTest, error) {
	request := llm.CreateChatCompletionRequest{
		Messages: []llm.ChatCompletionMessage{
			{
				Role:    "system",
				Content: context.InitialPrompt,
			},
		},
	}

	for _, followUp := range context.FollowUps {
		request.Messages = append(request.Messages, llm.ChatCompletionMessage{
			Role:    "user",
			Content: followUp,
		})
	}

	response, err := t.llm.CreateChatCompletion(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("t.llm.CreateChatCompletion(): %w", err)
	}

	llmTest, err := parse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("parse(): %w", err)
	}

	return llmTest, nil
}

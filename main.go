package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"pleto.dev/chattest/internal/chattest"
	"pleto.dev/chattest/internal/integrations/llm/openai"
)

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	openApiKey, found := os.LookupEnv("OPENAI_API_KEY")
	if !found {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	openAIClient := openai.NewClient(openai.Config{
		APIKey: openApiKey,
	})

	cfg, err := chattest.NewConfig()
	if err != nil {
		return err
	}

	if err := chattest.Run(ctx, cfg, openAIClient, w); err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

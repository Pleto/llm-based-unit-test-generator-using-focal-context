package chattest

import (
	"context"
	"fmt"
	"io"

	"pleto.dev/chattest/internal/integrations/llm"
)

func Run(ctx context.Context, cfg *Config, llm llm.LLM, w io.Writer) error {
	project, err := LoadPackages(cfg)
	if err != nil {
		return fmt.Errorf("LoadPackages(): %w", err)
	}

	llmContext := NewLLMTestContext()
	llmContext.AddTestPrompt(project)
	llmTestGenerator := NewLLMTestGenerator(llm)

	for i := 0; true; i++ {
		llmTest, err := llmTestGenerator.Generate(ctx, llmContext)
		if err != nil {
			return fmt.Errorf("llmTestGenerator.Generate(): %w", err)
		}

		test := NewTest(llmTest, project.FocalMethod.InferTestLocation(), cfg.RepoPath)

		if err := test.Save(); err != nil {
			return fmt.Errorf("Save(): %w", err)
		}

		result, err := test.Run()
		if err != nil {
			return fmt.Errorf("Run(): %w", err)
		}

		if !result.TestFailed {
			fmt.Fprintln(w, "Test passed")
			break
		}

		if i >= cfg.RepairRounds {
			break
		}

		llmContext.AddRepairPrompt(llmTest, result)
		fmt.Fprintln(w, "Test failed, trying to repair it")
	}

	return nil
}

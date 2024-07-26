package chattest

import (
	"flag"
	"fmt"
)

type Config struct {
	RepoPath        string
	FuncPath        string
	FuncName        string
	RepairRounds    int
	RandomTestCount int
	UseFuncTestFile bool
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	if err := cfg.flags(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) flags() error {
	flag.StringVar(&c.RepoPath, "repo", "", "path to the repository")
	flag.StringVar(&c.FuncPath, "func-path", "", "file path to the function under test")
	flag.StringVar(&c.FuncName, "func", "", "name of the function under test")
	flag.IntVar(&c.RepairRounds, "rounds", 0, "number of repair rounds")
	flag.IntVar(&c.RandomTestCount, "test-count", 0, "number of random tests to pick for prompt augmentation")
	flag.BoolVar(&c.UseFuncTestFile, "use-func-test-file", false, "if it exists, use the test file of the function under test for prompt augmentation")

	flag.Parse()

	if c.RepoPath == "" {
		return fmt.Errorf("missing required flag: -repo")
	}

	if c.FuncPath == "" {
		return fmt.Errorf("missing either of the required flags: -pkg-path, -func-path")
	}

	if c.FuncName == "" {
		return fmt.Errorf("missing required flag: -func")
	}

	return nil
}

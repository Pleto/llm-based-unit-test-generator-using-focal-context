package chattest

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TODO: Add the line which failed to the func result.
// It's possible to do that using the message from 'Error Trace'
func ExtractTestFailure(goTestResult string) (bool, string) {
	failPattern := regexp.MustCompile("--- FAIL:")

	if failPattern.MatchString(goTestResult) {
		errorPattern := regexp.MustCompile(`Error:\s*(.*)`)

		errorMatch := errorPattern.FindStringSubmatch(goTestResult)
		if len(errorMatch) > 1 {
			return true, strings.TrimSpace(errorMatch[1])
		}
		return true, "Error message not found"
	}

	return false, ""
}

func ExtractCompileErrors(stderr bytes.Buffer) (string, error) {
	re := regexp.MustCompile(`^[^#].*\.go:[0-9]+:[0-9]+:\s+(.*)$`)
	scanner := bufio.NewScanner(&stderr)

	var compileErrors strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if match := re.FindStringSubmatch(line); match != nil {
			compileErrors.WriteString(match[1])
			compileErrors.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner.Scan(): %w", err)
	}

	return compileErrors.String(), nil
}

type TestRunResult struct {
	TestFailed    bool
	FailedMessage string
	CompileError  string
}

func (t *Test) Run() (*TestRunResult, error) {
	cmd := exec.Command("go", "test", filepath.Dir(t.Path), "-v", "-run", fmt.Sprintf("^%s$", t.Name))
	cmd.Dir = t.RepoPath

	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		if errOut.Len() > 0 {
			compileError, err := ExtractCompileErrors(errOut)
			if err != nil {
				return nil, fmt.Errorf("ExtractCompileErrors(): %w", err)
			}
			return &TestRunResult{
				TestFailed:   true,
				CompileError: compileError,
			}, nil
		}

		testFailed, message := ExtractTestFailure(out.String())
		if testFailed {
			return &TestRunResult{
				TestFailed:    true,
				FailedMessage: message,
			}, nil
		}

		return nil, fmt.Errorf("cmd.Run(): %w", err)
	}

	return &TestRunResult{
		TestFailed: false,
	}, nil
}

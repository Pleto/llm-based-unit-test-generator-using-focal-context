package chattest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/imports"
)

func (t *Test) Save() error {
	file, err := cropen(t.Path)
	if err != nil {
		return fmt.Errorf("cropen(): %w", err)
	}
	defer file.Close()

	isIn, err := t.isIn(file)
	if err != nil {
		return fmt.Errorf("isIn(): %w", err)
	}

	if isIn {
		err := t.removeFrom(file)
		if err != nil {
			return fmt.Errorf("removeFrom(): %w", err)
		}
	}

	if err := t.writeTo(file); err != nil {
		return fmt.Errorf("writeTo(): %w", err)
	}

	if err := fixImports(file); err != nil {
		return fmt.Errorf("fixImports(): %w", err)
	}

	return nil
}

// cropen creates a file if it does not exist, or opens an existing file in append mode.
func cropen(path string) (*os.File, error) {
	if !fileExists(path) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("os.Create(): %w", err)
		}
		return file, nil
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile(): %w", err)
	}
	return file, nil
}

func (t *Test) isIn(file *os.File) (bool, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false, fmt.Errorf("file.Seek(0, io.SeekStart): %w", err)
	}

	if contains, err := t.scanLinesForTestFunc(file); err != nil {
		return false, fmt.Errorf("scanLinesForTestFunc(): %w", err)
	} else if contains {
		return true, nil
	}
	return false, nil
}

func (t *Test) scanLinesForTestFunc(file *os.File) (bool, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := t.matchesTestFunc(scanner.Text())
		if matches {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (t *Test) matchesTestFunc(line string) bool {
	pattern := `^func ` + t.Name + `\(.* \*testing\.T\)`
	testFuncPattern := regexp.MustCompile(pattern)
	return testFuncPattern.MatchString(line)
}

func (t *Test) removeFrom(file *os.File) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("file.Seek(0, io.SeekStart): %w", err)
	}

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	inTestFunction := false
	braceCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if inTestFunction {
			if strings.Contains(trimmedLine, "{") {
				braceCount++
			}
			if strings.Contains(trimmedLine, "}") {
				braceCount--
			}
			if braceCount == 0 {
				inTestFunction = false
			}
			continue
		}

		if t.matchesTestFunc(trimmedLine) {
			inTestFunction = true
			if strings.Contains(trimmedLine, "{") {
				braceCount++
			}
			continue
		}

		buffer.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner.Err(): %w", err)
	}

	err := os.WriteFile(t.Path, buffer.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile(): %w", err)
	}

	return nil
}

func (t *Test) writeTo(file *os.File) error {
	content := t.Test + "\n"

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("file.Stat(): %w", err)
	}

	if stat.Size() == 0 {
		content = fmt.Sprintf("package %s\n\n%s", t.Package.Name, content)
	}

	if err := writeContentToFile(file, content); err != nil {
		return fmt.Errorf("writeContentToFile(): %w", err)
	}

	return nil
}

func writeContentToFile(file *os.File, content string) error {
	n, err := file.WriteString(content)
	if err != nil {
		return fmt.Errorf("file.WriteString(): %w", err)
	}
	if n != len(content) {
		return fmt.Errorf("file.WriteString(): wrote %d bytes, expected %d", n, len(content))
	}
	return nil
}

func fixImports(file *os.File) error {
	fileContentWithImports, err := imports.Process(file.Name(), nil, nil)
	if err != nil {
		return fmt.Errorf("imports.Process(): %w", err)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("file.Truncate(): %w", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("file.Seek(): %w", err)
	}

	if err := writeContentToFile(file, string(fileContentWithImports)); err != nil {
		return fmt.Errorf("writeContentToFile(): %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

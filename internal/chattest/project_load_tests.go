package chattest

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"golang.org/x/tools/go/packages"
)

// ExampleTestFile represents a test file selected from the project
// to assist in prompt augmentation for generating the test for the focal method.
type ExampleTestFile struct {
	File    OsPath
	Content string
}

func selectTests(focalMethod *FocalMethod, pkgs []*packages.Package, cfg *Config) ([]*ExampleTestFile, error) {
	focalPkgID := focalMethod.Pkg.ID
	if focalPkgID == "" {
		return nil, errors.New("focal method package ID is empty")
	}

	rootPaths := extractRootPaths(focalPkgID, 3)
	testCandidates := collectAllTests(pkgs, rootPaths)
	pickedTests, err := pickRandomExampleTests(testCandidates, cfg.RandomTestCount)
	if err != nil {
		return nil, fmt.Errorf("pickRandomTests(): %w", err)
	}

	if cfg.UseFuncTestFile {
		funcTestFile, err := focalMethodTestFile(focalMethod)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("pickAsExampleTest(): %w", err)
		}
		pickedTests = append(pickedTests, funcTestFile)
	}

	return pickedTests, nil
}

func extractRootPaths(focalPkgID PackageID, depth int) []string {
	parts := strings.Split(string(focalPkgID), "/")
	var result []string
	for i := 0; i < depth && len(parts)-i > 0; i++ {
		rootPath := strings.Join(parts[:len(parts)-i], "/")
		result = append(result, rootPath)
	}
	return result
}

type testFileCandidate struct {
	distanceToFM int
	file         OsPath
}

func collectAllTests(pkgs []*packages.Package, rootPaths []string) []*testFileCandidate {
	resultSet := make(map[string]testFileCandidate)
	for _, pkg := range pkgs {
		for i, rootPath := range rootPaths {
			if strings.HasPrefix(pkg.ID, rootPath) {
				for _, file := range pkg.GoFiles {
					if strings.HasSuffix(file, "_test.go") && testedFileExists(file) {
						resultSet[file] = testFileCandidate{
							distanceToFM: i,
							file:         file,
						}
					}
				}
			}
		}
	}
	var result []*testFileCandidate
	for _, test := range resultSet {
		result = append(result, &test)
	}
	return result
}

func testedFileExists(testFile OsPath) bool {
	testedFile, _ := strings.CutSuffix(testFile, "_test.go")
	testedFile += ".go"
	return fileExists(testedFile)
}

func focalMethodTestFile(focalMethod *FocalMethod) (*ExampleTestFile, error) {
	content, err := readContent(focalMethod.InferTestLocation().Path)
	if err != nil {
		return nil, fmt.Errorf("readContent(): %w", err)
	}
	return &ExampleTestFile{
		File:    focalMethod.InferTestLocation().Path,
		Content: content,
	}, nil
}

func pickRandomExampleTests(testCandidates []*testFileCandidate, count int) ([]*ExampleTestFile, error) {
	rand.Seed(uint64(time.Now().UnixNano()))
	rand.Shuffle(len(testCandidates), func(i, j int) {
		testCandidates[i], testCandidates[j] = testCandidates[j], testCandidates[i]
	})

	var result []*ExampleTestFile
	for i := 0; i < count && i < len(testCandidates); i++ {
		content, err := readContent(testCandidates[i].file)
		if err != nil {
			return nil, fmt.Errorf("readContent(): %w", err)
		}
		result = append(result, &ExampleTestFile{
			File:    testCandidates[i].file,
			Content: content,
		})
	}
	return result, nil
}

func readContent(file OsPath) (string, error) {
	data, err := os.ReadFile(string(file))
	if err != nil {
		return "", fmt.Errorf("os.ReadFile(): %w", err)
	}
	return string(data), nil
}

package chattest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type LLMGeneratedTest struct {
	Name string
	Test string
}

func parse(llmTest string) (*LLMGeneratedTest, error) {
	test := extractTest(llmTest)
	ast, err := parseAST(test)
	if err != nil {
		return nil, fmt.Errorf("parseAST(): %w", err)
	}

	return &LLMGeneratedTest{
		Name: ast.Name.Name,
		Test: test,
	}, nil
}

func extractTest(llmTest string) string {
	llmTest = strings.TrimSpace(llmTest)

	preffixesToTrim := []string{"```golang", "```go", "```"}
	for _, preffix := range preffixesToTrim {
		llmTest = strings.TrimPrefix(llmTest, preffix)
	}

	suffixesToTrim := []string{"```"}
	for _, suffix := range suffixesToTrim {
		llmTest = strings.TrimSuffix(llmTest, suffix)
	}

	return llmTest
}

func parseAST(test string) (*ast.FuncDecl, error) {
	// we need to add a dummy package declaration to use go/parser
	test = decorateWithPackage(test)

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, "", test, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parser.ParseFile(): %w", err)
	}

	var testFunc *ast.FuncDecl
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			// in golang, tests are functions whose names begin with Test
			if strings.HasPrefix(fn.Name.Name, "Test") {
				testFunc = fn
			}
		}
		return true
	})

	if testFunc == nil {
		return nil, fmt.Errorf("test function not found")
	}
	return testFunc, nil
}

func decorateWithPackage(test string) string {
	return fmt.Sprintf("package main\n%s", test)
}

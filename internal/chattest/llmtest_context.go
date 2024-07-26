package chattest

import "strings"

// LLMTestContext keeps the prompts used to generate the test.
// InitialPrompt is self-explanatory.
// In FollowUps are the prompts used to repair the generated test.
type LLMTestContext struct {
	InitialPrompt string
	FollowUps     []string
}

func NewLLMTestContext() *LLMTestContext {
	return &LLMTestContext{}
}

func (c *LLMTestContext) AddTestPrompt(project *Project) {
	c.InitialPrompt = c.testPrompt(project)
}

func (c *LLMTestContext) testPrompt(project *Project) string {
	return newTestPromptBuilder().
		addTests(project.TestFiles).
		addDefinitions(project.FocalMethod.GroupDefinitionsByPackage()).
		addFocalMethod(project.FocalMethod).
		addInstructions().
		build()
}

type testPromptBuilder struct {
	sb strings.Builder
}

func newTestPromptBuilder() *testPromptBuilder {
	return &testPromptBuilder{}
}

func (b *testPromptBuilder) addTests(tests []*ExampleTestFile) *testPromptBuilder {
	if len(tests) == 0 {
		return b
	}
	b.sb.WriteString("Using these tests as example:\n")
	b.sb.WriteString(sprintTests(tests))
	b.sb.WriteString("\n\n")
	return b
}

func (b *testPromptBuilder) addDefinitions(defs map[string][]Definition) *testPromptBuilder {
	if len(defs) == 0 {
		return b
	}
	b.sb.WriteString("Given:\n")
	b.sb.WriteString(sprintDefinitions(defs))
	b.sb.WriteString("\n")
	return b
}

func (b *testPromptBuilder) addFocalMethod(focalMethod *FocalMethod) *testPromptBuilder {
	b.sb.WriteString("Write a golang test for:\n")
	b.sb.WriteString("```\n")
	b.sb.WriteString(focalMethod.Body)
	b.sb.WriteString("\n")
	b.sb.WriteString("```\n")
	return b
}

func (b *testPromptBuilder) addInstructions() *testPromptBuilder {
	b.sb.WriteString(writeTestInstructions())
	return b
}

func (b *testPromptBuilder) build() string {
	return b.sb.String()
}

func sprintTests(tests []*ExampleTestFile) string {
	var sb strings.Builder
	for _, test := range tests {
		sb.WriteString("```\n")
		sb.WriteString(test.Content)
		sb.WriteString("```\n\n")
	}
	return sb.String()
}

func sprintDefinitions(defs map[string][]Definition) string {
	var sb strings.Builder
	for pkg, defs := range defs {
		sb.WriteString("```\n")
		sb.WriteString("package " + pkg + "\n\n")
		for _, def := range defs {
			if def.Body() != nil {
				sb.WriteString(*def.Body())
				sb.WriteString("\n\n")
			}
		}
		sb.WriteString("```\n\n")
	}
	return sb.String()
}

func (c *LLMTestContext) AddRepairPrompt(llmTest *LLMGeneratedTest, testRun *TestRunResult) {
	if testRun.CompileError != "" {
		c.FollowUps = append(c.FollowUps, c.compileErrorPrompt(llmTest.Test, testRun.CompileError))
	}
	if testRun.FailedMessage != "" {
		c.FollowUps = append(c.FollowUps, c.runtimeFailMessagePrompt(llmTest.Test, testRun.FailedMessage))
	}
}

func (c *LLMTestContext) runtimeFailMessagePrompt(llmTest, message string) string {
	var sb strings.Builder
	sb.WriteString("The test you generated\n")
	sb.WriteString(llmTest)
	sb.WriteString("\n")
	sb.WriteString("failed with the message:\n")
	sb.WriteString("```\n")
	sb.WriteString(message)
	sb.WriteString("\n")
	sb.WriteString("```\n")
	sb.WriteString("Please fix the test.\n")
	sb.WriteString(writeTestInstructions())
	return sb.String()
}

func (c *LLMTestContext) compileErrorPrompt(llmTest, compileError string) string {
	var sb strings.Builder
	sb.WriteString("The test you generated\n")
	sb.WriteString(llmTest)
	sb.WriteString("\n")
	sb.WriteString("failed with the compile error:\n")
	sb.WriteString("```\n")
	sb.WriteString(compileError)
	sb.WriteString("\n")
	sb.WriteString("```\n")
	sb.WriteString("Please fix the test.\n")
	sb.WriteString(writeTestInstructions())
	return sb.String()
}

func writeTestInstructions() string {
	return "Don't mock, use fakes. Write only the test, only one, with no imports or explanations.\n"
}

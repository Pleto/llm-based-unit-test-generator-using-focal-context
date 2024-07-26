package chattest

type Pkg struct {
	ID   PackageID
	Name string
}

type Test struct {
	*LLMGeneratedTest
	RepoPath string
	Package  Pkg
	Path     OsPath
}

func NewTest(llmGenTest *LLMGeneratedTest, details *TestLocation, repoPath string) *Test {
	return &Test{
		LLMGeneratedTest: llmGenTest,
		RepoPath:         repoPath,
		Package:          details.Pkg,
		Path:             details.Path,
	}
}

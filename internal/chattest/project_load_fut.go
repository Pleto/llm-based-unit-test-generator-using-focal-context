package chattest

import (
	"fmt"
	"strings"
)

// function under test (fut)
type FocalMethod struct {
	ID   QualifiedName
	Name string
	Body string
	Pkg  Pkg
	File OsPath
	// Uses contains all definitions of the identifiers used in the focal method
	Uses map[QualifiedName]Definition
}

type TestLocation struct {
	Path OsPath
	Pkg  Pkg
}

func (fm *FocalMethod) InferTestLocation() *TestLocation {
	pathWithoutGoExt := strings.TrimSuffix(fm.File, ".go")
	testPath := fmt.Sprintf("%s_test.go", pathWithoutGoExt)
	return &TestLocation{
		Path: testPath,
		Pkg:  fm.Pkg,
	}
}

func (fm *FocalMethod) GroupDefinitionsByPackage() map[string][]Definition {
	pkgDefs := make(map[string][]Definition)
	for _, def := range fm.Uses {
		pkgName := def.Package().Name
		pkgDefs[pkgName] = append(pkgDefs[pkgName], def)
	}
	return pkgDefs
}

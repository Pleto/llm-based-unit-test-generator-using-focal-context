package chattest

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

type Project struct {
	Path        string
	FocalMethod *FocalMethod
	TestFiles   []*ExampleTestFile
}

func LoadPackages(cfg *Config) (*Project, error) {
	// TODO: find out what we actually need
	const loadMode = packages.NeedName |
		packages.NeedFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	pkgsLoadCfg := packages.Config{
		Mode: loadMode,
		Dir:  cfg.RepoPath,
		// Ensure Go modules are enabled
		Env:   append(os.Environ(), "GO111MODULE=on"),
		Tests: true,
	}

	pkgs, err := packages.Load(&pkgsLoadCfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("packages.Load(): %w", err)
	}

	err = packageErrors(pkgs)
	if err != nil {
		return nil, fmt.Errorf("packageErrors(): %w", err)
	}

	focalMethod, err := new(focalMethodParser).Parse(pkgs, cfg)
	if err != nil {
		return nil, fmt.Errorf("focalMethodParser.Parse(): %w", err)
	}

	tests, err := selectTests(focalMethod, pkgs, cfg)
	if err != nil {
		return nil, fmt.Errorf("selectRandomTests(): %w", err)
	}

	return &Project{
		FocalMethod: focalMethod,
		Path:        cfg.RepoPath,
		TestFiles:   tests,
	}, nil
}

type multiError struct {
	errors []error
}

// stupid func that collects all errors and wrap them as only one error.
func (m *multiError) Error() error {
	if len(m.errors) == 0 {
		return nil
	}
	var aggregatedErr error
	for _, err := range m.errors {
		aggregatedErr = errors.Join(aggregatedErr, err)
	}
	return aggregatedErr
}

func (m *multiError) Append(err error) {
	if err != nil {
		m.errors = append(m.errors, err)
	}
}

func packageErrors(pkgs []*packages.Package) error {
	errors := multiError{}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			errors.Append(err)
		}
	})
	return errors.Error()
}

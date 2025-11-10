package analyzer

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"

	"github.com/Arneball/goasted/rules"
)

// Analyzer analyzes Go source code for rule violations
type Analyzer struct {
	registry *rules.Registry
}

// New creates a new Analyzer with the given rule registry
func New(registry *rules.Registry) *Analyzer {
	return &Analyzer{
		registry: registry,
	}
}

// Analyze analyzes the given path (file or directory) and returns violations
func (a *Analyzer) Analyze(path string) ([]rules.Violation, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	var violations []rules.Violation

	if info.IsDir() {
		violations, err = a.analyzeDirectory(path)
	} else {
		violations, err = a.analyzeFile(path)
	}

	if err != nil {
		return nil, err
	}

	return violations, nil
}

// analyzeDirectory recursively analyzes all Go files in a directory
func (a *Analyzer) analyzeDirectory(dir string) ([]rules.Violation, error) {
	var violations []rules.Violation

	// Try to load packages with type information
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   dir,
		Tests: true, // Include test files
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil || len(pkgs) == 0 {
		// Fallback to file-by-file analysis if package loading fails
		return a.analyzeDirectoryFallback(dir)
	}

	// Analyze each file concurrently
	violationsChan := make(chan []rules.Violation, 100)
	var wg sync.WaitGroup

	for _, pkg := range pkgs {
		// Skip packages with errors (but continue analyzing others)
		if len(pkg.Errors) > 0 {
			// Try individual files as fallback
			for _, file := range pkg.GoFiles {
				wg.Add(1)

				go func() {
					defer wg.Done()
					if fileViolations, err := a.analyzeFile(file); err == nil {
						violationsChan <- fileViolations
					}
				}()
			}
			continue
		}

		// Analyze each file in the package with full type information
		getRules := a.registry.GetRules()
		for i, file := range pkg.Syntax {
			wg.Add(len(getRules))

			ctx := &rules.Context{
				FileSet:  pkg.Fset,
				File:     file,
				Filename: pkg.GoFiles[i],
				TypeInfo: pkg.TypesInfo,
			}

			for _, rule := range getRules {
				go func() {
					defer wg.Done()
					violationsChan <- rule.Check(ctx)
				}()
			}
		}
	}

	go func() {
		wg.Wait()
		close(violationsChan)
	}()

	for v := range violationsChan {
		violations = append(violations, v...)
	}

	return violations, nil
}

// analyzeDirectoryFallback is the fallback for when package loading fails
func (a *Analyzer) analyzeDirectoryFallback(dir string) ([]rules.Violation, error) {
	var violations []rules.Violation

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only analyze Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileViolations, err := a.analyzeFile(path)
		if err != nil {
			return fmt.Errorf("failed to analyze %s: %w", path, err)
		}

		violations = append(violations, fileViolations...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return violations, nil
}

// analyzeFile analyzes a single Go file
func (a *Analyzer) analyzeFile(filename string) ([]rules.Violation, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Compute type information for the file
	typeInfo := a.computeTypeInfo(fset, node)

	ctx := &rules.Context{
		FileSet:  fset,
		File:     node,
		Filename: filename,
		TypeInfo: typeInfo,
	}

	var violations []rules.Violation

	// Apply all rules to the file
	for _, rule := range a.registry.GetRules() {
		ruleViolations := rule.Check(ctx)
		violations = append(violations, ruleViolations...)
	}

	return violations, nil
}

// computeTypeInfo computes type information for a single file
func (a *Analyzer) computeTypeInfo(fset *token.FileSet, file *ast.File) *types.Info {
	// Create a type checker configuration
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			// Ignore type checking errors - we do best effort
		},
	}

	// Prepare info to collect type information
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	// Type check the file
	// Note: This checks a single file in isolation, so some type info may be incomplete
	// For more accurate checking, we'd need to load the entire package
	conf.Check("", fset, []*ast.File{file}, info)

	return info
}

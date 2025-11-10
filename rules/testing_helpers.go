package rules

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

// parseTestCode parses Go source code and creates a Context for testing
func parseTestCode(t *testing.T, filename string, src string) *Context {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	ctx := &Context{
		FileSet:  fset,
		File:     node,
		Filename: filename,
	}

	return ctx
}

// parseTestCodeWithTypes parses Go source code with type checking and creates a Context for testing
func parseTestCodeWithTypes(t *testing.T, filename string, src string) *Context {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Perform type checking
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			// Ignore errors in tests
		},
	}

	typeInfo := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf.Check("", fset, []*ast.File{node}, typeInfo)

	ctx := &Context{
		FileSet:  fset,
		File:     node,
		Filename: filename,
		TypeInfo: typeInfo,
	}

	return ctx
}

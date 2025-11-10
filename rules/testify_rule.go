package rules

import (
	"go/ast"
	"go/token"
	"strings"
)

// TestifyRule checks if test code is calling into github.com/stretchr/testify
type TestifyRule struct{}

// Name returns the rule name
func (r TestifyRule) Name() string {
	return "testify-usage"
}

// Description returns the rule description
func (r TestifyRule) Description() string {
	return "Detects usage of github.com/stretchr/testify in test files"
}

// Check checks the file for all testify imports and usages
func (r TestifyRule) Check(ctx *Context) []Violation {
	// Only check test files
	if !strings.HasSuffix(ctx.Filename, "_test.go") {
		return nil
	}

	var violations []Violation
	violationChan := make(chan Violation)
	// Check for import of testify
	importInspector := func(node ast.Node) bool {
		if node == nil {
			return false
		}
		importSpec, ok := node.(*ast.ImportSpec)
		var importPath string
		var pos token.Position
		if !ok {
			return true
		}
		if importSpec.Path == nil {
			return true
		}
		importPath = strings.Trim(importSpec.Path.Value, `"`)
		if !strings.HasPrefix(importPath, "github.com/stretchr/testify") {
			return true
		}
		pos = ctx.FileSet.Position(importSpec.Pos())
		violationChan <- Violation{
			File:    ctx.Filename,
			Line:    pos.Line,
			Column:  pos.Column,
			Rule:    r.Name(),
			Message: "Test file imports testify package: " + importPath,
		}
		return true
	}

	// Check for selector expressions that might be testify calls
	// For example: assert.Equal, require.NoError, suite.Run, etc.
	selectorInspector := func(node ast.Node) bool {
		if node == nil {
			return false
		}
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		// Check if the selector is a known testify package identifier
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		// Common testify package aliases
		testifyPkgs := map[string]bool{
			"assert":  true,
			"require": true,
			"suite":   true,
			"mock":    true,
		}

		if !testifyPkgs[ident.Name] {
			return true
		}

		// Verify this identifier is actually imported from testify
		// by checking the imports
		for _, imp := range ctx.File.Imports {
			if imp.Path == nil {
				continue
			}
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(importPath, "github.com/stretchr/testify") {
				continue
			}
			// Check if the import alias or package name matches
			var pkgName string
			if imp.Name != nil {
				pkgName = imp.Name.Name
			} else {
				// Extract package name from path
				parts := strings.Split(importPath, "/")
				if len(parts) > 0 {
					pkgName = parts[len(parts)-1]
				}
			}

			if pkgName != ident.Name {
				continue
			}
			pos := ctx.FileSet.Position(sel.Pos())
			violations = append(violations, Violation{
				File:    ctx.Filename,
				Line:    pos.Line,
				Column:  pos.Column,
				Rule:    r.Name(),
				Message: "Test code calls testify method: " + ident.Name + "." + sel.Sel.Name,
			})
		}

		return true
	}

	// Walk the AST to find all violations
	go func() {
		ast.Inspect(ctx.File, importInspector)
		close(violationChan)
	}()
	ast.Inspect(ctx.File, selectorInspector)
	for violation := range violationChan {
		violations = append(violations, violation)
	}
	return violations
}

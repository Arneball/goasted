package rules

import (
	"go/ast"
	"strings"
)

// GokitRule checks if code is using github.com/go-kit/kit
type GokitRule struct{}

// Name returns the rule name
func (r GokitRule) Name() string {
	return "gokit-usage"
}

// Description returns the rule description
func (r GokitRule) Description() string {
	return "Detects usage of github.com/go-kit/kit in code"
}

// Check checks the file for all gokit imports and usages
func (r GokitRule) Check(ctx *Context) []Violation {
	var violations []Violation

	// Walk the AST to find all violations
	ast.Inspect(ctx.File, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Check for import of gokit
		var importSpec *ast.ImportSpec
		var ok bool
		if importSpec, ok = node.(*ast.ImportSpec); !ok {
			return true
		}
		if importSpec.Path == nil {
			return true
		}
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		if !strings.HasPrefix(importPath, "github.com/go-kit/kit") {
			return true
		}
		pos := ctx.FileSet.Position(importSpec.Pos())
		violations = append(violations, Violation{
			File:    ctx.Filename,
			Line:    pos.Line,
			Column:  pos.Column,
			Rule:    r.Name(),
			Message: "File imports go-kit package: " + importPath,
		})

		return true
	})

	return violations
}

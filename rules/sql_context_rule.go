package rules

import (
	"go/ast"
	"go/types"
)

// SqlContextRule checks if code calls sql.DB or sql.Tx methods without context
// when a context-aware version exists
type SqlContextRule struct{}

// Name returns the rule name
func (r SqlContextRule) Name() string {
	return "sql-context-required"
}

// Description returns the rule description
func (r SqlContextRule) Description() string {
	return "Detects calls to database/sql methods that should use context-aware versions"
}

// methodsWithContextOverload maps method names to their context-aware equivalents
var dbMethodsWithContextOverload = map[string]string{
	"Exec":     "ExecContext",
	"Query":    "QueryContext",
	"QueryRow": "QueryRowContext",
	"Prepare":  "PrepareContext",
	"Begin":    "BeginTx",
	"Ping":     "PingContext",
}

var txMethodsWithContextOverload = map[string]string{
	"Exec":     "ExecContext",
	"Query":    "QueryContext",
	"QueryRow": "QueryRowContext",
	"Prepare":  "PrepareContext",
}

// Check checks the file for sql.DB or sql.Tx method calls without context
func (r SqlContextRule) Check(ctx *Context) []Violation {
	var violations []Violation

	// Walk the AST to find method calls
	ast.Inspect(ctx.File, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Look for method call expressions
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a selector expression (e.g., db.Exec)
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		methodName := selExpr.Sel.Name

		// Check if this is a method that has a context overload
		contextMethod, isDbMethod := dbMethodsWithContextOverload[methodName]
		if !isDbMethod {
			contextMethod, isDbMethod = txMethodsWithContextOverload[methodName]
		}

		if !isDbMethod {
			return true
		}

		// Use type information to check the receiver type
		if ctx.TypeInfo == nil {
			// Skip if no type info available
			return true
		}

		exprType := ctx.TypeInfo.TypeOf(selExpr.X)
		if exprType == nil {
			// Skip if type cannot be determined
			return true
		}

		// Check if the receiver is *sql.DB or *sql.Tx
		if !typeIsDbOrTx(exprType) {
			// Not a sql type, skip
			return true
		}

		// Found a sql.DB or sql.Tx call without context
		pos := ctx.FileSet.Position(callExpr.Pos())
		typeName := exprType.String()
		violations = append(violations, Violation{
			File:    ctx.Filename,
			Line:    pos.Line,
			Column:  pos.Column,
			Rule:    r.Name(),
			Message: "Use " + contextMethod + " instead of " + methodName + " (called on " + getReceiverName(selExpr.X) + " of type " + typeName + ")",
		})

		return true
	})

	return violations
}

// getReceiverName extracts a readable name from the receiver expression
func getReceiverName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	case *ast.CallExpr:
		if funcIdent, ok := x.Fun.(*ast.Ident); ok {
			return funcIdent.Name + "()"
		} else if funcSel, ok := x.Fun.(*ast.SelectorExpr); ok {
			return funcSel.Sel.Name + "()"
		}
	}
	return "receiver"
}

// typeIsDbOrTx checks if a type is *sql.DB or *sql.Tx
func typeIsDbOrTx(t types.Type) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}

	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil || pkg.Path() != "database/sql" {
		return false
	}

	name := obj.Name()
	return name == "DB" || name == "Tx"
}

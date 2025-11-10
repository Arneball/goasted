package rules

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Context provides context information for rule checking
type Context struct {
	FileSet  *token.FileSet
	File     *ast.File
	Filename string
	TypeInfo *types.Info // Type information for the file (may be nil)
}

// Violation represents a rule violation
type Violation struct {
	File    string
	Line    int
	Column  int
	Rule    string
	Message string
}

// Rule defines the interface that all rules must implement
type Rule interface {
	// Name returns the unique name of the rule
	Name() string

	// Description returns a human-readable description of the rule
	Description() string

	// Check checks a file and returns all violations found
	Check(ctx *Context) []Violation
}

// Registry manages a collection of rules
type Registry []Rule

// NewRegistry creates a new rule registry
func NewRegistry() *Registry {
	return &Registry{}
}

// Register registers a new rule
func (r *Registry) Register(rule Rule) {
	*r = append(*r, rule)
}

// GetRules returns all registered rules
func (r *Registry) GetRules() []Rule {
	return *r
}

// GetRule returns a rule by name, or nil if not found
func (r *Registry) GetRule(name string) Rule {
	for _, rule := range *r {
		if rule.Name() == name {
			return rule
		}
	}
	return nil
}

// Filter returns a new registry containing only the rules with the specified names
func (r *Registry) Filter(names []string) *Registry {
	filtered := NewRegistry()
	for _, name := range names {
		if rule := r.GetRule(name); rule != nil {
			filtered.Register(rule)
		}
	}
	return filtered
}

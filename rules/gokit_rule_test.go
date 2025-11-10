package rules

import (
	"testing"
)

func TestGokitRule_DetectsImports(t *testing.T) {
	src := `package main

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func main() {
	// Some code using go-kit
}
`

	ctx := parseTestCode(t, "main.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	// Should detect 2 imports
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}

	// Verify the violations are for go-kit imports
	for _, v := range violations {
		if v.Rule != "gokit-usage" {
			t.Errorf("Expected rule 'gokit-usage', got '%s'", v.Rule)
		}
	}
}

func TestGokitRule_NoViolationsWithoutGokit(t *testing.T) {
	src := `package main

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("No go-kit here")
}
`

	ctx := parseTestCode(t, "main.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations without go-kit, got %d", len(violations))
	}
}

func TestGokitRule_DetectsSingleImport(t *testing.T) {
	src := `package service

import (
	"github.com/go-kit/kit/endpoint"
)

type Service struct{}
`

	ctx := parseTestCode(t, "service.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Message != "File imports go-kit package: github.com/go-kit/kit/endpoint" {
		t.Errorf("Unexpected violation message: %s", violations[0].Message)
	}
}

func TestGokitRule_WorksInTestFiles(t *testing.T) {
	// Unlike testify rule, gokit rule should detect go-kit in any file type
	src := `package service

import (
	"testing"
	"github.com/go-kit/kit/log"
)

func TestService(t *testing.T) {
	// test code
}
`

	ctx := parseTestCode(t, "service_test.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation in test file, got %d", len(violations))
	}
}

func TestGokitRule_HandlesEmptyImportPath(t *testing.T) {
	// Test edge case where import spec might have nil path
	src := `package service

import (
	"fmt"
)

type Service struct{}
`

	ctx := parseTestCode(t, "service.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for standard imports, got %d", len(violations))
	}
}

func TestGokitRule_IgnoresSimilarPackageNames(t *testing.T) {
	// Test that only exact go-kit imports are detected
	src := `package service

import (
	"github.com/some-other/kit"
	"github.com/go-kit-fake/endpoint"
)

type Service struct{}
`

	ctx := parseTestCode(t, "service.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for non-go-kit packages, got %d", len(violations))
	}
}

func TestGokitRule_DetectsMultipleVariants(t *testing.T) {
	// Test various go-kit subpackages
	src := `package service

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	"github.com/go-kit/kit/metrics"
)

type Service struct{}
`

	ctx := parseTestCode(t, "service.go", src)
	rule := NewGokitRule()
	violations := rule.Check(ctx)

	if len(violations) != 4 {
		t.Errorf("Expected 4 violations for different go-kit subpackages, got %d", len(violations))
	}
}

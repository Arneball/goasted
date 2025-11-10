package rules

import (
	"strings"
	"testing"
)

func TestTestifyRule_DetectsImports(t *testing.T) {
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert.Equal(t, 1, 1)
}
`

	ctx := parseTestCode(t, "test_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import + 1 usage
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}
}

func TestTestifyRule_IgnoresNonTestFiles(t *testing.T) {
	src := `package main

import (
	"github.com/stretchr/testify/assert"
)

func Something() {
	// This shouldn't be flagged because it's not a test file
}
`

	ctx := parseTestCode(t, "regular.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for non-test file, got %d", len(violations))
	}
}

func TestTestifyRule_NoViolationsForStandardTesting(t *testing.T) {
	src := `package main

import (
	"testing"
)

func TestSomething(t *testing.T) {
	if 1 != 1 {
		t.Error("Math is broken")
	}
}
`

	ctx := parseTestCode(t, "standard_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for standard testing, got %d", len(violations))
	}
}

func TestTestifyRule_DetectsMultipleSubpackages(t *testing.T) {
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestSomething(t *testing.T) {
	assert.Equal(t, 1, 1)
	require.NotNil(t, "foo")
}
`

	ctx := parseTestCode(t, "multi_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 3 imports + 2 usages
	if len(violations) != 5 {
		t.Errorf("Expected 5 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Message)
		}
	}
}

func TestTestifyRule_DetectsAliasedImport(t *testing.T) {
	src := `package main

import (
	"testing"
	a "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	a.Equal(t, 1, 1)
}
`

	ctx := parseTestCode(t, "alias_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import (usage with custom alias 'a' is not detected,
	// which is by design to avoid false positives)
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (import only) with aliased import, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Message)
		}
	}

	// Verify it's the import violation
	if len(violations) > 0 && !strings.Contains(violations[0].Message, "imports testify package") {
		t.Errorf("Expected import violation, got: %s", violations[0].Message)
	}
}

func TestTestifyRule_IgnoresNonTestifyIdentifiers(t *testing.T) {
	// Test that identifiers named "assert" but not from testify are ignored
	src := `package main

import (
	"testing"
)

type assert struct{}

func (a *assert) Equal(x, y int) bool {
	return x == y
}

func TestSomething(t *testing.T) {
	a := &assert{}
	a.Equal(1, 1)
}
`

	ctx := parseTestCode(t, "fake_assert_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should not detect any violations (no testify import)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for non-testify assert, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Message)
		}
	}
}

func TestTestifyRule_DetectsMockPackage(t *testing.T) {
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func TestWithMock(t *testing.T) {
	m := new(MockService)
	m.On("Method").Return(nil)
}
`

	ctx := parseTestCode(t, "mock_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import + 1 usage (mock.On)
	if len(violations) < 1 {
		t.Errorf("Expected at least 1 violation for mock usage, got %d", len(violations))
	}
}

func TestTestifyRule_DetectsSuitePackage(t *testing.T) {
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/suite"
)

type MySuite struct {
	suite.Suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(MySuite))
}
`

	ctx := parseTestCode(t, "suite_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import + 1 usage (suite.Run)
	if len(violations) < 1 {
		t.Errorf("Expected at least 1 violation for suite usage, got %d", len(violations))
	}
}

func TestTestifyRule_HandlesRequirePackage(t *testing.T) {
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestWithRequire(t *testing.T) {
	require.NotNil(t, "something")
	require.NoError(t, nil)
}
`

	ctx := parseTestCode(t, "require_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import + 2 usages
	if len(violations) != 3 {
		t.Errorf("Expected 3 violations for require usage, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Message)
		}
	}
}

func TestTestifyRule_HandlesNestedSelectors(t *testing.T) {
	// Test that we only catch direct testify calls, not nested expressions
	src := `package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type Wrapper struct {
	assert *assert.Assertions
}

func TestNested(t *testing.T) {
	assert.Equal(t, 1, 1)
	w := Wrapper{}
	// This should still be detected as it's a direct assert call
}
`

	ctx := parseTestCode(t, "nested_test.go", src)
	rule := NewTestifyRule()
	violations := rule.Check(ctx)

	// Should detect: 1 import + 1 direct usage
	if len(violations) < 2 {
		t.Errorf("Expected at least 2 violations, got %d", len(violations))
	}
}

package rules

import (
	"testing"
)

func TestSqlContextRule_DetectsDbExec(t *testing.T) {
	src := `package main

import "database/sql"

func example(db *sql.DB) {
	db.Exec("SELECT 1")
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

}

func TestSqlContextRule_DetectsDbQuery(t *testing.T) {
	src := `package main

import "database/sql"

func example(db *sql.DB) {
	db.Query("SELECT * FROM users")
	db.QueryRow("SELECT name FROM users WHERE id = 1")
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should detect Query and QueryRow
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}
}

func TestSqlContextRule_DetectsTxMethods(t *testing.T) {
	src := `package main

import "database/sql"

func example(tx *sql.Tx) {
	tx.Exec("INSERT INTO logs VALUES (?)", "test")
	tx.Query("SELECT * FROM logs")
	tx.Prepare("SELECT * FROM users WHERE id = ?")
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should detect Exec, Query, and Prepare on tx
	if len(violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(violations))
	}
}

func TestSqlContextRule_DetectsStructFieldAccess(t *testing.T) {
	src := `package main

import "database/sql"

type Service struct {
	db *sql.DB
}

func (s *Service) DoWork() {
	s.db.Exec("UPDATE users SET active = true")
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

}

func TestSqlContextRule_IgnoresContextMethods(t *testing.T) {
	src := `package main

import (
	"context"
	"database/sql"
)

func example(ctx context.Context, db *sql.DB) {
	db.ExecContext(ctx, "SELECT 1")
	db.QueryContext(ctx, "SELECT * FROM users")
	db.QueryRowContext(ctx, "SELECT name FROM users")
	db.PrepareContext(ctx, "SELECT * FROM users WHERE id = ?")
	db.PingContext(ctx)
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should not detect any violations - all methods use context
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Unexpected violation: %s", v.Message)
		}
	}
}

func TestSqlContextRule_IgnoresNonSqlImports(t *testing.T) {
	src := `package main

import "fmt"

func example() {
	// This shouldn't be flagged - no database/sql import
	fmt.Println("Exec something")
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for non-sql code, got %d", len(violations))
	}
}

func TestSqlContextRule_DetectsBegin(t *testing.T) {
	src := `package main

import "database/sql"

func example(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	tx.Commit()
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should detect Begin (should use BeginTx instead)
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

}

func TestSqlContextRule_DetectsPing(t *testing.T) {
	src := `package main

import "database/sql"

func healthCheck(database *sql.DB) error {
	return database.Ping()
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should detect Ping (should use PingContext)
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

}

func TestSqlContextRule_TypeCheckingWorks(t *testing.T) {
	// Test that type checking properly identifies sql.DB even with non-standard names
	src := `package main

import "database/sql"

type Repository struct {
	connection *sql.DB
}

func (r *Repository) Query() error {
	// Field named "connection" (not "db"), type checking should catch this
	_, err := r.connection.Query("SELECT 1")
	return err
}
`

	ctx := parseTestCodeWithTypes(t, "test.go", src)
	rule := NewSqlContextRule()
	violations := rule.Check(ctx)

	// Should detect Query via type checking, not heuristics
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

}

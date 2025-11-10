package examples

import (
	"context"
	"database/sql"
	"fmt"
)

// GoodDatabaseUsage demonstrates SQL calls WITH context
type GoodDatabaseUsage struct {
	db *sql.DB
}

func (g *GoodDatabaseUsage) GetUser(ctx context.Context, id int) error {
	// GOOD: Using ExecContext with context
	_, err := g.db.ExecContext(ctx, "UPDATE users SET last_seen = NOW() WHERE id = ?", id)
	if err != nil {
		return err
	}

	// GOOD: Using QueryContext with context
	rows, err := g.db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	// GOOD: Using QueryRowContext with context
	var name string
	err = g.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}

	fmt.Println("User:", name)
	return nil
}

func (g *GoodDatabaseUsage) PrepareStatement(ctx context.Context) error {
	// GOOD: Using PrepareContext with context
	stmt, err := g.db.PrepareContext(ctx, "SELECT * FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	return nil
}

func (g *GoodDatabaseUsage) Transaction(ctx context.Context) error {
	// GOOD: Using BeginTx with context
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// GOOD: Using ExecContext on transaction with context
	_, err = tx.ExecContext(ctx, "INSERT INTO logs (message) VALUES (?)", "test")
	if err != nil {
		return err
	}

	// GOOD: Using QueryContext on transaction with context
	rows, err := tx.QueryContext(ctx, "SELECT * FROM logs")
	if err != nil {
		return err
	}
	defer rows.Close()

	return tx.Commit()
}

func ProcessWithDBGood(ctx context.Context, db *sql.DB) error {
	// GOOD: Using PingContext with context
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

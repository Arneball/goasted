package examples

import (
	"database/sql"
	"fmt"
)

// BadDatabaseUsage demonstrates SQL calls without context
type BadDatabaseUsage struct {
	db *sql.DB
}

func (b *BadDatabaseUsage) GetUser(id int) error {
	// BAD: Using Exec without context
	_, err := b.db.Exec("UPDATE users SET last_seen = NOW() WHERE id = ?", id)
	if err != nil {
		return err
	}

	// BAD: Using Query without context
	rows, err := b.db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	// BAD: Using QueryRow without context
	var name string
	err = b.db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}

	fmt.Println("User:", name)
	return nil
}

func (b *BadDatabaseUsage) PrepareStatement() error {
	// BAD: Using Prepare without context
	stmt, err := b.db.Prepare("SELECT * FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	return nil
}

func (b *BadDatabaseUsage) Transaction() error {
	// BAD: Using Begin without context
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// BAD: Using Exec on transaction without context
	_, err = tx.Exec("INSERT INTO logs (message) VALUES (?)", "test")
	if err != nil {
		return err
	}

	// BAD: Using Query on transaction without context
	rows, err := tx.Query("SELECT * FROM logs")
	if err != nil {
		return err
	}
	defer rows.Close()

	return tx.Commit()
}

func ProcessWithDBBad(db *sql.DB) error {
	// BAD: Using Ping without context
	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}

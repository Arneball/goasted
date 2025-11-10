package examples

import (
	"database/sql"
)

// TypedDatabaseUsage uses non-standard variable names
// This should be caught by type checking, not heuristics
type TypedDatabaseUsage struct {
	connection *sql.DB
	trans      *sql.Tx
}

func (t *TypedDatabaseUsage) ProcessWithConnection() error {
	// Should be flagged via type checking (variable named "connection", not "db")
	_, err := t.connection.Exec("UPDATE users SET active = true")
	return err
}

func (t *TypedDatabaseUsage) ProcessWithTransaction() error {
	// Should be flagged via type checking (variable named "trans", not "tx")
	_, err := t.trans.Query("SELECT * FROM users")
	if err != nil {
		return err
	}
	return nil
}

func ProcessWithWeirdName(myDatabaseHandle *sql.DB) error {
	// Should be flagged via type checking (non-standard name)
	return myDatabaseHandle.Ping()
}

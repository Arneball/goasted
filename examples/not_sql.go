package examples

// MyExecutor is a custom type with an Exec method
// This should NOT be flagged
type MyExecutor struct {
	name string
}

func (m *MyExecutor) Exec(query string) error {
	// Some custom execution logic
	return nil
}

func UseCustomExecutor() {
	db := &MyExecutor{name: "custom"}

	// This should NOT be flagged - db is not *sql.DB
	// Type checking will correctly identify this
	db.Exec("some command")
}

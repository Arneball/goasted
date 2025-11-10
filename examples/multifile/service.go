package multifile

// Service uses Repository but doesn't import database/sql directly
type Service struct {
	repo *Repository
}

// GetUser calls a non-context method on the Connection field
// This file does NOT import database/sql
func (s *Service) GetUser(id int) error {
	// BAD: This should be flagged even though this file doesn't import database/sql
	// Field is named "Connection" (not "db"), so heuristics won't catch it
	_, err := s.repo.Connection.Query("SELECT * FROM users WHERE id = ?", id)
	return err
}

// UpdateUser also calls a non-context method
func (s *Service) UpdateUser(id int, name string) error {
	// BAD: This should also be flagged
	// Field is named "Connection" (not "db"), so heuristics won't catch it
	_, err := s.repo.Connection.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
	return err
}

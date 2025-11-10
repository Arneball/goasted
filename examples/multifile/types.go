package multifile

import "database/sql"

// Repository holds a database connection
type Repository struct {
	Connection *sql.DB
}

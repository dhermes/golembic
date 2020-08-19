package golembic

import (
	"database/sql"
)

// OptUpFromSQL returns an option that sets the `up` function to execute an
// SQL statement.
func OptUpFromSQL(statement string) Option {
	up := func(tx *sql.Tx) error {
		_, err := tx.Exec(statement)
		return err
	}

	return func(m *Migration) error {
		m.Up = up
		return nil
	}
}

// OptDownFromSQL returns an option that sets the `down` function to execute an
// SQL statement.
func OptDownFromSQL(statement string) Option {
	down := func(tx *sql.Tx) error {
		_, err := tx.Exec(statement)
		return err
	}

	return func(m *Migration) error {
		m.Down = down
		return nil
	}
}

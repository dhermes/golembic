package golembic

import (
	"context"
	"database/sql"
)

// OptUpFromSQL returns an option that sets the `up` function to execute an
// SQL statement.
func OptUpFromSQL(statement string) Option {
	up := func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, statement)
		return err
	}

	return func(m *Migration) error {
		m.Up = up
		return nil
	}
}

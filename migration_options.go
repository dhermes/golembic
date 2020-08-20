package golembic

import (
	"context"
	"database/sql"
)

// MigrationOption describes options used to create a new migration.
type MigrationOption = func(*Migration) error

// OptParent sets the parent on a migration.
func OptParent(parent string) MigrationOption {
	return func(m *Migration) error {
		m.Parent = parent
		return nil
	}
}

// OptRevision sets the revision on a migration.
func OptRevision(revision string) MigrationOption {
	return func(m *Migration) error {
		if revision == "" {
			return ErrMissingRevision
		}

		m.Revision = revision
		return nil
	}
}

// OptDescription sets the description on a migration.
func OptDescription(description string) MigrationOption {
	return func(m *Migration) error {
		m.Description = description
		return nil
	}
}

// OptUp sets the `up` function on a migration.
func OptUp(up UpMigration) MigrationOption {
	return func(m *Migration) error {
		if up == nil {
			return ErrNilInterface
		}

		m.Up = up
		return nil
	}
}

// OptUpFromSQL returns an option that sets the `up` function to execute an
// SQL statement.
func OptUpFromSQL(statement string) MigrationOption {
	up := func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, statement)
		return err
	}

	return func(m *Migration) error {
		m.Up = up
		return nil
	}
}
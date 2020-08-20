package golembic

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	// DefaultMetadataTable is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTable = "golembic_migrations"
)

// NewManager creates a new manager for orchestrating migrations. The variadic
// input `table` can be used
func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{MetadataTable: DefaultMetadataTable}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Manager orchestrates database operations done via `UpMigration` as well as
// supporting operations such as creating a table for migration metadata and
// writing rows into that metadata table during an `UpMigration.`
type Manager struct {
	MetadataTable string
	Connection    *sql.DB
	Provider      EngineProvider
	Sequence      *Migrations
}

// EnsureConnection returns a cached database connection (if already set) or
// creates and validates a new one.
func (m *Manager) EnsureConnection() (*sql.DB, error) {
	if m.Connection != nil {
		return m.Connection, nil
	}

	db, err := m.Provider.Open()
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	m.Connection = db
	return m.Connection, nil
}

// EnsureMigrationsTable checks that the migrations metadata table exists
// and creates it if not.
func (m *Manager) EnsureMigrationsTable(ctx context.Context) error {
	db, err := m.EnsureConnection()
	if err != nil {
		return err
	}

	return CreateMigrationsTable(ctx, db, m.Provider, m.MetadataTable)
}

// InsertMigration inserts a migration into the migrations metadata table.
func (m *Manager) InsertMigration(ctx context.Context, tx *sql.Tx, migration Migration) error {
	if migration.Parent == "" {
		statement := fmt.Sprintf(
			"INSERT INTO %s (parent, revision) VALUES (NULL, $1);",
			m.Provider.QuoteIdentifier(m.MetadataTable),
		)
		_, err := tx.ExecContext(ctx, statement, migration.Revision)
		return err
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (parent, revision) VALUES ($1, $2);",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	_, err := tx.ExecContext(ctx, statement, migration.Parent, migration.Revision)
	return err
}

// IsApplied checks if a migration has already been applied.
//
// NOTE: This assumes, but does not check, that the migrations metadata table
// exists.
func (m *Manager) IsApplied(ctx context.Context, tx *sql.Tx, revision string) (bool, error) {
	migration := m.Sequence.Get(revision)
	if migration == nil {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, revision)
		return false, err
	}

	query := fmt.Sprintf(
		"SELECT parent, revision FROM %s WHERE revision = $1;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query, revision)
	if err != nil {
		return false, err
	}

	return verifyMigration(rows, *migration)
}

func verifyMigration(rows []Migration, migration Migration) (bool, error) {
	if len(rows) == 0 {
		return false, nil
	}

	// NOTE: We don't verify that `len(rows) == 1` since we trust the UNIQUE
	//       index in the `revision` column.
	if rows[0].Parent != migration.Parent {
		err := fmt.Errorf(
			"%w; revision: %q, registered parent %q does not match parent %q from migrations table",
			ErrMigrationMismatch,
			migration.Revision,
			migration.Parent,
			rows[0].Parent,
		)
		return false, err
	}

	return true, nil
}

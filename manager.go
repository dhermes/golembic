package golembic

import (
	"context"
	"database/sql"
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

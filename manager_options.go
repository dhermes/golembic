package golembic

import (
	"database/sql"
)

// ManagerOption describes options used to create a new manager.
type ManagerOption = func(*Manager) error

// OptManagerMetadataTable sets the metadata table name on a manager.
func OptManagerMetadataTable(table string) ManagerOption {
	return func(m *Manager) error {
		m.MetadataTable = table
		return nil
	}
}

// OptManagerConnection sets the connection on a manager.
func OptManagerConnection(connection *sql.DB) ManagerOption {
	return func(m *Manager) error {
		m.Connection = connection
		return nil
	}
}

// OptManagerProvider sets the provider on a manager.
func OptManagerProvider(provider EngineProvider) ManagerOption {
	return func(m *Manager) error {
		m.Provider = provider
		return nil
	}
}

// OptManagerSequence sets the migrations sequence on a manager.
func OptManagerSequence(migrations *Migrations) ManagerOption {
	return func(m *Manager) error {
		m.Sequence = migrations
		return nil
	}
}

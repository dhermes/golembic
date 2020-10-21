package golembic

import (
	"database/sql"
)

// OptManagerMetadataTable sets the metadata table name on a manager.
func OptManagerMetadataTable(table string) ManagerOption {
	return func(m *Manager) error {
		m.MetadataTable = table
		return nil
	}
}

// OptManagerConnectionPool sets the connection pool on a manager.
func OptManagerConnectionPool(pool *sql.DB) ManagerOption {
	return func(m *Manager) error {
		m.ConnectionPool = pool
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

// OptManagerLog sets the logger interface on a manager. If `log` is `nil`
// the option will return an error.
func OptManagerLog(log PrintfReceiver) ManagerOption {
	return func(m *Manager) error {
		if log == nil {
			return ErrNilInterface
		}

		m.Log = log
		return nil
	}
}

// OptDevelopmentMode sets the development mode flag on a manager.
func OptDevelopmentMode(mode bool) ManagerOption {
	return func(m *Manager) error {
		m.DevelopmentMode = mode
		return nil
	}
}

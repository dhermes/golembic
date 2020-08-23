package golembic

import (
	"context"
	"database/sql"
)

// UpMigration defines a function interface to be used for up / forward
// migrations. The SQL transaction will be started **before** `UpMigration`
// is invoked and will be committed **after** the `UpMigration` exits without
// error. In addition to the contents of `UpMigration`, a row will be written
// to the migrations metadata table as part of the transaction.
//
// The expectation is that the migration runs SQL statements within the
// transaction. If a migration cannot run inside a transaction, e.g. a
// `CREATE UNIQUE INDEX CONCURRENTLY` statement, then the `UpMigration`
// interface should be used.
type UpMigration = func(context.Context, *sql.Tx) error

// UpMigrationConn defines a function interface to be used for up / forward
// migrations. This is the non-transactional form of `UpMigration` and
// should only be used in rare situations.
type UpMigrationConn = func(context.Context, *sql.Conn) error

// migrationsFilter defines a function interface that filters migrations
// based on the `latest` revision. It's expected that a migrations filter
// will enclose other state such as a `Manager`.
type migrationsFilter = func(latest string) ([]Migration, error)

// EngineProvider describes the interface required for a database engine.
type EngineProvider interface {
	// QuoteIdentifier quotes an identifier, such as a table name, for usage
	// in a query.
	QuoteIdentifier(name string) string
	// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
	// in a query.
	QuoteLiteral(literal string) string
	// Open creates a database connection for the engine provider.
	Open() (*sql.DB, error)
	// TableExistsSQL returns a SQL query that can be used to determine if a
	// table exists. It is expected to use a clause such as `WHERE tablename = $1`
	// to filter results.
	TableExistsSQL() string
}

// PrintfReceiver is a generic interface for logging and printing.
type PrintfReceiver interface {
	Printf(format string, a ...interface{}) (n int, err error)
}

// ManagerOption describes options used to create a new manager.
type ManagerOption = func(*Manager) error

// MigrationOption describes options used to create a new migration.
type MigrationOption = func(*Migration) error

// ApplyOption describes options used to create a apply configuration.
type ApplyOption = func(*ApplyConfig) error

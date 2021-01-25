package golembic

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
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
// will enclose other state such as a `Manager`. In addition to returning
// a slice of filtered migrations, it will also return a count of the
// number of existing migrations that were filtered out.
type migrationsFilter = func(latest string) (int, []Migration, error)

// EngineProvider describes the interface required for a database engine.
type EngineProvider interface {
	// QueryParameter produces a placeholder like `$1` or `?` for a numbered
	// parameter in a SQL query.
	QueryParameter(index int) string
	// NewCreateTableParameters produces column types and constraints for the
	// `CREATE TABLE` statement used to create the migrations table.
	NewCreateTableParameters() CreateTableParameters
	// TimestampColumn produces a value that can be used for reading / writing
	// the `created_at` column.
	TimestampColumn() TimestampColumn
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
	// or `WHERE tablename = ?` to filter results.
	TableExistsSQL() string
}

// PrintfReceiver is a generic interface for logging and printing. In cases
// where a trailing newline is desired (e.g. STDOUT), the type implemented
// `PrintfReceiver` must add the newline explicitly.
type PrintfReceiver interface {
	Printf(format string, a ...interface{}) (n int, err error)
}

// ManagerOption describes options used to create a new manager.
type ManagerOption = func(*Manager) error

// MigrationOption describes options used to create a new migration.
type MigrationOption = func(*Migration) error

// ApplyOption describes options used to create an apply configuration.
type ApplyOption = func(*ApplyConfig) error

// CreateTableOption describes options used for create table parameters.
type CreateTableOption = func(*CreateTableParameters)

// Column represents an abstract SQL column value; it requires a `Scan()`
// interface for reading and a `Value()` interface for writing.
type Column interface {
	sql.Scanner
	driver.Valuer
}

// TimestampColumn represents an abstract SQL column that stores a timestamp.
type TimestampColumn interface {
	Pointer() interface{}
	Timestamp() time.Time
}

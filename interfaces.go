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
// should only be used in rare situations. Note that the second argument is
// a `Conn` (vs. a `DB`). A DB is concurrency-safe because it represents a
// connection pool. However, we pass in a **single** connection because we
// need to guarantee that
type UpMigrationConn = func(context.Context, *sql.Conn) error

// NewConnection defines a function interface that can generate a new
// connection on demand.
type NewConnection = func(context.Context) (*sql.Conn, error)

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
	// SetConnTimeouts sets timeouts on a database connection to ensure that a
	// migration doesn't get stuck or cause the application to get blocked while
	// migrations are running.
	SetConnTimeouts(context.Context, *sql.Conn) error
	// SetTxTimeouts sets timeouts on a transaction to ensure that a migration
	// doesn't get stuck or cause the application to get blocked while
	// migrations are running.
	SetTxTimeouts(context.Context, *sql.Tx) error
	// TableExistsSQL returns a SQL query that can be used to determine if a
	// table exists. It is expected to use a clause such as `WHERE tablename = $1`
	// to filter results.
	TableExistsSQL() string
}

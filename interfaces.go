package golembic

import (
	"context"
	"database/sql"
)

// UpMigration defines a function interface that must be satisfied by
// up / forward migrations. The expectation as that the migration runs SQL
// statements within the transaction but this is not required. The SQL
// transaction will be started **before** `UpMigration` is invoked and will
// be committed **after** the `UpMigration` exits without error. In addition to
// the contents of `UpMigration`, a row will be written to the migrations
// metadata table as part of the transaction.
type UpMigration = func(context.Context, *sql.Tx) error

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
	// SetTxTimeouts sets timeouts on a transaction to ensure that a migration
	// doesn't get stuck or cause the application to get blocked while
	// migrations are running.
	SetTxTimeouts(context.Context, *sql.Tx) error
	// TableExistsSQL returns a SQL query that can be used to determine if a
	// table exists. It is expected to use a clause such as `WHERE tablename = $1`
	// to filter results.
	TableExistsSQL() string
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// NOTE: Importing `pq` comes with side effects, see
	//       https://github.com/lib/pq/blob/v1.8.0/conn.go#L51-L53.
	"github.com/lib/pq"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//       * `SQLProvider` satisfies `golembic.EngineProvider`.
var (
	_ golembic.EngineProvider = (*SQLProvider)(nil)
)

// New creates a PostgreSQL-specific database engine provider from some
// options.
func New(opts ...Option) (*SQLProvider, error) {
	cfg := &Config{
		Host:             DefaultHost,
		Port:             DefaultPort,
		Database:         DefaultDatabase,
		Schema:           DefaultSchema,
		LockTimeout:      DefaultLockTimeout,
		StatementTimeout: DefaultStatementTimeout,
		IdleConnections:  DefaultIdleConnections,
		MaxConnections:   DefaultMaxConnections,
		MaxLifetime:      DefaultMaxLifetime,
	}
	for _, opt := range opts {
		err := opt(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &SQLProvider{Config: cfg}, nil
}

// SQLProvider is a PostgreSQL-specific database engine provider.
type SQLProvider struct {
	Config *Config
}

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
func (sp *SQLProvider) QuoteIdentifier(name string) string {
	return pq.QuoteIdentifier(name)
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
func (sp *SQLProvider) QuoteLiteral(literal string) string {
	return pq.QuoteLiteral(literal)
}

// Open creates a database connection to a PostgreSQL instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	// NOTE: This requires that the `postgres` driver has been registered with
	//       the `sql` package.
	db, err := sql.Open("postgres", sp.Config.GetConnectionString())
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(sp.Config.MaxLifetime)
	db.SetMaxIdleConns(sp.Config.IdleConnections)
	db.SetMaxOpenConns(sp.Config.MaxConnections)
	return db, nil
}

// TableExistsSQL returns a SQL query that can be used to determine if a
// table exists.
func (sp *SQLProvider) TableExistsSQL() string {
	return fmt.Sprintf(
		"SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1 AND schemaname = %s",
		sp.QuoteLiteral(sp.Config.Schema),
	)
}

// SetConnTimeouts sets timeouts on a database connection to ensure that a
// migration doesn't get stuck or cause the application to get blocked while
// migrations are running.
func (sp *SQLProvider) SetConnTimeouts(ctx context.Context, conn *sql.Conn) error {
	err := sp.setLockTimeoutConn(ctx, conn)
	if err != nil {
		return err
	}

	return sp.setStatementTimeoutConn(ctx, conn)
}

// SetTxTimeouts sets timeouts on a transaction to ensure that a migration
// doesn't get stuck or cause the application to get blocked while migrations
// are running.
func (sp *SQLProvider) SetTxTimeouts(ctx context.Context, tx *sql.Tx) error {
	err := sp.setLockTimeoutTx(ctx, tx)
	if err != nil {
		return err
	}

	return sp.setStatementTimeoutTx(ctx, tx)
}

// setLockTimeoutTx invokes a `SET LOCAL lock_timeout` statement within a
// transaction.
//
//   golembic=> BEGIN;
//   BEGIN
//   golembic=> SET LOCAL lock_timeout TO '4000ms';
//   SET
//   golembic=> SHOW lock_timeout;
//    lock_timeout
//   --------------
//    4s
//   (1 row)
//   --
//   golembic=> SET LOCAL lock_timeout TO '4500ms';
//   SET
//   golembic=> SHOW lock_timeout;
//    lock_timeout
//   --------------
//    4500ms
//   (1 row)
//   --
//   golembic=> COMMIT;
//   COMMIT
//
// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
//
// Valid units for this parameter in PostgresSQL are "ms", "s", "min", "h"
// and "d" and the value should be between 0 and 2147483647ms. We explicitly
// cast to milliseconds but leave validation on the value to PostgreSQL.
func (sp *SQLProvider) setLockTimeoutTx(ctx context.Context, tx *sql.Tx) error {
	ms, err := toMilliseconds(sp.Config.LockTimeout)
	if err != nil {
		return err
	}

	timeout := fmt.Sprintf("%dms", ms)
	statement := fmt.Sprintf("SET LOCAL lock_timeout TO %s;", sp.QuoteLiteral(timeout))
	_, err = tx.ExecContext(ctx, statement)
	return err
}

// setLockTimeoutConn is the database connection equivalent of `setLockTimeoutTx`.
func (sp *SQLProvider) setLockTimeoutConn(ctx context.Context, conn *sql.Conn) error {
	ms, err := toMilliseconds(sp.Config.LockTimeout)
	if err != nil {
		return err
	}

	timeout := fmt.Sprintf("%dms", ms)
	statement := fmt.Sprintf("SET lock_timeout TO %s;", sp.QuoteLiteral(timeout))
	_, err = conn.ExecContext(ctx, statement)
	return err
}

// setStatementTimeoutTx invokes a `SET LOCAL statement_timeout` statement within
// a transaction.
//
// For more information on valid units, consult `setLockTimeoutTx()`.
//
// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
func (sp *SQLProvider) setStatementTimeoutTx(ctx context.Context, tx *sql.Tx) error {
	ms, err := toMilliseconds(sp.Config.StatementTimeout)
	if err != nil {
		return err
	}

	timeout := fmt.Sprintf("%dms", ms)
	statement := fmt.Sprintf("SET LOCAL statement_timeout TO %s;", sp.QuoteLiteral(timeout))
	_, err = tx.ExecContext(ctx, statement)
	return err
}

// setStatementTimeoutConn is the database connection equivalent of
// `setStatementTimeoutTx`.
func (sp *SQLProvider) setStatementTimeoutConn(ctx context.Context, conn *sql.Conn) error {
	ms, err := toMilliseconds(sp.Config.StatementTimeout)
	if err != nil {
		return err
	}

	timeout := fmt.Sprintf("%dms", ms)
	statement := fmt.Sprintf("SET statement_timeout TO %s;", sp.QuoteLiteral(timeout))
	_, err = conn.ExecContext(ctx, statement)
	return err
}

// toMilliseconds converts a duration to the **exact** number of milliseconds
// or errors if round off is required.
func toMilliseconds(d time.Duration) (int64, error) {
	remainder := d % time.Millisecond
	if remainder != 0 {
		err := fmt.Errorf("%w; duration: %s", ErrNotMilliseconds, d)
		return 0, err
	}

	ms := int64(d / time.Millisecond)
	return ms, nil
}

package postgres

import (
	"database/sql"
	"fmt"

	// NOTE: Importing `pq` comes with side effects, see
	//       https://github.com/lib/pq/blob/v1.8.0/conn.go#L51-L53.
	"github.com/lib/pq"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//       * `SQLProvider` satisfies `golembic.EngineProvider `.
var (
	_ golembic.EngineProvider = (*SQLProvider)(nil)
)

// New creates a PostgreSQL-specific database engine provider from some
// options.
func New(opts ...Option) (*SQLProvider, error) {
	cfg := &Config{
		Host:            DefaultHost,
		Port:            DefaultPort,
		Database:        DefaultDatabase,
		Schema:          DefaultSchema,
		LockTimeout:     DefaultLockTimeout,
		IdleConnections: DefaultIdleConnections,
		MaxConnections:  DefaultMaxConnections,
		MaxLifetime:     DefaultMaxLifetime,
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
func (sp *SQLProvider) TableExistsSQL(table string) string {
	return fmt.Sprintf(
		"SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = %s AND schemaname = %s",
		sp.QuoteLiteral(table),
		sp.QuoteLiteral(sp.Config.Schema),
	)
}

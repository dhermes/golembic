package postgres

import (
	"database/sql"
	"fmt"
	"strings"

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
		DriverName:       DefaultDriverName,
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
//
// This implementation is vendored in here to avoid the side effects of
// importing `github.com/lib/pq`. See:
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L1564-L1581
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L51-L53.
func (sp *SQLProvider) QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
//
// This implementation is vendored in here to avoid the side effects of
// importing `github.com/lib/pq`. See:
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L1583-L1614
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L51-L53.
func (sp *SQLProvider) QuoteLiteral(literal string) string {
	literal = strings.Replace(literal, `'`, `''`, -1)
	if strings.Contains(literal, `\`) {
		literal = strings.Replace(literal, `\`, `\\`, -1)
		literal = ` E'` + literal + `'`
	} else {
		literal = `'` + literal + `'`
	}
	return literal
}

// Open creates a database connection pool to a PostgreSQL instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	cs, err := sp.Config.GetConnectionString()
	if err != nil {
		return nil, err
	}

	// NOTE: This requires that the `postgres` driver has been registered with
	//       the `sql` package.
	pool, err := sql.Open(sp.Config.DriverName, cs)
	if err != nil {
		return nil, err
	}

	pool.SetConnMaxLifetime(sp.Config.MaxLifetime)
	pool.SetMaxIdleConns(sp.Config.IdleConnections)
	pool.SetMaxOpenConns(sp.Config.MaxConnections)
	return pool, nil
}

// TableExistsSQL returns a SQL query that can be used to determine if a
// table exists.
func (sp *SQLProvider) TableExistsSQL() string {
	if sp.Config.Schema != "" {
		return fmt.Sprintf(
			"SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1 AND schemaname = %s",
			sp.QuoteLiteral(sp.Config.Schema),
		)
	}

	return "SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1"
}

package postgres

import (
	"database/sql"
	"fmt"

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

// QueryParameter produces a placeholder like `$1` for a numbered
// parameter in a PostgreSQL query.
func (*SQLProvider) QueryParameter(index int) string {
	return fmt.Sprintf("$%d", index)
}

// NewCreateTableParameters produces the SQL expressions used in the
// the `CREATE TABLE` statement used to create the migrations table.
func (*SQLProvider) NewCreateTableParameters() golembic.CreateTableParameters {
	return golembic.NewCreateTableParameters(
		golembic.OptCreateTableCreatedAt("TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP"),
	)
}

// TimestampColumn produces a value that can be used for reading / writing
// a `TIMESTAMP` column to a `time.Time` in PostgreSQL.
func (*SQLProvider) TimestampColumn() golembic.TimestampColumn {
	return &golembic.TimeColumnPointer{}
}

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
func (*SQLProvider) QuoteIdentifier(name string) string {
	return golembic.QuoteIdentifier(name)
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
func (*SQLProvider) QuoteLiteral(literal string) string {
	return golembic.QuoteLiteral(literal)
}

// Open creates a database connection pool to a PostgreSQL instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	cs, err := sp.Config.GetConnectionString()
	if err != nil {
		return nil, err
	}

	// NOTE: This requires that the `DriverName` (`postgres`) driver has been
	//       registered with the `sql` package.
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

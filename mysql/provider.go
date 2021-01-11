package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//       * `SQLProvider` satisfies `golembic.EngineProvider`.
var (
	_ golembic.EngineProvider = (*SQLProvider)(nil)
)

// New creates a MySQL-specific database engine provider from some
// options.
func New(opts ...Option) (*SQLProvider, error) {
	sp := &SQLProvider{Config: &Config{ParseTime: true}}
	for _, opt := range opts {
		err := opt(sp)
		if err != nil {
			return nil, err
		}
	}

	return sp, nil
}

// SQLProvider is a MySQL-specific database engine provider.
type SQLProvider struct {
	Config *Config

	// IdleConnections is the number of idle connections.
	IdleConnections int
	// MaxConnections is the maximum number of connections.
	MaxConnections int
	// MaxLifetime is the maximum time a connection can be open.
	MaxLifetime time.Duration
}

// QueryParameter produces the placeholder `?` for a numbered
// parameter in a MySQL query.
func (*SQLProvider) QueryParameter(_ int) string {
	return "?"
}

// NewCreateTableParameters produces the SQL expressions used in the
// the `CREATE TABLE` statement used to create the migrations table.
func (*SQLProvider) NewCreateTableParameters() golembic.CreateTableParameters {
	return golembic.NewCreateTableParameters(
		golembic.OptCreateTableCreatedAt("TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6)"),
	)
}

// TimestampColumn produces a value that can be used for reading / writing
// a `TIMESTAMP` column to a `time.Time` in MySQL.
func (*SQLProvider) TimestampColumn() golembic.TimestampColumn {
	return &golembic.TimeColumnPointer{}
}

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
func (*SQLProvider) QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
func (*SQLProvider) QuoteLiteral(literal string) string {
	return golembic.QuoteLiteral(literal)
}

// Open creates a database connection pool to a MySQL instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	dsn := sp.Config.FormatDSN()

	// NOTE: This requires that the `mysql` driver has been registered with
	//       the `sql` package.
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	pool.SetConnMaxLifetime(sp.MaxLifetime)
	pool.SetMaxIdleConns(sp.IdleConnections)
	pool.SetMaxOpenConns(sp.MaxConnections)
	return pool, nil
}

// TableExistsSQL returns a SQL query that can be used to determine if a
// table exists.
func (sp *SQLProvider) TableExistsSQL() string {
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_name = ? AND table_schema = %s",
		sp.QuoteLiteral(sp.Config.DBName),
	)
}

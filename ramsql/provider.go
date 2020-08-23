package ramsql

import (
	"database/sql"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//       * `SQLProvider` satisfies `golembic.EngineProvider`.
var (
	_ golembic.EngineProvider = (*SQLProvider)(nil)
)

// New creates a RamSQL-specific database engine provider from some
// options.
func New(opts ...Option) (*SQLProvider, error) {
	cfg := &Config{
		DriverName: "ramsql",
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

// QuoteIdentifier is a no-op.
func (sp *SQLProvider) QuoteIdentifier(name string) string {
	return name
}

// QuoteLiteral is a no-op.
func (sp *SQLProvider) QuoteLiteral(literal string) string {
	return literal
}

// Open creates a database connection pool to a PostgreSQL instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	return sql.Open(sp.Config.DriverName, sp.Config.DataSourceName)
}

// TableExistsSQL is TODO.
func (sp *SQLProvider) TableExistsSQL() string {
	return "(TODO)"
}

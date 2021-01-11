package sqlite3

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

// New creates a SQLite-specific database engine provider from some
// options.
func New(opts ...Option) (*SQLProvider, error) {
	cfg := &Config{
		DataSourceName: DefaultDataSourceName,
		DriverName:     DefaultDriverName,
	}
	for _, opt := range opts {
		err := opt(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &SQLProvider{Config: cfg}, nil
}

// SQLProvider is a SQLite-specific database engine provider.
type SQLProvider struct {
	Config *Config
}

// QueryParameter produces the placeholder `?NNN` for a numbered
// parameter in a SQLite query.
//
// See: https://sqlite.org/lang_expr.html#parameters
func (*SQLProvider) QueryParameter(index int) string {
	return fmt.Sprintf("?%d", index)
}

// NewCreateTableParameters produces the SQL expressions used in the
// the `CREATE TABLE` statement used to create the migrations table.
func (*SQLProvider) NewCreateTableParameters() golembic.CreateTableParameters {
	ctp := golembic.NewCreateTableParameters(
		// H/T: https://stackoverflow.com/a/3693112/1068170
		golembic.OptCreateTableCreatedAt("INTEGER DEFAULT (CAST((julianday('now') - 2440587.5) * 86400.0 * 1000000 AS INTEGER))"),
		golembic.OptCreateTableSkip(true),
	)

	return ctp
}

// TimestampColumn produces a value that can be used for reading / writing
// an `INTEGER` column to a `time.Time` in SQLite.
func (*SQLProvider) TimestampColumn() golembic.TimestampColumn {
	return &TimeFromInteger{}
}

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
// See: https://www.sqlite.org/lang_keywords.html
func (sp *SQLProvider) QuoteIdentifier(name string) string {
	return golembic.QuoteIdentifier(name)
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
// See: https://www.sqlite.org/lang_keywords.html
func (sp *SQLProvider) QuoteLiteral(literal string) string {
	return golembic.QuoteLiteral(literal)
}

// Open creates a database connection pool to a SQLite instance.
func (sp *SQLProvider) Open() (*sql.DB, error) {
	return sql.Open(sp.Config.DriverName, sp.Config.DataSourceName)
}

// TableExistsSQL returns a SQL query that can be used to determine if a
// table exists.
//
// See:
// https://www.sqlite.org/fileformat2.html#storage_of_the_sql_database_schema
func (sp *SQLProvider) TableExistsSQL() string {
	return "SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?1;"
}

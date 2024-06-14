package sqlite3

const (
	// DefaultDataSourceName is the DSN / connection string for completely
	// in-memory values. Richer support for SQLite connection strings can be
	// added if desired; for reference: https://www.sqlite.org/uri.html
	DefaultDataSourceName = "file::memory:?cache=shared"

	// DefaultDriverName is the default SQL driver to be used when creating
	// a new database connection pool via `sql.Open()`. This default driver
	// is expected to be registered by importing one of
	// - github.com/gwenn/gosqlite
	// - github.com/mattn/go-sqlite3
	// - github.com/mxk/go-sqlite
	// - github.com/rsc/sqlite
	// - modernc.org/sqlite
	DefaultDriverName = "sqlite3"
)

// Config is a set of connection config options.
type Config struct {
	// DataSourceName is the DSN or connection string for a SQLite connection.
	DataSourceName string

	// DriverName specifies the name of SQL driver to be used when creating
	// a new database connection pool via `sql.Open()`. The default driver
	// is expected to be registered by importing `modernc.org/sqlite`,
	// though `https://github.com/golang/go/wiki/SQLDrivers` lists (as of
	// June 13, 2024) four other implementations that all register the
	// same driver name
	// - github.com/gwenn/gosqlite
	// - github.com/mattn/go-sqlite3
	// - github.com/mxk/go-sqlite
	// - github.com/rsc/sqlite
	DriverName string
}

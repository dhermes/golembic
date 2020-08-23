package ramsql

// Config is a set of connection config options.
type Config struct {
	// DataSourceName is a fully formed connection string.
	DataSourceName string

	// DriverName specifies the name of SQL driver to be used when creating
	// a new database connection pool via `sql.Open()`. The default driver
	// is expected to be registered by importing `github.com/proullon/ramsql/driver`,
	// however we may want to support other drivers that are compatible, such
	// as a driver wrapped with `github.com/ngrok/sqlmw`.
	DriverName string
}

package postgres

import (
	"fmt"
	"net/url"
	"time"

	"github.com/dhermes/golembic"
)

const (
	// DefaultHost is the default database hostname, typically used
	// when developing locally.
	DefaultHost = "localhost"
	// DefaultPort is the default postgres port.
	DefaultPort = "5432"
	// DefaultDatabase is the default database to connect to, we use
	// `postgres` to not pollute the template databases.
	DefaultDatabase = "postgres"

	// DefaultDriverName is the default SQL driver to be used when creating
	// a new database connection pool via `sql.Open()`. This default driver
	// is expected to be registered by importing `github.com/lib/pq`.
	DefaultDriverName = "postgres"

	// DefaultLockTimeout is the default timeout to use when attempting to
	// acquire a lock.
	DefaultLockTimeout = 4 * time.Second
	// DefaultStatementTimeout is the default timeout to use when invoking a
	// SQL statement.
	DefaultStatementTimeout = 5 * time.Second
	// DefaultIdleConnections is the default number of idle connections.
	DefaultIdleConnections = 16
	// DefaultMaxConnections is the default maximum number of connections.
	DefaultMaxConnections = 32
	// DefaultMaxLifetime is the default maximum lifetime of driver connections.
	//
	// If max lifetime <= 0, connections are not closed due to a connection's age.
	// See: https://github.com/golang/go/blob/go1.15/src/database/sql/sql.go#L940
	DefaultMaxLifetime = time.Duration(0)
)

// Config is a set of connection config options.
type Config struct {
	// ConnectionString is a fully formed connection string.
	ConnectionString string
	// Host is the server to connect to.
	Host string
	// Port is the port to connect to.
	Port string
	// Database is the database name
	Database string
	// Schema is the application schema within the database, defaults to `public`.
	Schema string
	// Username is the username for the connection via password auth.
	Username string
	// Password is the password for the connection via password auth.
	Password string
	// SSLMode is the SSL mode for the connection.
	SSLMode string

	// DriverName specifies the name of SQL driver to be used when creating
	// a new database connection pool via `sql.Open()`. The default driver
	// is expected to be registered by importing `github.com/lib/pq`, however
	// we may want to support other drivers that are wire compatible, such
	// as `github.com/jackc/pgx`.
	DriverName string

	// ConnectTimeout determines the maximum wait for connection. The minimum
	// allowed timeout is 2 seconds, so anything below is treated the same
	// as unset.
	//
	// See: https://www.postgresql.org/docs/10/libpq-connect.html#LIBPQ-CONNECT-CONNECT-TIMEOUT
	ConnectTimeout time.Duration
	// LockTimeout is the timeout to use when attempting to acquire a lock.
	//
	// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
	LockTimeout time.Duration
	// StatementTimeout is the timeout to use when invoking a SQL statement.
	//
	// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
	StatementTimeout time.Duration
	// IdleConnections is the number of idle connections.
	IdleConnections int
	// MaxConnections is the maximum number of connections.
	MaxConnections int
	// MaxLifetime is the maximum time a connection can be open.
	MaxLifetime time.Duration
}

// GetConnectionString creates a PostgreSQL connection string from the config.
// If `ConnectionString` is already cached on the `Config`, it will be returned
// immediately.
func (c Config) GetConnectionString() (string, error) {
	if c.ConnectionString != "" {
		return c.ConnectionString, nil
	}

	host := c.Host
	if c.Port != "" {
		host = host + ":" + c.Port
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   c.Database,
	}

	if len(c.Username) > 0 {
		if len(c.Password) > 0 {
			u.User = url.UserPassword(c.Username, c.Password)
		} else {
			u.User = url.User(c.Username)
		}
	}

	q := url.Values{}
	if len(c.SSLMode) > 0 {
		q.Add("sslmode", c.SSLMode)
	}
	if c.ConnectTimeout > 0 {
		err := SetTimeoutSeconds(q, "connect_timeout", c.ConnectTimeout)
		if err != nil {
			return "", err
		}
	}
	if c.LockTimeout > 0 {
		err := SetTimeoutMilliseconds(q, "lock_timeout", c.LockTimeout)
		if err != nil {
			return "", err
		}
	}
	if c.StatementTimeout > 0 {
		err := SetTimeoutMilliseconds(q, "statement_timeout", c.StatementTimeout)
		if err != nil {
			return "", err
		}
	}

	// NOTE: If no schema is specified, `postgres` will connect to the
	//       `"public"` schema.
	if c.Schema != "" {
		q.Add("search_path", c.Schema)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SetTimeoutMilliseconds sets a timeout value in connection string query parameters.
//
// Valid units for this parameter in PostgresSQL are "ms", "s", "min", "h"
// and "d" and the value should be between 0 and 2147483647ms. We explicitly
// cast to milliseconds but leave validation on the value to PostgreSQL.
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
// See:
// - https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
// - https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
func SetTimeoutMilliseconds(q url.Values, name string, d time.Duration) error {
	ms, err := golembic.ToRoundDuration(d, time.Millisecond)
	if err != nil {
		return err
	}

	q.Add(name, fmt.Sprintf("%dms", ms))
	return nil
}

// SetTimeoutSeconds sets a timeout value in connection string query parameters.
//
// This timeout is expected to be an exact number of seconds (as an integer)
// so we convert `d` to an integer first and set the value as a query parameter
// without units.
//
// See:
// - https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
func SetTimeoutSeconds(q url.Values, name string, d time.Duration) error {
	s, err := golembic.ToRoundDuration(d, time.Second)
	if err != nil {
		return err
	}

	q.Add(name, fmt.Sprintf("%d", s))
	return nil
}

package postgres

import (
	"net/url"
	"strconv"
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
	// ConnectTimeout is the connection timeout in seconds.
	ConnectTimeout int
	// SSLMode is the sslmode for the connection.
	SSLMode string
}

// HostOrDefault returns the postgres host for the connection or a default.
func (c Config) HostOrDefault() string {
	if c.Host != "" {
		return c.Host
	}
	return DefaultHost
}

// PortOrDefault returns the port for a connection if it is not the standard postgres port.
func (c Config) PortOrDefault() string {
	if c.Port != "" {
		return c.Port
	}
	return DefaultPort
}

// DatabaseOrDefault returns the connection database or a default.
func (c Config) DatabaseOrDefault(inherited ...string) string {
	if c.Database != "" {
		return c.Database
	}
	return DefaultDatabase
}

// GetConnectionString creates a PostgreSQL connection string from the config.
// If `ConnectionString` is already cached on the `Config`, it will be returned
// immediately.
func (c Config) GetConnectionString() string {
	if c.ConnectionString != "" {
		return c.ConnectionString
	}

	host := c.HostOrDefault()
	port := c.PortOrDefault()
	if port != "" {
		host = host + ":" + port
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   c.DatabaseOrDefault(),
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
		q.Add("connect_timeout", strconv.Itoa(c.ConnectTimeout))
	}
	if c.Schema != "" {
		q.Add("search_path", c.Schema)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

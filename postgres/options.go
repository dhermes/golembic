package postgres

// Option describes options used to create a new config for a SQL provider.
type Option = func(*Config) error

// OptHost sets the `Host` on a `Config`.
func OptHost(host string) Option {
	return func(cfg *Config) error {
		cfg.Host = host
		return nil
	}
}

// OptPort sets the `Port` on a `Config`.
func OptPort(port string) Option {
	return func(cfg *Config) error {
		cfg.Port = port
		return nil
	}
}

// OptDatabase sets the `Database` on a `Config`.
func OptDatabase(database string) Option {
	return func(cfg *Config) error {
		cfg.Database = database
		return nil
	}
}

// OptSchema sets the `Schema` on a `Config`.
func OptSchema(schema string) Option {
	return func(cfg *Config) error {
		cfg.Schema = schema
		return nil
	}
}

// OptUsername sets the `Username` on a `Config`.
func OptUsername(username string) Option {
	return func(cfg *Config) error {
		cfg.Username = username
		return nil
	}
}

// OptPassword sets the `Password` on a `Config`.
func OptPassword(password string) Option {
	return func(cfg *Config) error {
		cfg.Password = password
		return nil
	}
}

// OptConnectTimeout sets the `ConnectTimeout` on a `Config`.
func OptConnectTimeout(timeout int) Option {
	return func(cfg *Config) error {
		cfg.ConnectTimeout = timeout
		return nil
	}
}

// OptSSLMode sets the `SSLMode` on a `Config`.
func OptSSLMode(sslMode string) Option {
	return func(cfg *Config) error {
		cfg.SSLMode = sslMode
		return nil
	}
}

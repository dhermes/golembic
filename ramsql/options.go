package ramsql

// Option describes options used to create a new config for a SQL provider.
type Option = func(*Config) error

// OptDataSourceName sets the `DataSourceName` on a `Config`.
func OptDataSourceName(name string) Option {
	return func(cfg *Config) error {
		cfg.DataSourceName = name
		return nil
	}
}

// OptDriverName sets the `DriverName` on a `Config`.
func OptDriverName(name string) Option {
	return func(cfg *Config) error {
		cfg.DriverName = name
		return nil
	}
}

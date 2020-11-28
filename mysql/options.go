package mysql

import (
	"fmt"
	"time"
)

// Option describes options used to create a new config for a SQL provider.
type Option = func(*SQLProvider) error

// OptUser sets the `User` on a `Config`.
func OptUser(user string) Option {
	return func(sp *SQLProvider) error {
		sp.Config.User = user
		return nil
	}
}

// OptPassword sets the `Passwd` on a `Config`.
func OptPassword(password string) Option {
	return func(sp *SQLProvider) error {
		sp.Config.Passwd = password
		return nil
	}
}

// OptNet sets the `Net` on a `Config`.
func OptNet(net string) Option {
	return func(sp *SQLProvider) error {
		sp.Config.Net = net
		return nil
	}
}

// OptHostPort sets the `Addr` on a `Config`.
func OptHostPort(host string, port int) Option {
	return func(sp *SQLProvider) error {
		sp.Config.Addr = fmt.Sprintf("%s:%d", host, port)
		return nil
	}
}

// OptDBName sets the `DBName` on a `Config`.
func OptDBName(database string) Option {
	return func(sp *SQLProvider) error {
		sp.Config.DBName = database
		return nil
	}
}

// OptConfig sets the `Config` on a `SQLProvider`.
func OptConfig(cfg *Config) Option {
	return func(sp *SQLProvider) error {
		sp.Config = cfg
		return nil
	}
}

// OptIdleConnections sets the `IdleConnections` on a `SQLProvider`.
func OptIdleConnections(count int) Option {
	return func(sp *SQLProvider) error {
		sp.IdleConnections = count
		return nil
	}
}

// OptMaxConnections sets the `MaxConnections` on a `SQLProvider`.
func OptMaxConnections(count int) Option {
	return func(sp *SQLProvider) error {
		sp.MaxConnections = count
		return nil
	}
}

// OptMaxLifetime sets the `MaxLifetime` on a `SQLProvider`.
func OptMaxLifetime(d time.Duration) Option {
	return func(sp *SQLProvider) error {
		sp.MaxLifetime = d
		return nil
	}
}

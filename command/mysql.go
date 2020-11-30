package command

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/mysql"
	"github.com/spf13/cobra"
)

const (
	// EnvVarMySQLPassword is the environment variable used to supply a MySQL
	// password. Due to the sensitive nature of passwords, we don't support a
	// `--password` flag for passing along a password in plain text.
	EnvVarMySQLPassword = "DB_PASSWORD"
)

func mysqlPasswordFromEnv(cfg *mysql.Config) {
	password, exists := os.LookupEnv(EnvVarMySQLPassword)
	if !exists {
		return
	}

	cfg.Passwd = password
}

func mysqlSubCommand(manager *golembic.Manager, parent *cobra.Command, engine *string) (*cobra.Command, error) {
	provider, err := mysql.New()
	if err != nil {
		return nil, err
	}

	short := "Manage database migrations for a MySQL database"
	long := strings.Join([]string{
		short + ".",
		"",
		fmt.Sprintf("Use the %s environment variable to set the password for the database connection.", EnvVarMySQLPassword),
	}, "\n")
	cfg := provider.Config
	cfg.Net = "tcp" // Default to `tcp`
	host := "localhost"
	port := int(3306)
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: short,
		Long:  long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			*engine = "mysql"

			// NOTE: Manually invoke `PersistentPreRunE` on the parent to enable
			//       chaining (the behavior in `cobra` is to replace as the
			//       tree is traversed). See:
			//       - https://github.com/spf13/cobra/issues/216
			//       - https://github.com/spf13/cobra/issues/252
			if parent != nil && parent.PersistentPreRunE != nil {
				err := parent.PersistentPreRunE(cmd, args)
				if err != nil {
					return err
				}
			}

			opt := mysql.OptHostPort(host, port)
			err := opt(provider)
			if err != nil {
				return err
			}

			manager.Provider = provider
			mysqlPasswordFromEnv(cfg)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&cfg.Net,
		"protocol",
		cfg.Net,
		"The network protocol to use when connecting to MySQL",
	)
	cmd.PersistentFlags().StringVar(
		&host,
		"host",
		host,
		"The host to use when connecting to MySQL",
	)
	cmd.PersistentFlags().IntVar(
		&port,
		"port",
		port,
		"The port to use when connecting to MySQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.DBName,
		"dbname",
		cfg.DBName,
		"The database name to use when connecting to MySQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.User,
		"user",
		cfg.User,
		"The user to use when connecting to MySQL",
	)
	cmd.PersistentFlags().Var(
		&RoundDuration{Base: time.Millisecond, Value: &cfg.Timeout},
		"dial-timeout",
		"The timeout to use when waiting on a new connection to MySQL, must be exactly convertible to milliseconds",
	)
	// TODO: lock-timeout (if possible)
	// TODO: statement-timeout (if possible)
	cmd.PersistentFlags().IntVar(
		&provider.IdleConnections,
		"idle-connections",
		provider.IdleConnections,
		"The maximum number of idle connections (in a connection pool) to MySQL",
	)
	cmd.PersistentFlags().IntVar(
		&provider.MaxConnections,
		"max-connections",
		provider.MaxConnections,
		"The maximum number of connections (in a connection pool) to MySQL",
	)
	cmd.PersistentFlags().DurationVar(
		&provider.MaxLifetime,
		"max-lifetime",
		provider.MaxLifetime,
		"The maximum time a connection (from a connection pool) to MySQL can remain open",
	)

	return cmd, nil
}

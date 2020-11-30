package command

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/postgres"
)

const (
	// EnvVarPostgresPassword is the environment variable officially supported by
	// `psql` for a password. Due to the sensitive nature of passwords, we
	// don't support a `--password` flag for passing along a password in plain
	// text.
	EnvVarPostgresPassword = "PGPASSWORD"
)

func postgresPasswordFromEnv(cfg *postgres.Config) {
	password, exists := os.LookupEnv(EnvVarPostgresPassword)
	if !exists {
		return
	}

	cfg.Password = password
}

func postgresSubCommand(manager *golembic.Manager, parent *cobra.Command, engine *string) (*cobra.Command, error) {
	provider, err := postgres.New()
	if err != nil {
		return nil, err
	}

	short := "Manage database migrations for a PostgreSQL database"
	long := strings.Join([]string{
		short + ".",
		"",
		fmt.Sprintf("Use the %s environment variable to set the password for the database connection.", EnvVarPostgresPassword),
	}, "\n")
	cfg := provider.Config
	cmd := &cobra.Command{
		Use:   "postgres",
		Short: short,
		Long:  long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			*engine = "postgres"

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

			manager.Provider = provider
			postgresPasswordFromEnv(cfg)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&cfg.Host,
		"host",
		cfg.Host,
		"The host to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Port,
		"port",
		cfg.Port,
		"The port to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Database,
		"dbname",
		cfg.Database,
		"The database name to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Schema,
		"schema",
		cfg.Schema,
		"The schema to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Username,
		"username",
		cfg.Username,
		"The username to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.SSLMode,
		"ssl-mode",
		cfg.SSLMode,
		"The SSL mode to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.DriverName,
		"driver-name",
		cfg.DriverName,
		"The name of SQL driver to be used when creating a new database connection pool",
	)
	cmd.PersistentFlags().Var(
		&RoundDuration{Base: time.Second, Value: &cfg.ConnectTimeout},
		"connect-timeout",
		"The timeout to use when waiting on a new connection to PostgreSQL, must be exactly convertible to seconds",
	)
	cmd.PersistentFlags().Var(
		&RoundDuration{Base: time.Millisecond, Value: &cfg.LockTimeout},
		"lock-timeout",
		"The lock timeout to use when connecting to PostgreSQL, must be exactly convertible to milliseconds",
	)
	cmd.PersistentFlags().Var(
		&RoundDuration{Base: time.Millisecond, Value: &cfg.StatementTimeout},
		"statement-timeout",
		"The statement timeout to use when connecting to PostgreSQL, must be exactly convertible to milliseconds",
	)
	cmd.PersistentFlags().IntVar(
		&cfg.IdleConnections,
		"idle-connections",
		cfg.IdleConnections,
		"The maximum number of idle connections (in a connection pool) to PostgreSQL",
	)
	cmd.PersistentFlags().IntVar(
		&cfg.MaxConnections,
		"max-connections",
		cfg.MaxConnections,
		"The maximum number of connections (in a connection pool) to PostgreSQL",
	)
	cmd.PersistentFlags().DurationVar(
		&cfg.MaxLifetime,
		"max-lifetime",
		cfg.MaxLifetime,
		"The maximum time a connection (from a connection pool) to PostgreSQL can remain open",
	)

	return cmd, nil
}

package command

import (
	"fmt"
	"os"
	"strings"

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

func postgresSubCommand(manager *golembic.Manager, parent *cobra.Command) (*cobra.Command, error) {
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
		&cfg.Schema,
		"schema",
		cfg.Schema,
		"The schema to use when connecting to PostgreSQL",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.DriverName,
		"driver-name",
		cfg.DriverName,
		"The name of SQL driver to be used when creating a new database connection pool",
	)

	return cmd, nil
}

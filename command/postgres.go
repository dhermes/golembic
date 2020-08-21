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
	// TODO: Consider using `github.com/spf13/viper` for reading from
	//       environment variables.
	password, exists := os.LookupEnv(EnvVarPostgresPassword)
	if !exists {
		return
	}

	cfg.Password = password
}

func postgresSubCommand(manager *golembic.Manager, sqlDirectory *string) (*cobra.Command, error) {
	provider, err := postgres.New()
	if err != nil {
		return nil, err
	}

	short := "Manage database migrations for a PostgreSQL database"
	// TODO: Consider using `github.com/spf13/viper` for reading from
	//       environment variables. It's unclear if using viper would eliminate
	//       the need for the extra bit at the end of `long` below.
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			manager.Provider = provider
			postgresPasswordFromEnv(cfg)
		},
	}

	cmd.PersistentFlags().StringVar(
		&cfg.Host,
		"host",
		cfg.Host,
		"The host to use when connecting to PostgreSQL.",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Port,
		"port",
		cfg.Port,
		"The port to use when connecting to PostgreSQL.",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Database,
		"dbname",
		cfg.Database,
		"The database name to use when connecting to PostgreSQL.",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.Username,
		"username",
		cfg.Username,
		"The username to use when connecting to PostgreSQL.",
	)
	cmd.PersistentFlags().StringVar(
		&cfg.SSLMode,
		"ssl-mode",
		cfg.SSLMode,
		"The SSL mode to use when connecting to PostgreSQL.",
	)

	return cmd, nil
}

package command

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/dhermes/golembic"
)

// MakeRootCommand creates a `cobra` command that is bound to a sequence of
// migrations. The flags for the root command and relevant subcommands will
// be used to configure a `Manager`.
func MakeRootCommand(rm RegisterMigrations) (*cobra.Command, error) {
	if rm == nil {
		return nil, errors.New("Root command requires a non-nil migrations sequence")
	}

	manager, err := golembic.NewManager()
	engine := ""
	if err != nil {
		return nil, err
	}

	sqlDirectory := ""
	cmd := &cobra.Command{
		Use:           "golembic",
		Short:         "Manage database migrations for Go codebases",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: This will be replaced if any child commands also define
			//       `PersistentPreRunE`; this is written with the expectation
			//       that child commands will maintain a reference to parents
			//       if necessary and invoke this function. See:
			//       - https://github.com/spf13/cobra/issues/216
			//       - https://github.com/spf13/cobra/issues/252
			migrations, err := rm(sqlDirectory, engine)
			if err != nil {
				return err
			}

			manager.Sequence = migrations
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&manager.MetadataTable,
		"metadata-table",
		golembic.DefaultMetadataTable,
		"The name of the table that stores migration metadata",
	)

	cmd.PersistentFlags().StringVar(
		&sqlDirectory,
		"sql-directory",
		"",
		"Path to a directory containing \".sql\" migration files",
	)

	cmd.PersistentFlags().BoolVar(
		&manager.DevelopmentMode,
		"dev",
		false,
		"Flag indicating that the migrations should be run in development mode",
	)

	// Add PostgreSQL specific sub-commands.
	postgres, err := postgresSubCommand(manager, cmd, &engine)
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(postgres)
	registerProviderSubcommands(postgres, manager)
	// Add MySQL specific sub-commands.
	mysql, err := mysqlSubCommand(manager, cmd, &engine)
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(mysql)
	registerProviderSubcommands(mysql, manager)

	return cmd, nil
}

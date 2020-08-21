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

	sqlDirectory := ""
	manager := &golembic.Manager{}
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
			//       https://github.com/spf13/cobra/issues/216
			migrations, err := rm(sqlDirectory)
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
		"The name of the table that stores migration metadata.",
	)

	cmd.PersistentFlags().StringVar(
		&sqlDirectory,
		"sql-directory",
		"",
		"Path to a directory containing \".sql\" migration files.",
	)

	// Add provider specific sub-commands.
	postgres, err := postgresSubCommand(manager, cmd)
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(postgres)
	registerProviderSubcommands(postgres, manager)

	return cmd, nil
}

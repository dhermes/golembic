package command

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/dhermes/golembic"
)

// MakeRootCommand creates a `cobra` command that is bound to a sequence of
// migrations. The flags for the root command and relevant subcommands will
// be used to configure a `Manager`.
func MakeRootCommand(migrations *golembic.Migrations) (*cobra.Command, error) {
	if migrations == nil {
		return nil, errors.New("Root command requires a non-nil migrations sequence")
	}

	cmd := &cobra.Command{
		Use:          "golembic",
		Short:        "Manage database migrations for Go codebases",
		SilenceUsage: true,
	}

	manager := &golembic.Manager{Sequence: migrations}
	cmd.PersistentFlags().StringVar(
		&manager.MetadataTable,
		"metadata-table",
		golembic.DefaultMetadataTable,
		"The name of the table that stores migration metadata.",
	)

	sqlDirectory := ""
	cmd.PersistentFlags().StringVar(
		&sqlDirectory,
		"sql-directory",
		"",
		"Path to a directory containing \".sql\" migration files.",
	)

	// Add provider specific sub-commands.
	postgres, err := postgresSubCommand(manager, &sqlDirectory)
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(postgres)
	registerProviderSubcommands(postgres, manager)

	return cmd, nil
}

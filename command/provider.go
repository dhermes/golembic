package command

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/dhermes/golembic"
)

func registerProviderSubcommands(cmd *cobra.Command, manager *golembic.Manager) {
	cmd.AddCommand(
		describeSubCommand(manager),
		upSubCommand(manager),
		upOneSubCommand(manager),
		upToSubCommand(manager),
		verifySubCommand(manager),
		versionSubCommand(manager),
	)
}

func describeSubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the registered sequence of migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.Describe(ctx)
		},
	}
	return cmd
}

func upSubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all migrations that have not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.Up(ctx)
		},
	}
	return cmd
}

func upOneSubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up-one",
		Short: "Run the first migration that has not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.UpOne(ctx)
		},
	}
	return cmd
}

func upToSubCommand(manager *golembic.Manager) *cobra.Command {
	revision := ""
	cmd := &cobra.Command{
		Use:   "up-to",
		Short: "Run all the migrations up to a fixed revision that have not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.UpTo(ctx, revision)
		},
	}

	cmd.PersistentFlags().StringVar(
		&revision,
		"revision",
		"",
		"The revision to run migrations up to.",
	)
	cobra.MarkFlagRequired(cmd.PersistentFlags(), "revision")

	return cmd
}

func verifySubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the stored migration metadata against the registered sequence",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.Verify(ctx)
		},
	}
	return cmd
}

func versionSubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display the revision of the most recent migration to be applied",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return manager.Version(ctx)
		},
	}
	return cmd
}

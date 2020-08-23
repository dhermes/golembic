package command

import (
	"context"
	"fmt"
	"strings"

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
	short := "Describe the registered sequence of migrations"
	long := strings.Join([]string{
		short + ".",
		"",
		"This does not make any connection to the database. Use the",
		"`verify` command to compare registered migrations to history.",
	}, "\n")
	cmd := &cobra.Command{
		Use:   "describe",
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.Describe(ctx)
			return
		},
	}
	return cmd
}

func upSubCommand(manager *golembic.Manager) *cobra.Command {
	verifyHistory := false
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all migrations that have not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.Up(ctx, golembic.OptApplyVerifyHistory(verifyHistory))
			return
		},
	}

	addVerifyHistory(cmd, &verifyHistory)
	return cmd
}

func addVerifyHistory(cmd *cobra.Command, verifyHistory *bool) {
	cmd.PersistentFlags().BoolVar(
		verifyHistory,
		"verify-history",
		false,
		"If set, verify that all of the migration history matches the registered migrations",
	)
}

func upOneSubCommand(manager *golembic.Manager) *cobra.Command {
	verifyHistory := false
	cmd := &cobra.Command{
		Use:   "up-one",
		Short: "Run the first migration that has not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.UpOne(ctx, golembic.OptApplyVerifyHistory(verifyHistory))
			return
		},
	}

	addVerifyHistory(cmd, &verifyHistory)
	return cmd
}

func upToSubCommand(manager *golembic.Manager) *cobra.Command {
	verifyHistory := false
	revision := ""
	cmd := &cobra.Command{
		Use:   "up-to",
		Short: "Run all the migrations up to a fixed revision that have not yet been applied",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.UpTo(
				ctx,
				golembic.OptApplyRevision(revision),
				golembic.OptApplyVerifyHistory(verifyHistory),
			)
			return
		},
	}

	cmd.PersistentFlags().StringVar(
		&revision,
		"revision",
		"",
		"The revision to run migrations up to",
	)
	cobra.MarkFlagRequired(cmd.PersistentFlags(), "revision")

	addVerifyHistory(cmd, &verifyHistory)
	return cmd
}

func verifySubCommand(manager *golembic.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the stored migration metadata against the registered sequence",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.Verify(ctx)
			return
		},
	}
	return cmd
}

func versionSubCommand(manager *golembic.Manager) *cobra.Command {
	verifyHistory := false
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display the revision of the most recent migration to be applied",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				err = poolFinalize(manager, err)
			}()

			ctx := context.Background()
			err = manager.Version(ctx, golembic.OptApplyVerifyHistory(verifyHistory))
			return
		},
	}

	addVerifyHistory(cmd, &verifyHistory)
	return cmd
}

// poolFinalize is intended to be used in `defer` blocks to ensure that a SQL
// database connection pool on a manager is always closed after the manager is
// used.
func poolFinalize(manager *golembic.Manager, err error) error {
	closeErr := manager.CloseConnectionPool()
	return maybeWrap(err, closeErr, "failed to close connection pool")

}

// maybeWrap attempts to wrap a secondary error inside a primary one. If
// one (or both) of the errors if `nil`, then no wrapping is necessary.
//
// This has been copied directly from `github.com/dhermes/golembic:sql.go`
func maybeWrap(primary, secondary error, message string) error {
	if primary == nil {
		return secondary
	}
	if secondary == nil {
		return primary
	}

	return fmt.Errorf("%w; %s: %v", primary, message, secondary)
}

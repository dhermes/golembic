package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/postgres"
)

func allMigrations() (*golembic.Migrations, error) {
	sqlDir := mustEnvVar("GOLEMBIC_SQL_DIR")
	root := golembic.MustNewMigration(
		golembic.OptRevision("c9b52448285b"),
		golembic.OptDescription("Create users table"),
		golembic.OptUpFromFile(filepath.Join(sqlDir, "0001_create_users_table.sql")),
	)

	migrations, err := golembic.NewSequence(root)
	if err != nil {
		return nil, err
	}
	err = migrations.RegisterMany(
		golembic.MustNewMigration(
			golembic.OptParent("c9b52448285b"),
			golembic.OptRevision("f1be62155239"),
			golembic.OptDescription("Seed data in users table"),
			golembic.OptUpFromFile(filepath.Join(sqlDir, "0002_seed_users_table.sql")),
		),
		golembic.MustNewMigration(
			golembic.OptParent("f1be62155239"),
			golembic.OptRevision("dce8812d7b6f"),
			golembic.OptDescription("Add city column to users table"),
			golembic.OptUpFromFile(filepath.Join(sqlDir, "0003_add_users_city_column.sql")),
		),
		golembic.MustNewMigration(
			golembic.OptParent("dce8812d7b6f"),
			golembic.OptRevision("0430566018cc"),
			golembic.OptDescription("Rename the root user"),
			golembic.OptUpFromFile(filepath.Join(sqlDir, "0004_rename_root.sql")),
		),
		// https://github.com/dhermes/golembic/issues/10
		golembic.MustNewMigration(
			golembic.OptParent("0430566018cc"),
			golembic.OptRevision("0501ccd1d98c"),
			golembic.OptDescription("Add index on user emails"),
			golembic.OptUpFromSQL(`
ALTER TABLE users
  ADD CONSTRAINT uq_users_email UNIQUE (email);
`),
		),
		golembic.MustNewMigration(
			golembic.OptParent("0501ccd1d98c"),
			golembic.OptRevision("e2d4eecb1841"),
			golembic.OptDescription("Create books table"),
			golembic.OptUpFromFile(filepath.Join(sqlDir, "0006_create_books_table.sql")),
		),
		golembic.MustNewMigration(
			golembic.OptParent("e2d4eecb1841"),
			golembic.OptRevision("432f690fcbda"),
			golembic.OptDescription("Create movies table"),
			golembic.OptUpFromFile(filepath.Join(sqlDir, "0007_create_movies_table.sql")),
		),
	)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func mustEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable missing: %q", key)
	}
	return value
}

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type cmdMetadata struct {
	UpToRevision string
	RedoRevision string
}

func golembicCommand() (string, cmdMetadata) {
	cmdWith := mustEnvVar("GOLEMBIC_CMD")
	parts := strings.SplitN(cmdWith, ":", 2)
	cmd := parts[0]
	if len(parts) == 1 {
		return cmd, cmdMetadata{}
	}

	metadata := cmdMetadata{}
	switch cmd {
	case "up-to":
		metadata.UpToRevision = parts[1]
	case "redo":
		metadata.RedoRevision = parts[1]
	}
	return cmd, metadata
}

func main() {
	migrations, err := allMigrations()
	mustNil(err)

	provider, err := postgres.New(
		postgres.OptHost(mustEnvVar("DB_HOST")),
		postgres.OptPort(mustEnvVar("DB_PORT")),
		postgres.OptDatabase(mustEnvVar("DB_NAME")),
		postgres.OptUsername(mustEnvVar("DB_ADMIN_USER")),
		postgres.OptPassword(mustEnvVar("DB_ADMIN_PASSWORD")),
		postgres.OptSSLMode(mustEnvVar("DB_SSLMODE")),
	)
	m, err := golembic.NewManager(
		golembic.OptManagerProvider(provider),
		golembic.OptManagerSequence(migrations),
	)
	mustNil(err)

	cmd, metadata := golembicCommand()
	switch cmd {
	case "up":
		ctx := context.Background()
		err = m.Up(ctx)
		mustNil(err)
	case "up-one":
		ctx := context.Background()
		err = m.UpOne(ctx)
		mustNil(err)
	case "up-to":
		ctx := context.Background()
		err = m.UpTo(ctx, metadata.UpToRevision)
		mustNil(err)
	case "redo":
		ctx := context.Background()
		migration := m.Sequence.Get(metadata.RedoRevision)
		if migration == nil {
			mustNil(fmt.Errorf("Migration does not exist %q", metadata.RedoRevision))
		}
		err = m.ApplyMigration(ctx, *migration)
		mustNil(err)
	case "version":
		ctx := context.Background()
		migration, err := m.Version(ctx)
		mustNil(err)
		// TODO: https://github.com/dhermes/golembic/issues/1
		if migration == nil {
			log.Println("No migrations have been run")
		} else {
			log.Printf("%s: %s\n", migration.Revision, migration.Description)
		}
	case "describe":
		fmt.Println(migrations.Describe())
	default:
		err = fmt.Errorf("Invalid command: %q", cmd)
		mustNil(err)
	}
}

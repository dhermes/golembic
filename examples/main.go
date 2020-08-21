package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/command"
)

func mustEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable missing: %q", key)
	}
	return value
}

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
		golembic.MustNewMigration(
			golembic.OptParent("0430566018cc"),
			golembic.OptRevision("0501ccd1d98c"),
			golembic.OptDescription("Add index on user emails (concurrently)"),
			golembic.OptUpConnFromFile(filepath.Join(sqlDir, "0005_add_users_email_index_concurrently.sql")),
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

func mustNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	migrations, err := allMigrations()
	mustNil(err)

	cmd, err := command.MakeRootCommand(migrations)
	mustNil(err)
	err = cmd.Execute()
	mustNil(err)
}

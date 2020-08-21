package examples

import (
	"path/filepath"

	"github.com/dhermes/golembic"
)

// AllMigrations returns a sequence of migrations based on a directory
// containing `.sql` files.
func AllMigrations(sqlDirectory string) (*golembic.Migrations, error) {
	root, err := golembic.NewMigration(
		golembic.OptRevision("c9b52448285b"),
		golembic.OptDescription("Create users table"),
		golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0001_create_users_table.sql")),
	)
	if err != nil {
		return nil, err
	}

	migrations, err := golembic.NewSequence(*root)
	if err != nil {
		return nil, err
	}
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptParent("c9b52448285b"),
			golembic.OptRevision("f1be62155239"),
			golembic.OptDescription("Seed data in users table"),
			golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0002_seed_users_table.sql")),
		},
		[]golembic.MigrationOption{
			golembic.OptParent("f1be62155239"),
			golembic.OptRevision("dce8812d7b6f"),
			golembic.OptDescription("Add city column to users table"),
			golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0003_add_users_city_column.sql")),
		},
		[]golembic.MigrationOption{
			golembic.OptParent("dce8812d7b6f"),
			golembic.OptRevision("0430566018cc"),
			golembic.OptDescription("Rename the root user"),
			golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0004_rename_root.sql")),
		},
		[]golembic.MigrationOption{
			golembic.OptParent("0430566018cc"),
			golembic.OptRevision("0501ccd1d98c"),
			golembic.OptDescription("Add index on user emails (concurrently)"),
			golembic.OptUpConnFromFile(filepath.Join(sqlDirectory, "0005_add_users_email_index_concurrently.sql")),
		},
		[]golembic.MigrationOption{
			golembic.OptParent("0501ccd1d98c"),
			golembic.OptRevision("e2d4eecb1841"),
			golembic.OptDescription("Create books table"),
			golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0006_create_books_table.sql")),
		},
		[]golembic.MigrationOption{
			golembic.OptParent("e2d4eecb1841"),
			golembic.OptRevision("432f690fcbda"),
			golembic.OptDescription("Create movies table"),
			golembic.OptUpFromFile(filepath.Join(sqlDirectory, "0007_create_movies_table.sql")),
		},
	)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

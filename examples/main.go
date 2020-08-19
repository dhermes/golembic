package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dhermes/golembic"
	"github.com/dhermes/golembic/postgres"
)

func allMigrations() (*golembic.Migrations, error) {
	root := golembic.MustNewMigration(
		golembic.OptRevision("c9b52448285b"),
		golembic.OptDescription("Create users table"),
		golembic.OptUpFromSQL(`
CREATE TABLE users (
  user_id integer unique,
  name    varchar(40),
  email   varchar(40)
);
`),
	)

	migrations, err := golembic.NewSequence(root)
	if err != nil {
		return nil, err
	}
	err = migrations.RegisterMany(
		golembic.MustNewMigration(
			golembic.OptParent("c9b52448285b"),
			golembic.OptRevision("dce8812d7b6f"),
			golembic.OptDescription("Add city to users"),
			golembic.OptUpFromSQL(`
ALTER TABLE users ADD COLUMN city varchar(100);
`),
		),
		golembic.MustNewMigration(
			golembic.OptParent("dce8812d7b6f"),
			golembic.OptRevision("0501ccd1d98c"),
			golembic.OptDescription("Add index on user emails"),
			golembic.OptUpFromSQL(`
CREATE UNIQUE INDEX CONCURRENTLY uq_users_email ON users (email);
`),
		),
		golembic.MustNewMigration(
			golembic.OptParent("0501ccd1d98c"),
			golembic.OptRevision("e2d4eecb1841"),
			golembic.OptDescription("Create books table"),
			golembic.OptUpFromSQL(`
CREATE TABLE books (
  user_id integer,
  name    varchar(40),
  author  varchar(40)
);
`),
		),
		golembic.MustNewMigration(
			golembic.OptParent("e2d4eecb1841"),
			golembic.OptRevision("432f690fcbda"),
			golembic.OptDescription("Create movies table"),
			golembic.OptUpFromSQL(`
CREATE TABLE movies (
  user_id   integer,
  name      varchar(40),
  director  varchar(40)
);
`),
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
		panic(err)
	}
}

func main() {
	migrations, err := allMigrations()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(migrations.Describe())

	provider, err := postgres.New(
		postgres.OptHost(mustEnvVar("DB_HOST")),
		postgres.OptPort(mustEnvVar("DB_PORT")),
		postgres.OptDatabase(mustEnvVar("DB_NAME")),
		postgres.OptUsername(mustEnvVar("DB_ADMIN_USER")),
		postgres.OptPassword(mustEnvVar("DB_ADMIN_PASSWORD")),
		postgres.OptSSLMode(mustEnvVar("DB_SSLMODE")),
	)
	mustNil(err)

	ctx := context.Background()
	err = golembic.CreateMigrationsTable(ctx, provider)
	mustNil(err)
}

package main

import (
	"fmt"
	"log"

	"github.com/dhermes/golembic"
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
		golembic.OptDownFromSQL(`
DROP TABLE IF EXISTS users;
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
			golembic.OptDownFromSQL(`
ALTER TABLE users DROP COLUMN IF EXISTS city;
`),
		),
		golembic.MustNewMigration(
			golembic.OptParent("dce8812d7b6f"),
			golembic.OptRevision("0501ccd1d98c"),
			golembic.OptDescription("Add index on user emails"),
			golembic.OptUpFromSQL(`
CREATE UNIQUE INDEX CONCURRENTLY uq_users_email ON users (email);
`),
			golembic.OptDownFromSQL(`
DROP INDEX IF EXISTS uq_users_email;
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
			golembic.OptDownFromSQL(`
DROP TABLE IF EXISTS books;
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
			golembic.OptDownFromSQL(`
DROP TABLE IF EXISTS movies;
`),
		),
	)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func main() {
	migrations, err := allMigrations()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(migrations.Describe())
}

# golembic

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][11]][12]

## Usage

If a `*golembic.Migrations` sequence has been formed, then a binary can be
created as follows:

```go
func main() {
	migrations, err := allMigrations()
	mustNil(err)

	cmd, err := command.MakeRootCommand(migrations)
	mustNil(err)
	err = cmd.Execute()
	mustNil(err)
}
```

The root command of this binary has a subcommand for each provider

```
$ go build -o golembic ./examples/main.go
$ GOLEMBIC_SQL_DIR=./examples/sql/ ./golembic --help
Manage database migrations for Go codebases

Usage:
  golembic [command]

Available Commands:
  help        Help about any command
  postgres    Manage database migrations for a PostgreSQL database

Flags:
  -h, --help                    help for golembic
      --metadata-table string   The name of the table that stores migration metadata. (default "golembic_migrations")
      --sql-directory string    Path to a directory containing ".sql" migration files.

Use "golembic [command] --help" for more information about a command.
```

and the given subcommands have the same set of actions they can perform
(as subcommands)

```
$ GOLEMBIC_SQL_DIR=./examples/sql/ ./golembic postgres --help
Manage database migrations for a PostgreSQL database.

Use the PGPASSWORD environment variable to set the password for the database connection.

Usage:
  golembic postgres [command]

Available Commands:
  describe    Describe the registered sequence of migrations
  up          Run all migrations that have not yet been applied
  up-one      Run the first migration that has not yet been applied
  up-to       Run the all migrations up to a fixed revision that have not yet been applied
  verify      Verify the stored migration metadata against the registered sequence
  version     Display the revision of the most recent migration to be applied

Flags:
      --dbname string     The database name to use when connecting to PostgreSQL. (default "postgres")
  -h, --help              help for postgres
      --host string       The host to use when connecting to PostgreSQL. (default "localhost")
      --port string       The port to use when connecting to PostgreSQL. (default "5432")
      --ssl-mode string   The SSL mode to use when connecting to PostgreSQL.
      --username string   The username to use when connecting to PostgreSQL.

Global Flags:
      --metadata-table string   The name of the table that stores migration metadata. (default "golembic_migrations")
      --sql-directory string    Path to a directory containing ".sql" migration files.

Use "golembic postgres [command] --help" for more information about a command.
```

Some of the "leaf" commands have their own flags as well, but this is
uncommon:

```
$ GOLEMBIC_SQL_DIR=./examples/sql/ ./golembic postgres up-to --help
Run the all migrations up to a fixed revision that have not yet been applied

Usage:
  golembic postgres up-to [flags]

Flags:
  -h, --help              help for up-to
      --revision string   The revision to run migrations up to.

Global Flags:
      --dbname string           The database name to use when connecting to PostgreSQL. (default "postgres")
      --host string             The host to use when connecting to PostgreSQL. (default "localhost")
      --metadata-table string   The name of the table that stores migration metadata. (default "golembic_migrations")
      --port string             The port to use when connecting to PostgreSQL. (default "5432")
      --sql-directory string    Path to a directory containing ".sql" migration files.
      --ssl-mode string         The SSL mode to use when connecting to PostgreSQL.
      --username string         The username to use when connecting to PostgreSQL.
```

## Examples

### `up`

> **NOTE**: If `GOLEMBIC_CMD` is not provided to the `Makefile`, the default
> is `up`.

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 19:10:05 Applying c9b52448285b: Create users table
2020/08/20 19:10:05 Applying f1be62155239: Seed data in users table
2020/08/20 19:10:05 Applying dce8812d7b6f: Add city column to users table
2020/08/20 19:10:05 Applying 0430566018cc: Rename the root user
2020/08/20 19:10:05 Applying 0501ccd1d98c: Add index on user emails (concurrently)
2020/08/20 19:10:05 Applying e2d4eecb1841: Create books table
2020/08/20 19:10:05 Applying 432f690fcbda: Create movies table
```

After creation, the next run does nothing

```
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 19:10:26 No migrations to run; latest revision: 432f690fcbda
```

If we manually delete one, the last migration will get run

```
$ make psql-db
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '432f690fcbda';
DELETE 1
golembic=> DROP TABLE movies;
DROP TABLE
golembic=> \q
$
$
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 19:10:45 Applying 432f690fcbda: Create movies table
```

### `up-one`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:03 Applying c9b52448285b: Create users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:05 Applying f1be62155239: Seed data in users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:06 Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:08 Applying 0430566018cc: Rename the root user
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:10 Applying 0501ccd1d98c: Add index on user emails (concurrently)
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:11 Applying e2d4eecb1841: Create books table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:13 Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:11:16 No migrations to run; latest revision: 432f690fcbda
```

### `up-to`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision dce8812d7b6f"
2020/08/20 19:12:23 Applying c9b52448285b: Create users table
2020/08/20 19:12:23 Applying f1be62155239: Seed data in users table
2020/08/20 19:12:23 Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 0501ccd1d98c"
2020/08/20 19:12:41 Applying 0430566018cc: Rename the root user
2020/08/20 19:12:41 Applying 0501ccd1d98c: Add index on user emails (concurrently)
$
$ # TODO: Fix the way this is searched / the interval is determined
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 0430566018cc"
Error: No migration registered for revision; revision: "0501ccd1d98c"
Usage:
  golembic postgres up-to [flags]
...
2020/08/20 19:13:02 No migration registered for revision; revision: "0501ccd1d98c"
exit status 1
make: *** [run-examples-main] Error 1
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 432f690fcbda"
2020/08/20 19:13:41 Applying e2d4eecb1841: Create books table
2020/08/20 19:13:41 Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 432f690fcbda"
2020/08/20 19:13:43 No migrations to run; latest revision: 432f690fcbda
```

### `version`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 19:14:17 No migrations have been run
```

Then run **all** of the migrations and check the version

```
$ make run-examples-main GOLEMBIC_CMD=up
...
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 19:14:27 432f690fcbda: Create movies table (applied 2020-08-21 00:14:25.679836 +0000 UTC)
```

### `verify`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 19:14:40 :: 0 | c9b52448285b | Create users table (not yet applied)
2020/08/20 19:14:40 :: 1 | f1be62155239 | Seed data in users table (not yet applied)
2020/08/20 19:14:40 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 19:14:40 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 19:14:40 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
2020/08/20 19:14:40 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 19:14:40 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:14:51 Applying c9b52448285b: Create users table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 19:14:58 :: 0 | c9b52448285b | Create users table (applied 2020-08-21 00:14:51.490791 +0000 UTC)
2020/08/20 19:14:58 :: 1 | f1be62155239 | Seed data in users table (not yet applied)
2020/08/20 19:14:58 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 19:14:58 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 19:14:58 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
2020/08/20 19:14:58 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 19:14:58 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 19:15:08 Applying f1be62155239: Seed data in users table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 19:15:12 :: 0 | c9b52448285b | Create users table (applied 2020-08-21 00:14:51.490791 +0000 UTC)
2020/08/20 19:15:12 :: 1 | f1be62155239 | Seed data in users table (applied 2020-08-21 00:15:08.390287 +0000 UTC)
2020/08/20 19:15:12 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 19:15:12 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 19:15:12 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
2020/08/20 19:15:12 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 19:15:12 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 19:15:25 Applying dce8812d7b6f: Add city column to users table
2020/08/20 19:15:25 Applying 0430566018cc: Rename the root user
2020/08/20 19:15:25 Applying 0501ccd1d98c: Add index on user emails (concurrently)
2020/08/20 19:15:25 Applying e2d4eecb1841: Create books table
2020/08/20 19:15:25 Applying 432f690fcbda: Create movies table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 19:15:28 :: 0 | c9b52448285b | Create users table (applied 2020-08-21 00:14:51.490791 +0000 UTC)
2020/08/20 19:15:28 :: 1 | f1be62155239 | Seed data in users table (applied 2020-08-21 00:15:08.390287 +0000 UTC)
2020/08/20 19:15:28 :: 2 | dce8812d7b6f | Add city column to users table (applied 2020-08-21 00:15:25.87271 +0000 UTC)
2020/08/20 19:15:28 :: 3 | 0430566018cc | Rename the root user (applied 2020-08-21 00:15:25.890171 +0000 UTC)
2020/08/20 19:15:28 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently) (applied 2020-08-21 00:15:25.89843 +0000 UTC)
2020/08/20 19:15:28 :: 5 | e2d4eecb1841 | Create books table (applied 2020-08-21 00:15:25.914567 +0000 UTC)
2020/08/20 19:15:28 :: 6 | 432f690fcbda | Create movies table (applied 2020-08-21 00:15:25.921886 +0000 UTC)
```

We can artificially introduce a "new" migration and see failure to verify

```
$ make psql-db
...
golembic=> INSERT INTO golembic_migrations (parent, revision) VALUES ('432f690fcbda', 'not-in-sequence');
INSERT 0 1
golembic=> \q
$
$ make run-examples-main GOLEMBIC_CMD=verify
Error: Migration stored in SQL doesn't match sequence; sequence has 7 migrations but 8 are stored in the table
Usage:
  golembic postgres verify [flags]
...
2020/08/20 19:16:02 Migration stored in SQL doesn't match sequence; sequence has 7 migrations but 8 are stored in the table
exit status 1
make: *** [run-examples-main] Error 1
```

Similarly, if we can introduce an unknown entry "in sequence"

```
$ make psql-db
...
golembic=> DELETE FROM golembic_migrations WHERE revision = 'not-in-sequence';
DELETE 1
golembic=> DELETE FROM golembic_migrations WHERE revision = '432f690fcbda';
DELETE 1
golembic=> INSERT INTO golembic_migrations (parent, revision) VALUES ('e2d4eecb1841', 'not-in-sequence');
INSERT 0 1
golembic=> \q
$
$ make run-examples-main GOLEMBIC_CMD=verify
Error: Migration stored in SQL doesn't match sequence; stored migration 6: "not-in-sequence:e2d4eecb1841" does not match migration "432f690fcbda:e2d4eecb1841" in sequence
Usage:
  golembic postgres verify [flags]
...
2020/08/20 19:16:35 Migration stored in SQL doesn't match sequence; stored migration 6: "not-in-sequence:e2d4eecb1841" does not match migration "432f690fcbda:e2d4eecb1841" in sequence
exit status 1
make: *** [run-examples-main] Error 1
```

Luckily more painful cases such as one migration being deleted "in the middle"
are protected by the constraints on the table:

```
$ make psql-db
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '0430566018cc';
ERROR:  update or delete on table "golembic_migrations" violates foreign key constraint "fk_golembic_migrations_parent" on table "golembic_migrations"
DETAIL:  Key (revision)=(0430566018cc) is still referenced from table "golembic_migrations".
golembic=> \q
```

### `describe`

```
$ make run-examples-main GOLEMBIC_CMD=describe
2020/08/20 19:17:00 :: 0 | c9b52448285b | Create users table
2020/08/20 19:17:00 :: 1 | f1be62155239 | Seed data in users table
2020/08/20 19:17:00 :: 2 | dce8812d7b6f | Add city column to users table
2020/08/20 19:17:00 :: 3 | 0430566018cc | Rename the root user
2020/08/20 19:17:00 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently)
2020/08/20 19:17:00 :: 5 | e2d4eecb1841 | Create books table
2020/08/20 19:17:00 :: 6 | 432f690fcbda | Create movies table
```

## Development

```
$ make
Makefile for `golembic` project

Usage:
   make dev-deps               Install (or upgrade) development time dependencies
   make vet                    Run `go vet` over source tree
   make start-docker-db        Starts a PostgreSQL database running in a Docker container
   make superuser-migration    Run superuser migration
   make run-migrations         Run all migrations
   make start-db               Run start-docker-db, and migration target(s)
   make stop-db                Stops the PostgreSQL database running in a Docker container
   make restart-db             Stops the PostgreSQL database (if running) and starts a fresh Docker container
   make require-db             Determine if PostgreSQL database is running; fail if not
   make psql-db                Connects to currently running PostgreSQL DB via `psql`
   make run-examples-main      Run `./examples/main.go`

```

## Resources and Inspiration

-   `alembic` [tutorial][1]
-   `goose` [package][2]
-   Blog [post][3]: Move fast and migrate things: how we automated migrations
    in Postgres (in particular, the notes about lock timeouts)
-   Blog [post][4]: Update your Database Schema Without Downtime
-   Blog [post][5]: Multiple heads in alembic migrations - what to do
-   StackOverflow [answer][7] about setting a [lock timeout][8] and
    [statement timeout][9] in Postgres
    ```sql
    BEGIN;
    SET LOCAL lock_timeout TO '4s';
    SET LOCAL statement_timeout TO '5s';
    SELECT * FROM users;
    COMMIT;
    ```
-   Blog [post][10]: When Postgres blocks: 7 tips for dealing with locks

![Multiple Revision Heads][6]

[1]: https://alembic.sqlalchemy.org/en/latest/tutorial.html
[2]: https://github.com/pressly/goose
[3]: https://benchling.engineering/move-fast-and-migrate-things-how-we-automated-migrations-in-postgres-d60aba0fc3d4
[4]: https://thorben-janssen.com/update-database-schema-without-downtime/
[5]: https://blog.jerrycodes.com/multiple-heads-in-alembic-migrations/
[6]: images/multiple-heads.png
[7]: https://stackoverflow.com/a/20963803/1068170
[8]: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
[9]: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
[10]: https://www.citusdata.com/blog/2018/02/22/seven-tips-for-dealing-with-postgres-locks/
[11]: https://godoc.org/github.com/dhermes/golembic?status.svg
[12]: https://godoc.org/github.com/dhermes/golembic

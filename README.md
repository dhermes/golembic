# golembic

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][11]][12]

## Usage

If a `*golembic.Migrations` sequence has been formed, then a binary can be
created as follows:

```go
func main() {
	cmd, err := command.MakeRootCommand(allMigrations)
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
```

The root command of this binary has a subcommand for each provider

```
$ go build -o golembic ./examples/main.go
$ ./golembic --help
Manage database migrations for Go codebases

Usage:
  golembic [command]

Available Commands:
  help        Help about any command
  postgres    Manage database migrations for a PostgreSQL database

Flags:
  -h, --help                    help for golembic
      --metadata-table string   The name of the table that stores migration metadata (default "golembic_migrations")
      --sql-directory string    Path to a directory containing ".sql" migration files

Use "golembic [command] --help" for more information about a command.
```

and the given subcommands have the same set of actions they can perform
(as subcommands)

```
$ ./golembic postgres --help
Manage database migrations for a PostgreSQL database.

Use the PGPASSWORD environment variable to set the password for the database connection.

Usage:
  golembic postgres [command]

Available Commands:
  describe    Describe the registered sequence of migrations
  up          Run all migrations that have not yet been applied
  up-one      Run the first migration that has not yet been applied
  up-to       Run all the migrations up to a fixed revision that have not yet been applied
  verify      Verify the stored migration metadata against the registered sequence
  version     Display the revision of the most recent migration to be applied

Flags:
      --dbname string        The database name to use when connecting to PostgreSQL (default "postgres")
      --driver-name string   The name of SQL driver to be used when creating a new database connection pool (default "postgres")
  -h, --help                 help for postgres
      --host string          The host to use when connecting to PostgreSQL (default "localhost")
      --port string          The port to use when connecting to PostgreSQL (default "5432")
      --ssl-mode string      The SSL mode to use when connecting to PostgreSQL
      --username string      The username to use when connecting to PostgreSQL

Global Flags:
      --metadata-table string   The name of the table that stores migration metadata (default "golembic_migrations")
      --sql-directory string    Path to a directory containing ".sql" migration files

Use "golembic postgres [command] --help" for more information about a command.
```

Some of the "leaf" commands have their own flags as well, but this is
uncommon:

```
$ ./golembic postgres up-to --help
Run all the migrations up to a fixed revision that have not yet been applied

Usage:
  golembic postgres up-to [flags]

Flags:
  -h, --help              help for up-to
      --revision string   The revision to run migrations up to

Global Flags:
      --dbname string           The database name to use when connecting to PostgreSQL (default "postgres")
      --driver-name string      The name of SQL driver to be used when creating a new database connection pool (default "postgres")
      --host string             The host to use when connecting to PostgreSQL (default "localhost")
      --metadata-table string   The name of the table that stores migration metadata (default "golembic_migrations")
      --port string             The port to use when connecting to PostgreSQL (default "5432")
      --sql-directory string    Path to a directory containing ".sql" migration files
      --ssl-mode string         The SSL mode to use when connecting to PostgreSQL
      --username string         The username to use when connecting to PostgreSQL
```

## Examples

### `up`

> **NOTE**: If `GOLEMBIC_CMD` is not provided to the `Makefile`, the default
> is `up`.

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up
Applying c9b52448285b: Create users table
Applying f1be62155239: Seed data in users table
Applying dce8812d7b6f: Add city column to users table
Applying 0430566018cc: Rename the root user
Applying 0501ccd1d98c: Add index on user emails (concurrently)
Applying e2d4eecb1841: Create books table
Applying 432f690fcbda: Create movies table
```

After creation, the next run does nothing

```
$ make run-examples-main GOLEMBIC_CMD=up
No migrations to run; latest revision: 432f690fcbda
```

If we manually delete one, the last migration will get run

```
$ make psql
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '432f690fcbda';
DELETE 1
golembic=> DROP TABLE movies;
DROP TABLE
golembic=> \q
$
$
$ make run-examples-main GOLEMBIC_CMD=up
Applying 432f690fcbda: Create movies table
```

### `up-one`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying c9b52448285b: Create users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying f1be62155239: Seed data in users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying 0430566018cc: Rename the root user
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying 0501ccd1d98c: Add index on user emails (concurrently)
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying e2d4eecb1841: Create books table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
No migrations to run; latest revision: 432f690fcbda
```

### `up-to`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision dce8812d7b6f"
Applying c9b52448285b: Create users table
Applying f1be62155239: Seed data in users table
Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 0501ccd1d98c"
Applying 0430566018cc: Rename the root user
Applying 0501ccd1d98c: Add index on user emails (concurrently)
$
$ # TODO: Fix the way this is searched / the interval is determined
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 0430566018cc"
2020/08/21 00:33:28 No migration registered for revision; revision: "0501ccd1d98c"
exit status 1
make: *** [run-examples-main] Error 1
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 432f690fcbda"
Applying e2d4eecb1841: Create books table
Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-to GOLEMBIC_ARGS="--revision 432f690fcbda"
No migrations to run; latest revision: 432f690fcbda
```

### `version`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=version
 No migrations have been run
```

Then run **all** of the migrations and check the version

```
$ make run-examples-main GOLEMBIC_CMD=up
...
$ make run-examples-main GOLEMBIC_CMD=version
432f690fcbda: Create movies table (applied 2020-08-21 05:34:41.98568 +0000 UTC)
```

### `verify`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=verify
0 | c9b52448285b | Create users table (not yet applied)
1 | f1be62155239 | Seed data in users table (not yet applied)
2 | dce8812d7b6f | Add city column to users table (not yet applied)
3 | 0430566018cc | Rename the root user (not yet applied)
4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
5 | e2d4eecb1841 | Create books table (not yet applied)
6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying c9b52448285b: Create users table
$ make run-examples-main GOLEMBIC_CMD=verify
0 | c9b52448285b | Create users table (applied 2020-08-21 05:35:08.846686 +0000 UTC)
1 | f1be62155239 | Seed data in users table (not yet applied)
2 | dce8812d7b6f | Add city column to users table (not yet applied)
3 | 0430566018cc | Rename the root user (not yet applied)
4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
5 | e2d4eecb1841 | Create books table (not yet applied)
6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
Applying f1be62155239: Seed data in users table
$ make run-examples-main GOLEMBIC_CMD=verify
0 | c9b52448285b | Create users table (applied 2020-08-21 05:35:08.846686 +0000 UTC)
1 | f1be62155239 | Seed data in users table (applied 2020-08-21 05:35:19.509364 +0000 UTC)
2 | dce8812d7b6f | Add city column to users table (not yet applied)
3 | 0430566018cc | Rename the root user (not yet applied)
4 | 0501ccd1d98c | Add index on user emails (concurrently) (not yet applied)
5 | e2d4eecb1841 | Create books table (not yet applied)
6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up
Applying dce8812d7b6f: Add city column to users table
Applying 0430566018cc: Rename the root user
Applying 0501ccd1d98c: Add index on user emails (concurrently)
Applying e2d4eecb1841: Create books table
Applying 432f690fcbda: Create movies table
$ make run-examples-main GOLEMBIC_CMD=verify
0 | c9b52448285b | Create users table (applied 2020-08-21 05:35:08.846686 +0000 UTC)
1 | f1be62155239 | Seed data in users table (applied 2020-08-21 05:35:19.509364 +0000 UTC)
2 | dce8812d7b6f | Add city column to users table (applied 2020-08-21 05:35:56.149084 +0000 UTC)
3 | 0430566018cc | Rename the root user (applied 2020-08-21 05:35:56.159952 +0000 UTC)
4 | 0501ccd1d98c | Add index on user emails (concurrently) (applied 2020-08-21 05:35:56.169768 +0000 UTC)
5 | e2d4eecb1841 | Create books table (applied 2020-08-21 05:35:56.190074 +0000 UTC)
6 | 432f690fcbda | Create movies table (applied 2020-08-21 05:35:56.199978 +0000 UTC)
```

We can artificially introduce a "new" migration and see failure to verify

```
$ make psql
...
golembic=> INSERT INTO golembic_migrations (parent, revision) VALUES ('432f690fcbda', 'not-in-sequence');
INSERT 0 1
golembic=> \q
$
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/21 00:36:21 Migration stored in SQL doesn't match sequence; sequence has 7 migrations but 8 are stored in the table
exit status 1
make: *** [run-examples-main] Error 1
```

Similarly, if we can introduce an unknown entry "in sequence"

```
$ make psql
...
golembic=> DELETE FROM golembic_migrations WHERE revision IN ('not-in-sequence', '432f690fcbda');
DELETE 2
golembic=> INSERT INTO golembic_migrations (parent, revision) VALUES ('e2d4eecb1841', 'not-in-sequence');
INSERT 0 1
golembic=> \q
$
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/21 00:36:41 Migration stored in SQL doesn't match sequence; stored migration 6: "not-in-sequence:e2d4eecb1841" does not match migration "432f690fcbda:e2d4eecb1841" in sequence
exit status 1
make: *** [run-examples-main] Error 1
```

Luckily more painful cases such as one migration being deleted "in the middle"
are protected by the constraints on the table:

```
$ make psql
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '0430566018cc';
ERROR:  update or delete on table "golembic_migrations" violates foreign key constraint "fk_golembic_migrations_parent" on table "golembic_migrations"
DETAIL:  Key (revision)=(0430566018cc) is still referenced from table "golembic_migrations".
golembic=> \q
```

### `describe`

```
$ make run-examples-main GOLEMBIC_CMD=describe
0 | c9b52448285b | Create users table
1 | f1be62155239 | Seed data in users table
2 | dce8812d7b6f | Add city column to users table
3 | 0430566018cc | Rename the root user
4 | 0501ccd1d98c | Add index on user emails (concurrently)
5 | e2d4eecb1841 | Create books table
6 | 432f690fcbda | Create movies table
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
   make psql                   Connects to currently running PostgreSQL DB via `psql`
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

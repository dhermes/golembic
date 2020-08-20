# golembic

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][11]][12]

## Usage

### `up`

> **NOTE**: If `GOLEMBIC_CMD` is not provided, the default is `up`.

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 10:04:41 Applying c9b52448285b: Create users table
2020/08/20 10:04:41 Applying f1be62155239: Seed data in users table
2020/08/20 10:04:41 Applying dce8812d7b6f: Add city column to users table
2020/08/20 10:04:41 Applying 0430566018cc: Rename the root user
2020/08/20 10:04:41 Applying 0501ccd1d98c: Add index on user emails
2020/08/20 10:04:41 Applying e2d4eecb1841: Create books table
2020/08/20 10:04:41 Applying 432f690fcbda: Create movies table
```

After creation, the next run does nothing

```
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 10:04:54 No migrations to run; latest revision: 432f690fcbda
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
2020/08/20 10:05:24 Applying 432f690fcbda: Create movies table
```

### `up-one`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:43 Applying c9b52448285b: Create users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:46 Applying f1be62155239: Seed data in users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:48 Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:50 Applying 0430566018cc: Rename the root user
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:52 Applying 0501ccd1d98c: Add index on user emails
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:54 Applying e2d4eecb1841: Create books table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:56 Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 10:05:58 No migrations to run; latest revision: 432f690fcbda
```

### `up-to`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up-to:dce8812d7b6f
2020/08/20 10:07:02 Applying c9b52448285b: Create users table
2020/08/20 10:07:02 Applying f1be62155239: Seed data in users table
2020/08/20 10:07:02 Applying dce8812d7b6f: Add city column to users table
$
$ make run-examples-main GOLEMBIC_CMD=up-to:0501ccd1d98c
2020/08/20 10:07:12 Applying 0430566018cc: Rename the root user
2020/08/20 10:07:12 Applying 0501ccd1d98c: Add index on user emails
$
$ # TODO: Fix the way this is searched / the interval is determined
$ make run-examples-main GOLEMBIC_CMD=up-to:0430566018cc
2020/08/20 10:08:09 No migration registered for revision; revision: "0501ccd1d98c"
exit status 1
make: *** [run-examples-main] Error 1
$
$ make run-examples-main GOLEMBIC_CMD=up-to:432f690fcbda
2020/08/20 10:08:21 Applying e2d4eecb1841: Create books table
2020/08/20 10:08:21 Applying 432f690fcbda: Create movies table
$
$ make run-examples-main GOLEMBIC_CMD=up-to:432f690fcbda
2020/08/20 10:08:23 No migrations to run; latest revision: 432f690fcbda
```

### `redo`

First apply all of the migrations

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=up
...
```

Then try to redo (won't work if the tables are already there)

```
$ make run-examples-main GOLEMBIC_CMD=redo:432f690fcbda
2020/08/20 10:09:00 Applying 432f690fcbda: Create movies table
2020/08/20 10:09:00 pq: relation "movies" already exists
exit status 1
make: *** [run-examples-main] Error 1
```

Manually drop the table, **then** redo

```
$ make psql-db
...
golembic=> DROP TABLE movies;
DROP TABLE
golembic=> \q
$
$ # TODO: Add a flag that changes this to `UPDATE` instead of `INSERT`.
$ make run-examples-main GOLEMBIC_CMD=redo:432f690fcbda
2020/08/20 10:09:35 Applying 432f690fcbda: Create movies table
2020/08/20 10:09:35 pq: duplicate key value violates unique constraint "pk_golembic_migrations_revision"
exit status 1
make: *** [run-examples-main] Error 1
```

Go one step further and actually delete the row (in addition to the table)

```
$ make psql-db
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '432f690fcbda';
DELETE 1
golembic=> \q
$
$ make run-examples-main GOLEMBIC_CMD=redo:432f690fcbda
2020/08/20 10:09:53 Applying 432f690fcbda: Create movies table
```

Failure mode

```
$ make run-examples-main GOLEMBIC_CMD=redo:sentinel
2020/08/20 10:10:04 Migration does not exist "sentinel"
exit status 1
make: *** [run-examples-main] Error 1
```

### `version`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 10:37:05 No migrations have been run
```

Then run **all** of the migrations and check the version

```
$ make run-examples-main GOLEMBIC_CMD=up
...
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 10:37:22 432f690fcbda: Create movies table (applied 2020-08-20 15:37:18.740333 +0000 UTC)
```

### `verify`

```
$ make restart-db
...
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 11:04:10 :: 0 | c9b52448285b | Create users table (not yet applied)
2020/08/20 11:04:10 :: 1 | f1be62155239 | Seed data in users table (not yet applied)
2020/08/20 11:04:10 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 11:04:10 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 11:04:10 :: 4 | 0501ccd1d98c | Add index on user emails (not yet applied)
2020/08/20 11:04:10 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 11:04:10 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 11:04:21 Applying c9b52448285b: Create users table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 11:04:25 :: 0 | c9b52448285b | Create users table (applied 2020-08-20 16:04:21.869794 +0000 UTC)
2020/08/20 11:04:25 :: 1 | f1be62155239 | Seed data in users table (not yet applied)
2020/08/20 11:04:25 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 11:04:25 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 11:04:25 :: 4 | 0501ccd1d98c | Add index on user emails (not yet applied)
2020/08/20 11:04:25 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 11:04:25 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up-one
2020/08/20 11:04:30 Applying f1be62155239: Seed data in users table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 11:04:32 :: 0 | c9b52448285b | Create users table (applied 2020-08-20 16:04:21.869794 +0000 UTC)
2020/08/20 11:04:32 :: 1 | f1be62155239 | Seed data in users table (applied 2020-08-20 16:04:30.783097 +0000 UTC)
2020/08/20 11:04:32 :: 2 | dce8812d7b6f | Add city column to users table (not yet applied)
2020/08/20 11:04:32 :: 3 | 0430566018cc | Rename the root user (not yet applied)
2020/08/20 11:04:32 :: 4 | 0501ccd1d98c | Add index on user emails (not yet applied)
2020/08/20 11:04:32 :: 5 | e2d4eecb1841 | Create books table (not yet applied)
2020/08/20 11:04:32 :: 6 | 432f690fcbda | Create movies table (not yet applied)
$
$
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 11:04:36 Applying dce8812d7b6f: Add city column to users table
2020/08/20 11:04:36 Applying 0430566018cc: Rename the root user
2020/08/20 11:04:36 Applying 0501ccd1d98c: Add index on user emails
2020/08/20 11:04:36 Applying e2d4eecb1841: Create books table
2020/08/20 11:04:36 Applying 432f690fcbda: Create movies table
$ make run-examples-main GOLEMBIC_CMD=verify
2020/08/20 11:04:38 :: 0 | c9b52448285b | Create users table (applied 2020-08-20 16:04:21.869794 +0000 UTC)
2020/08/20 11:04:38 :: 1 | f1be62155239 | Seed data in users table (applied 2020-08-20 16:04:30.783097 +0000 UTC)
2020/08/20 11:04:38 :: 2 | dce8812d7b6f | Add city column to users table (applied 2020-08-20 16:04:36.121622 +0000 UTC)
2020/08/20 11:04:38 :: 3 | 0430566018cc | Rename the root user (applied 2020-08-20 16:04:36.129109 +0000 UTC)
2020/08/20 11:04:38 :: 4 | 0501ccd1d98c | Add index on user emails (applied 2020-08-20 16:04:36.136544 +0000 UTC)
2020/08/20 11:04:38 :: 5 | e2d4eecb1841 | Create books table (applied 2020-08-20 16:04:36.144876 +0000 UTC)
2020/08/20 11:04:38 :: 6 | 432f690fcbda | Create movies table (applied 2020-08-20 16:04:36.152191 +0000 UTC)
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
2020/08/20 11:07:12 Migration stored in SQL doesn't match sequence; sequence has 7 migrations but 8 are stored in the table
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
2020/08/20 11:12:12 Migration stored in SQL doesn't match sequence; stored migration 6: "not-in-sequence:e2d4eecb1841" does not match migration "432f690fcbda:e2d4eecb1841" in sequence
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
2020/08/20 11:02:35 :: 0 | c9b52448285b | Create users table
2020/08/20 11:02:35 :: 1 | f1be62155239 | Seed data in users table
2020/08/20 11:02:35 :: 2 | dce8812d7b6f | Add city column to users table
2020/08/20 11:02:35 :: 3 | 0430566018cc | Rename the root user
2020/08/20 11:02:35 :: 4 | 0501ccd1d98c | Add index on user emails
2020/08/20 11:02:35 :: 5 | e2d4eecb1841 | Create books table
2020/08/20 11:02:35 :: 6 | 432f690fcbda | Create movies table
```

### Invalid Command

```
$ make run-examples-main GOLEMBIC_CMD=baz
2020/08/20 10:11:10 Invalid command: "baz"
exit status 1
make: *** [run-examples-main] Error 1
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

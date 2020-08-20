# golembic

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][11]][12]

## Usage

## `up`

> **NOTE**: If `GOLEMBIC_CMD` is not provided, the default is `up`.

```
$ make restart-db
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 00:46:14 Applying c9b52448285b: Create users table
2020/08/20 00:46:14 Applying dce8812d7b6f: Add city to users
2020/08/20 00:46:14 Applying 0501ccd1d98c: Add index on user emails
2020/08/20 00:46:14 Applying e2d4eecb1841: Create books table
2020/08/20 00:46:14 Applying 432f690fcbda: Create movies table
```

After creation, the next run does nothing

```
$ make run-examples-main GOLEMBIC_CMD=up
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
2020/08/20 00:49:06 Applying 432f690fcbda: Create movies table
```

## `describe`

```
$ make run-examples-main GOLEMBIC_CMD=describe
0 | c9b52448285b | Create users table
1 | dce8812d7b6f | Add city to users
2 | 0501ccd1d98c | Add index on user emails
3 | e2d4eecb1841 | Create books table
4 | 432f690fcbda | Create movies table
```

## Invalid Command

```
$ make run-examples-main GOLEMBIC_CMD=baz
2020/08/20 00:51:59 Invalid command: "baz"
exit status 1
make: *** [run-examples-main] Error 1
```

## Development

```
$ make
Makefile for `golembic` project

Usage:
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

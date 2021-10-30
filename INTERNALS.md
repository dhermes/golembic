# Internals of `golembic`

## Starting a Container

```
$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
$
$
$ make start-postgres
Network dev-network-golembic created.
Container dev-postgres-golembic started on port 18426.
Container dev-postgres-golembic added to network dev-network-golembic.
Container dev-postgres-golembic accepting Postgres connections.
CREATE ROLE golembic_admin
...
REVOKE ROLE
$
$
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED          STATUS          PORTS                     NAMES
5510cfdc52e4   postgres:13.1-alpine   "docker-entrypoint.s…"   22 seconds ago   Up 20 seconds   0.0.0.0:18426->5432/tcp   dev-postgres-golembic
```

## Run `./examples/cmd/main.go`

```
$ make psql
...
golembic=> \dt
Did not find any relations.
golembic=> \q
$
$
$ make run-postgres-cmd GOLEMBIC_CMD=describe
0 | c9b52448285b | Create users table
1 | f1be62155239 | Seed data in users table
2 | dce8812d7b6f | Add city column to users table
3 | 0430566018cc | Rename the root user [MILESTONE]
4 | 0501ccd1d98c | Add index on user emails (concurrently)
5 | e2d4eecb1841 | Create books table
6 | 432f690fcbda | Create movies table
$ make run-postgres-cmd GOLEMBIC_CMD=version
No migrations have been run
```

## Migrations Metadata Table Created by Version Check

```
$ make psql
...
golembic=> \dt
                   List of relations
 Schema |        Name         | Type  |     Owner
--------+---------------------+-------+----------------
 public | golembic_migrations | table | golembic_admin
(1 row)

golembic=> \d+ golembic_migrations
                                            Table "public.golembic_migrations"
   Column   |           Type           | Collation | Nullable |      Default      | Storage  | Stats target | Description
------------+--------------------------+-----------+----------+-------------------+----------+--------------+-------------
 serial_id  | integer                  |           | not null |                   | plain    |              |
 revision   | character varying(32)    |           | not null |                   | extended |              |
 previous   | character varying(32)    |           |          |                   | extended |              |
 created_at | timestamp with time zone |           |          | CURRENT_TIMESTAMP | plain    |              |
Indexes:
    "pk_golembic_migrations_revision" PRIMARY KEY, btree (revision)
    "uq_golembic_migrations_previous" UNIQUE CONSTRAINT, btree (previous)
    "uq_golembic_migrations_serial_id" UNIQUE CONSTRAINT, btree (serial_id)
Check constraints:
    "chk_golembic_migrations_null_previous" CHECK (serial_id = 0 AND previous IS NULL OR serial_id <> 0 AND previous IS NOT NULL)
    "chk_golembic_migrations_previous_neq_revision" CHECK (previous::text <> revision::text)
    "chk_golembic_migrations_serial_id" CHECK (serial_id >= 0)
Foreign-key constraints:
    "fk_golembic_migrations_previous" FOREIGN KEY (previous) REFERENCES golembic_migrations(revision)
Referenced by:
    TABLE "golembic_migrations" CONSTRAINT "fk_golembic_migrations_previous" FOREIGN KEY (previous) REFERENCES golembic_migrations(revision)
Access method: heap

golembic=> SELECT * FROM golembic_migrations;
 serial_id | revision | previous | created_at
-----------+----------+----------+------------
(0 rows)

golembic=> \q
```

## Run Some Migrations

```
$ make run-postgres-cmd GOLEMBIC_CMD=up
Applying c9b52448285b: Create users table
Applying f1be62155239: Seed data in users table
Applying dce8812d7b6f: Add city column to users table
Applying 0430566018cc: Rename the root user [MILESTONE]
Applying 0501ccd1d98c: Add index on user emails (concurrently)
Applying e2d4eecb1841: Create books table
Applying 432f690fcbda: Create movies table
```

Observe the tables created by the migrations

```
$ make psql
...
golembic=> \dt
                   List of relations
 Schema |        Name         | Type  |     Owner
--------+---------------------+-------+----------------
 public | books               | table | golembic_admin
 public | golembic_migrations | table | golembic_admin
 public | movies              | table | golembic_admin
 public | users               | table | golembic_admin
(4 rows)

golembic=> \d+ users
                                             Table "public.users"
   Column   |          Type          | Collation | Nullable | Default | Storage  | Stats target | Description
------------+------------------------+-----------+----------+---------+----------+--------------+-------------
 id         | integer                |           | not null |         | plain    |              |
 email      | character varying(40)  |           |          |         | extended |              |
 first_name | character varying(40)  |           | not null |         | extended |              |
 last_name  | character varying(40)  |           | not null |         | extended |              |
 city       | character varying(100) |           |          |         | extended |              |
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "uq_users_email" UNIQUE, btree (email)
Access method: heap

golembic=> SELECT * FROM users;
   id   |        email         | first_name | last_name | city
--------+----------------------+------------+-----------+------
  83917 | dhermes@mail.invalid | Danny      | Hermes    |
 109203 |                      | admin      |           |
(2 rows)

golembic=> \d+ books
                                           Table "public.books"
 Column  |         Type          | Collation | Nullable | Default | Storage  | Stats target | Description
---------+-----------------------+-----------+----------+---------+----------+--------------+-------------
 user_id | integer               |           |          |         | plain    |              |
 name    | character varying(40) |           |          |         | extended |              |
 author  | character varying(40) |           |          |         | extended |              |

golembic=> SELECT * FROM books;
 user_id | name | author
---------+------+--------
(0 rows)

golembic=> \d+ movies
                                           Table "public.movies"
  Column  |         Type          | Collation | Nullable | Default | Storage  | Stats target | Description
----------+-----------------------+-----------+----------+---------+----------+--------------+-------------
 user_id  | integer               |           |          |         | plain    |              |
 name     | character varying(40) |           |          |         | extended |              |
 director | character varying(40) |           |          |         | extended |              |

golembic=> SELECT * FROM movies;
 user_id | name | director
---------+------+----------
(0 rows)

golembic=> \q
```

And see how these migrations are tracked

```
$ make run-postgres-cmd GOLEMBIC_CMD=version
432f690fcbda: Create movies table (applied 2021-01-25 17:27:55.015148 +0000 UTC)
$
$ make psql
...
golembic=> SELECT * FROM golembic_migrations;
 serial_id |   revision   |   previous   |          created_at
-----------+--------------+--------------+-------------------------------
         0 | c9b52448285b |              | 2021-01-25 17:27:54.951363+00
         1 | f1be62155239 | c9b52448285b | 2021-01-25 17:27:54.962251+00
         2 | dce8812d7b6f | f1be62155239 | 2021-01-25 17:27:54.970523+00
         3 | 0430566018cc | dce8812d7b6f | 2021-01-25 17:27:54.979727+00
         4 | 0501ccd1d98c | 0430566018cc | 2021-01-25 17:27:54.987617+00
         5 | e2d4eecb1841 | 0501ccd1d98c | 2021-01-25 17:27:55.006433+00
         6 | 432f690fcbda | e2d4eecb1841 | 2021-01-25 17:27:55.015148+00
(7 rows)

golembic=> \q
```

## Stop the Database

```
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED         STATUS         PORTS                     NAMES
5510cfdc52e4   postgres:13.1-alpine   "docker-entrypoint.s…"   4 minutes ago   Up 4 minutes   0.0.0.0:18426->5432/tcp   dev-postgres-golembic
$ make stop-postgres
Container dev-postgres-golembic stopped.
Network dev-network-golembic stopped.
$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
$ make stop-postgres
Container dev-postgres-golembic is not currently running.
Network dev-network-golembic is not currently running.
```

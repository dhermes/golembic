# Internals of `golembic`

## Starting a Container

```
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
$
$
$ make start-db
Container dev-postgres-golembic started on port 18426.
Container dev-postgres-golembic accepting Postgres connections.
CREATE ROLE golembic_admin
...
REVOKE ROLE
$
$
$ docker ps
CONTAINER ID        IMAGE                  COMMAND                  CREATED             STATUS              PORTS                     NAMES
9849e034b79d        postgres:10.6-alpine   "docker-entrypoint.s…"   4 seconds ago       Up 3 seconds        0.0.0.0:18426->5432/tcp   dev-postgres-golembic
```

## Run `./examples/main.go`

```
$ make psql
...
golembic=> \dt
Did not find any relations.
golembic=> \q
$
$
$ make run-examples-main GOLEMBIC_CMD=describe
2020/08/20 19:18:08 :: 0 | c9b52448285b | Create users table
2020/08/20 19:18:08 :: 1 | f1be62155239 | Seed data in users table
2020/08/20 19:18:08 :: 2 | dce8812d7b6f | Add city column to users table
2020/08/20 19:18:08 :: 3 | 0430566018cc | Rename the root user
2020/08/20 19:18:08 :: 4 | 0501ccd1d98c | Add index on user emails (concurrently)
2020/08/20 19:18:08 :: 5 | e2d4eecb1841 | Create books table
2020/08/20 19:18:08 :: 6 | 432f690fcbda | Create movies table
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 19:18:18 No migrations have been run
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
 revision   | character varying(32)    |           | not null |                   | extended |              |
 parent     | character varying(32)    |           |          |                   | extended |              |
 created_at | timestamp with time zone |           |          | CURRENT_TIMESTAMP | plain    |              |
Indexes:
    "pk_golembic_migrations_revision" PRIMARY KEY, btree (revision)
    "uq_golembic_migrations_parent" UNIQUE CONSTRAINT, btree (parent)
Check constraints:
    "chk_golembic_migrations_parent_neq_revision" CHECK (parent::text <> revision::text)
Foreign-key constraints:
    "fk_golembic_migrations_parent" FOREIGN KEY (parent) REFERENCES golembic_migrations(revision)
Referenced by:
    TABLE "golembic_migrations" CONSTRAINT "fk_golembic_migrations_parent" FOREIGN KEY (parent) REFERENCES golembic_migrations(revision)

golembic=> SELECT * FROM golembic_migrations;
 revision | parent | created_at
----------+--------+------------
(0 rows)

golembic=> \q
```

## Run Some Migrations

```
$ make run-examples-main GOLEMBIC_CMD=up
2020/08/20 19:18:45 Applying c9b52448285b: Create users table
2020/08/20 19:18:45 Applying f1be62155239: Seed data in users table
2020/08/20 19:18:45 Applying dce8812d7b6f: Add city column to users table
2020/08/20 19:18:45 Applying 0430566018cc: Rename the root user
2020/08/20 19:18:45 Applying 0501ccd1d98c: Add index on user emails (concurrently)
2020/08/20 19:18:45 Applying e2d4eecb1841: Create books table
2020/08/20 19:18:45 Applying 432f690fcbda: Create movies table
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
 Column  |          Type          | Collation | Nullable | Default | Storage  | Stats target | Description
---------+------------------------+-----------+----------+---------+----------+--------------+-------------
 user_id | integer                |           |          |         | plain    |              |
 name    | character varying(40)  |           |          |         | extended |              |
 email   | character varying(40)  |           |          |         | extended |              |
 city    | character varying(100) |           |          |         | extended |              |
Indexes:
    "uq_users_email" UNIQUE CONSTRAINT, btree (email)
    "users_user_id_key" UNIQUE CONSTRAINT, btree (user_id)

golembic=> SELECT * FROM users;
 user_id |  name   |        email         | city
---------+---------+----------------------+------
       1 | dhermes | dhermes@mail.invalid |
       0 | admin   |                      |
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
$ make run-examples-main GOLEMBIC_CMD=version
2020/08/20 19:19:47 432f690fcbda: Create movies table (applied 2020-08-21 00:18:45.134451 +0000 UTC)
$
$ make psql
...
golembic=> SELECT * FROM golembic_migrations;
   revision   |    parent    |          created_at
--------------+--------------+-------------------------------
 c9b52448285b |              | 2020-08-21 00:18:45.072003+00
 f1be62155239 | c9b52448285b | 2020-08-21 00:18:45.082318+00
 dce8812d7b6f | f1be62155239 | 2020-08-21 00:18:45.090277+00
 0430566018cc | dce8812d7b6f | 2020-08-21 00:18:45.097912+00
 0501ccd1d98c | 0430566018cc | 2020-08-21 00:18:45.105813+00
 e2d4eecb1841 | 0501ccd1d98c | 2020-08-21 00:18:45.126928+00
 432f690fcbda | e2d4eecb1841 | 2020-08-21 00:18:45.134451+00
(7 rows)

golembic=> \q
```

## Stop the Database

```
$ docker ps
CONTAINER ID        IMAGE                  COMMAND                  CREATED             STATUS              PORTS                     NAMES
9849e034b79d        postgres:10.6-alpine   "docker-entrypoint.s…"   2 minutes ago       Up 2 minutes        0.0.0.0:18426->5432/tcp   dev-postgres-golembic
$ make stop-db
Container dev-postgres-golembic stopped.
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
$ make stop-db
Container dev-postgres-golembic is not currently running.
```

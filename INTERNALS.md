# Internals of `golembic`

```
$ make start-db
Container dev-postgres-golembic started on port 18426.
Container dev-postgres-golembic accepting Postgres connections.
CREATE ROLE golembic_admin
...
REVOKE ROLE
$
$
$ make run-examples-main
0 | c9b52448285b | Create users table
1 | dce8812d7b6f | Add city to users
2 | 0501ccd1d98c | Add index on user emails
3 | e2d4eecb1841 | Create books table
4 | 432f690fcbda | Create movies table
$
$
$ make psql-db
...
golembic=> \dt
                   List of relations
 Schema |        Name         | Type  |     Owner
--------+---------------------+-------+----------------
 public | golembic_migrations | table | golembic_admin
(1 row)

golembic=> \d+ golembic_migrations
                                                           Table "public.golembic_migrations"
   Column   |           Type           | Collation | Nullable |                     Default                     | Storage  | Stats target | Description
------------+--------------------------+-----------+----------+-------------------------------------------------+----------+--------------+-------------
 id         | integer                  |           | not null | nextval('golembic_migrations_id_seq'::regclass) | plain    |              |
 revision   | character varying(32)    |           | not null |                                                 | extended |              |
 created_at | timestamp with time zone |           |          | CURRENT_TIMESTAMP                               | plain    |              |
Indexes:
    "pk_golembic_migrations_id" PRIMARY KEY, btree (id)

golembic=> \q
```

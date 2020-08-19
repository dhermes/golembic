package golembic

import (
	"database/sql"
)

// UpMigration defines a function interface that must be satisfied by
// up / forward migrations. The expectation as that the migration runs SQL
// statements within the transaction but this is not required.
type UpMigration = func(*sql.Tx) error

// DownMigration defines a function interface that must be satisfied by
// down / reverse / rollback migrations. The expectation as that the migration
// runs SQL statements within the transaction but this is not required.
type DownMigration = func(*sql.Tx) error

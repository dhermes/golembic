package command

import (
	"github.com/dhermes/golembic"
)

// RegisterMigrations defines a function interface that registers an entire
// sequence of migrations. The inputs are a directory where `.sql` files
// may be stored and the database engine (i.e. `postgres` or `mysql`). Functions
// satisfying this interface are intended to be used to lazily create migrations
// after flag parsing provides the SQL directory and engine as input.
type RegisterMigrations = func(sqlDirectory, engine string) (*golembic.Migrations, error)

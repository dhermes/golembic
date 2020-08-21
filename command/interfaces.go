package command

import (
	"github.com/dhermes/golembic"
)

// RegisterMigrations defines a function interface that registers an entire
// sequence of migrations. The only input is a directory where `.sql` files
// may be stored. Functions satisfying this interface are intended to be used
// to lazily create migrations after flag parsing provides the SQL directory
// as input.
type RegisterMigrations = func(sqlDirectory string) (*golembic.Migrations, error)

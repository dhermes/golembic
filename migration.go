package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migration represents an individual migration to be applied; typically as
// a set of SQL statements.
type Migration struct {
	// Parent is the revision identifier for the migration immediately
	// preceding this one. If absent, this indicates that this migration is
	// the "base" or "root" migration.
	Parent string
	// Revision is an opaque name that uniquely identifies a migration. It
	// is required for a migration to be valid.
	Revision string
	// Description is a long form description of why the migration is being
	// performed. It is intended to be used in "describe" scenarios where
	// a long form "history" of changes is presented.
	Description string
	// Up is the function to be executed when a migration is being applied. It
	// is required for a migration to be valid.
	Up UpMigration
	// UpConn is the non-transactional form of `Up`. This should be used in
	// rare situations where a migration cannot run inside a transaction, e.g.
	// a `CREATE UNIQUE INDEX CONCURRENTLY` statement.
	UpConn UpMigrationConn
	// CreatedAt is intended to be used for migrations retrieved via a SQL
	// query to the migrations metadata table.
	CreatedAt time.Time
}

// NewMigration creates a new migration from a variadic slice of options.
func NewMigration(opts ...MigrationOption) (*Migration, error) {
	m := &Migration{}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Like is "almost" an equality check, it compares the `Parent` and `Revision`.
func (m Migration) Like(other Migration) bool {
	return m.Parent == other.Parent && m.Revision == other.Revision
}

// Compact gives a "limited" representation of the migration
func (m Migration) Compact() string {
	if m.Parent == "" {
		return fmt.Sprintf("%s:NULL", m.Revision)
	}

	return fmt.Sprintf("%s:%s", m.Revision, m.Parent)
}

// InvokeUp dispatches to `Up` or `UpConn`, depending on which is set. If both
// or neither is set, that is considered an error. If `UpConn` needs to be invoked,
// this lazily invokes a helper to create a new connection. It's crucial that
// this helper sets the relevant timeouts on that connection to make sure
// migrations don't cause disruptions in application performance due to accidentally
// holding locks for an extended period.
func (m Migration) InvokeUp(ctx context.Context, nc NewConnection, tx *sql.Tx) error {
	// Handle the `UpConn` case first.
	if m.UpConn != nil {
		if m.Up != nil {
			return fmt.Errorf("%w; both Up and UpConn are set", ErrCannotInvokeUp)
		}

		conn, err := nc(ctx)
		if err != nil {
			return err
		}

		return m.UpConn(ctx, conn)
	}

	// If neither `UpConn` nor `Up` is set, we can't invoke anything.
	if m.Up == nil {
		return fmt.Errorf("%w; neither Up nor UpConn are set", ErrCannotInvokeUp)
	}

	return m.Up(ctx, tx)
}

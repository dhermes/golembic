package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	milestoneSuffix = " [MILESTONE]"
)

// Migration represents an individual migration to be applied; typically as
// a set of SQL statements.
type Migration struct {
	// Previous is the revision identifier for the migration immediately
	// preceding this one. If absent, this indicates that this migration is
	// the "base" or "root" migration.
	Previous string
	// Revision is an opaque name that uniquely identifies a migration. It
	// is required for a migration to be valid.
	Revision string
	// Description is a long form description of why the migration is being
	// performed. It is intended to be used in "describe" scenarios where
	// a long form "history" of changes is presented.
	Description string
	// Milestone is a flag indicating if the current migration is a milestone.
	// A milestone is a special migration that **must** be the last migration
	// in a sequence whenever applied. This is intended to be used in situations
	// where a change must be "staged" in two (or more parts) and one part
	// must run and "stabilize" before the next migration runs. For example, in
	// a rolling update deploy strategy some changes may not be compatible with
	// "old" and "new" versions of the code that may run simultaneously, so a
	// milestone marks the last point where old / new versions of application
	// code should be expected to be able to interact with the current schema.
	Milestone bool
	// Up is the function to be executed when a migration is being applied. Either
	// this field or `UpConn` are required (not both) and this field should be
	// the default choice in most cases. This function will be run in a transaction
	// that also writes a row to the migrations metadata table to signify that
	// this migration was applied.
	Up UpMigration
	// UpConn is the non-transactional form of `Up`. This should be used in
	// rare situations where a migration cannot run inside a transaction, e.g.
	// a `CREATE UNIQUE INDEX CONCURRENTLY` statement.
	UpConn UpMigrationConn
	// createdAt is stored in the migrations metadata table and represents the
	// moment when the migration was inserted into the table.  It is **not**
	// exported because it is internal to the implementation and should not be
	// specified by calling code.
	createdAt time.Time
	// serialID is an integer used for sorting migrations and will be stored
	// in the migrations table. It is intended to be used for migrations
	// retrieved via a SQL query to the migrations metadata table. It is
	// **not** exported because it is internal to the implementation and should
	// not be specified by calling code.
	serialID uint32
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

// ExtendedDescription is an extended form of `m.Description` that also
// incorporates other information like whether `m` is a milestone.
func (m Migration) ExtendedDescription() string {
	if m.Milestone {
		return m.Description + milestoneSuffix
	}

	return m.Description
}

// Like is "almost" an equality check, it compares the `Previous` and `Revision`.
func (m Migration) Like(other Migration) bool {
	return m.Previous == other.Previous && m.Revision == other.Revision
}

// Compact gives a "limited" representation of the migration
func (m Migration) Compact() string {
	if m.Previous == "" {
		return fmt.Sprintf("%s:NULL", m.Revision)
	}

	return fmt.Sprintf("%s:%s", m.Revision, m.Previous)
}

// InvokeUp dispatches to `Up` or `UpConn`, depending on which is set. If both
// or neither is set, that is considered an error. If `UpConn` needs to be invoked,
// this lazily creates a new connection from a pool. It's crucial that the pool
// sets the relevant timeouts when creating a new connection to make sure
// migrations don't cause disruptions in application performance due to
// accidentally holding locks for an extended period.
func (m Migration) InvokeUp(ctx context.Context, pool *sql.DB, tx *sql.Tx) error {
	// Handle the `UpConn` case first.
	if m.UpConn != nil {
		if m.Up != nil {
			return fmt.Errorf("%w; both Up and UpConn are set", ErrCannotInvokeUp)
		}

		conn, err := pool.Conn(ctx)
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

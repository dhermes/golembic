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

// MustNewMigration is the "must" form of `NewMigration()`. It panics if the
// migration could not be created.
func MustNewMigration(opts ...MigrationOption) Migration {
	m, err := NewMigration(opts...)
	if err != nil {
		panic(err)
	}

	return *m
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
// or neither is set, that is considered an error.
func (m Migration) InvokeUp(ctx context.Context, conn *sql.Conn, tx *sql.Tx) error {
	// Handle the `UpConn` case first.
	if m.UpConn != nil {
		if m.Up != nil {
			return fmt.Errorf("%w; both Up and UpConn are set", ErrCannotInvokeUp)
		}

		return m.UpConn(ctx, conn)
	}

	// If neither `UpConn` nor `Up` is set, we can't invoke anything.
	if m.Up == nil {
		return fmt.Errorf("%w; neither Up nor UpConn are set", ErrCannotInvokeUp)
	}

	return m.Up(ctx, tx)
}

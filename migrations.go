package golembic

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrNotRoot is the error returned when attempting to start a sequence of
	// migration with a non-root migration.
	ErrNotRoot = errors.New("Root migration cannot have a parent")
	// ErrMissingRevision is the error returned when attempting to register a migration
	// with no revision.
	ErrMissingRevision = errors.New("Cannot register a migration with no revision")
	// ErrNoParent is the error returned when attempting to register a migration
	// with no parent.
	ErrNoParent = errors.New("Cannot register a migration with no parent")
	// ErrAlreadyRegistered is the error return when a migration has already been
	// registered.
	ErrAlreadyRegistered = errors.New("Migration has already been registered")
)

// Migrations represents a sequence of migrations to be applied.
type Migrations struct {
	sequence map[string]Migration
	lock     sync.Mutex
}

// NewSequence creates a new sequence of migrations rooted in a single
// base / root migration.
func NewSequence(root Migration) (*Migrations, error) {
	if root.Parent != "" {
		err := fmt.Errorf(
			"%w; parent: %q, revision: %q",
			ErrNotRoot, root.Parent, root.Revision,
		)
		return nil, err
	}

	if root.Revision == "" {
		return nil, ErrMissingRevision
	}

	m := &Migrations{
		sequence: map[string]Migration{
			root.Revision: root,
		},
		lock: sync.Mutex{},
	}
	return m, nil
}

// Register adds a new migration to an existing sequence of migrations, if
// possible. The new migration must have a parent and have a valid revision
// that is not already registered.
func (m *Migrations) Register(migration Migration) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if migration.Parent == "" {
		return fmt.Errorf("%w; revision: %q", ErrNoParent, migration.Revision)
	}

	if migration.Revision == "" {
		return fmt.Errorf("%w; parent: %q", ErrMissingRevision, migration.Parent)
	}

	if _, ok := m.sequence[migration.Revision]; ok {
		return fmt.Errorf("%w; revision: %q", ErrAlreadyRegistered, migration.Revision)
	}

	m.sequence[migration.Revision] = migration
	return nil
}

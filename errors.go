package golembic

import (
	"errors"
)

var (
	// ErrDurationConversion is the error returned when a duration cannot be
	// converted to multiple of some base (e.g. milliseconds or seconds)
	// without round off.
	ErrDurationConversion = errors.New("Cannot convert duration")
	// ErrNotRoot is the error returned when attempting to start a sequence of
	// migration with a non-root migration.
	ErrNotRoot = errors.New("Root migration cannot have a previous migration set")
	// ErrMissingRevision is the error returned when attempting to register a migration
	// with no revision.
	ErrMissingRevision = errors.New("A migration must have a revision")
	// ErrNoPrevious is the error returned when attempting to register a migration
	// with no previous.
	ErrNoPrevious = errors.New("Cannot register a migration with no previous migration")
	// ErrPreviousNotRegistered is the error returned when attempting to register
	// a migration with a previous that is not yet registered.
	ErrPreviousNotRegistered = errors.New("Cannot register a migration until previous migration is registered")
	// ErrAlreadyRegistered is the error returned when a migration has already been
	// registered.
	ErrAlreadyRegistered = errors.New("Migration has already been registered")
	// ErrNilInterface is the error returned when a value satisfying an interface
	// is nil in a context where it is not allowed.
	ErrNilInterface = errors.New("Value satisfying interface was nil")
	// ErrMigrationNotRegistered is the error returned when no migration has been
	// registered for a given revision.
	ErrMigrationNotRegistered = errors.New("No migration registered for revision")
	// ErrMigrationMismatch is the error returned when the migration stored in
	// SQL does not match the registered migration.
	ErrMigrationMismatch = errors.New("Migration stored in SQL doesn't match sequence")
	// ErrCannotInvokeUp is the error returned when a migration cannot invoke the
	// up function (e.g. if it is `nil`).
	ErrCannotInvokeUp = errors.New("Cannot invoke up function for a migration")
	// ErrCannotPassMilestone is the error returned when a migration sequence
	// contains a milestone migration that is **NOT** the last step.
	ErrCannotPassMilestone = errors.New("If a migration sequence contains a milestone, it must be the last migration")
)

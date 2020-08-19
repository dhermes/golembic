package golembic

// Migration represents an individual up / down migration to be applied or
// rolled back (depending on the context).
type Migration struct {
	// Revision is an opaque name that uniquely identifies a migration. It
	// is required for a migration to be valid.
	Revision string
	// Parent is the revision identifier for the migration immediately
	// preceding this one. If absent, this indicates that this migration is
	// the "base" or "root" migration.
	Parent string
	// Description is a long form description of why the migration is being
	// performed. It is intended to be used in "describe" scenarios where
	// a long form "history" of changes is presented.
	Description string
	// Up is the function to be executed when a migration is being applied. It
	// is required for a migration to be valid.
	Up UpMigration
	// Down is the function to be executed when a migration is being rolled
	// back. It is required for a migration to be valid.
	Down DownMigration
}

// NewMigration creates a new migration from a variadic slice of options.
func NewMigration(opts ...Option) (*Migration, error) {
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
func MustNewMigration(opts ...Option) *Migration {
	m, err := NewMigration(opts...)
	if err != nil {
		panic(err)
	}

	return m
}

// OptRevision sets the revision on a migration.
func OptRevision(revision string) Option {
	return func(m *Migration) error {
		if revision == "" {
			return ErrMissingRevision
		}

		m.Revision = revision
		return nil
	}
}

// OptParent sets the parent on a migration.
func OptParent(parent string) Option {
	return func(m *Migration) error {
		m.Parent = parent
		return nil
	}
}

// OptDescription sets the description on a migration.
func OptDescription(description string) Option {
	return func(m *Migration) error {
		m.Description = description
		return nil
	}
}

// OptUp sets the `up` function on a migration.
func OptUp(up UpMigration) Option {
	return func(m *Migration) error {
		if up == nil {
			return ErrNilInterface
		}

		m.Up = up
		return nil
	}
}

// OptDown sets the `down` function on a migration.
func OptDown(down DownMigration) Option {
	return func(m *Migration) error {
		if down == nil {
			return ErrNilInterface
		}

		m.Down = down
		return nil
	}
}

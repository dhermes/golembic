package golembic

import (
	"fmt"
	"sync"
)

// NOTE: Ensure that
//       * `Migrations.Since` satisfies `migrationsFilter`.
//       * `Migrations.Until` satisfies `migrationsFilter`.
var (
	_ migrationsFilter = (*Migrations)(nil).Since
	_ migrationsFilter = (*Migrations)(nil).Until
)

// Migrations represents a sequence of migrations to be applied.
type Migrations struct {
	sequence map[string]Migration
	lock     sync.Mutex
}

// NewSequence creates a new sequence of migrations rooted in a single
// base / root migration.
func NewSequence(root Migration) (*Migrations, error) {
	if root.Previous != "" {
		err := fmt.Errorf(
			"%w; previous: %q, revision: %q",
			ErrNotRoot, root.Previous, root.Revision,
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
// possible. The new migration must have a previous migration and have a valid
// revision that is not already registered.
func (m *Migrations) Register(migration Migration) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if migration.Previous == "" {
		return fmt.Errorf("%w; revision: %q", ErrNoPrevious, migration.Revision)
	}

	if _, ok := m.sequence[migration.Previous]; !ok {
		return fmt.Errorf(
			"%w; revision: %q, previous: %q",
			ErrPreviousNotRegistered, migration.Revision, migration.Previous,
		)
	}

	if migration.Revision == "" {
		return fmt.Errorf("%w; previous: %q", ErrMissingRevision, migration.Previous)
	}

	if _, ok := m.sequence[migration.Revision]; ok {
		return fmt.Errorf("%w; revision: %q", ErrAlreadyRegistered, migration.Revision)
	}

	// NOTE: This crucially relies on `m.sequence` being locked.
	migration.serialID = uint32(len(m.sequence))
	m.sequence[migration.Revision] = migration
	return nil
}

// RegisterMany attempts to register multiple migrations (in order) with an
// existing sequence.
func (m *Migrations) RegisterMany(ms ...Migration) error {
	for _, migration := range ms {
		err := m.Register(migration)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterManyOpt attempts to register multiple migrations (in order) with an
// existing sequence. It differs from `RegisterMany()` in that the construction
// of `Migration` objects is handled directly here by taking a slice of
// option slices.
func (m *Migrations) RegisterManyOpt(manyOpts ...[]MigrationOption) error {
	for _, opts := range manyOpts {
		migration, err := NewMigration(opts...)
		if err != nil {
			return err
		}

		err = m.Register(*migration)
		if err != nil {
			return err
		}
	}

	return nil
}

// Root does a linear scan of every migration in the sequence and returns
// the root migration. In the "general" case such a scan would be expensive, but
// the number of migrations should always be a small number.
//
// NOTE: This does not verify or enforce the invariant that there must be
// exactly one migration without a previous migration. This invariant is enforced
// by the exported methods such as `Register()` and `RegisterMany()` and the
// constructor `NewSequence()`.
func (m *Migrations) Root() Migration {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, migration := range m.sequence {
		if migration.Previous == "" {
			return migration
		}
	}

	return Migration{}
}

// All produces the migrations in the sequence, in order.
//
// NOTE: This does not verify or enforce the invariant that there must be
//       exactly one migration without a previous migration. This invariant is
//       enforced by the exported methods such as `Register()` and
//       `RegisterMany()` and the constructor `NewSequence()`.
func (m *Migrations) All() []Migration {
	root := m.Root()

	m.lock.Lock()
	defer m.lock.Unlock()
	result := []Migration{root}
	// Find the unique revision (without validation) that points at the
	// current `previous`.
	previous := root.Revision
	for i := 0; i < len(m.sequence)-1; i++ {
		for _, migration := range m.sequence {
			if migration.Previous != previous {
				continue
			}

			result = append(result, migration)
			previous = migration.Revision
			break
		}
	}

	return result
}

// Since returns the migrations that occur **after** `revision`.
//
// This utilizes `All()` and returns all migrations after the one that
// matches `revision`. If none match, an error will be returned. If
// `revision` is the **last** migration, the migrations returned will be an
// empty slice.
func (m *Migrations) Since(revision string) (int, []Migration, error) {
	all := m.All()
	found := false

	result := []Migration{}
	pastMigrationCount := 0
	for _, migration := range all {
		if found {
			result = append(result, migration)
			continue
		}

		pastMigrationCount++
		if migration.Revision == revision {
			found = true
		}
	}

	if !found {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, revision)
		return 0, nil, err
	}

	return pastMigrationCount, result, nil
}

// Until returns the migrations that occur **before** `revision`.
//
// This utilizes `All()` and returns all migrations up to and including the one
// that matches `revision`. If none match, an error will be returned.
func (m *Migrations) Until(revision string) (int, []Migration, error) {
	all := m.All()
	found := false

	result := []Migration{}
	for _, migration := range all {
		result = append(result, migration)
		if migration.Revision == revision {
			found = true
			break
		}
	}

	if !found {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, revision)
		return 0, nil, err
	}

	// I.e. we are not filtering any migrations from the beginning of the
	// sequence.
	pastMigrationCount := 0
	return pastMigrationCount, result, nil
}

// Between returns the migrations that occur between two revisions.
//
// This can be seen as a combination of `Since()` and `Until()`.
func (m *Migrations) Between(since, until string) (int, []Migration, error) {
	all := m.All()
	foundSince := false
	foundUntil := false

	result := []Migration{}
	pastMigrationCount := 0
	for _, migration := range all {
		if foundSince {
			if foundUntil {
				break
			}
			result = append(result, migration)
		}

		pastMigrationCount++
		if migration.Revision == since {
			foundSince = true
		}

		if migration.Revision == until {
			foundUntil = true
		}
	}

	if !foundSince {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, since)
		return 0, nil, err
	}

	if !foundUntil {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, until)
		return 0, nil, err
	}

	return pastMigrationCount, result, nil
}

// Revisions produces the revisions in the sequence, in order.
//
// This utilizes `All()` and just extracts the revisions.
func (m *Migrations) Revisions() []string {
	result := []string{}
	for _, migration := range m.All() {
		result = append(result, migration.Revision)
	}
	return result
}

type describeMetadata struct {
	Revision    string
	Description string
}

// Describe displays all of the registered migrations (with descriptions).
func (m *Migrations) Describe(log PrintfReceiver) {
	revisions := m.Revisions()

	m.lock.Lock()
	defer m.lock.Unlock()
	dms := []describeMetadata{}
	revisionWidth := 0
	for _, revision := range revisions {
		migration := m.sequence[revision]
		dms = append(
			dms,
			describeMetadata{
				Revision:    revision,
				Description: migration.ExtendedDescription(),
			},
		)
		if len(revision) > revisionWidth {
			revisionWidth = len(revision)
		}
	}

	indexWidth := len(fmt.Sprintf("%d", len(dms)-1))
	format := ("%" + fmt.Sprintf("%d", indexWidth) + "d " +
		"| %" + fmt.Sprintf("%d", revisionWidth) + "s " +
		"| %s")
	for i, dm := range dms {
		log.Printf(format, i, dm.Revision, dm.Description)
	}
}

// Get retrieves a revision from the sequence, if present. If not, returns
// `nil`.
func (m *Migrations) Get(revision string) *Migration {
	m.lock.Lock()
	defer m.lock.Unlock()

	migration, ok := m.sequence[revision]
	if ok {
		return &migration
	}

	return nil
}

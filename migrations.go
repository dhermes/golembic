package golembic

import (
	"fmt"
	"strings"
	"sync"
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

// RegisterMany attempts to register multiple migrations (in order) with an
// existing sequence
func (m *Migrations) RegisterMany(ms ...Migration) error {
	for _, migration := range ms {
		err := m.Register(migration)
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
// exactly one migration without a parent. This invariant is enforced by the
// exported methods such as `Register()` and `RegisterMany()` and the constructor
// `NewSequence()`.
func (m *Migrations) Root() Migration {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, migration := range m.sequence {
		if migration.Parent == "" {
			return migration
		}
	}

	return Migration{}
}

// All produces the migrations in the sequence, in order.
//
// NOTE: This does not verify or enforce the invariant that there must be
// exactly one migration without a parent. This invariant is enforced by the
// exported methods such as `Register()` and `RegisterMany()` and the constructor
// `NewSequence()`.
func (m *Migrations) All() []Migration {
	root := m.Root()

	m.lock.Lock()
	defer m.lock.Unlock()
	result := []Migration{root}
	// Find the unique revision (without validation) that points at the
	// current `parent`.
	parent := root.Revision
	for i := 0; i < len(m.sequence)-1; i++ {
		for _, migration := range m.sequence {
			if migration.Parent != parent {
				continue
			}

			result = append(result, migration)
			parent = migration.Revision
			break
		}
	}

	return result
}

// Since returns the migrations that occur **after** `revision`. The
//
// This utilizes `All()` and returns all migrations after the one that
// matches `revision`. If none match, an error will be returned. If
// `revision` is the **last** migration, the migrations returned will be an
// empty slice.
func (m *Migrations) Since(revision string) ([]Migration, error) {
	all := m.All()
	found := false

	result := []Migration{}
	for _, migration := range all {
		if found {
			result = append(result, migration)
			continue
		}

		if migration.Revision == revision {
			found = true
		}
	}

	if !found {
		err := fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, revision)
		return nil, err
	}

	return result, nil
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
func (m *Migrations) Describe() string {
	revisions := m.Revisions()
	lines := []string{}

	m.lock.Lock()
	defer m.lock.Unlock()
	dms := []describeMetadata{}
	revisionWidth := 0
	for _, revision := range revisions {
		migration := m.sequence[revision]
		dms = append(dms, describeMetadata{Revision: revision, Description: migration.Description})
		if len(revision) > revisionWidth {
			revisionWidth = len(revision)
		}
	}

	indexWidth := len(fmt.Sprintf("%d", len(dms)-1))
	format := ("%" + fmt.Sprintf("%d", indexWidth) + "d " +
		"| %" + fmt.Sprintf("%d", revisionWidth) + "s " +
		"| %s")
	for i, dm := range dms {
		line := fmt.Sprintf(format, i, dm.Revision, dm.Description)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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

package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	// DefaultMetadataTable is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTable = "golembic_migrations"
)

// NewManager creates a new manager for orchestrating migrations.
func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		MetadataTable: DefaultMetadataTable,
		Log:           &stdoutPrintf{},
	}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Manager orchestrates database operations done via `Up` / `UpConn` as well as
// supporting operations such as creating a table for migration metadata and
// writing rows into that metadata table during a migration.
type Manager struct {
	// MetadataTable is the name of the table that stores migration metadata.
	// The expected default value (`DefaultMetadataTable`) is
	// "golembic_migrations".
	MetadataTable string
	// ConnectionPool is a cache-able pool of connections to the database.
	ConnectionPool *sql.DB
	// Provider delegates all actions to an abstract SQL database engine, with
	// the expectation that the provider also encodes connection information.
	Provider EngineProvider
	// Sequence is the collection of registered migrations to be applied,
	// verified, described, etc. by this manager.
	Sequence *Migrations
	// Log is used for printing output
	Log PrintfReceiver
}

// NewConnectionPool creates a new database connection pool and validates that
// it can ping the DB.
func (m *Manager) NewConnectionPool(ctx context.Context) (*sql.DB, error) {
	pool, err := m.Provider.Open()
	if err != nil {
		return nil, err
	}

	err = pool.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// EnsureConnectionPool returns a cached database connection pool (if already
// set) or invokes `NewConnection()` to create a new one.
func (m *Manager) EnsureConnectionPool(ctx context.Context) (*sql.DB, error) {
	if m.ConnectionPool != nil {
		return m.ConnectionPool, nil
	}

	pool, err := m.NewConnectionPool(ctx)
	if err != nil {
		return nil, err
	}

	m.ConnectionPool = pool
	return m.ConnectionPool, nil
}

// EnsureMigrationsTable checks that the migrations metadata table exists
// and creates it if not.
func (m *Manager) EnsureMigrationsTable(ctx context.Context) error {
	return CreateMigrationsTable(ctx, m)
}

// InsertMigration inserts a migration into the migrations metadata table.
func (m *Manager) InsertMigration(ctx context.Context, tx *sql.Tx, migration Migration) error {
	if migration.Parent == "" {
		statement := fmt.Sprintf(
			"INSERT INTO %s (parent, revision) VALUES (NULL, $1);",
			m.Provider.QuoteIdentifier(m.MetadataTable),
		)
		_, err := tx.ExecContext(ctx, statement, migration.Revision)
		return err
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (parent, revision) VALUES ($1, $2);",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	_, err := tx.ExecContext(ctx, statement, migration.Parent, migration.Revision)
	return err
}

// NewTx creates a new transaction after ensuring there is an existing
// connection.
func (m *Manager) NewTx(ctx context.Context) (*sql.Tx, error) {
	pool, err := m.EnsureConnectionPool(ctx)
	if err != nil {
		return nil, err
	}

	return pool.BeginTx(ctx, nil)
}

// ApplyMigration creates a transaction that runs the "Up" migration.
func (m *Manager) ApplyMigration(ctx context.Context, migration Migration) (err error) {
	var tx *sql.Tx
	defer func() {
		err = txFinalize(tx, err)
	}()

	m.Log.Printf("Applying %s: %s\n", migration.Revision, migration.Description)
	pool, err := m.EnsureConnectionPool(ctx)
	if err != nil {
		return
	}

	tx, err = m.NewTx(ctx)
	if err != nil {
		return
	}

	err = migration.InvokeUp(ctx, pool, tx)
	if err != nil {
		return
	}

	err = m.InsertMigration(ctx, tx, migration)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

// filterMigrations applies a filter function that takes the revision of the
// last applied migration to determine a set of migrations to run.
func (m *Manager) filterMigrations(ctx context.Context, filter migrationsFilter) ([]Migration, error) {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return nil, err
	}

	latest, _, err := m.Latest(ctx)
	if err != nil {
		return nil, err
	}

	migrations, err := filter(latest)
	if err != nil {
		return nil, err
	}

	if len(migrations) == 0 {
		m.Log.Printf("No migrations to run; latest revision: %s\n", latest)
		return nil, nil
	}

	return migrations, nil
}

// Up applies all migrations that have not yet been applied.
func (m *Manager) Up(ctx context.Context, opts ...ApplyOption) error {
	_, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	migrations, err := m.filterMigrations(ctx, m.sinceOrAll)
	if err != nil {
		return err
	}

	if migrations == nil {
		return nil
	}

	for _, migration := range migrations {
		err = m.ApplyMigration(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) sinceOrAll(revision string) ([]Migration, error) {
	if revision == "" {
		return m.Sequence.All(), nil
	}

	return m.Sequence.Since(revision)
}

// UpOne applies the **next** migration that has yet been applied, if any.
func (m *Manager) UpOne(ctx context.Context, opts ...ApplyOption) error {
	_, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	migrations, err := m.filterMigrations(ctx, m.sinceOrAll)
	if err != nil {
		return err
	}

	if migrations == nil {
		return nil
	}

	migration := migrations[0]
	return m.ApplyMigration(ctx, migration)
}

// UpTo applies all migrations that have yet to be applied up to (and
// including) `revision`, if any.
func (m *Manager) UpTo(ctx context.Context, revision string, opts ...ApplyOption) error {
	_, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	filter := func(latest string) ([]Migration, error) {
		return m.betweenOrUntil(latest, revision)
	}

	migrations, err := m.filterMigrations(ctx, filter)
	if err != nil {
		return err
	}

	if migrations == nil {
		return nil
	}

	for _, migration := range migrations {
		err = m.ApplyMigration(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) betweenOrUntil(latest string, revision string) ([]Migration, error) {
	if latest == "" {
		return m.Sequence.Until(revision)
	}

	return m.Sequence.Between(latest, revision)
}

// Latest determines the revision and timestamp of the most recently applied
// migration.
//
// NOTE: This assumes, but does not check, that the migrations metadata table
// exists.
func (m *Manager) Latest(ctx context.Context) (revision string, createdAt time.Time, err error) {
	var tx *sql.Tx
	defer func() {
		err = txFinalize(tx, err)
	}()

	tx, err = m.NewTx(ctx)
	if err != nil {
		return
	}

	query := fmt.Sprintf(
		"SELECT parent, revision, created_at FROM %s ORDER BY created_at DESC LIMIT 1;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query)
	if err != nil {
		return
	}

	if len(rows) == 0 {
		return
	}

	// NOTE: Here we trust that the query is sufficient to guarantee that
	//       `len(rows) == 1`.
	revision = rows[0].Revision
	createdAt = rows[0].CreatedAt
	return
}

// GetVersion returns the migration that corresponds to the version that was
// most recently applied.
func (m *Manager) GetVersion(ctx context.Context) (*Migration, error) {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return nil, err
	}

	revision, createdAt, err := m.Latest(ctx)
	if err != nil {
		return nil, err
	}

	if revision == "" {
		return nil, nil
	}

	migration := m.Sequence.Get(revision)
	if migration == nil {
		err = fmt.Errorf("%w; revision: %q", ErrMigrationNotRegistered, revision)
		return nil, err
	}

	withCreated := &Migration{
		Parent:      migration.Parent,
		Revision:    migration.Revision,
		Description: migration.Description,
		CreatedAt:   createdAt,
	}
	return withCreated, nil
}

// Verify checks that the rows in the migrations metadata table match the
// sequence.
func (m *Manager) Verify(ctx context.Context) (err error) {
	var tx *sql.Tx
	defer func() {
		err = txFinalize(tx, err)
	}()

	err = m.EnsureMigrationsTable(ctx)
	if err != nil {
		return
	}

	tx, err = m.NewTx(ctx)
	if err != nil {
		return
	}

	query := fmt.Sprintf(
		"SELECT parent, revision, created_at FROM %s ORDER BY created_at ASC;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query)
	if err != nil {
		return
	}

	all := m.Sequence.All()
	if len(rows) > len(all) {
		err = fmt.Errorf(
			"%w; sequence has %d migrations but %d are stored in the table",
			ErrMigrationMismatch, len(all), len(rows),
		)
		return
	}

	// Do a first pass for correctness.
	for i, row := range rows {
		expected := all[i]
		if !row.Like(expected) {
			err = fmt.Errorf(
				"%w; stored migration %d: %q does not match migration %q in sequence",
				ErrMigrationMismatch, i, row.Compact(), expected.Compact(),
			)
			return
		}
	}

	// Do a second pass for display purposes.
	for i, fromAll := range all {
		if i < len(rows) {
			row := rows[i]
			m.Log.Printf(
				"%d | %s | %s (applied %s)\n",
				i, fromAll.Revision, fromAll.Description, row.CreatedAt,
			)
		} else {
			m.Log.Printf(
				"%d | %s | %s (not yet applied)\n",
				i, fromAll.Revision, fromAll.Description,
			)
		}
	}

	return
}

// Describe displays all of the registered migrations (with descriptions).
func (m *Manager) Describe(_ context.Context) error {
	m.Sequence.Describe(m.Log)
	return nil
}

// Version displays the revision of the most recent migration to be applied
func (m *Manager) Version(ctx context.Context) error {
	migration, err := m.GetVersion(ctx)
	if err != nil {
		return err
	}

	if migration == nil {
		m.Log.Printf("No migrations have been run\n")
	} else {
		m.Log.Printf(
			"%s: %s (applied %s)\n",
			migration.Revision, migration.Description, migration.CreatedAt,
		)
	}
	return nil
}

// IsApplied checks if a migration has already been applied.
//
// NOTE: This assumes, but does not check, that the migrations metadata table
// exists.
func (m *Manager) IsApplied(ctx context.Context, tx *sql.Tx, migration Migration) (bool, error) {
	query := fmt.Sprintf(
		"SELECT parent, revision, created_at FROM %s WHERE revision = $1;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query, migration.Revision)
	if err != nil {
		return false, err
	}

	return verifyMigration(rows, migration)
}

func verifyMigration(rows []Migration, migration Migration) (bool, error) {
	if len(rows) == 0 {
		return false, nil
	}

	// NOTE: We don't verify that `len(rows) == 1` since we trust the UNIQUE
	//       index in the `revision` column.
	if rows[0].Parent != migration.Parent {
		err := fmt.Errorf(
			"%w; revision: %q, registered parent %q does not match parent %q from migrations table",
			ErrMigrationMismatch,
			migration.Revision,
			migration.Parent,
			rows[0].Parent,
		)
		return false, err
	}

	return true, nil
}

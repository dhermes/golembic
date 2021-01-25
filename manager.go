package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NOTE: Ensure that
//       * `Manager.sinceOrAll` satisfies `migrationsFilter`.
var (
	_ migrationsFilter = (*Manager)(nil).sinceOrAll
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
	// DevelopmentMode is a flag indicating that this manager is currently
	// being run in development mode, so things like extra validation should
	// intentionally be disabled. This is intended for use in testing and
	// development, where an entire database is spun up locally (e.g. in Docker)
	// and migrations will be applied from scratch (including milestones that
	// may not come at the end).
	DevelopmentMode bool
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

// CloseConnectionPool closes the connection pool and removes it, if one is
// set / cached on the current manager.
func (m *Manager) CloseConnectionPool() error {
	if m.ConnectionPool == nil {
		return nil
	}

	err := m.ConnectionPool.Close()
	if err != nil {
		return err
	}

	m.ConnectionPool = nil
	return nil
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
	if migration.Previous == "" {
		statement := fmt.Sprintf(
			"INSERT INTO %s (serial_id, revision, previous) VALUES (0, %s, NULL)",
			m.Provider.QuoteIdentifier(m.MetadataTable),
			m.Provider.QueryParameter(1),
		)
		_, err := tx.ExecContext(ctx, statement, migration.Revision)
		return err
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (serial_id, revision, previous) VALUES (%s, %s, %s)",
		m.Provider.QuoteIdentifier(m.MetadataTable),
		m.Provider.QueryParameter(1),
		m.Provider.QueryParameter(2),
		m.Provider.QueryParameter(3),
	)
	_, err := tx.ExecContext(
		ctx,
		statement,
		migration.serialID, // Parameter 1
		migration.Revision, // Parameter 2
		migration.Previous, // Parameter 3
	)
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

	m.Log.Printf("Applying %s: %s", migration.Revision, migration.ExtendedDescription())
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
func (m *Manager) filterMigrations(ctx context.Context, filter migrationsFilter, verifyHistory bool) (int, []Migration, error) {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return 0, nil, err
	}

	latest, _, err := m.latestMaybeVerify(ctx, verifyHistory)
	if err != nil {
		return 0, nil, err
	}

	pastMigrationCount, migrations, err := filter(latest)
	if err != nil {
		return 0, nil, err
	}

	if len(migrations) == 0 {
		format := "No migrations to run; latest revision: %s"

		// Add `milestoneSuffix`, if we can detect `latest` is a milestone.
		migration := m.Sequence.Get(latest)
		if migration != nil && migration.Milestone {
			format += milestoneSuffix
		}

		m.Log.Printf(format, latest)
		return pastMigrationCount, nil, nil
	}

	return pastMigrationCount, migrations, nil
}

func (m *Manager) validateMilestones(pastMigrationCount int, migrations []Migration) error {
	// Early exit if no migrations have been run yet. This **assumes** that the
	// database is being brought up from scratch.
	if pastMigrationCount == 0 {
		return nil
	}

	count := len(migrations)
	// Ensure all (but the last) are not a milestone.
	for i := 0; i < count-1; i++ {
		migration := migrations[i]
		if !migration.Milestone {
			continue
		}

		err := fmt.Errorf(
			"%w; revision %s (%d / %d migrations)",
			ErrCannotPassMilestone, migration.Revision, i+1, count,
		)

		// In development mode, log the error message but don't return an error.
		if m.DevelopmentMode {
			m.Log.Printf("Ignoring error in development mode")
			m.Log.Printf("  %s", err)
			continue
		}

		return err
	}

	return nil
}

// Up applies all migrations that have not yet been applied.
func (m *Manager) Up(ctx context.Context, opts ...ApplyOption) error {
	ac, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	pastMigrationCount, migrations, err := m.filterMigrations(ctx, m.sinceOrAll, ac.VerifyHistory)
	if err != nil {
		return err
	}

	if migrations == nil {
		return nil
	}

	err = m.validateMilestones(pastMigrationCount, migrations)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		err = m.ApplyMigration(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) sinceOrAll(revision string) (int, []Migration, error) {
	if revision == "" {
		return 0, m.Sequence.All(), nil
	}

	return m.Sequence.Since(revision)
}

// UpOne applies the **next** migration that has yet been applied, if any.
func (m *Manager) UpOne(ctx context.Context, opts ...ApplyOption) error {
	ac, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	_, migrations, err := m.filterMigrations(ctx, m.sinceOrAll, ac.VerifyHistory)
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
// including) a revision, if any. This expects the `ApplyConfig` revision to
// be set in `opts`.
func (m *Manager) UpTo(ctx context.Context, opts ...ApplyOption) error {
	ac, err := NewApplyConfig(opts...)
	if err != nil {
		return err
	}

	var filter migrationsFilter = func(latest string) (int, []Migration, error) {
		return m.betweenOrUntil(latest, ac.Revision)
	}

	pastMigrationCount, migrations, err := m.filterMigrations(ctx, filter, ac.VerifyHistory)
	if err != nil {
		return err
	}

	if migrations == nil {
		return nil
	}

	err = m.validateMilestones(pastMigrationCount, migrations)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		err = m.ApplyMigration(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) betweenOrUntil(latest string, revision string) (int, []Migration, error) {
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
		"SELECT revision, previous, created_at FROM %s ORDER BY serial_id DESC LIMIT 1",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	tc := m.Provider.TimestampColumn()
	rows, err := readAllMigration(ctx, tx, query, tc)
	if err != nil {
		return
	}

	if len(rows) == 0 {
		return
	}

	// NOTE: Here we trust that the query is sufficient to guarantee that
	//       `len(rows) == 1`.
	revision = rows[0].Revision
	createdAt = rows[0].createdAt
	return
}

// latestMaybeVerify determines the latest applied migration and verifies all of the
// migration history if `verifyHistory` is true.
func (m *Manager) latestMaybeVerify(ctx context.Context, verifyHistory bool) (revision string, createdAt time.Time, err error) {
	if !verifyHistory {
		revision, createdAt, err = m.Latest(ctx)
		return
	}

	var tx *sql.Tx
	defer func() {
		err = txFinalize(tx, err)
	}()

	tx, err = m.NewTx(ctx)
	if err != nil {
		return
	}

	history, _, err := m.verifyHistory(ctx, tx)
	if err != nil {
		return
	}

	if len(history) == 0 {
		return
	}

	revision = history[len(history)-1].Revision
	createdAt = history[len(history)-1].createdAt
	err = tx.Commit()
	return
}

// GetVersion returns the migration that corresponds to the version that was
// most recently applied.
func (m *Manager) GetVersion(ctx context.Context, opts ...ApplyOption) (*Migration, error) {
	ac, err := NewApplyConfig(opts...)
	if err != nil {
		return nil, err
	}

	err = m.EnsureMigrationsTable(ctx)
	if err != nil {
		return nil, err
	}

	revision, createdAt, err := m.latestMaybeVerify(ctx, ac.VerifyHistory)
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
		Previous:    migration.Previous,
		Revision:    migration.Revision,
		Description: migration.Description,
		createdAt:   createdAt,
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

	history, registered, err := m.verifyHistory(ctx, tx)
	if err != nil {
		return
	}

	for i, migration := range registered {
		description := migration.ExtendedDescription()
		if i < len(history) {
			applied := history[i]
			m.Log.Printf(
				"%d | %s | %s (applied %s)",
				i, migration.Revision, description, applied.createdAt,
			)
		} else {
			m.Log.Printf(
				"%d | %s | %s (not yet applied)",
				i, migration.Revision, description,
			)
		}
	}

	return
}

// verifyHistory retrieves a full history of migrations and compares it against
// the sequence of registered migrations. If they match (up to the end of the
// history, the registered sequence can be longer), this will return with no
// error and include slices of the history and the registered migrations.
func (m *Manager) verifyHistory(ctx context.Context, tx *sql.Tx) (history, registered []Migration, err error) {
	query := fmt.Sprintf(
		"SELECT revision, previous, created_at FROM %s ORDER BY serial_id ASC",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	tc := m.Provider.TimestampColumn()
	history, err = readAllMigration(ctx, tx, query, tc)
	if err != nil {
		return
	}

	registered = m.Sequence.All()
	if len(history) > len(registered) {
		err = fmt.Errorf(
			"%w; sequence has %d migrations but %d are stored in the table",
			ErrMigrationMismatch, len(registered), len(history),
		)
		return
	}

	for i, row := range history {
		expected := registered[i]
		if !row.Like(expected) {
			err = fmt.Errorf(
				"%w; stored migration %d: %q does not match migration %q in sequence",
				ErrMigrationMismatch, i, row.Compact(), expected.Compact(),
			)
			return
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
func (m *Manager) Version(ctx context.Context, opts ...ApplyOption) error {
	migration, err := m.GetVersion(ctx, opts...)
	if err != nil {
		return err
	}

	if migration == nil {
		m.Log.Printf("No migrations have been run")
	} else {
		m.Log.Printf(
			"%s: %s (applied %s)",
			migration.Revision, migration.Description, migration.createdAt,
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
		"SELECT revision, previous, created_at FROM %s WHERE revision = %s",
		m.Provider.QuoteIdentifier(m.MetadataTable),
		m.Provider.QueryParameter(1),
	)
	tc := m.Provider.TimestampColumn()
	rows, err := readAllMigration(ctx, tx, query, tc, migration.Revision)
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
	if rows[0].Previous != migration.Previous {
		err := fmt.Errorf(
			"%w; revision: %q, registered previous migration %q does not match %q from migrations table",
			ErrMigrationMismatch,
			migration.Revision,
			migration.Previous,
			rows[0].Previous,
		)
		return false, err
	}

	return true, nil
}

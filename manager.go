package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	// DefaultMetadataTable is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTable = "golembic_migrations"
)

// NewManager creates a new manager for orchestrating migrations. The variadic
// input `table` can be used
func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{MetadataTable: DefaultMetadataTable}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Manager orchestrates database operations done via `UpMigration` as well as
// supporting operations such as creating a table for migration metadata and
// writing rows into that metadata table during an `UpMigration.`
type Manager struct {
	MetadataTable string
	Connection    *sql.Conn
	Provider      EngineProvider
	Sequence      *Migrations
}

// EnsureConnection returns a cached database connection (if already set) or
// creates a new one, validates the connection can ping the DB and sets timeouts
// via `Provider.SetConnTimeouts()`.
//
// We use a `sql.Conn` vs. a `sql.DB` because we can guarantee that connection
// timeouts are set on a connection whereas a `sql.DB` may use a different
// connection from the pool. Though `sql.Conn` is **not** concurrency safe,
// this isn't a problem for us because migrations should run in series.
func (m *Manager) EnsureConnection(ctx context.Context) (*sql.Conn, error) {
	if m.Connection != nil {
		return m.Connection, nil
	}

	db, err := m.Provider.Open()
	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	err = m.Provider.SetConnTimeouts(ctx, conn)
	if err != nil {
		return nil, err
	}

	m.Connection = conn
	return m.Connection, nil
}

// EnsureMigrationsTable checks that the migrations metadata table exists
// and creates it if not.
func (m *Manager) EnsureMigrationsTable(ctx context.Context) error {
	conn, err := m.EnsureConnection(ctx)
	if err != nil {
		return err
	}

	return CreateMigrationsTable(ctx, conn, m.Provider, m.MetadataTable)
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

// ApplyMigration creates a transaction that runs the "Up" migration.
func (m *Manager) ApplyMigration(ctx context.Context, migration Migration) error {
	// TODO: https://github.com/dhermes/golembic/issues/1
	log.Printf("Applying %s: %s\n", migration.Revision, migration.Description)

	conn, err := m.EnsureConnection(ctx)
	if err != nil {
		return err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackAndLog(tx)

	// Make sure to "guard" against long locks by setting timeouts within the
	// transaction before doing any work.
	err = m.Provider.SetTxTimeouts(ctx, tx)
	if err != nil {
		return err
	}

	err = migration.Up(ctx, tx)
	if err != nil {
		return err
	}

	err = m.InsertMigration(ctx, tx, migration)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Up applies all migrations that have not yet been applied.
func (m *Manager) Up(ctx context.Context) error {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return err
	}

	latest, _, err := m.Latest(ctx)
	if err != nil {
		return err
	}

	migrations, err := m.sinceOrAll(latest)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		// TODO: https://github.com/dhermes/golembic/issues/1
		log.Printf("No migrations to run; latest revision: %s\n", latest)
		return nil
	}

	// TODO: Re-factor the above into a helper that is common to `Up`, `UpOne`
	//       and `UpTo`.
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
func (m *Manager) UpOne(ctx context.Context) error {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return err
	}

	latest, _, err := m.Latest(ctx)
	if err != nil {
		return err
	}

	migrations, err := m.sinceOrAll(latest)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		// TODO: https://github.com/dhermes/golembic/issues/1
		log.Printf("No migrations to run; latest revision: %s\n", latest)
		return nil
	}

	// TODO: Re-factor the above into a helper that is common to `Up`, `UpOne`
	//       and `UpTo`.
	migration := migrations[0]
	return m.ApplyMigration(ctx, migration)
}

// UpTo applies all migrations that have yet to be applied up to (and
// including) `revision`, if any.
func (m *Manager) UpTo(ctx context.Context, revision string) error {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return err
	}

	latest, _, err := m.Latest(ctx)
	if err != nil {
		return err
	}

	migrations, err := m.betweenOrUntil(latest, revision)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		// TODO: https://github.com/dhermes/golembic/issues/1
		log.Printf("No migrations to run; latest revision: %s\n", latest)
		return nil
	}

	// TODO: Re-factor the above into a helper that is common to `Up`, `UpOne`
	//       and `UpTo`.
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
func (m *Manager) Latest(ctx context.Context) (string, time.Time, error) {
	conn, err := m.EnsureConnection(ctx)
	if err != nil {
		return "", time.Time{}, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	defer rollbackAndLog(tx)

	// Make sure to "guard" against long locks by setting timeouts within the
	// transaction before doing any work.
	err = m.Provider.SetTxTimeouts(ctx, tx)
	if err != nil {
		return "", time.Time{}, err
	}

	query := fmt.Sprintf(
		"SELECT parent, revision, created_at FROM %s ORDER BY created_at DESC LIMIT 1;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query)
	if err != nil {
		return "", time.Time{}, err
	}

	if len(rows) == 0 {
		return "", time.Time{}, nil
	}

	// NOTE: Here we trust that the query is sufficient to guarantee that
	//       `len(rows) == 1`.
	return rows[0].Revision, rows[0].CreatedAt, nil
}

// Version returns the migration that corresponds to the version that was
// most recently applied.
func (m *Manager) Version(ctx context.Context) (*Migration, error) {
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
func (m *Manager) Verify(ctx context.Context) error {
	err := m.EnsureMigrationsTable(ctx)
	if err != nil {
		return err
	}

	conn, err := m.EnsureConnection(ctx)
	if err != nil {
		return err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackAndLog(tx)

	// Make sure to "guard" against long locks by setting timeouts within the
	// transaction before doing any work.
	err = m.Provider.SetTxTimeouts(ctx, tx)
	if err != nil {
		return err
	}

	// TODO: All of the code above gets copy-pasted quite a bit; try to
	//       re-factor to make it easier for re-use (e.g. `StartTx()`). It
	//       is a wee bit delicate because of the need to `defer` in this block
	//       vs. an invoked method or function.
	query := fmt.Sprintf(
		"SELECT parent, revision, created_at FROM %s ORDER BY created_at ASC;",
		m.Provider.QuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, tx, query)
	if err != nil {
		return err
	}

	all := m.Sequence.All()
	if len(rows) > len(all) {
		err := fmt.Errorf(
			"%w; sequence has %d migrations but %d are stored in the table",
			ErrMigrationMismatch, len(all), len(rows),
		)
		return err
	}

	// Do a first pass for correctness.
	for i, row := range rows {
		expected := all[i]
		if !row.Like(expected) {
			err := fmt.Errorf(
				"%w; stored migration %d: %q does not match migration %q in sequence",
				ErrMigrationMismatch, i, row.Compact(), expected.Compact(),
			)
			return err
		}
	}

	// Do a second pass for display purposes.
	for i, fromAll := range all {
		if i < len(rows) {
			row := rows[i]
			log.Printf(
				":: %d | %s | %s (applied %s)\n",
				i, fromAll.Revision, fromAll.Description, row.CreatedAt,
			)
		} else {
			log.Printf(
				":: %d | %s | %s (not yet applied)\n",
				i, fromAll.Revision, fromAll.Description,
			)
		}
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

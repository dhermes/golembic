package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// readAllInt performs a SQL query and reads all rows into an `int` slice,
// under the assumption that a single (INTEGER) column is being returned for
// the query.
func readAllInt(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (result []int, err error) {
	var rows *sql.Rows
	defer func() {
		err = rowsClose(rows, err)
	}()

	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		return
	}

	for rows.Next() {
		var value int
		err = rows.Scan(&value)
		if err != nil {
			return
		}
		result = append(result, value)
	}

	return
}

// migrationFromQuery is intended to be used to construct a metadata row
// from a pair of values read off of a `sql.Rows`.
func migrationFromQuery(previous sql.NullString, revision string, createdAt time.Time) Migration {
	if previous.Valid {
		return Migration{Previous: previous.String, Revision: revision, CreatedAt: createdAt}
	}

	// Handle NULL.
	return Migration{Revision: revision, CreatedAt: createdAt}
}

// readAllMigration performs a SQL query and reads all rows into a
// `Migration` slice, under the assumption that three columns -- revision,
// previous and created_at -- are being returned for the query (in that order).
// For example, the query
//
//   SELECT revision, previous, created_at FROM golembic_migrations
//
// would satisfy this. A more "focused" query would return the latest migration
// applied
//
//   SELECT
//     revision,
//     previous,
//     created_at
//   FROM
//     golembic_migrations
//   ORDER BY
//     serial_id DESC
//   LIMIT 1
func readAllMigration(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (result []Migration, err error) {
	var rows *sql.Rows
	defer func() {
		err = rowsClose(rows, err)
	}()

	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		return
	}

	for rows.Next() {
		var revision string
		var previous sql.NullString
		var createdAt time.Time
		err = rows.Scan(&revision, &previous, &createdAt)
		if err != nil {
			return
		}
		result = append(result, migrationFromQuery(previous, revision, createdAt))
	}

	return
}

// rowsClose is intended to be used in `defer` blocks to ensure that a SQL
// query `Rows` iterator is always closed after being consumed (or abandonded
// during iteration).
func rowsClose(rows *sql.Rows, err error) error {
	if rows == nil {
		return err
	}

	closeErr := rows.Close()
	return maybeWrap(err, closeErr, "failed to close rows")
}

// txFinalize is intended to be used in `defer` blocks to ensure that a SQL
// transaction is always rolled back after being started. In cases when the
// transaction was successfully committed, the rollback will fail with
// `sql.ErrTxDone` which will be ignored here.
func txFinalize(tx *sql.Tx, err error) error {
	if tx == nil {
		return err
	}

	rollbackErr := ignoreTxDone(tx.Rollback())
	return maybeWrap(err, rollbackErr, "failed to rollback transaction")
}

// ignoreTxDone converts a `sql.ErrTxDone` error to `nil` (and leaves alone
// all other errors).
func ignoreTxDone(err error) error {
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}

// maybeWrap attempts to wrap a secondary error inside a primary one. If
// one (or both) of the errors if `nil`, then no wrapping is necessary.
func maybeWrap(primary, secondary error, message string) error {
	if primary == nil {
		return secondary
	}
	if secondary == nil {
		return primary
	}

	return fmt.Errorf("%w; %s: %v", primary, message, secondary)
}

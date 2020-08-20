package golembic

import (
	"context"
	"database/sql"
	"fmt"
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
func migrationFromQuery(parent sql.NullString, revision string) Migration {
	if parent.Valid {
		return Migration{Parent: parent.String, Revision: revision}
	}

	// Handle NULL.
	return Migration{Revision: revision}
}

// readAllMigration performs a SQL query and reads all rows into a
// `Migration` slice, under the assumption that two columns -- parent and
// revision -- are being returned for the query (in that order). For example,
// the query
//
//   SELECT parent, revision FROM golembic_migrations;
//
// would satisfy this. A more "focused" query would return the latest migration
// applied
//
//   SELECT
//       parent,
//       revision
//   FROM
//       golembic_migrations
//   ORDER BY
//       created_at DESC
//   LIMIT 1;
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
		var parent sql.NullString
		var revision string
		err = rows.Scan(&parent, &revision)
		if err != nil {
			return
		}
		result = append(result, migrationFromQuery(parent, revision))
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
	if closeErr == nil {
		return err
	}

	return fmt.Errorf("%w; failed to close rows: %v", err, closeErr)
}

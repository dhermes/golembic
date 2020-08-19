package golembic

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS %s (
  parent VARCHAR(32),
  revision VARCHAR(32) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`
	addPrimaryKeyMigrationsTableSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s PRIMARY KEY (revision);
`
)

// CreateMigrationsTable invokes SQL statements required to create the metadata
// table used to track migrations. If the table already exists (as detected by
// `provider.TableExistsSQL()`), this function will not attempt to create a
// table or any constraints.
func CreateMigrationsTable(ctx context.Context, db *sql.DB, provider EngineProvider, table string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollbackAndLog(tx)

	// Early exit if the table exists.
	exists, err := tableExists(ctx, tx, provider, table)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = tx.ExecContext(ctx, createMigrationsSQL(provider, table))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPrimaryKeyMigrationsSQL(provider, table))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func createMigrationsSQL(provider EngineProvider, table string) string {
	return fmt.Sprintf(createMigrationsTableSQL, provider.QuoteIdentifier(table))
}

func addPrimaryKeyMigrationsSQL(provider EngineProvider, table string) string {
	constraint := fmt.Sprintf("pk_%s_revision", table)
	return fmt.Sprintf(
		addPrimaryKeyMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		provider.QuoteIdentifier(constraint),
	)
}

func tableExists(ctx context.Context, tx *sql.Tx, provider EngineProvider, table string) (bool, error) {
	query := provider.TableExistsSQL(table)
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return false, err
	}

	anyRows := rows.Next()
	err = rows.Err()
	if err != nil {
		return false, err
	}

	return anyRows, rows.Close()
}

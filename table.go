package golembic

import (
	"context"
	"database/sql"
)

const (
	// DefaultMetadataTableName is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTableName = "golembic_migrations"

	// TODO: https://github.com/dhermes/golembic/issues/2
	createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS golembic_migrations (
  id SERIAL,
  revision VARCHAR(32) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`
	addPrimaryKeyMigrationsTableSQL = `
ALTER TABLE golembic_migrations
  ADD CONSTRAINT pk_golembic_migrations_id PRIMARY KEY (id);
`
)

// CreateMigrationsTable invokes SQL statements required to create the metadata
// table used to track migrations. If the table already exists (as detected by
// `provider.TableExistsSQL()`), this function will not attempt to create a
// table or any constraints.
func CreateMigrationsTable(ctx context.Context, db *sql.DB, provider EngineProvider) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollbackAndLog(tx)

	// Early exit if the table exists.
	exists, err := tableExists(ctx, tx, provider, DefaultMetadataTableName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = tx.ExecContext(ctx, createMigrationsTableSQL)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPrimaryKeyMigrationsTableSQL)
	if err != nil {
		return err
	}

	return tx.Commit()
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

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
CREATE TABLE golembic_migrations (
    id SERIAL,
    revision VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`
	addPrimaryKeyMigrationsTableSQL = `
ALTER TABLE golembic_migrations ADD CONSTRAINT pk_golembic_migrations_id PRIMARY KEY (id);
`
)

// CreateMigrationsTable invokes SQL statements required to create the metadata
// table used to track migrations.
func CreateMigrationsTable(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollbackAndLog(tx)

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

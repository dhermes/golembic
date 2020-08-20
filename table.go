package golembic

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS %[1]s (
  revision VARCHAR(32) NOT NULL,
  parent VARCHAR(32),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s PRIMARY KEY (revision);
ALTER TABLE %[1]s
  ADD CONSTRAINT %[3]s FOREIGN KEY (parent)
  REFERENCES %[1]s(revision);
ALTER TABLE %[1]s
  ADD CONSTRAINT %[4]s UNIQUE (parent);
ALTER TABLE %[1]s
  ADD CONSTRAINT %[5]s CHECK (parent != revision);
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

	// Make sure to "guard" against long locks by setting timeouts within the
	// transaction before doing any work.
	err = provider.SetTxTimeouts(ctx, tx)
	if err != nil {
		return err
	}

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

	return tx.Commit()
}

func createMigrationsSQL(provider EngineProvider, table string) string {
	pkConstraint := fmt.Sprintf("pk_%s_revision", table)
	fkConstraint := fmt.Sprintf("fk_%s_parent", table)
	uqConstraint := fmt.Sprintf("uq_%s_parent", table)
	chkConstraint := fmt.Sprintf("chk_%s_parent_neq_revision", table)
	return fmt.Sprintf(
		createMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		provider.QuoteIdentifier(pkConstraint),
		provider.QuoteIdentifier(fkConstraint),
		provider.QuoteIdentifier(uqConstraint),
		provider.QuoteIdentifier(chkConstraint),
	)
}

func tableExists(ctx context.Context, tx *sql.Tx, provider EngineProvider, table string) (bool, error) {
	query := provider.TableExistsSQL()
	rows, err := readAllInt(ctx, tx, query, table)
	if err != nil {
		return false, err
	}

	// NOTE: Here we trust that the query is sufficient to determine existence
	//       by just having some results. We could go further and check that
	//       `rows == []int{1}` but we elect not to do so.
	return len(rows) > 0, nil
}

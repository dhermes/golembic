package golembic

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	createMigrationsTableSQL = `
CREATE TABLE %s (
  serial_id  INTEGER NOT NULL,
  revision   VARCHAR(32) NOT NULL,
  previous   VARCHAR(32),
  created_at %s
)
`
	pkMigrationsTableSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s PRIMARY KEY (revision)
`
	fkPreviousMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s FOREIGN KEY (previous)
  REFERENCES %[1]s(revision)
`
	uqPreviousMigrationsTableSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s UNIQUE (previous)
`
	uqSerialIDSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s UNIQUE (serial_id)
`
	nonNegativeSerialIDSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s CHECK (serial_id >= 0)
`
	noCyclesMigrationsTableSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s CHECK (previous != revision)
`
	singleRootMigrationsTableSQL = `
ALTER TABLE %s
  ADD CONSTRAINT %s CHECK
  (
    (serial_id = 0 AND previous IS NULL) OR
    (serial_id != 0 AND previous IS NOT NULL)
  )
`
)

// CreateMigrationsTable invokes SQL statements required to create the metadata
// table used to track migrations. If the table already exists (as detected by
// `provider.TableExistsSQL()`), this function will not attempt to create a
// table or any constraints.
func CreateMigrationsTable(ctx context.Context, manager *Manager) (err error) {
	var tx *sql.Tx
	defer func() {
		err = txFinalize(tx, err)
	}()

	tx, err = manager.NewTx(ctx)
	if err != nil {
		return
	}

	// Early exit if the table exists.
	exists, err := tableExists(ctx, tx, manager)
	if err != nil {
		return
	}
	if exists {
		return
	}

	_, err = tx.ExecContext(ctx, createMigrationsSQL(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, pkMigrationsSQL(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, fkPreviousMigrationsSQL(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, uqSerialID(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, nonNegativeSerialID(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, uqPreviousMigrationsSQL(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, noCyclesMigrationsSQL(manager))
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, singleRootMigrationsSQL(manager))
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func createMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	provider := manager.Provider
	timestampColumn := provider.TimestampColumn()
	return fmt.Sprintf(createMigrationsTableSQL, provider.QuoteIdentifier(table), timestampColumn)
}

func pkMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	pkConstraint := fmt.Sprintf("pk_%s_revision", table)

	provider := manager.Provider
	return fmt.Sprintf(
		pkMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		pkConstraint,
	)
}

func fkPreviousMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	fkConstraint := fmt.Sprintf("fk_%s_previous", table)

	provider := manager.Provider
	return fmt.Sprintf(
		fkPreviousMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		fkConstraint,
	)
}

func uqSerialID(manager *Manager) string {
	table := manager.MetadataTable
	uqConstraint := fmt.Sprintf("uq_%s_serial_id", table)

	provider := manager.Provider
	return fmt.Sprintf(
		uqSerialIDSQL,
		provider.QuoteIdentifier(table),
		uqConstraint,
	)
}

func nonNegativeSerialID(manager *Manager) string {
	table := manager.MetadataTable
	chkConstraint := fmt.Sprintf("chk_%s_serial_id", table)

	provider := manager.Provider
	return fmt.Sprintf(
		nonNegativeSerialIDSQL,
		provider.QuoteIdentifier(table),
		chkConstraint,
	)
}

func uqPreviousMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	uqConstraint := fmt.Sprintf("uq_%s_previous", table)

	provider := manager.Provider
	return fmt.Sprintf(
		uqPreviousMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		uqConstraint,
	)
}

func noCyclesMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	chkConstraint := fmt.Sprintf("chk_%s_previous_neq_revision", table)

	provider := manager.Provider
	return fmt.Sprintf(
		noCyclesMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		chkConstraint,
	)
}

func singleRootMigrationsSQL(manager *Manager) string {
	table := manager.MetadataTable
	nullPreviousIndex := fmt.Sprintf("chk_%s_null_previous", table)

	provider := manager.Provider
	return fmt.Sprintf(
		singleRootMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		provider.QuoteIdentifier(nullPreviousIndex),
	)
}

func tableExists(ctx context.Context, tx *sql.Tx, manager *Manager) (bool, error) {
	query := manager.Provider.TableExistsSQL()
	rows, err := readAllInt(ctx, tx, query, manager.MetadataTable)
	if err != nil {
		return false, err
	}

	// NOTE: If `len(rows) > 1` it could be due to an insufficiently specific
	//       query. For example, if no database schema is specified in the
	//       provider configuration, but the table name exists in multiple
	//       schemas / schemata.
	return len(rows) == 1, nil
}

package golembic

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	createMigrationsTableSQL = `
CREATE TABLE %s (
  serial_id  %s,
  revision   %s,
  previous   %s,
  created_at %s%s
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
	createTableInlineConstraintsSQL = `,

FOREIGN KEY(previous) REFERENCES revision
CHECK (
  (serial_id = 0 AND previous IS NULL) OR
  (serial_id != 0 AND previous IS NOT NULL)
)
`
)

// CreateTableParameters specifies a set of parameters that are intended
// to be used in a `CREATE TABLE` statement. This allows providers to
// specify custom column types or add custom checks / constraints that are
// engine specific.
type CreateTableParameters struct {
	SerialID                 string
	Revision                 string
	Previous                 string
	CreatedAt                string
	Constraints              string
	SkipConstraintStatements bool
}

// NewCreateTableParameters populates a `CreateTableParameters` with a
// basic set of defaults and allows optional overrides for all fields.
func NewCreateTableParameters(opts ...CreateTableOption) CreateTableParameters {
	ctp := CreateTableParameters{}
	for _, opt := range opts {
		opt(&ctp)
	}

	ctp.ensureSerialID()
	ctp.ensureRevision()
	ctp.ensurePrevious()
	ctp.ensureConstraints()

	return ctp
}

// ensureSerialID makes sure that `SerialID` is set on the current
// `CreateTableParameters` receiver.
func (ctp *CreateTableParameters) ensureSerialID() {
	// Early exit if already set.
	if ctp.SerialID != "" {
		return
	}

	// If we are allowed to use `ALTER TABLE ... ADD CONSTRAINT ...` statements
	// then the default column type can be simple.
	if !ctp.SkipConstraintStatements {
		ctp.SerialID = "INTEGER NOT NULL"
		return
	}

	ctp.SerialID = "INTEGER NOT NULL UNIQUE CHECK (serial_id >= 0)"
	return
}

// ensureSerialID makes sure that `Revision` is set on the current
// `CreateTableParameters` receiver.
func (ctp *CreateTableParameters) ensureRevision() {
	// Early exit if already set.
	if ctp.Revision != "" {
		return
	}

	// If we are allowed to use `ALTER TABLE ... ADD CONSTRAINT ...` statements
	// then the default column type can be simple.
	if !ctp.SkipConstraintStatements {
		ctp.Revision = "VARCHAR(32) NOT NULL"
		return
	}

	ctp.Revision = "VARCHAR(32) NOT NULL PRIMARY KEY"
	return
}

// ensureSerialID makes sure that `Previous` is set on the current
// `CreateTableParameters` receiver.
func (ctp *CreateTableParameters) ensurePrevious() {
	// Early exit if already set.
	if ctp.Previous != "" {
		return
	}

	// If we are allowed to use `ALTER TABLE ... ADD CONSTRAINT ...` statements
	// then the default column type can be simple.
	if !ctp.SkipConstraintStatements {
		ctp.Previous = "VARCHAR(32)"
		return
	}

	ctp.Previous = "VARCHAR(32) UNIQUE CHECK (previous != revision)"
	return
}

// ensureConstraints makes sure that `Constraints` is set on the current
// `CreateTableParameters` receiver.
func (ctp *CreateTableParameters) ensureConstraints() {
	// Early exit if already set.
	if ctp.Constraints != "" {
		return
	}

	// If we are allowed to use `ALTER TABLE ... ADD CONSTRAINT ...` statements
	// then no constraints are needed in the `CREATE TABLE` statement.
	if !ctp.SkipConstraintStatements {
		return
	}

	ctp.Constraints = createTableInlineConstraintsSQL
	return
}

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

	ctp, createStatement := createMigrationsSQL(manager)
	_, err = tx.ExecContext(ctx, createStatement)
	if err != nil {
		return
	}

	if ctp.SkipConstraintStatements {
		err = tx.Commit()
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

func createMigrationsSQL(manager *Manager) (CreateTableParameters, string) {
	table := manager.MetadataTable
	provider := manager.Provider
	ctp := provider.NewCreateTableParameters()

	statement := fmt.Sprintf(
		createMigrationsTableSQL,
		provider.QuoteIdentifier(table),
		ctp.SerialID,
		ctp.Revision,
		ctp.Previous,
		ctp.CreatedAt,
		ctp.Constraints,
	)
	return ctp, statement
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

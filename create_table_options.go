package golembic

// OptCreateTableSerialID sets the `SerialID` field in create table options.
func OptCreateTableSerialID(serialID string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.SerialID = serialID
		return
	}
}

// OptCreateTableRevision sets the `Revision` field in create table options.
func OptCreateTableRevision(revision string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.Revision = revision
		return
	}
}

// OptCreateTablePrevious sets the `Previous` field in create table options.
func OptCreateTablePrevious(previous string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.Previous = previous
		return
	}
}

// OptCreateTableCreatedAt sets the `CreatedAt` field in create table options.
func OptCreateTableCreatedAt(createdAt string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.CreatedAt = createdAt
		return
	}
}

// OptCreateTableConstraints sets the `Constraints` field in create table options.
func OptCreateTableConstraints(constraints string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.Constraints = constraints
		return
	}
}

// OptCreateTableSkip sets the `SkipConstraintStatements` field in create
// table options.
func OptCreateTableSkip(skip bool) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.SkipConstraintStatements = skip
		return
	}
}

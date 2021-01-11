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

package golembic

// OptCreateTableCreatedAt sets the `CreatedAt` field in create table options.
func OptCreateTableCreatedAt(createdAt string) CreateTableOption {
	return func(ctp *CreateTableParameters) {
		ctp.CreatedAt = createdAt
		return
	}
}

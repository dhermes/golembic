package golembic

// Migration represents an individual up / down migration to be applied or
// rolled back (depending on the context).
type Migration struct {
	// Revision is an opaque name that uniquely identifies a migration. It
	// is required for a migration to be valid.
	Revision string
	// Parent is the revision identifier for the migration immediately
	// preceding this one. If absent, this indicates that this migration is
	// the "base" or "root" migration.
	Parent string
	// Description is a long form description of why the migration is being
	// performed. It is intended to be used in "describe" scenarios where
	// a long form "history" of changes is presented.
	Description string
	// Up is the function to be executed when a migration is being applied. It
	// is required for a migration to be valid.
	Up UpMigration
	// Down is the function to be executed when a migration is being rolled
	// back. It is required for a migration to be valid.
	Down DownMigration
}

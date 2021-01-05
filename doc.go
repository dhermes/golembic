// Package golembic is intended to provide tooling for managing SQL migrations in Go.
//
// The underlying principles of golembic are as follows:
//
// - Migrations should follow a straight line (from some root); this line should
// be verified to avoid merge conflicts causing "branching"
//
// - The "current" state of the world will be tracked in a `migrations` table
// containing all migrations that have been applied.
//
// - A series of migrations should be easy to use both in a script or as part
// of a larger piece of Go code
//
// - Avoid all import time side effects caused either by importing a package that
// uses `init()` or by requiring migrations files to use `init()`
//
// - Down migrations are not supported. The idea being that the risk of data loss
// from a down migration is not worth it and that writing down migrations can
// be more challenging than writing up migrations.
//
// The design allows for running "arbitrary" code inside migrations
// so that even non-SQL tasks can be tracked as a "run-once" migration.
//
// The name is a portmanteau of Go (the programming language) and `alembic`, the
// Python migrations package. The name `alembic` itself is motivated by the
// foundational ORM SQLAlchemy (an alembic is a distilling apparatus used by
// alchemists).
package golembic

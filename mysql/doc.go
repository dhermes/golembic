// Package mysql provides MySQL helpers for golembic.
//
// The DSN `Config` helper struct from `github.com/go-sql-driver/mysql` has
// been vendored into this package (license headers intact) in the files
// prefixed with `vendor_`. This was done to avoid invoking `init()` in that
// package and registering a driver that would not be used.
package mysql

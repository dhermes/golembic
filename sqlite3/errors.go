package sqlite3

import (
	"errors"
)

var (
	// ErrTimestampNotInteger is the error returned when a TimeFromInteger column
	// is expected but the value in the database is not an integer.
	ErrTimestampNotInteger = errors.New("Timestamp was not stored as an integer")
	// ErrTimestampRounding is the error returned when a TimeFromInteger column
	// can't be converted an integer (e.g. microseconds since the epoch) without
	// rounding.
	ErrTimestampRounding = errors.New("Timestamp cannot be converted to integer without rounding")
)

package postgres

import (
	"errors"
)

var (
	// ErrNotMilliseconds is the error returned when a duration cannot be
	// converted to milliseconds without round off.
	ErrNotMilliseconds = errors.New("Cannot convert to milliseconds")
)

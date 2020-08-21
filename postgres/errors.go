package postgres

import (
	"errors"
)

var (
	// ErrNotMilliseconds is the error returned when a duration cannot be
	// converted to milliseconds without round off.
	ErrNotMilliseconds = errors.New("Cannot convert to milliseconds")
	// ErrNegativeTimeout is the error returned when a timeout duration cannot
	// be negative.
	ErrNegativeTimeout = errors.New("Negative values not allowed for timeouts")
	// ErrNegativeCount is the error returned when a configured count cannot
	// be negative.
	ErrNegativeCount = errors.New("Negative values not allowed for count")
)

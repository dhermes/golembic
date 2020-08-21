package postgres

import (
	"errors"
)

var (
	// ErrDurationConversion is the error returned when a duration cannot be
	// converted to multiple of some base (e.g. milliseconds or seconds)
	// without round off.
	ErrDurationConversion = errors.New("Cannot convert duration")
	// ErrNegativeTimeout is the error returned when a timeout duration cannot
	// be negative.
	ErrNegativeTimeout = errors.New("Negative values not allowed for timeouts")
	// ErrNegativeCount is the error returned when a configured count cannot
	// be negative.
	ErrNegativeCount = errors.New("Negative values not allowed for count")
)

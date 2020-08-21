package postgres

import (
	"errors"
)

var (
	// ErrNegativeTimeout is the error returned when a timeout duration cannot
	// be negative.
	ErrNegativeTimeout = errors.New("Negative values not allowed for timeouts")
	// ErrNegativeCount is the error returned when a configured count cannot
	// be negative.
	ErrNegativeCount = errors.New("Negative values not allowed for count")
)

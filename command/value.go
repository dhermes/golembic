package command

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//       * `RoundDuration` satisfies `pflag.Value`.
var (
	_ pflag.Value = (*RoundDuration)(nil)
)

// RoundDuration wraps a `time.Duration` as a value that can be used as flag
// with `cobra` / `pflag`, but one that must be convertible to a multiple of
// some base duration.
type RoundDuration struct {
	Base  time.Duration
	Value *time.Duration
}

// String is the string representation of the stored value.
func (dv *RoundDuration) String() string {
	if dv.Value == nil || *dv.Value == 0 {
		return ""
	}
	return fmt.Sprintf("%s", dv.Value)
}

// Set sets the duration based on a string input.
func (dv *RoundDuration) Set(value string) error {
	d, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	_, err = golembic.ToRoundDuration(d, dv.Base)
	if err != nil {
		return err
	}

	*dv.Value = d
	return nil
}

// Type is a human readable "description" of the underlying type being
// represented.
func (dv *RoundDuration) Type() string {
	return "duration"
}

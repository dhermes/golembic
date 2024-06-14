package sqlite3

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/dhermes/golembic"
)

// NOTE: Ensure that
//   - `TimeFromInteger` satisfies `golembic.Column`.
//   - `TimeFromInteger` satisfies `golembic.TimestampColumn`.
var (
	_ golembic.Column          = (*TimeFromInteger)(nil)
	_ golembic.TimestampColumn = (*TimeFromInteger)(nil)
)

// TimeFromInteger represents a `time.Time` stored in a database as an
// `INTEGER` number of microseconds (in UTC) since the epoch.
//
// This is **necessary** because SQLite doesn't have rich support for
// timestamps and the drivers can't paper over this issue, e.g.
// https://github.com/mattn/go-sqlite3/issues/142
type TimeFromInteger struct {
	Stored time.Time
}

// Scan implements the Scanner interface.
func (tfi *TimeFromInteger) Scan(src interface{}) error {
	if src == nil {
		tfi.Stored = time.Time{}
		return nil
	}

	srcInteger, ok := src.(int64)
	if !ok {
		return fmt.Errorf("%w; value type: %T", ErrTimestampNotInteger, src)
	}

	seconds := srcInteger / 1000000
	microseconds := seconds % 1000000

	tfi.Stored = time.Unix(seconds, 1000*microseconds)
	return nil
}

// Value implements the driver Valuer interface.
func (tfi TimeFromInteger) Value() (driver.Value, error) {
	if tfi.Stored.IsZero() {
		return nil, nil
	}

	unixNano := tfi.Stored.UnixNano()
	if unixNano%1000 != 0 {
		return nil, fmt.Errorf("%w; precision: microseconds, value: %s", ErrTimestampRounding, tfi.Stored)
	}

	unixMicro := unixNano / 1000
	return unixMicro, nil
}

// Pointer returns the current pointer receiver.
func (tfi *TimeFromInteger) Pointer() interface{} {
	return tfi
}

// Timestamp returns the stored timestamp value.
func (tfi TimeFromInteger) Timestamp() time.Time {
	return tfi.Stored
}

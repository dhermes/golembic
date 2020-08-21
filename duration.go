package golembic

import (
	"fmt"
	"time"
)

// ToRoundDuration converts a duration to an **exact** multiple of some base
// duration or errors if round off is required.
func ToRoundDuration(d, base time.Duration) (int64, error) {
	remainder := d % base
	if remainder != 0 {
		err := fmt.Errorf("%w; duration %s is not a multiple of %s", ErrDurationConversion, d, base)
		return 0, err
	}

	ms := int64(d / base)
	return ms, nil
}

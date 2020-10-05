package golembic

import (
	"fmt"
)

// NOTE: Ensure that
//       * `stdoutPrintf` satisfies `PrintfReceiver`.
var (
	_ PrintfReceiver = (*stdoutPrintf)(nil)
)

// stdoutPrintf implements `PrintfReceiver` and just prints to STDOUT.
type stdoutPrintf struct{}

func (sp *stdoutPrintf) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format+"\n", a...)
}

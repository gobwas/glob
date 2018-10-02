// +build globdebug

package debug

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

const Enabled = true

var i = new(int32)

func Logf(f string, args ...interface{}) {
	n := int(atomic.LoadInt32(i))
	fmt.Fprint(os.Stderr,
		strings.Repeat("  ", n),
		fmt.Sprintf("(%d) ", n),
		fmt.Sprintf(f, args...),
		"\n",
	)
}

func Enter() {
	atomic.AddInt32(i, 1)
}

func Leave() {
	atomic.AddInt32(i, -1)
}

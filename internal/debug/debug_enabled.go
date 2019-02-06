// +build globdebug

package debug

import (
	"fmt"
	"os"
	"strings"
)

const Enabled = true

var (
	i      = 0
	prefix = map[int]string{}
)

func Logf(f string, args ...interface{}) {
	if f != "" && prefix[i] != "" {
		f = ": " + f
	}
	fmt.Fprint(os.Stderr,
		strings.Repeat("  ", i),
		fmt.Sprintf("(%d) ", i),
		prefix[i],
		fmt.Sprintf(f, args...),
		"\n",
	)
}

func Indexing(name, s string) func(int, []int) {
	EnterPrefix("%s: index: %q", name, s)
	return func(index int, segments []int) {
		Logf("-> %d, %v", index, segments)
		LeavePrefix()
	}
}

func Matching(name, s string) func(bool) {
	EnterPrefix("%s: match %q", name, s)
	return func(ok bool) {
		Logf("-> %t", ok)
		LeavePrefix()
	}
}

func EnterPrefix(s string, args ...interface{}) {
	Enter()
	prefix[i] = fmt.Sprintf(s, args...)
	Logf("")
}

func LeavePrefix() {
	prefix[i] = ""
	Leave()
}

func Enter() {
	i++
}

func Leave() {
	i--
}

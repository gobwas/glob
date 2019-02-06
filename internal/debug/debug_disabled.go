// +build !globdebug

package debug

const Enabled = false

func Logf(string, ...interface{})        {}
func Enter()                             {}
func Leave()                             {}
func EnterPrefix(string, ...interface{}) {}
func LeavePrefix()                       {}
func Indexing(n, s string) func(int, []int) {
	panic("must never be called")
}
func Matching(n, s string) func(bool) {
	panic("must never be called")
}

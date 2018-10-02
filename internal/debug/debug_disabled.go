// +build !globdebug

package debug

const Enabled = false

func Logf(_ string, _ ...interface{}) {}
func Enter()                          {}
func Leave()                          {}

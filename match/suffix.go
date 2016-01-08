package match

import (
	"fmt"
	"strings"
)

type Suffix struct {
	Suffix string
}

func (self Suffix) Kind() Kind {
	return KindSuffix
}

func (self Suffix) Len() int {
	return -1
}

func (self Suffix) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Suffix) Match(s string) bool {
	return strings.HasSuffix(s, self.Suffix)
}

func (self Suffix) String() string {
	return fmt.Sprintf("[suffix:%s]", self.Suffix)
}

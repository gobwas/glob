package match

import (
	"fmt"
	"strings"
)

type Prefix struct {
	Prefix string
}

func (self Prefix) Kind() Kind {
	return KindPrefix
}

func (self Prefix) Len() int {
	return -1
}

func (self Prefix) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Prefix) Match(s string) bool {
	return strings.HasPrefix(s, self.Prefix)
}

func (self Prefix) String() string {
	return fmt.Sprintf("[prefix:%s]", self.Prefix)
}

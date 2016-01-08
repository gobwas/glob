package match

import (
	"fmt"
	"strings"
)

type List struct {
	List string
	Not  bool
}

func (self List) Kind() Kind {
	return KindList
}

func (self List) Match(s string) bool {
	if len([]rune(s)) != 1 {
		return false
	}

	inList := strings.Index(self.List, s) != -1

	return inList == !self.Not
}

func (self List) Index(s string) (index, min, max int) {
	for i, r := range []rune(s) {
		if self.Not == (strings.IndexRune(self.List, r) == -1) {
			return i, 1, 1
		}
	}

	return -1, 0, 0
}

func (self List) String() string {
	return fmt.Sprintf("[list:list=%s not=%t]", self.List, self.Not)
}

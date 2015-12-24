package match

import (
	"fmt"
)

type Between struct {
	Lo, Hi rune
	Not    bool
}

func (self Between) Kind() Kind {
	return KindRangeBetween
}

func (self Between) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Between) Match(s string) bool {
	r := []rune(s)

	if (len(r) != 1) {
		return false
	}

	inRange := r[0] >= self.Lo && r[0] <= self.Hi

	return inRange == !self.Not
}

func (self Between) String() string {
	return fmt.Sprintf("[range_between:%s-%s(%t)]", self.Lo, self.Hi, self.Not)
}
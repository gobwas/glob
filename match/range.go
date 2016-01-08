package match

import (
	"fmt"
)

type Range struct {
	Lo, Hi rune
	Not    bool
}

func (self Range) Kind() Kind {
	return KindRange
}

func (self Range) Match(s string) bool {
	r := []rune(s)

	if len(r) != 1 {
		return false
	}

	inRange := r[0] >= self.Lo && r[0] <= self.Hi

	return inRange == !self.Not
}

func (self Range) Index(s string) (index, min, max int) {
	for i, r := range []rune(s) {
		if self.Not != (r >= self.Lo && r <= self.Hi) {
			return i, 1, 1
		}
	}

	return -1, 0, 0
}

func (self Range) String() string {
	return fmt.Sprintf("[range_between:%s-%s(%t)]", self.Lo, self.Hi, self.Not)
}

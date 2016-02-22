package match

import (
	"fmt"
	"unicode/utf8"
)

type Range struct {
	Lo, Hi rune
	Not    bool
}

// todo make factory
// todo make range table inside factory

func (self Range) Len() int {
	return lenOne
}

func (self Range) Match(s string) bool {
	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}

	inRange := r >= self.Lo && r <= self.Hi

	return inRange == !self.Not
}

func (self Range) Index(s string, segments []int) (int, []int) {
	for i, r := range s {
		if self.Not != (r >= self.Lo && r <= self.Hi) {
			return i, append(segments, utf8.RuneLen(r))
		}
	}

	return -1, segments
}

func (self Range) String() string {
	var not string
	if self.Not {
		not = "!"
	}
	return fmt.Sprintf("<range:%s[%s-%s]>", not, string(self.Lo), string(self.Hi))
}

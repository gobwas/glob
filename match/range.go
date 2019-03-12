package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/internal/debug"
)

type Range struct {
	Lo, Hi rune
	Not    bool
}

func NewRange(lo, hi rune, not bool) Range {
	return Range{lo, hi, not}
}

func (self Range) MinLen() int {
	return 1
}

func (self Range) RunesCount() int {
	return 1
}

func (self Range) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("range", s)
		defer func() { done(ok) }()
	}
	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}

	inRange := r >= self.Lo && r <= self.Hi

	return inRange == !self.Not
}

func (self Range) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("range", s)
		defer func() { done(index, segments) }()
	}
	for i, r := range s {
		if self.Not != (r >= self.Lo && r <= self.Hi) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}

	return -1, nil
}

func (self Range) String() string {
	var not string
	if self.Not {
		not = "!"
	}
	return fmt.Sprintf("<range:%s[%s,%s]>", not, string(self.Lo), string(self.Hi))
}

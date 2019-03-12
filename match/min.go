package match

import (
	"fmt"
	"unicode/utf8"
)

type Min struct {
	n int
}

func NewMin(n int) Min {
	return Min{n}
}

func (m Min) Match(s string) bool {
	var n int
	for range s {
		n += 1
		if n >= m.n {
			return true
		}
	}
	return false
}

func (m Min) Index(s string) (int, []int) {
	var count int

	c := len(s) - m.n + 1
	if c <= 0 {
		return -1, nil
	}
	segments := acquireSegments(c)
	for i, r := range s {
		count++
		if count >= m.n {
			segments = append(segments, i+utf8.RuneLen(r))
		}
	}
	if len(segments) == 0 {
		return -1, nil
	}
	return 0, segments
}

func (m Min) MinLen() int {
	return m.n
}

func (m Min) String() string {
	return fmt.Sprintf("<min:%d>", m.n)
}

package match

import (
	"fmt"
	"unicode/utf8"
)

type Max struct {
	n int
}

func NewMax(n int) Max {
	return Max{n}
}

func (m Max) Match(s string) bool {
	var n int
	for range s {
		n += 1
		if n > m.n {
			return false
		}
	}
	return true
}

func (m Max) Index(s string) (int, []int) {
	segments := acquireSegments(m.n + 1)
	segments = append(segments, 0)
	var count int
	for i, r := range s {
		count++
		if count > m.n {
			break
		}
		segments = append(segments, i+utf8.RuneLen(r))
	}

	return 0, segments
}

func (m Max) MinLen() int {
	return 0
}

func (m Max) String() string {
	return fmt.Sprintf("<max:%d>", m.n)
}

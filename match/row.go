package match

import (
	"fmt"
)

type Row struct {
	Matchers    Matchers
	RunesLength int
	Segments    []int
}

func NewRow(len int, m ...Matcher) Row {
	return Row{
		Matchers:    Matchers(m),
		RunesLength: len,
		Segments:    []int{len},
	}
}

func (self Row) matchAll(s string) bool {
	var idx int
	for _, m := range self.Matchers {
		length := m.Len()

		var next, i int
		for next = range s[idx:] {
			i++
			if i == length {
				break
			}
		}

		end := idx + next + 1
		if i < length || end > len(s) || !m.Match(s[idx:end]) {
			return false
		}

		idx = end
	}

	return true
}

func (self Row) lenOk(s string) bool {
	var i int
	for range s {
		i++
		if i > self.RunesLength {
			return false
		}
	}
	return self.RunesLength == i
}

func (self Row) Match(s string) bool {
	return self.lenOk(s) && self.matchAll(s)
}

func (self Row) Len() (l int) {
	return self.RunesLength
}

func (self Row) Index(s string) (int, []int) {
	for i := range s {
		if len(s[i:]) < self.RunesLength {
			break
		}
		if self.matchAll(s[i:]) {
			return i, self.Segments
		}
	}
	return -1, nil
}

func (self Row) String() string {
	return fmt.Sprintf("<row_%d:[%s]>", self.RunesLength, self.Matchers)
}

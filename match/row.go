package match

import (
	"fmt"
	"unicode/utf8"
)

type Row struct {
	Matchers    Matchers
	RunesLength int
}

func (self Row) matchAll(s string) bool {
	var idx int
	for _, m := range self.Matchers {
		l := m.Len()
		if !m.Match(s[idx : idx+l]) {
			return false
		}

		idx += l
	}

	return true
}

func (self Row) Match(s string) bool {
	if utf8.RuneCountInString(s) < self.RunesLength {
		return false
	}

	return self.matchAll(s)
}

func (self Row) Len() (l int) {
	return self.RunesLength
}

func (self Row) Index(s string) (int, []int) {
	l := utf8.RuneCountInString(s)
	if l < self.RunesLength {
		return -1, nil
	}

	for i := range s {
		sub := s[i:]
		if self.matchAll(sub) {
			return i, []int{self.RunesLength}
		}

		l -= 1
		if l < self.RunesLength {
			return -1, nil
		}
	}

	return -1, nil
}

func (self Row) String() string {
	return fmt.Sprintf("<row_%d:[%s]>", self.RunesLength, self.Matchers)
}

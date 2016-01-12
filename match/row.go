package match

import (
	"fmt"
)

type Row struct {
	Matchers Matchers
	len      int
}

func (self *Row) Add(m Matcher) error {
	if l := m.Len(); l == -1 {
		return fmt.Errorf("matcher should have fixed length")
	}

	self.Matchers = append(self.Matchers, m)
	return nil
}

func (self Row) Match(s string) bool {
	if len(s) < self.Len() {
		return false
	}

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

func (self Row) Len() (l int) {
	if self.len == 0 {
		for _, m := range self.Matchers {
			self.len += m.Len()
		}
	}

	return self.len
}

func (self Row) Index(s string) (int, []int) {
	for i := range s {
		sub := s[i:]
		if self.Match(sub) {
			return i, []int{self.Len()}
		}
	}

	return -1, nil
}

func (self Row) Kind() Kind {
	return KindMin
}

func (self Row) String() string {
	return fmt.Sprintf("<row:[%s]>", self.Matchers)
}

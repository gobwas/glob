package match

import (
	"fmt"
)

type Every struct {
	Matchers Matchers
}

func (self *Every) Add(m Matcher) error {
	self.Matchers = append(self.Matchers, m)
	return nil
}

func (self Every) Len() (l int) {
	for _, m := range self.Matchers {
		if ml := m.Len(); l > 0 {
			l += ml
		} else {
			return -1
		}
	}

	return
}

func (self Every) Match(s string) bool {
	for _, m := range self.Matchers {
		if !m.Match(s) {
			return false
		}
	}

	return true
}

func (self Every) Kind() Kind {
	return KindEveryOf
}

func (self Every) String() string {
	return fmt.Sprintf("[every_of:%s]", self.Matchers)
}

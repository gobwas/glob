package match

import (
	"fmt"
)

type Every struct {
	Matchers Matchers
}

func (self *Every) Add(m Matcher) {
	self.Matchers = append(self.Matchers, m)
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

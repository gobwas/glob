package match

import (
	"fmt"
)

type AnyOf struct {
	Matchers Matchers
}

func (self *AnyOf) Add(m Matcher) error {
	self.Matchers = append(self.Matchers, m)
	return nil
}

func (self AnyOf) Match(s string) bool {
	for _, m := range self.Matchers {
		if m.Match(s) {
			return true
		}
	}

	return false
}

func (self AnyOf) Len() int {
	return -1
}

func (self AnyOf) Kind() Kind {
	return KindAnyOf
}

func (self AnyOf) String() string {
	return fmt.Sprintf("[any_of:%s]", self.Matchers)
}

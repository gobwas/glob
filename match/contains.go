package match

import (
	"fmt"
	"strings"
)

type Contains struct {
	Needle string
	Not    bool
}

func (self Contains) Match(s string) bool {
	return strings.Contains(s, self.Needle) != self.Not
}

func (self Contains) Kind() Kind {
	return KindContains
}

func (self Contains) String() string {
	return fmt.Sprintf("[contains:needle=%s not=%t]", self.Needle, self.Not)
}

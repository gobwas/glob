package match

import (
	"fmt"
	"strings"
)

type Any struct {
	Separators string
}

func (self Any) Match(s string) bool {
	return strings.IndexAny(s, self.Separators) == -1
}

func (self Any) Search(s string) (i, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Any) Kind() Kind {
	if self.Separators == "" {
		return KindMultipleSuper
	} else {
		return KindMultipleSeparated
	}
}

func (self Any) String() string {
	return fmt.Sprintf("[multiple:%s]", self.Separators)
}

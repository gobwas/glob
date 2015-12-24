package match

import (
	"strings"
	"fmt"
)

// multiple represents *
type Multiple struct {
	Separators string
}

func (self Multiple) Match(s string) bool {
	return strings.IndexAny(s, self.Separators) == -1
}

func (self Multiple) Search(s string) (i, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Multiple) Kind() Kind {
	if self.Separators == "" {
		return KindMultipleSuper
	} else {
		return KindMultipleSeparated
	}
}

func (self Multiple) String() string {
	return fmt.Sprintf("[multiple:%s]", self.Separators)
}
package match

import (
	"fmt"
	"strings"
)

// single represents ?
type Single struct {
	Separators string
}

func (self Single) Match(s string) bool {
	return len([]rune(s)) == 1 && strings.IndexAny(s, self.Separators) == -1
}

func (self Single) Index(s string) (index, min, max int) {
	for i, c := range []rune(s) {
		if strings.IndexRune(self.Separators, c) == -1 {
			return i, 1, 1
		}
	}

	return -1, 0, 0
}

func (self Single) Kind() Kind {
	return KindSingle
}

func (self Single) String() string {
	return fmt.Sprintf("[single:%s]", self.Separators)
}

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

func (self Any) Index(s string) (index, min, max int) {
	index = -1

	for i, r := range []rune(s) {
		if strings.IndexRune(self.Separators, r) == -1 {
			if index == -1 {
				index = i
			}
			max++
			continue
		}

		if index != -1 {
			break
		}
	}

	return
}

func (self Any) Kind() Kind {
	return KindAny
}

func (self Any) String() string {
	return fmt.Sprintf("[any:%s]", self.Separators)
}

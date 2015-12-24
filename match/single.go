package match

import (
	"strings"
	"fmt"
)


// single represents ?
type Single struct {
	Separators string
}

func (self Single) Match(s string) bool {
	return len([]rune(s)) == 1 && strings.IndexAny(s, self.Separators) == -1
}

func (self Single) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Single) Kind() Kind {
	return KindSingle
}


func (self Single) String() string {
	return fmt.Sprintf("[single:%s]", self.Separators)
}

package match

import (
	"strings"
	"fmt"
)

// raw represents raw string to match
type Raw struct {
	Str string
}

func (self Raw) Match(s string) bool {
	return self.Str == s
}

func (self Raw) Kind() Kind {
	return KindRaw
}

func (self Raw) Search(s string) (i int, l int, ok bool) {
	index := strings.Index(s, self.Str)
	if index == -1 {
		return
	}

	i = index
	l = len(self.Str)
	ok = true

	return
}

func (self Raw) String() string {
	return fmt.Sprintf("[raw:%s]", self.Str)
}

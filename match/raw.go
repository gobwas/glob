package match

import (
	"fmt"
	"strings"
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

func (self Raw) Index(s string) (index, min, max int) {
	index = strings.Index(s, self.Str)
	if index == -1 {
		return
	}

	min = len(self.Str)
	max = min

	return
}

func (self Raw) String() string {
	return fmt.Sprintf("[raw:%s]", self.Str)
}

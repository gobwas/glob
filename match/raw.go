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

func (self Raw) Len() int {
	return len(self.Str)
}

func (self Raw) Kind() Kind {
	return KindRaw
}

func (self Raw) Index(s string) (index int, segments []int) {
	index = strings.Index(s, self.Str)
	if index == -1 {
		return
	}

	segments = []int{len(self.Str)}

	return
}

func (self Raw) String() string {
	return fmt.Sprintf("[raw:%s]", self.Str)
}

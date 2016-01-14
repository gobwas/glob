package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// raw represents raw string to match
type Raw struct {
	Str    string
	Length int
}

func NewRaw(s string) Raw {
	return Raw{
		Str:    s,
		Length: utf8.RuneCountInString(s),
	}
}

func (self Raw) Match(s string) bool {
	return self.Str == s
}

func (self Raw) Len() int {
	return self.Length
}

func (self Raw) Index(s string) (index int, segments []int) {
	index = strings.Index(s, self.Str)
	if index == -1 {
		return
	}

	segments = []int{self.Length}

	return
}

func (self Raw) String() string {
	return fmt.Sprintf("<raw:%s>", self.Str)
}

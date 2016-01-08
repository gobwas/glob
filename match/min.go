package match

import (
	"fmt"
	"unicode/utf8"
)

type Min struct {
	Limit int
}

func (self Min) Match(s string) bool {
	return utf8.RuneCountInString(s) >= self.Limit
}

func (self Min) Len() int {
	return -1
}

func (self Min) Search(s string) (int, int, bool) {
	return 0, 0, false
}

func (self Min) Kind() Kind {
	return KindMin
}

func (self Min) String() string {
	return fmt.Sprintf("[min:%d]", self.Limit)
}

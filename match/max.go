package match

import (
	"fmt"
	"unicode/utf8"
)

type Max struct {
	Limit int
}

func (self Max) Match(s string) bool {
	return utf8.RuneCountInString(s) <= self.Limit
}

func (self Max) Len() int {
	return -1
}

func (self Max) Search(s string) (int, int, bool) {
	return 0, 0, false
}

func (self Max) Kind() Kind {
	return KindMax
}

func (self Max) String() string {
	return fmt.Sprintf("[max:%d]", self.Limit)
}

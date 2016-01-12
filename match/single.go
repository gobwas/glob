package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// single represents ?
type Single struct {
	Separators string
}

func (self Single) Match(s string) bool {
	return utf8.RuneCountInString(s) == 1 && strings.IndexAny(s, self.Separators) == -1
}

func (self Single) Len() int {
	return 1
}

func (self Single) Index(s string) (int, []int) {
	for i, r := range s {
		if strings.IndexRune(self.Separators, r) == -1 {
			return i, []int{utf8.RuneLen(r)}
		}
	}

	return -1, nil
}

func (self Single) Kind() Kind {
	return KindSingle
}

func (self Single) String() string {
	return fmt.Sprintf("<single:![%s]>", self.Separators)
}

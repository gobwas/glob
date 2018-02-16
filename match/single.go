package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

// single represents ?
type Single struct {
	sep []rune
}

func NewSingle(s []rune) Single {
	return Single{s}
}

func (s Single) Match(v string) bool {
	r, w := utf8.DecodeRuneInString(v)
	if len(v) > w {
		return false
	}
	return runes.IndexRune(s.sep, r) == -1
}

func (s Single) MinLen() int {
	return 1
}

func (s Single) RunesCount() int {
	return 1
}

func (s Single) Index(v string) (int, []int) {
	for i, r := range v {
		if runes.IndexRune(s.sep, r) == -1 {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}
	return -1, nil
}

func (s Single) String() string {
	return fmt.Sprintf("<single:![%s]>", string(s.sep))
}

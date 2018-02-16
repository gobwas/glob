package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

type List struct {
	rs  []rune
	not bool
}

func NewList(rs []rune, not bool) List {
	return List{rs, not}
}

func (l List) Match(s string) bool {
	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		// Invalid rune.
		return false
	}
	inList := runes.IndexRune(l.rs, r) != -1
	return inList == !l.not
}

func (l List) MinLen() int {
	return 1
}

func (l List) Index(s string) (int, []int) {
	for i, r := range s {
		if l.not == (runes.IndexRune(l.rs, r) == -1) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}
	return -1, nil
}

func (l List) String() string {
	var not string
	if l.not {
		not = "!"
	}
	return fmt.Sprintf("<list:%s[%s]>", not, string(l.rs))
}

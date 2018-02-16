package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

type Row struct {
	ms    []MatchIndexSizer
	runes int
	seg   []int
}

func NewRow(ms []MatchIndexSizer) Row {
	var r int
	for _, m := range ms {
		r += m.RunesCount()
	}
	return Row{
		ms:    ms,
		runes: r,
		seg:   []int{r},
	}
}

func (r Row) Match(s string) bool {
	if !runes.ExactlyRunesCount(s, r.runes) {
		return false
	}
	return r.matchAll(s)
}

func (r Row) MinLen() int {
	return r.runes
}

func (r Row) RunesCount() int {
	return r.runes
}

func (r Row) Index(s string) (int, []int) {
	for j := 0; j < len(s)-r.runes; {
		i, _ := r.ms[0].Index(s[j:])
		if i == -1 {
			return -1, nil
		}
		if r.matchAll(s[i:]) {
			return j + i, r.seg
		}
		_, x := utf8.DecodeRuneInString(s[i:])
		j += x
	}
	return -1, nil
}

func (r Row) String() string {
	return fmt.Sprintf("<row_%d:[%s]>", r.runes, r.ms)
}

func (r Row) matchAll(s string) bool {
	var i int
	for _, m := range r.ms {
		n := m.RunesCount()
		sub := runes.Head(s[i:], n)
		if !m.Match(sub) {
			return false
		}
		i += len(sub)
	}
	return true
}

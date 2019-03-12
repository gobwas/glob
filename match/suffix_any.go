package match

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

type SuffixAny struct {
	s      string
	sep    []rune
	minLen int
}

func NewSuffixAny(s string, sep []rune) SuffixAny {
	return SuffixAny{s, sep, utf8.RuneCountInString(s)}
}

func (s SuffixAny) Index(v string) (int, []int) {
	idx := strings.Index(v, s.s)
	if idx == -1 {
		return -1, nil
	}

	i := runes.LastIndexAnyRune(v[:idx], s.sep) + 1

	return i, []int{idx + len(s.s) - i}
}

func (s SuffixAny) MinLen() int {
	return s.minLen
}

func (s SuffixAny) Match(v string) bool {
	if !strings.HasSuffix(v, s.s) {
		return false
	}
	return runes.IndexAnyRune(v[:len(v)-len(s.s)], s.sep) == -1
}

func (s SuffixAny) String() string {
	return fmt.Sprintf("<suffix_any:![%s]%s>", string(s.sep), s.s)
}

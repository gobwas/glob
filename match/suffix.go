package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Suffix struct {
	s      string
	minLen int
}

func NewSuffix(s string) Suffix {
	return Suffix{s, utf8.RuneCountInString(s)}
}

func (s Suffix) MinLen() int {
	return s.minLen
}

func (s Suffix) Match(v string) bool {
	return strings.HasSuffix(v, s.s)
}

func (s Suffix) Index(v string) (int, []int) {
	idx := strings.Index(v, s.s)
	if idx == -1 {
		return -1, nil
	}
	return 0, []int{idx + len(s.s)}
}

func (s Suffix) String() string {
	return fmt.Sprintf("<suffix:%s>", s.s)
}

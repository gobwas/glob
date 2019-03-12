package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Prefix struct {
	s       string
	minSize int
}

func NewPrefix(p string) Prefix {
	return Prefix{
		s:       p,
		minSize: utf8.RuneCountInString(p),
	}
}

func (p Prefix) Index(s string) (int, []int) {
	idx := strings.Index(s, p.s)
	if idx == -1 {
		return -1, nil
	}

	length := len(p.s)
	var sub string
	if len(s) > idx+length {
		sub = s[idx+length:]
	} else {
		sub = ""
	}

	segments := acquireSegments(len(sub) + 1)
	segments = append(segments, length)
	for i, r := range sub {
		segments = append(segments, length+i+utf8.RuneLen(r))
	}

	return idx, segments
}

func (p Prefix) MinLen() int {
	return p.minSize
}

func (p Prefix) Match(s string) bool {
	return strings.HasPrefix(s, p.s)
}

func (p Prefix) String() string {
	return fmt.Sprintf("<prefix:%s>", p.s)
}

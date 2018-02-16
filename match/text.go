package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// raw represents raw string to match
type Text struct {
	s     string
	runes int
	bytes int
	seg   []int
}

func NewText(s string) Text {
	return Text{
		s:     s,
		runes: utf8.RuneCountInString(s),
		bytes: len(s),
		seg:   []int{len(s)},
	}
}

func (t Text) Match(s string) bool {
	return t.s == s
}

func (t Text) Index(s string) (int, []int) {
	i := strings.Index(s, t.s)
	if i == -1 {
		return -1, nil
	}
	return i, t.seg
}

func (t Text) MinLen() int {
	return t.runes
}

func (t Text) BytesCount() int {
	return t.bytes
}

func (t Text) RunesCount() int {
	return t.runes
}

func (t Text) String() string {
	return fmt.Sprintf("<text:`%v`>", t.s)
}

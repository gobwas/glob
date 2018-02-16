package match

// todo common table of rune's length

import (
	"fmt"
	"strings"
)

type Matcher interface {
	Match(string) bool
	MinLen() int
}

type Indexer interface {
	Index(string) (int, []int)
}

type Sizer interface {
	RunesCount() int
}

type MatchIndexer interface {
	Matcher
	Indexer
}

type MatchSizer interface {
	Matcher
	Sizer
}

type MatchIndexSizer interface {
	Matcher
	Indexer
	Sizer
}

type Container interface {
	Content() []Matcher
}

func MatchIndexers(ms []Matcher) ([]MatchIndexer, bool) {
	for _, m := range ms {
		if _, ok := m.(Indexer); !ok {
			return nil, false
		}
	}
	mis := make([]MatchIndexer, len(ms))
	for i := range mis {
		mis[i] = ms[i].(MatchIndexer)
	}
	return mis, true
}

type Matchers []Matcher

func (m Matchers) String() string {
	var s []string
	for _, matcher := range m {
		s = append(s, fmt.Sprint(matcher))
	}

	return fmt.Sprintf("%s", strings.Join(s, ","))
}

// appendMerge merges and sorts given already SORTED and UNIQUE segments.
func appendMerge(target, sub []int) []int {
	lt, ls := len(target), len(sub)
	out := make([]int, 0, lt+ls)

	for x, y := 0, 0; x < lt || y < ls; {
		if x >= lt {
			out = append(out, sub[y:]...)
			break
		}

		if y >= ls {
			out = append(out, target[x:]...)
			break
		}

		xValue := target[x]
		yValue := sub[y]

		switch {

		case xValue == yValue:
			out = append(out, xValue)
			x++
			y++

		case xValue < yValue:
			out = append(out, xValue)
			x++

		case yValue < xValue:
			out = append(out, yValue)
			y++

		}
	}

	target = append(target[:0], out...)

	return target
}

func reverseSegments(input []int) {
	l := len(input)
	m := l / 2

	for i := 0; i < m; i++ {
		input[i], input[l-i-1] = input[l-i-1], input[i]
	}
}

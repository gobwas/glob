package match

import (
	"fmt"
	"strings"
)

type Kind int

// todo use String for Kind, and self.Kind() in every matcher.String()
const (
	KindRaw Kind = iota
	KindEveryOf
	KindAnyOf
	KindAny
	KindSuper
	KindSingle
	KindComposition
	KindPrefix
	KindSuffix
	KindPrefixSuffix
	KindRange
	KindList
	KindMin
	KindMax
	KindBTree
	KindContains
)

type Matcher interface {
	Match(string) bool
	Index(string) (int, []int)
	Len() int
	String() string
}

type Matchers []Matcher

func (m Matchers) String() string {
	var s []string
	for _, matcher := range m {
		s = append(s, fmt.Sprint(matcher))
	}

	return fmt.Sprintf("%s", strings.Join(s, ","))
}

func appendIfNotAsPrevious(target []int, val int) []int {
	l := len(target)
	if l != 0 && target[l-1] == val {
		return target
	}

	return append(target, val)
}

// mergeSegments merges and sorts given already SORTED and UNIQUE segments.
func mergeSegments(segments [][]int) []int {
	var current []int
	for _, s := range segments {
		if current == nil {
			current = s
			continue
		}

		var next []int
		for x, y := 0, 0; x < len(current) || y < len(s); {
			if x >= len(current) {
				next = append(next, s[y:]...)
				break
			}

			if y >= len(s) {
				next = append(next, current[x:]...)
				break
			}

			xValue := current[x]
			yValue := s[y]

			switch {

			case xValue == yValue:
				x++
				y++
				next = appendIfNotAsPrevious(next, xValue)

			case xValue < yValue:
				next = appendIfNotAsPrevious(next, xValue)
				x++

			case yValue < xValue:
				next = appendIfNotAsPrevious(next, yValue)
				y++

			}
		}

		current = next
	}

	return current
}

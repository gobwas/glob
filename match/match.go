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
}

type Primitive interface {
	Index(string) (int, int, int)
}

type Matchers []Matcher

func (m Matchers) String() string {
	var s []string
	for _, matcher := range m {
		s = append(s, fmt.Sprint(matcher))
	}

	return fmt.Sprintf("matchers[%s]", strings.Join(s, ","))
}

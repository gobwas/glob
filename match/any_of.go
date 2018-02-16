package match

import (
	"fmt"
)

type AnyOf struct {
	ms  []Matcher
	min int
}

func NewAnyOf(ms ...Matcher) Matcher {
	a := AnyOf{ms, minLen(ms)}
	if mis, ok := MatchIndexers(ms); ok {
		return IndexedAnyOf{a, mis}
	}
	return a
}

func (a AnyOf) Match(s string) bool {
	for _, m := range a.ms {
		if m.Match(s) {
			return true
		}
	}
	return false
}

func (a AnyOf) MinLen() (n int) {
	return a.min
}

func (a AnyOf) Content() []Matcher {
	return a.ms
}

func (a AnyOf) String() string {
	return fmt.Sprintf("<any_of:[%s]>", Matchers(a.ms))
}

type IndexedAnyOf struct {
	AnyOf
	ms []MatchIndexer
}

func (a IndexedAnyOf) Index(s string) (int, []int) {
	index := -1
	segments := acquireSegments(len(s))
	for _, m := range a.ms {
		i, seg := m.Index(s)
		if i == -1 {
			continue
		}
		if index == -1 || i < index {
			index = i
			segments = append(segments[:0], seg...)
			continue
		}
		if i > index {
			continue
		}
		// here i == index
		segments = appendMerge(segments, seg)
	}
	if index == -1 {
		releaseSegments(segments)
		return -1, nil
	}
	return index, segments
}

func (a IndexedAnyOf) String() string {
	return fmt.Sprintf("<indexed_any_of:[%s]>", a.ms)
}

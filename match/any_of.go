package match

import (
	"fmt"

	"github.com/gobwas/glob/internal/debug"
)

type AnyOf struct {
	ms  []Matcher
	min int
}

func NewAnyOf(ms ...Matcher) Matcher {
	a := AnyOf{ms, minLen(ms)}
	if mis, ok := MatchIndexers(ms); ok {
		x := IndexedAnyOf{a, mis}
		if msz, ok := MatchIndexSizers(ms); ok {
			sz := -1
			for _, m := range msz {
				n := m.RunesCount()
				if sz == -1 {
					sz = n
				} else if sz != n {
					sz = -1
					break
				}
			}
			if sz != -1 {
				return IndexedSizedAnyOf{x, sz}
			}
		}
		return x
	}
	return a
}

func MustIndexedAnyOf(ms ...Matcher) MatchIndexer {
	return NewAnyOf(ms...).(MatchIndexer)
}

func MustIndexedSizedAnyOf(ms ...Matcher) MatchIndexSizer {
	return NewAnyOf(ms...).(MatchIndexSizer)
}

func (a AnyOf) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("any_of", s)
		defer func() { done(ok) }()
	}
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

func (a AnyOf) Content(cb func(Matcher)) {
	for _, m := range a.ms {
		cb(m)
	}
}

func (a AnyOf) String() string {
	return fmt.Sprintf("<any_of:[%s]>", Matchers(a.ms))
}

type IndexedAnyOf struct {
	AnyOf
	ms []MatchIndexer
}

func (a IndexedAnyOf) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("any_of", s)
		defer func() { done(index, segments) }()
	}
	index = -1
	segments = acquireSegments(len(s))
	for _, m := range a.ms {
		if debug.Enabled {
			debug.Logf("indexing: any_of: trying %s", m)
		}
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

type IndexedSizedAnyOf struct {
	IndexedAnyOf
	runes int
}

func (a IndexedSizedAnyOf) RunesCount() int {
	return a.runes
}

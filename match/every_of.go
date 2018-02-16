package match

import (
	"fmt"
)

type EveryOf struct {
	ms  []Matcher
	min int
}

func NewEveryOf(ms []Matcher) Matcher {
	e := EveryOf{ms, minLen(ms)}
	if mis, ok := MatchIndexers(ms); ok {
		return IndexedEveryOf{e, mis}
	}
	return e
}

func (e EveryOf) MinLen() (n int) {
	return e.min
}

func (e EveryOf) Match(s string) bool {
	for _, m := range e.ms {
		if !m.Match(s) {
			return false
		}
	}
	return true
}

func (e EveryOf) String() string {
	return fmt.Sprintf("<every_of:[%s]>", e.ms)
}

type IndexedEveryOf struct {
	EveryOf
	ms []MatchIndexer
}

func (e IndexedEveryOf) Index(s string) (int, []int) {
	var index int
	var offset int

	// make `in` with cap as len(s),
	// cause it is the maximum size of output segments values
	next := acquireSegments(len(s))
	current := acquireSegments(len(s))

	sub := s
	for i, m := range e.ms {
		idx, seg := m.Index(sub)
		if idx == -1 {
			releaseSegments(next)
			releaseSegments(current)
			return -1, nil
		}

		if i == 0 {
			// we use copy here instead of `current = seg`
			// cause seg is a slice from reusable buffer `in`
			// and it could be overwritten in next iteration
			current = append(current, seg...)
		} else {
			// clear the next
			next = next[:0]

			delta := index - (idx + offset)
			for _, ex := range current {
				for _, n := range seg {
					if ex+delta == n {
						next = append(next, n)
					}
				}
			}

			if len(next) == 0 {
				releaseSegments(next)
				releaseSegments(current)
				return -1, nil
			}

			current = append(current[:0], next...)
		}

		index = idx + offset
		sub = s[index:]
		offset += idx
	}

	releaseSegments(next)

	return index, current
}

func (e IndexedEveryOf) String() string {
	return fmt.Sprintf("<indexed_every_of:[%s]>", e.ms)
}

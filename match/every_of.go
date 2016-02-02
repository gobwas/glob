package match

import (
	"fmt"
)

type EveryOf struct {
	Matchers Matchers
}

func (self *EveryOf) Add(m Matcher) error {
	self.Matchers = append(self.Matchers, m)
	return nil
}

func (self EveryOf) Len() (l int) {
	for _, m := range self.Matchers {
		if ml := m.Len(); l > 0 {
			l += ml
		} else {
			return -1
		}
	}

	return
}

func max(a, b int) int {
	if a >= b {
		return a
	}

	return b
}

func (self EveryOf) Index(s string, out []int) (int, []int) {
	var index int
	var offset int
	var current []int

	sub := s
	for i, m := range self.Matchers {
		in := acquireSegments(len(sub))
		idx, seg := m.Index(sub, in)
		if idx == -1 {
			releaseSegments(in)
			if cap(current) > 0 {
				releaseSegments(current)
			}
			return -1, nil
		}

		next := acquireSegments(max(len(seg), len(current)))
		if i == 0 {
			next = append(next, seg...)
		} else {
			delta := index - (idx + offset)
			for _, ex := range current {
				for _, n := range seg {
					if ex+delta == n {
						next = append(next, n)
					}
				}
			}
		}

		if cap(current) > 0 {
			releaseSegments(current)
		}
		releaseSegments(in)

		if len(next) == 0 {
			releaseSegments(next)
			return -1, nil
		}

		current = next

		index = idx + offset
		sub = s[index:]
		offset += idx
	}

	out = append(out, current...)
	releaseSegments(current)

	return index, out
}

func (self EveryOf) Match(s string) bool {
	for _, m := range self.Matchers {
		if !m.Match(s) {
			return false
		}
	}

	return true
}

func (self EveryOf) String() string {
	return fmt.Sprintf("<every_of:[%s]>", self.Matchers)
}

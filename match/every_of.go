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

func (self EveryOf) Index(s string, out []int) (int, []int) {
	var index int
	var offset int

	// make `in` with cap as len(s),
	// cause it is the maximum size of output segments values
	//	seg := acquireSegments(len(s))
	//	next := acquireSegments(len(s))
	//	current := acquireSegments(len(s))
	//	defer func() {
	//		releaseSegments(seg)
	//		releaseSegments(next)
	//		releaseSegments(current)
	//	}()
	var (
		seg     []int
		next    []int
		current []int
	)

	sub := s
	for i, m := range self.Matchers {
		var idx int
		idx, seg = m.Index(sub, seg[:0])
		if idx == -1 {
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
				return -1, nil
			}

			current = append(current[:0], next...)
		}

		index = idx + offset
		sub = s[index:]
		offset += idx
	}

	// copy result in `out` to prevent
	// allocation `current` on heap
	out = append(out, current...)

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

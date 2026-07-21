package match

import "fmt"

type AnyOf struct {
	Matchers Matchers
}

func NewAnyOf(m ...Matcher) AnyOf {
	return AnyOf{Matchers(m)}
}

func (self *AnyOf) Add(m Matcher) error {
	self.Matchers = append(self.Matchers, m)
	return nil
}

func (self AnyOf) Match(s string) bool {
	for _, m := range self.Matchers {
		if m.Match(s) {
			return true
		}
	}

	return false
}

func (self AnyOf) Index(s string) (int, []int) {
	index := -1

	segments := acquireSegments(len(s))
	for _, m := range self.Matchers {
		idx, seg := m.Index(s)
		if idx == -1 {
			continue
		}

		if index == -1 || idx < index {
			index = idx
			segments = append(segments[:0], seg...)
			continue
		}

		if idx > index {
			continue
		}

		// here idx == index
		segments = appendMerge(segments, seg)
	}

	if index == -1 {
		releaseSegments(segments)
		return -1, nil
	}

	return index, segments
}

func (self AnyOf) Len() (l int) {
	if len(self.Matchers) == 0 {
		return 0
	}

	// Use an explicit "seen" flag. Previously l started at -1, which is also
	// the variable-length sentinel, so a leading variable-length alternative
	// (e.g. Suffix) was treated as "unset" and overwritten by a later fixed
	// length. That made "{**/x,y}" report a fixed length and get compiled into
	// a Row that can never match longer prefixes.
	l = self.Matchers[0].Len()
	for _, m := range self.Matchers[1:] {
		ml := m.Len()
		if l == -1 || ml == -1 || l != ml {
			return -1
		}
	}
	return l
}

func (self AnyOf) String() string {
	return fmt.Sprintf("<any_of:[%s]>", self.Matchers)
}

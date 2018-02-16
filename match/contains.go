package match

import (
	"fmt"
	"strings"
)

type Contains struct {
	s   string
	not bool
}

func NewContains(needle string) Contains {
	return Contains{needle, false}
}

func (c Contains) Match(s string) bool {
	return strings.Contains(s, c.s) != c.not
}

func (c Contains) Index(s string) (int, []int) {
	var offset int

	idx := strings.Index(s, c.s)

	if !c.not {
		if idx == -1 {
			return -1, nil
		}

		offset = idx + len(c.s)
		if len(s) <= offset {
			return 0, []int{offset}
		}
		s = s[offset:]
	} else if idx != -1 {
		s = s[:idx]
	}

	segments := acquireSegments(len(s) + 1)
	for i := range s {
		segments = append(segments, offset+i)
	}

	return 0, append(segments, offset+len(s))
}

func (c Contains) MinLen() int {
	return 0
}

func (c Contains) String() string {
	var not string
	if c.not {
		not = "!"
	}
	return fmt.Sprintf("<contains:%s[%s]>", not, c.s)
}

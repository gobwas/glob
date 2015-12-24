package match

import (
	"strings"
	"fmt"
)



// composite
type Composite struct {
	Chunks []Matcher
}


func (self Composite) Kind() Kind {
	return KindComposite
}

func (self Composite) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func m(chunks []Matcher, s string) bool {
	var prev Matcher
	for _, c := range chunks {
		if c.Kind() == KindRaw {
			i, l, ok := c.Search(s)
			if !ok {
				return false
			}

			if prev != nil {
				if !prev.Match(s[:i]) {
					return false
				}

				prev = nil
			}

			s = s[i+l:]
			continue
		}

		prev = c
	}

	if prev != nil {
		return prev.Match(s)
	}

	return len(s) == 0
}

func (self Composite) Match(s string) bool {
	return m(self.Chunks, s)
}

func (self Composite) String() string {
	var l []string
	for _, c := range self.Chunks {
		l = append(l, fmt.Sprint(c))
	}

	return fmt.Sprintf("[composite:%s]", strings.Join(l, ","))
}

package glob

import (
	"strings"
	"fmt"
)

const (
	Any = "*"
	SuperAny = "**"
	SingleAny = "?"
)

var Chars = []string{Any, SuperAny, SingleAny}

type Matcher interface {
	Match(string) bool
}

func firstIndexOfChars(p string, any []string) (min int, c string) {
	l := len(p)
	min = l
	weight := 0

	for _, s := range any {
		w := len(s)
		i := strings.Index(p, s)
		if i != -1 && i <= min && w > weight {
			min = i
			weight = w
			c = s
		}
	}

	if min == l {
		return -1, ""
	}

	return
}

func parse(p string, m []Matcher, d []string) ([]Matcher, error) {
	if len(p) == 0 {
		return m, nil
	}

	i, c := firstIndexOfChars(p, Chars)
	if i == -1 {
		return append(m, raw{p}), nil
	}

	if i > 0 {
		m = append(m, raw{p[0:i]})
	}

	switch c {
	case SuperAny:
		m = append(m, multiple{})
	case Any:
		m = append(m, multiple{d})
	case SingleAny:
		m = append(m, single{d})
	}

	return parse(p[i+len(c):], m, d)
}

func New(pattern string, d ...string) (Matcher, error) {
	chunks, err := parse(pattern, nil, d)
	if err != nil {
		return nil, err
	}

	if len(chunks) == 1 {
		return chunks[0], nil
	}

	return &composite{chunks}, nil
}

type raw struct {
	s string
}
func (self raw) Match(s string) bool {
	return self.s == s
}
func (self raw) String() string {
	return fmt.Sprintf("[raw:%s]", self.s)
}

type multiple struct {
	delimiters []string
}
func (self multiple) Match(s string) bool {
	i, _ := firstIndexOfChars(s, self.delimiters)
	return i == -1
}
func (self multiple) String() string {
	return fmt.Sprintf("[multiple:%s]", self.delimiters)
}

type single struct {
	delimiters []string
}
func (self single) Match(s string) bool {
	if len(s) != 1 {
		return false
	}

	i, _ := firstIndexOfChars(s, self.delimiters)

	return i == -1
}
func (self single) String() string {
	return fmt.Sprintf("[single:%s]", self.delimiters)
}

type composite struct {
	chunks []Matcher
}


func (self composite) Match(m string) bool {
	var prev Matcher

	for _, c := range self.chunks {
		if str, ok := c.(raw); ok {
			i := strings.Index(m, str.s)
			if i == -1 {
				return false
			}

			l := len(str.s)

			if prev != nil {
				if !prev.Match(m[:i]) {
					return false
				}

				prev = nil
			}

			m = m[i+l:]
			continue
		}

		prev = c
	}

	if prev != nil {
		return prev.Match(m)
	}

	return len(m) == 0
}

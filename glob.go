package glob

import (
	"strings"
	"fmt"
)

const (
	any       = `*`
	superAny  = `**`
	singleAny = `?`
	escape    = `\`
)

var chars = []string{any, superAny, singleAny, escape}

// Glob represents compiled glob pattern.
type Glob interface {
	Match(string) bool
}

// New creates Glob for given pattern and uses other given (if any) strings as separators.
// The pattern syntax is:
//
//	pattern:
//		{ term }
//	term:
//		`*`         matches any sequence of non-separator characters
//		`**`        matches any sequence of characters
//		`?`         matches any single non-separator character
//		c           matches character c (c != `*`, `**`, `?`, `\`)
//		`\` c       matches character c
func New(pattern string, d ...string) Glob {
	chunks := parse(pattern, nil, strings.Join(d, ""), false)

	if len(chunks) == 1 {
		return chunks[0]
	}

	return &composite{chunks}
}

func parse(p string, m []Glob, d string, esc bool) []Glob {
	var e bool

	if len(p) == 0 {
		return m
	}

	i, c := firstIndexOfChars(p, chars)
	if i == -1 {
		return append(m, raw{p})
	}

	if i > 0 {
		m = append(m, raw{p[0:i]})
	}

	if esc {
		m = append(m, raw{c})
	} else {
		switch c {
		case escape:
			e = true
		case superAny:
			m = append(m, multiple{})
		case any:
			m = append(m, multiple{d})
		case singleAny:
			m = append(m, single{d})
		}
	}

	return parse(p[i+len(c):], m, d, e)
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
	separators string
}

func (self multiple) Match(s string) bool {
	return strings.IndexAny(s, self.separators) == -1
}

func (self multiple) String() string {
	return fmt.Sprintf("[multiple:%s]", self.separators)
}

type single struct {
	separators string
}

func (self single) Match(s string) bool {
	return len(s) == 1 && strings.IndexAny(s, self.separators) == -1
}

func (self single) String() string {
	return fmt.Sprintf("[single:%s]", self.separators)
}

type composite struct {
	chunks []Glob
}


func (self composite) Match(m string) bool {
	var prev Glob

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

func firstIndexOfChars(p string, any []string) (min int, c string) {
	l := len(p)
	min = l
	weight := 0

	for _, s := range any {
		w := len(s)
		i := strings.Index(p, s)
		if i != -1 && i <= min && w >= weight {
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
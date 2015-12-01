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

type globKind int
const(
	glob_raw globKind = iota
	glob_multiple_separated
	glob_multiple_super
	glob_single
	glob_composite
	glob_prefix
	glob_suffix
	glob_prefix_suffix
)

// Glob represents compiled glob pattern.
type Glob interface {
	Match(string) bool
	search(string) (int, int, bool)
	kind() globKind
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
func New(pattern string, separators ...string) Glob {
	chunks := parse(pattern, nil, strings.Join(separators, ""), false)

	switch len(chunks) {
	case 1:
		return chunks[0].glob
	case 2:
		if chunks[0].glob.kind() == glob_raw && chunks[1].glob.kind() == glob_multiple_super {
			return &prefix{chunks[0].str}
		}
		if chunks[1].glob.kind() == glob_raw && chunks[0].glob.kind() == glob_multiple_super {
			return &suffix{chunks[1].str}
		}
	case 3:
		if chunks[0].glob.kind() == glob_raw && chunks[1].glob.kind() == glob_multiple_super && chunks[2].glob.kind() == glob_raw {
			return &prefix_suffix{chunks[0].str, chunks[2].str}
		}
	}

	var c []Glob
	for _, chunk := range chunks {
		c = append(c, chunk.glob)
	}

	return &composite{c}
}

type token struct {
	glob Glob
	str string
}

func parse(p string, m []token, d string, esc bool) []token {
	var e bool

	if len(p) == 0 {
		return m
	}

	i, c := firstIndexOfChars(p, chars)
	if i == -1 {
		return append(m, token{raw{p}, p})
	}

	if i > 0 {
		m = append(m, token{raw{p[0:i]}, p[0:i]})
	}

	if esc {
		m = append(m, token{raw{c}, c})
	} else {
		switch c {
		case escape:
			e = true
		case superAny:
			m = append(m, token{multiple{}, c})
		case any:
			m = append(m, token{multiple{d}, c})
		case singleAny:
			m = append(m, token{single{d}, c})
		}
	}

	return parse(p[i+len(c):], m, d, e)
}

// raw represents raw string to match
type raw struct {
	s string
}

func (self raw) Match(s string) bool {
	return self.s == s
}

func (self raw) kind() globKind {
	return glob_raw
}

func (self raw) search(s string) (i int, l int, ok bool) {
	index := strings.Index(s, self.s)
	if index == -1 {
		return
	}

	i = index
	l = len(self.s)
	ok = true

	return
}

func (self raw) String() string {
	return fmt.Sprintf("[raw:%s]", self.s)
}

// multiple represents *
type multiple struct {
	separators string
}

func (self multiple) Match(s string) bool {
	return strings.IndexAny(s, self.separators) == -1
}

func (self multiple) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self multiple) kind() globKind {
	if self.separators == "" {
		return glob_multiple_super
	} else {
		return glob_multiple_separated
	}
}

func (self multiple) String() string {
	return fmt.Sprintf("[multiple:%s]", self.separators)
}

// single represents ?
type single struct {
	separators string
}

func (self single) Match(s string) bool {
	return len(s) == 1 && strings.IndexAny(s, self.separators) == -1
}

func (self single) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, 1, true
	}

	return
}

func (self single) kind() globKind {
	return glob_single
}


func (self single) String() string {
	return fmt.Sprintf("[single:%s]", self.separators)
}


// composite
type composite struct {
	chunks []Glob
}


func (self composite) kind() globKind {
	return glob_composite
}

func (self composite) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func m(chunks []Glob, s string) bool {
	var prev Glob
	for _, c := range chunks {
		if c.kind() == glob_raw {
			i, l, ok := c.search(s)
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

func (self composite) Match(s string) bool {
	return m(self.chunks, s)
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

type prefix struct {
	s string
}

func (self prefix) kind() globKind {
	return glob_prefix
}

func (self prefix) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self prefix) Match(s string) bool {
	return strings.HasPrefix(s, self.s)
}

type suffix struct {
	s string
}

func (self suffix) kind() globKind {
	return glob_suffix
}

func (self suffix) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self suffix) Match(s string) bool {
	return strings.HasSuffix(s, self.s)
}

type prefix_suffix struct {
	p, s string
}

func (self prefix_suffix) kind() globKind {
	return glob_prefix_suffix
}

func (self prefix_suffix) search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self prefix_suffix) Match(s string) bool {
	return strings.HasPrefix(s, self.p) && strings.HasSuffix(s, self.s)
}




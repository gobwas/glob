package glob

import (
	"strings"
	"errors"
	"github.com/gobwas/glob/match"
)

const (
	any         = '*'
	single = '?'
	escape      = '\\'
	range_open  = '['
	range_close = ']'
)

const (
	inside_range_not = '!'
	inside_range_minus = '-'
)

var syntaxPhrases = string([]byte{any, single, escape, range_open, range_close})

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
func New(pattern string, separators ...string) (Glob, error) {
	chunks, err := parse(pattern, strings.Join(separators, ""), state{})
	if err != nil {
		return nil, err
	}

	switch len(chunks) {
	case 1:
		return chunks[0].matcher, nil
	case 2:
		if chunks[0].matcher.Kind() == match.KindRaw && chunks[1].matcher.Kind() == match.KindMultipleSuper {
			return &match.Prefix{chunks[0].str}, nil
		}
		if chunks[1].matcher.Kind() == match.KindRaw && chunks[0].matcher.Kind() == match.KindMultipleSuper {
			return &match.Suffix{chunks[1].str}, nil
		}
	case 3:
		if chunks[0].matcher.Kind() == match.KindRaw && chunks[1].matcher.Kind() == match.KindMultipleSuper && chunks[2].matcher.Kind() == match.KindRaw {
			return &match.PrefixSuffix{chunks[0].str, chunks[2].str}, nil
		}
	}

	var c []match.Matcher
	for _, chunk := range chunks {
		c = append(c, chunk.matcher)
	}

	return &match.Composite{c}, nil
}


// parse parsed given pattern into list of tokens
func parse(str string, sep string, st state) ([]token, error) {
	if len(str) == 0 {
		return st.tokens, nil
	}

	// if there are no syntax symbols - pattern is simple string
	i := strings.IndexAny(str, syntaxPhrases)
	if i == -1 {
		return append(st.tokens, token{match.Raw{str}, str}), nil
	}

	c := string(str[i])

	// if syntax symbol is not at the start of pattern - add raw part before it
	if i > 0 {
		st.tokens = append(st.tokens, token{match.Raw{str[0:i]}, str[0:i]})
	}

	// if we are in escape state
	if st.escape {
		st.tokens = append(st.tokens, token{match.Raw{c}, c})
		st.escape = false
	} else {
		switch str[i] {
		case range_open:
			closed := indexByteNonEscaped(str, range_close, escape, 0)
			if closed == -1 {
				return nil, errors.New("invalid format")
			}

			r := str[i+1:closed]
			g, err := parseRange(r)
			if err != nil {
				return nil, err
			}
			st.tokens = append(st.tokens, token{g, r})

			if closed == len(str) -1 {
				return st.tokens, nil
			}

			return parse(str[closed+1:], sep, st)

		case escape:
			st.escape = true
		case any:
			if len(str) > i+1 && str[i+1] == any {
				st.tokens = append(st.tokens, token{match.Multiple{}, c})
				return parse(str[i+len(c)+1:], sep, st)
			}

			st.tokens = append(st.tokens, token{match.Multiple{sep}, c})
		case single:
			st.tokens = append(st.tokens, token{match.Single{sep}, c})
		}
	}

	return parse(str[i+len(c):], sep, st)
}


func parseRange(def string) (match.Matcher, error) {
	var (
		not   bool
		esc   bool
		minus bool
		b   []byte
	)

	for i, c := range []byte(def) {
		if esc {
			b = append(b, c)
			esc = false
			continue
		}

		switch c{
		case inside_range_not:
			if i == 0 {
				not = true
			}
		case escape:
			if i == len(def) - 1 {
				return nil, errors.New("escape character without follower")
			}

			esc = true
		case inside_range_minus:
			minus = true
		default:
			b = append(b, c)
		}
	}

	def = string(b)

	if minus  {
		r := []rune(def)
		if len(r) != 3 || r[1] != inside_range_minus {
			return nil, errors.New("invalid range syntax")
		}

		return &match.Between{r[0], r[2], not}, nil
	}

	return &match.RangeList{def, not}, nil
}

type token struct {
	matcher match.Matcher
	str     string
}

type state struct {
	escape bool
	tokens []token
}

package glob

import (
	"github.com/gobwas/glob/syntax"
)

// Compile creates a new pattern instance for the given glob pattern.
// If given, additional separator characters may be set for the instance.
//
// The pattern syntax is:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
func Compile(pattern string, separators ...rune) (*Pattern, error) {
	return compile(pattern, separators)
}

// MustCompile is the same as Compile, except that if Compile returns error,
// this will panic.
func MustCompile(pattern string, separators ...rune) *Pattern {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

// QuoteMeta returns a copy of the s having all glob meta characters escaped.
func QuoteMeta(s string) string {
	// 2 is a pessimistic way of allocating an extra byte per each byte in s.
	b := make([]byte, 2*len(s))
	j := 0
	// A byte loop is correct here because all meta characters are ASCII.
	for i := 0; i < len(s); i++ {
		if syntax.IsSpecial(s[i]) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[0:j])
}

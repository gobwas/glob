package glob

import "strings"

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
func Compile(pattern string, separators ...string) (Glob, error) {
	ast, err := parse(newLexer(pattern))
	if err != nil {
		return nil, err
	}

	matcher, err := compile(ast, strings.Join(separators, ""))
	if err != nil {
		return nil, err
	}

	return matcher, nil
}

func MustCompile(pattern string, separators ...string) Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}

	return g
}

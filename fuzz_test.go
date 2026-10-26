package glob

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzMatchRegexp verifies Pattern.Match() against the regexp package.
//
// The fuzzer generates arbitrary (pattern, string, separator) inputs; each
// compilable pattern is translated into an equivalent anchored regular
// expression by translateGlob() -- a naive, independent re-implementation
// of the glob syntax -- and the match results are compared. Since the
// translator shares no code with Compile(), this validates the lexer, the
// parser, the compile-time rewrites and the matching engine all at once.
//
// Run with:
//
//	go test -fuzz=FuzzMatchRegexp
//
// Without -fuzz only the seed corpus is run, as a regression test.
func FuzzMatchRegexp(f *testing.F) {
	for _, seed := range []struct {
		pattern string
		fixture string
		dotSep  bool
	}{
		{
			pattern: "",
			fixture: "",
		},
		{
			pattern: "abc",
			fixture: "abc",
		},
		{
			pattern: "a*c",
			fixture: "abbbc",
		},
		{
			pattern: "a*a*a*a*b",
			fixture: "aaaaaaab",
		},
		{
			pattern: "**a*b",
			fixture: "axc.ab",
			dotSep:  true,
		},
		{
			pattern: "*a**b",
			fixture: "xa.b",
			dotSep:  true,
		},
		{
			pattern: "{*,**,?}",
			fixture: "a.b",
			dotSep:  true,
		},
		{
			pattern: "{a,ab}c",
			fixture: "abc",
		},
		{
			pattern: "{*.google.*,yandex.*}",
			fixture: "www.google.com",
			dotSep:  true,
		},
		{
			pattern: "{https://*gobwas.com,http://exclude.gobwas.com}",
			fixture: "https://safe.gobwas.com",
		},
		{
			pattern: "[!a]*",
			fixture: "this is a test3",
		},
		{
			pattern: "[a-z][!a-x]*cat*[h][!b]*eyes*",
			fixture: "my cat has very bright eyes",
		},
		{
			pattern: `\*`,
			fixture: "*",
		},
		{
			pattern: "*ä",
			fixture: "åä",
		},
		{
			pattern: "{,a}{a,}a",
			fixture: "a",
		},
		{
			pattern: "{{a,b},c}d",
			fixture: "bd",
		},
		{
			// A star inside an alternative must not discard the restart
			// point of a star outside of it; see the starsFloor handling
			// in matchContext.storeStar().
			pattern: "*{*0,}",
			fixture: "1",
		},
		{
			// U+FFFD is a valid character, not an invalid byte: the class
			// must match it, and the pattern must compile.
			pattern: "[!a]",
			fixture: "�",
		},
		{
			pattern: "[�]*",
			fixture: "�x",
		},
		{
			pattern: "@(ab|a)c",
			fixture: "abc",
		},
		{
			pattern: "?(a*)b",
			fixture: "axxb",
		},
		{
			pattern: "*(ab|c)",
			fixture: "cabab",
		},
		{
			pattern: "+(a*b|c)",
			fixture: "abcc",
		},
		{
			pattern: "*(a?(b)|bc)",
			fixture: "abbc",
		},
		{
			// The `(` binds to the star right before it: a `*` and a `*(`,
			// not a `**` and a literal `(`.
			pattern: "**(a)",
			fixture: "x.a",
			dotSep:  true,
		},
		{
			// A literal separator inside an alternative matches itself.
			pattern: "*(a.)b",
			fixture: "a.a.b",
			dotSep:  true,
		},
	} {
		f.Add(seed.pattern, seed.fixture, seed.dotSep)
	}
	f.Fuzz(func(t *testing.T, pattern, s string, dotSep bool) {
		if len(pattern) > 32 || len(s) > 64 {
			t.Skip("too large")
		}
		if !utf8.ValidString(pattern) || !utf8.ValidString(s) {
			// The regexp package normalizes invalid UTF-8; the comparison
			// makes no sense for it.
			t.Skip("invalid utf8")
		}
		if strings.Contains(pattern, "!(") {
			// A `!(...)` matches what its alternatives do not, and RE2 has
			// no negation to translate that into. The negated groups are
			// covered by TestGroupReference, against a backtracking
			// reference matcher instead of a regular expression.
			t.Skip("negation group")
		}
		var sep []rune
		if dotSep {
			sep = []rune{'.'}
			if strings.Count(pattern, "*") > 4 {
				// Backtracking over many separator-limited stars is
				// combinatorial in the worst case; keep the fuzzing fast.
				t.Skip("too many stars")
			}
		}
		p, err := Compile(pattern, sep...)
		if err != nil {
			t.Skip("does not compile")
		}
		reStr, err := translateGlob(pattern, sep)
		if err != nil {
			t.Fatalf("Compile() accepted %#q but the translator did not: %v", pattern, err)
		}
		re, err := regexp.Compile(reStr)
		if err != nil {
			t.Fatalf("bad regexp translation of %#q: %v", pattern, err)
		}
		var (
			got  = p.Match(s)
			want = re.MatchString(s)
		)
		if got != want {
			t.Errorf(
				"sep=%q pattern=%#q (regexp %#q) string=%#q: Match()=%t, regexp=%t",
				string(sep), pattern, re, s, got, want,
			)
		}
	})
}

// translateGlob translates the glob pattern into an equivalent anchored
// regular expression. It is intentionally independent from Compile(): a naive
// rune-by-rune re-implementation of the documented syntax, sharing no code
// with the lexer or the parser, to serve as the differential oracle.
func translateGlob(pattern string, sep []rune) (string, error) {
	var sb strings.Builder
	sb.WriteString("^(?:")
	rs := []rune(pattern)
	// nesting is the stack of the constructs the translator is inside of, the
	// innermost last: `{` for a `{...}` and the operator character for an
	// extended group. The innermost one tells which characters are special:
	// `,` and `}` inside a `{...}`, `|` and `)` inside a group.
	var nesting []rune
	innermost := func() rune {
		if n := len(nesting); n > 0 {
			return nesting[n-1]
		}
		return 0
	}
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		// A `(` opens an extended group only right after one of the five
		// operator characters; anywhere else it is a plain character.
		if i+1 < len(rs) && rs[i+1] == '(' && strings.ContainsRune("?*+@!", r) {
			if r == '!' {
				return "", errNegationGroup
			}
			nesting = append(nesting, r)
			sb.WriteString(`(?:`)
			i++
			continue
		}
		switch r {
		case '\\':
			// A trailing backslash does not compile, so it never gets here.
			if i+1 < len(rs) {
				i++
				sb.WriteString(regexp.QuoteMeta(string(rs[i])))
			}

		case '*':
			// The `(` binds to the `*` right before it, so `**(` is a `*`
			// followed by the `*(` group opener, not a `**` and a `(`.
			if i+1 < len(rs) && rs[i+1] == '*' && (i+2 >= len(rs) || rs[i+2] != '(') {
				i++
				sb.WriteString(`(?s:.*)`)
			} else {
				sb.WriteString(sepClassRegexp(sep) + `*`)
			}

		case '?':
			sb.WriteString(sepClassRegexp(sep))

		case '{':
			nesting = append(nesting, '{')
			sb.WriteString(`(?:`)

		case '}':
			if innermost() != '{' {
				// A literal outside of braces.
				sb.WriteString(regexp.QuoteMeta(`}`))
				break
			}
			nesting = nesting[:len(nesting)-1]
			sb.WriteString(`)`)

		case ',':
			if innermost() != '{' {
				// A literal outside of braces.
				sb.WriteString(`,`)
				break
			}
			sb.WriteString(`|`)

		case '|':
			if op := innermost(); op == 0 || op == '{' {
				// A literal outside of a group.
				sb.WriteString(regexp.QuoteMeta(`|`))
				break
			}
			sb.WriteString(`|`)

		case ')':
			op := innermost()
			if op == 0 || op == '{' {
				// A literal outside of a group.
				sb.WriteString(regexp.QuoteMeta(`)`))
				break
			}
			nesting = nesting[:len(nesting)-1]
			sb.WriteString(`)`)
			// `@(...)` is the bare alternation; the other operators are it
			// repeated, which is exactly what the regexp quantifiers do.
			switch op {
			case '?':
				sb.WriteString(`?`)
			case '*':
				sb.WriteString(`*`)
			case '+':
				sb.WriteString(`+`)
			}

		case '[':
			n, err := translateClass(&sb, rs[i:])
			if err != nil {
				return "", err
			}
			i += n - 1

		default:
			sb.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	if n := len(nesting); n > 0 {
		if nesting[n-1] == '{' {
			return "", errors.New("unclosed `{`")
		}
		return "", errors.New("unclosed `(`")
	}
	sb.WriteString(`)$`)
	return sb.String(), nil
}

// errNegationGroup is returned for the patterns translateGlob() can not
// handle; see the skip in FuzzMatchRegexp.
var errNegationGroup = errors.New("`!(...)` has no regexp equivalent")

// translateClass translates the `[...]` character class rs begins with and
// returns the number of runes consumed. Mirroring the syntax, a class is
// either a range `[a-z]` or a set `[abc]`, both optionally negated with the
// leading `!`; the range boundary characters are taken verbatim, the set
// characters may be escaped.
func translateClass(sb *strings.Builder, rs []rune) (int, error) {
	i := 1 // Skip the `[`.
	not := ""
	if i < len(rs) && rs[i] == '!' {
		not = "^"
		i++
	}
	if i+1 < len(rs) && rs[i+1] == '-' {
		// A range.
		if i+3 >= len(rs) || rs[i+3] != ']' {
			return 0, errors.New("unclosed range")
		}
		lo, hi := rs[i], rs[i+2]
		if hi < lo {
			return 0, errors.New("range hi character is less than lo")
		}
		fmt.Fprintf(sb, `[%s\x{%x}-\x{%x}]`, not, lo, hi)
		return i + 4, nil
	}
	// A set.
	var chars []rune
	escaped := false
	for ; i < len(rs); i++ {
		r := rs[i]
		switch {
		case escaped:
			chars = append(chars, r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ']':
			if len(chars) == 0 {
				return 0, errors.New("empty class")
			}
			sb.WriteString("[" + not)
			for _, c := range chars {
				fmt.Fprintf(sb, `\x{%x}`, c)
			}
			sb.WriteString("]")
			return i + 1, nil
		default:
			chars = append(chars, r)
		}
	}
	return 0, errors.New("unclosed `[`")
}

// sepClassRegexp returns a class matching any single rune except the
// separators, e.g. `[^\x{2e}]` -- or the match-all when there are none.
func sepClassRegexp(sep []rune) string {
	if len(sep) == 0 {
		return `(?s:.)`
	}
	var sb strings.Builder
	sb.WriteString("[^")
	for _, r := range sep {
		fmt.Fprintf(&sb, `\x{%x}`, r)
	}
	sb.WriteString("]")
	return sb.String()
}

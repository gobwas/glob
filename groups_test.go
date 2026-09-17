package glob

import (
	"errors"
	"fmt"
	"testing"
)

var sepDot = []rune{'.'}

type groupCase struct {
	sep []rune
	pat string
	str string
	exp bool
}

func testGroupCases(t *testing.T, cases []groupCase) {
	t.Helper()
	for _, c := range cases {
		g, err := Compile(c.pat, c.sep...)
		if err != nil {
			t.Errorf("Compile(%q, sep=%q): %v", c.pat, string(c.sep), err)
			continue
		}
		if got := g.Match(c.str); got != c.exp {
			t.Errorf(
				"sep=%q pattern=%q string=%q: Match() = %t; want %t",
				string(c.sep), c.pat, c.str, got, c.exp,
			)
		}
	}
}

// TestGroupExactlyOne covers `@(...)`, which matches a sub-string exactly one
// of its alternatives matches.
func TestGroupExactlyOne(t *testing.T) {
	testGroupCases(t, []groupCase{
		{pat: "@(a|b)", str: "a", exp: true},
		{pat: "@(a|b)", str: "b", exp: true},
		{pat: "@(a|b)", str: "ab", exp: false},
		{pat: "@(a|b)", str: "", exp: false},
		{pat: "@(a|b)", str: "@(a|b)", exp: false},
		{pat: "@(ab|a)c", str: "ac", exp: true},
		{pat: "@(ab|a)c", str: "abc", exp: true},
		{pat: "@(ab|a)c", str: "c", exp: false},
		{pat: "x@(a|b)y", str: "xay", exp: true},
		{pat: "x@(a|b)y", str: "xy", exp: false},
		{pat: "@(a|b)@(c|d)", str: "bd", exp: true},
		{pat: "@(a|b)@(c|d)", str: "bc", exp: true},
		{pat: "@(a|b)@(c|d)", str: "ba", exp: false},
		{pat: "@(a)", str: "a", exp: true},
		{pat: "@(a)", str: "aa", exp: false},
		{pat: "@()", str: "", exp: true},
		{pat: "@()", str: "a", exp: false},
		{pat: "@(a|)", str: "", exp: true},
		{pat: "@(a|)", str: "a", exp: true},
		{pat: "@(a|)", str: "aa", exp: false},
		{pat: "@(*|b)c", str: "xxc", exp: true},
		{pat: "@(*|b)c", str: "c", exp: true},
		{pat: "@(a*|b)c", str: "ac", exp: true},
		{pat: "@(a*|b)c", str: "axxc", exp: true},
		{pat: "@(a*|b)c", str: "abc", exp: true},
		{pat: "@(a*|b)c", str: "bc", exp: true},
		{pat: "@(a*|b)c", str: "c", exp: false},
		{pat: "@(?)", str: "a", exp: true},
		{pat: "@(?)", str: "", exp: false},
		{pat: "@(?)", str: "ab", exp: false},
		{pat: "@([ab]|c)", str: "a", exp: true},
		{pat: "@([ab]|c)", str: "c", exp: true},
		{pat: "@([ab]|c)", str: "d", exp: false},
		{pat: "@([!a]|a)", str: "a", exp: true},
		{pat: "@([!a]|a)", str: "b", exp: true},
		{pat: "@([!a]|a)", str: "bb", exp: false},
		{pat: "@(a|@(b|c))", str: "c", exp: true},
		{pat: "@(a|@(b|c))", str: "bc", exp: false},
		{pat: "@(a|{b,c})", str: "c", exp: true},
		{pat: "@(a|{b,c})", str: "a", exp: true},
		{pat: "{@(a|b),c}", str: "b", exp: true},
		{pat: "{@(a|b),c}", str: "c", exp: true},
		{pat: "{@(a|b),c}", str: "d", exp: false},
		{pat: "@(a|b)*", str: "axyz", exp: true},
		{pat: "@(a|b)*", str: "xyz", exp: false},
		{pat: "*@(a|b)", str: "xyzb", exp: true},
		{pat: "*@(a|b)", str: "xyz", exp: false},
		{pat: "*@(a|)", str: "xyz", exp: true},
		{pat: "*@(|b)", str: "a", exp: true},
	})
}

// TestGroupOptional covers `?(...)`, which matches what `@(...)` matches plus
// the empty sub-string.
func TestGroupOptional(t *testing.T) {
	testGroupCases(t, []groupCase{
		{pat: "?(a)b", str: "b", exp: true},
		{pat: "?(a)b", str: "ab", exp: true},
		{pat: "?(a)b", str: "aab", exp: false},
		{pat: "?(a)b", str: "?(a)b", exp: false},
		{pat: "?(ab|c)d", str: "d", exp: true},
		{pat: "?(ab|c)d", str: "abd", exp: true},
		{pat: "?(ab|c)d", str: "cd", exp: true},
		{pat: "?(ab|c)d", str: "abcd", exp: false},
		{pat: "?()", str: "", exp: true},
		{pat: "?()", str: "a", exp: false},
		{pat: "x?(a)?(b)y", str: "xy", exp: true},
		{pat: "x?(a)?(b)y", str: "xay", exp: true},
		{pat: "x?(a)?(b)y", str: "xby", exp: true},
		{pat: "x?(a)?(b)y", str: "xaby", exp: true},
		{pat: "x?(a)?(b)y", str: "xbay", exp: false},
		{pat: "?(a|b|c)", str: "", exp: true},
		{pat: "?(a|b|c)", str: "b", exp: true},
		{pat: "?(a|b|c)", str: "ab", exp: false},
		{pat: "?(a*)b", str: "b", exp: true},
		{pat: "?(a*)b", str: "ab", exp: true},
		{pat: "?(a*)b", str: "axxb", exp: true},
		{pat: "?(a*)b", str: "xb", exp: false},
		{pat: "??(a)", str: "x", exp: true},
		{pat: "??(a)", str: "xa", exp: true},
		{pat: "??(a)", str: "a", exp: true},
		{pat: "??(a)", str: "", exp: false},
		{pat: "??(a)", str: "xaa", exp: false},
		{pat: "*?(b)", str: "a", exp: true},
		{pat: "?(a)?(a)b", str: "aab", exp: true},
		{pat: "?(a)?(a)b", str: "b", exp: true},
		{pat: "?(a)?(a)b", str: "aaab", exp: false},
	})
}

// TestGroupRepeat covers `*(...)` and `+(...)`, which match a sequence of
// sub-strings, each matched by some alternative -- not necessarily the same
// one every time.
func TestGroupRepeat(t *testing.T) {
	testGroupCases(t, []groupCase{
		{pat: "*(a)", str: "", exp: true},
		{pat: "*(a)", str: "a", exp: true},
		{pat: "*(a)", str: "aaa", exp: true},
		{pat: "*(a)", str: "b", exp: false},
		{pat: "*(a)", str: "aab", exp: false},
		{pat: "*(a)", str: "*(a)", exp: false},
		{pat: "+(a)", str: "", exp: false},
		{pat: "+(a)", str: "a", exp: true},
		{pat: "+(a)", str: "aaa", exp: true},
		{pat: "+(a)", str: "+(a)", exp: false},
		{pat: "*(ab|c)", str: "", exp: true},
		{pat: "*(ab|c)", str: "abc", exp: true},
		{pat: "*(ab|c)", str: "cab", exp: true},
		{pat: "*(ab|c)", str: "abab", exp: true},
		{pat: "*(ab|c)", str: "aba", exp: false},
		{pat: "+(ab|c)d", str: "d", exp: false},
		{pat: "+(ab|c)d", str: "abd", exp: true},
		{pat: "+(ab|c)d", str: "cabd", exp: true},
		{pat: "*(ab)", str: "", exp: true},
		{pat: "*(ab)", str: "a", exp: false},
		{pat: "x*(ab)", str: "x", exp: true},
		{pat: "x*(ab)", str: "xabab", exp: true},
		{pat: "x*(ab)", str: "xaba", exp: false},
		{pat: "*()", str: "", exp: true},
		{pat: "*()", str: "a", exp: false},
		{pat: "+()", str: "", exp: true},
		{pat: "+()", str: "a", exp: false},
		{pat: "*(a|)", str: "", exp: true},
		{pat: "*(a|)", str: "aa", exp: true},
		{pat: "+(|a)", str: "", exp: true},
		{pat: "+(|a)", str: "a", exp: true},
		{pat: "*(a*)b", str: "b", exp: true},
		{pat: "*(a*)b", str: "ab", exp: true},
		{pat: "*(a*)b", str: "aab", exp: true},
		{pat: "*(a*)b", str: "axab", exp: true},
		{pat: "*(*a|b)", str: "", exp: true},
		{pat: "*(*a|b)", str: "ab", exp: true},
		{pat: "*(*a|b)", str: "ba", exp: true},
		{pat: "*(*a|b)", str: "bb", exp: true},
		{pat: "*(*a|b)", str: "xb", exp: false},
		{pat: "+(a*b|c)", str: "", exp: false},
		{pat: "+(a*b|c)", str: "c", exp: true},
		{pat: "+(a*b|c)", str: "ab", exp: true},
		{pat: "+(a*b|c)", str: "abcc", exp: true},
		{pat: "+(a*b|c)", str: "aabab", exp: true},
		{pat: "+(a*b|c)", str: "abca", exp: false},
		{pat: "*(a?(b)|bc)", str: "", exp: true},
		{pat: "*(a?(b)|bc)", str: "ab", exp: true},
		{pat: "*(a?(b)|bc)", str: "bc", exp: true},
		{pat: "*(a?(b)|bc)", str: "abc", exp: true},
		{pat: "*(a?(b)|bc)", str: "abbc", exp: true},
		{pat: "*(a?(b)|bc)", str: "ac", exp: false},
		{pat: "*(a?(b)|bc)", str: "abcb", exp: false},
		// Alternatives that match the empty string must terminate.
		{pat: "*(*)", str: "", exp: true},
		{pat: "*(*)", str: "abc", exp: true},
		{pat: "*(?(a))", str: "", exp: true},
		{pat: "*(?(a))", str: "aa", exp: true},
		{pat: "*(?(a))", str: "b", exp: false},
		{pat: "*(!(a))", str: "aa", exp: true},
		{pat: "*(!(a))", str: "a", exp: false},
		{pat: "+(a)+(a)", str: "a", exp: false},
		{pat: "+(a)+(a)", str: "aa", exp: true},
		{pat: "+(a)+(a)", str: "aaa", exp: true},
		{pat: "*(a)*(a)b", str: "aab", exp: true},
		{pat: "*(@(a|b)c)", str: "", exp: true},
		{pat: "*(@(a|b)c)", str: "ac", exp: true},
		{pat: "*(@(a|b)c)", str: "acbc", exp: true},
		{pat: "*(@(a|b)c)", str: "abc", exp: false},
		{pat: "+(ab|cb)", str: "abcb", exp: true},
		{pat: "+(ab|cb)", str: "b", exp: false},
		{pat: "+(ab|cb)", str: "", exp: false},
		{pat: "*(a)x*(b)", str: "aaxbb", exp: true},
		{pat: "*(a)x*(b)", str: "aaxbba", exp: false},
		{pat: "*(a|aa)", str: "aaaaa", exp: true},
		{pat: "+(aa)", str: "aaaaa", exp: false},
		{pat: "+(aa)", str: "aaaaaa", exp: true},
		{pat: "+(*)", str: "", exp: true},
		{pat: "**(a)", str: "aa", exp: true},
	})
}

// TestGroupNegation covers `!(...)`, which matches a sub-string -- without
// separators, like a `*` -- that none of its alternatives matches.
func TestGroupNegation(t *testing.T) {
	testGroupCases(t, []groupCase{
		{pat: "!(a)", str: "", exp: true},
		{pat: "!(a)", str: "a", exp: false},
		{pat: "!(a)", str: "b", exp: true},
		{pat: "!(a)", str: "aa", exp: true},
		{pat: "!(a)", str: "!(a)", exp: true},
		{pat: "!(a|b)", str: "a", exp: false},
		{pat: "!(a|b)", str: "b", exp: false},
		{pat: "!(a|b)", str: "c", exp: true},
		{pat: "!(a|b)", str: "ab", exp: true},
		{pat: "!(a)b", str: "ab", exp: false},
		{pat: "!(a)b", str: "b", exp: true},
		{pat: "!(a)b", str: "aab", exp: true},
		{pat: "!(a)b", str: "cb", exp: true},
		{pat: "!(a*)z", str: "z", exp: true},
		{pat: "!(a*)z", str: "bz", exp: true},
		{pat: "!(a*)z", str: "abz", exp: false},
		{pat: "!(a*)z", str: "az", exp: false},
		{pat: "!(a*)z", str: "baz", exp: true},
		{pat: "!(*a)", str: "", exp: true},
		{pat: "!(*a)", str: "a", exp: false},
		{pat: "!(*a)", str: "ba", exp: false},
		{pat: "!(*a)", str: "ab", exp: true},
		{pat: "!()", str: "", exp: false},
		{pat: "!()", str: "x", exp: true},
		{pat: "!()", str: "xy", exp: true},
		{pat: "!(!(a))", str: "a", exp: true},
		{pat: "!(!(a))", str: "b", exp: false},
		{pat: "!(!(a))", str: "", exp: false},
		{pat: "!(!(a))", str: "aa", exp: false},
		{pat: "!(?)x", str: "x", exp: true},
		{pat: "!(?)x", str: "ax", exp: false},
		{pat: "!(?)x", str: "aax", exp: true},
		{pat: "x!(a)y", str: "xy", exp: true},
		{pat: "x!(a)y", str: "xay", exp: false},
		{pat: "x!(a)y", str: "xby", exp: true},
		{pat: "x!(a)y", str: "xaay", exp: true},
		{pat: "!(a)!(a)", str: "", exp: true},
		{pat: "!(a)!(a)", str: "a", exp: false},
		{pat: "!(a)!(a)", str: "aa", exp: true},
		{pat: "!(a)!(a)", str: "b", exp: true},
		{pat: "!(**)", str: "", exp: false},
		{pat: "!(**)", str: "a", exp: false},
		{pat: "!(*)", str: "a", exp: false},
		{pat: "!(*)", str: "", exp: false},
		// The star takes the whole string and leaves the group the empty
		// sub-string, which is not "a".
		{pat: "*!(a)", str: "a", exp: true},
		{pat: "*!(a)", str: "", exp: true},
		{pat: "!(a)*", str: "a", exp: true},
		{pat: "!(ab|ba)c", str: "abc", exp: false},
		{pat: "!(ab|ba)c", str: "bac", exp: false},
		{pat: "!(ab|ba)c", str: "aac", exp: true},
		{pat: "!(ab|ba)c", str: "c", exp: true},
		{pat: "!({a,b})", str: "a", exp: false},
		{pat: "!({a,b})", str: "c", exp: true},
		{pat: "@(!(a)|a)", str: "a", exp: true},
		{pat: "?(!(a))b", str: "ab", exp: false},
		{pat: "?(!(a))b", str: "b", exp: true},
		{pat: "?(!(a))b", str: "xb", exp: true},
	})
}

// TestGroupSeparators verifies that the separators apply inside the groups
// exactly as they apply everywhere else: a `*` or a `?` in an alternative
// stops at a separator, a `**` crosses it, a literal separator matches
// itself -- and a `!(...)`, like a `*`, never spans one.
func TestGroupSeparators(t *testing.T) {
	sepSlash := []rune{'/'}
	testGroupCases(t, []groupCase{
		{sep: sepDot, pat: "@(a|b).c", str: "a.c", exp: true},
		{sep: sepDot, pat: "@(a|b).c", str: "ab.c", exp: false},
		{sep: sepDot, pat: "@(*)", str: "ab", exp: true},
		{sep: sepDot, pat: "@(*)", str: "a.b", exp: false},
		{sep: sepDot, pat: "@(**)", str: "a.b", exp: true},
		{sep: sepDot, pat: "@(a.b|c)", str: "a.b", exp: true},
		{sep: sepDot, pat: "*(a.)b", str: "a.a.b", exp: true},
		{sep: sepDot, pat: "*(a.)b", str: "a.ab", exp: false},
		{sep: sepDot, pat: "*(a.b)", str: "a.ba.b", exp: true},
		{sep: sepDot, pat: "*(a.b)", str: "a.b.", exp: false},
		{sep: sepDot, pat: "+(*.)x", str: "a.b.x", exp: true},
		{sep: sepDot, pat: "+(*.)x", str: "x", exp: false},
		{sep: sepDot, pat: "+(*.)x", str: ".x", exp: true},
		{sep: sepDot, pat: "*(*)", str: "a.b", exp: false},
		{sep: sepDot, pat: "*(**)", str: "a.b", exp: true},
		{sep: sepDot, pat: "*(?)", str: "ab", exp: true},
		{sep: sepDot, pat: "*(?)", str: "a.b", exp: false},
		{sep: sepDot, pat: "!(a)", str: "b", exp: true},
		{sep: sepDot, pat: "!(a)", str: "a.b", exp: false},
		{sep: sepDot, pat: "!(a)", str: ".", exp: false},
		{sep: sepDot, pat: "!(a).b", str: "a.b", exp: false},
		{sep: sepDot, pat: "!(a).b", str: "c.b", exp: true},
		{sep: sepDot, pat: "!(a).b", str: "aa.b", exp: true},
		{sep: sepDot, pat: "!(a).b", str: ".b", exp: true},
		{sep: sepDot, pat: "!(a).b", str: "x.y.b", exp: false},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "b/a", exp: true},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "a/b", exp: false},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "a/a", exp: false},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "b/b", exp: false},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "ab/ba", exp: true},
		{sep: sepSlash, pat: "!(a)/!(b)", str: "b/c/a", exp: false},
		{sep: sepDot, pat: "!(*)", str: "a.b", exp: false},
		{sep: sepDot, pat: "!(a.b)", str: "a.b", exp: false},
		{pat: "!(a.b)", str: "a.b", exp: false},
		{pat: "!(a.b)", str: "a.c", exp: true},
		// The `(` binds to the character right before it.
		{sep: sepDot, pat: "**(a)", str: "x.(a)", exp: false},
		{sep: sepDot, pat: "**(a)", str: "xa", exp: true},
		{sep: sepDot, pat: "***(a)", str: "x.a", exp: true},
		{sep: sepDot, pat: "**@(a|b)", str: "x.b", exp: true},
		{sep: sepDot, pat: "*@(a|b)", str: "x.b", exp: false},
		{sep: sepSlash, pat: "src/**/@(*.go|*.md)", str: "src/a/b/c.go", exp: true},
		{sep: sepSlash, pat: "src/**/@(*.go|*.md)", str: "src/a/b/c.rs", exp: false},
		{sep: sepSlash, pat: "src/**/!(*_test).go", str: "src/a/b.go", exp: true},
		{sep: sepSlash, pat: "src/**/!(*_test).go", str: "src/a/b_test.go", exp: false},
		{sep: sepSlash, pat: "src/**/!(*_test).go", str: "src/a/b_test/c.go", exp: true},
	})
}

// TestGroupLexing verifies that the group syntax is recognized exactly where
// it is meant to be: a `(` opens a group only right after one of the five
// operator characters, and the innermost open construct decides whether
// `,` `}` `|` `)` are special.
func TestGroupLexing(t *testing.T) {
	testGroupCases(t, []groupCase{
		{pat: "a|b", str: "a|b", exp: true},
		{pat: "a|b", str: "a", exp: false},
		{pat: "a)b", str: "a)b", exp: true},
		{pat: "(a)", str: "(a)", exp: true},
		{pat: "a(b|c)", str: "a(b|c)", exp: true},
		{pat: "a(b|c)", str: "ab", exp: false},
		{pat: "@@(a)", str: "@a", exp: true},
		{pat: "+@(a|b)", str: "+b", exp: true},
		{pat: "!@(a)", str: "!a", exp: true},
		{pat: "a+b", str: "a+b", exp: true},
		{pat: "a@b", str: "a@b", exp: true},
		{pat: "a!b", str: "a!b", exp: true},
		{pat: "a!", str: "a!", exp: true},
		{pat: `@(a\|b)`, str: "a|b", exp: true},
		{pat: `@(a\|b)`, str: "a", exp: false},
		{pat: `@(a\)b)`, str: "a)b", exp: true},
		{pat: "@(a(b|c)", str: "a(b", exp: true},
		{pat: "@(a(b|c)", str: "c", exp: true},
		{pat: `\@(a)`, str: "@(a)", exp: true},
		{pat: `\@(a)`, str: "a", exp: false},
		{pat: `@\(a)`, str: "@(a)", exp: true},
		{pat: `\*(a)`, str: "*(a)", exp: true},
		{pat: `*\(a)`, str: "x(a)", exp: true},
		{pat: `*\(a)`, str: "xa", exp: false},
		{pat: `?\(a)`, str: "x(a)", exp: true},
		{pat: `\!(a)`, str: "!(a)", exp: true},
		{pat: `\!(a)`, str: "b", exp: false},
		{pat: "[!(a)]", str: "b", exp: true},
		{pat: "[!(a)]", str: "(", exp: false},
		{pat: "{@(a,b),c}", str: "a,b", exp: true},
		{pat: "{@(a,b),c}", str: "c", exp: true},
		{pat: "{@(a,b),c}", str: "a", exp: false},
		{pat: "@(a|{b,c|d})", str: "c|d", exp: true},
		{pat: "@(a|{b,c|d})", str: "b", exp: true},
		{pat: "@(a|{b,c|d})", str: "c", exp: false},
		{pat: "{a,@(b|c}d)}", str: "c}d", exp: true},
		{pat: "{a,@(b|c}d)}", str: "a", exp: true},
		{pat: "@(a|b,c)", str: "b,c", exp: true},
		{pat: "**(a)", str: "", exp: true},
		{pat: "**(a)", str: "x", exp: true},
	})
}

// TestGroupQuoteMeta verifies that QuoteMeta neutralizes the group syntax:
// the quoted text matches itself, and nothing else.
func TestGroupQuoteMeta(t *testing.T) {
	for _, s := range []string{
		"@(a|b)", "!(x)", "*(a)", "+(b)", "?(c)",
		"a|b", "(a)", "a)b", "!x", `*?\[]{}()`,
	} {
		q := QuoteMeta(s)
		g, err := Compile(q)
		if err != nil {
			t.Errorf("Compile(QuoteMeta(%q) = %q): %v", s, q, err)
			continue
		}
		if !g.Match(s) {
			t.Errorf("Compile(QuoteMeta(%q) = %q).Match(%q) = false; want true", s, q, s)
		}
		if g.Match(s[1:]) {
			t.Errorf("Compile(QuoteMeta(%q) = %q).Match(%q) = true; want false", s, q, s[1:])
		}
	}
}

func TestGroupCompileErrors(t *testing.T) {
	for _, test := range []struct {
		pat    string
		offset int
		reason string
	}{
		{pat: "@(a|b", offset: 5, reason: "unclosed `(`"},
		{pat: "*(a", offset: 3, reason: "unclosed `(`"},
		{pat: "!(", offset: 2, reason: "unclosed `(`"},
		{pat: "{@(a,b}", offset: 7, reason: "unclosed `(`"},
		{pat: "a+(b|@(c)", offset: 9, reason: "unclosed `(`"},
		{pat: "@(a|{b)", offset: 7, reason: "unclosed `{`"},
		{pat: "@(a|[)", offset: 6, reason: "unexpected end of input"},
		{pat: `@(a\`, offset: 4, reason: "trailing backslash"},
	} {
		_, err := Compile(test.pat)
		var se *SyntaxError
		if !errors.As(err, &se) {
			t.Errorf("Compile(%q) = %v; want *SyntaxError", test.pat, err)
			continue
		}
		if se.Offset != test.offset || se.Reason != test.reason {
			t.Errorf(
				"Compile(%q) = (%d, %q); want (%d, %q)",
				test.pat, se.Offset, se.Reason, test.offset, test.reason,
			)
		}
	}
	// A group is not an error anywhere a term may appear.
	for _, pat := range []string{
		"@()", "!()", "*()", "+()", "?()", "@(|)", "@(a||b)", "{@(a|b)}",
		"@({a,b}|{c,d})", "[a]@(b)[c]", "@(a|b)*", "@(@(@(a)))", "!(!(!()))",
	} {
		if _, err := Compile(pat); err != nil {
			t.Errorf("Compile(%q): %v; want no error", pat, err)
		}
	}
}

// TestGroupReference compares Pattern.Match() against refMatcher over an
// exhaustive set of small group patterns and strings, with and without a
// separator. Like translateGlob() of FuzzMatchRegexp it is an independent
// re-implementation of the syntax -- needed here because RE2 has no
// negation, so `!(...)` cannot be translated into a regular expression.
func TestGroupReference(t *testing.T) {
	bodies := []string{
		"", "a", "b", "ab", "ba", "a*", "*a", "?", "*", "**",
		"[!a]", "a.b", ".", "{a,b}", "a?(b)", "!(a)", "*(a)",
	}
	var groups []string
	for _, op := range []string{"?(", "*(", "+(", "@(", "!("} {
		for i, x := range bodies {
			groups = append(groups, op+x+")")
			for _, y := range bodies[i+1:] {
				groups = append(groups, op+x+"|"+y+")")
			}
		}
	}
	small := []string{
		"a", "b", ".", "?", "*", "**", "[ab]", "[!a]", "{a,b}", "{a,}",
		"@(a|b)", "?(a)", "*(a)", "+(ab|b)", "!(a)", "!(a*)", "*(a?(b)|bc)",
		"@(a|*)", "!(?)", "*(*a|b)", "!(.)", "*(a.)", "+(**)", "!(**)",
		"?()", "*()", "+()", "@()", "!()", "!(a|b)",
	}
	tiny := []string{"a", "*", "?", "@(a|b)", "*(a)", "!(a)", "+(b)", "?(a)", "**", "."}

	var pats []string
	pats = append(pats, groups...)
	pats = append(pats, concatAll(small, 2)...)
	pats = append(pats, concatAll(tiny, 3)...)
	strs := allStrings("ab.", 4)

	seen := map[string]bool{}
	var patterns, comparisons, diffs int
	for _, p := range pats {
		if seen[p] {
			continue
		}
		seen[p] = true
		patterns++
		for _, sep := range [][]rune{nil, sepDot} {
			g, err := Compile(p, sep...)
			want, referr := refParse(p)
			if (err != nil) != (referr != nil) {
				t.Errorf("Compile(%q): err=%v, while the reference says err=%v", p, err, referr)
				continue
			}
			if err != nil {
				continue
			}
			m := refMatcher{sep: sep}
			for _, s := range strs {
				comparisons++
				got := g.Match(s)
				exp := m.match(want, []rune(s))
				if got != exp {
					diffs++
					if diffs <= 50 {
						t.Errorf(
							"sep=%q pattern=%q string=%q: Match() = %t; want %t",
							string(sep), p, s, got, exp,
						)
					}
				}
			}
		}
	}
	if diffs > 50 {
		t.Errorf("%d differences in total, the first 50 reported", diffs)
	}
	t.Logf("patterns=%d comparisons=%d diffs=%d", patterns, comparisons, diffs)
}

// allStrings returns every string of up to maxLen characters of the alphabet,
// the empty one included.
func allStrings(alphabet string, maxLen int) []string {
	out := []string{""}
	prev := []string{""}
	for l := 1; l <= maxLen; l++ {
		var cur []string
		for _, p := range prev {
			for _, c := range alphabet {
				cur = append(cur, p+string(c))
			}
		}
		out = append(out, cur...)
		prev = cur
	}
	return out
}

// concatAll returns every concatenation of up to n of the atoms.
func concatAll(atoms []string, n int) []string {
	out := []string{""}
	prev := []string{""}
	for l := 1; l <= n; l++ {
		var cur []string
		for _, p := range prev {
			for _, a := range atoms {
				cur = append(cur, p+a)
			}
		}
		out = append(out, cur...)
		prev = cur
	}
	return out
}

// A naive, independent re-implementation of the pattern syntax with the
// extended groups, used as the oracle of TestGroupReference: a
// recursive-descent parser into a tiny AST and an exhaustive backtracking
// matcher over runes. It shares no code with the lexer, the parser or the
// engine.

const (
	refLit = iota
	refAny
	refSuper
	refSingle
	refClass
	refAlt   // {a,b}
	refGroup // op(a|b)
)

type refNode struct {
	kind    int
	r       rune
	set     map[rune]bool
	lo, hi  rune
	isRange bool
	not     bool
	alts    [][]refNode
	op      rune
}

func refParse(pattern string) ([]refNode, error) {
	rs := []rune(pattern)
	i := 0
	seq, err := refParseSeq(rs, &i, 0)
	if err != nil {
		return nil, err
	}
	if i != len(rs) {
		return nil, fmt.Errorf("unexpected %q at %d", rs[i], i)
	}
	return seq, nil
}

// refRuneAt returns the i-th rune of rs, or -1 past its end.
func refRuneAt(rs []rune, i int) rune {
	if i < len(rs) {
		return rs[i]
	}
	return -1
}

// refParseSeq parses terms until the end of the input or, inside an open
// `{` or `(`, until the separator or the closer of that group.
func refParseSeq(rs []rune, i *int, open rune) ([]refNode, error) {
	var seq []refNode
	for *i < len(rs) {
		r := rs[*i]
		next := refRuneAt(rs, *i+1)
		switch {
		case open == '{' && (r == ',' || r == '}'):
			return seq, nil
		case open == '(' && (r == '|' || r == ')'):
			return seq, nil
		case r == '\\':
			if next == -1 {
				return nil, errors.New("trailing backslash")
			}
			seq = append(seq, refNode{kind: refLit, r: next})
			*i += 2
		case (r == '?' || r == '*' || r == '+' || r == '@' || r == '!') && next == '(':
			*i += 2
			alts, err := refParseAlts(rs, i, '(')
			if err != nil {
				return nil, err
			}
			seq = append(seq, refNode{kind: refGroup, op: r, alts: alts})
		case r == '*':
			if next == '*' && refRuneAt(rs, *i+2) != '(' {
				seq = append(seq, refNode{kind: refSuper})
				*i += 2
			} else {
				seq = append(seq, refNode{kind: refAny})
				*i++
			}
		case r == '?':
			seq = append(seq, refNode{kind: refSingle})
			*i++
		case r == '[':
			n, err := refParseClass(rs, i)
			if err != nil {
				return nil, err
			}
			seq = append(seq, n)
		case r == '{':
			*i++
			alts, err := refParseAlts(rs, i, '{')
			if err != nil {
				return nil, err
			}
			seq = append(seq, refNode{kind: refAlt, alts: alts})
		default:
			seq = append(seq, refNode{kind: refLit, r: r})
			*i++
		}
	}
	if open != 0 {
		return nil, fmt.Errorf("unclosed %q", open)
	}
	return seq, nil
}

func refParseAlts(rs []rune, i *int, open rune) ([][]refNode, error) {
	sep, closer := ',', '}'
	if open == '(' {
		sep, closer = '|', ')'
	}
	var alts [][]refNode
	for {
		seq, err := refParseSeq(rs, i, open)
		if err != nil {
			return nil, err
		}
		alts = append(alts, seq)
		switch refRuneAt(rs, *i) {
		case sep:
			*i++
		case closer:
			*i++
			return alts, nil
		default:
			return nil, fmt.Errorf("unclosed %q", open)
		}
	}
}

func refParseClass(rs []rune, i *int) (refNode, error) {
	n := refNode{kind: refClass, set: map[rune]bool{}}
	*i++ // Skip the `[`.
	if refRuneAt(rs, *i) == '!' {
		n.not = true
		*i++
	}
	if refRuneAt(rs, *i+1) == '-' {
		if refRuneAt(rs, *i+3) != ']' || refRuneAt(rs, *i+2) == -1 {
			return n, errors.New("bad range")
		}
		n.isRange, n.lo, n.hi = true, rs[*i], rs[*i+2]
		if n.hi < n.lo {
			return n, errors.New("range hi < lo")
		}
		*i += 4
		return n, nil
	}
	for *i < len(rs) {
		r := rs[*i]
		switch r {
		case '\\':
			if *i+1 >= len(rs) {
				return n, errors.New("trailing backslash")
			}
			n.set[rs[*i+1]] = true
			*i += 2
		case ']':
			*i++
			if len(n.set) == 0 {
				return n, errors.New("empty class")
			}
			return n, nil
		default:
			n.set[r] = true
			*i++
		}
	}
	return n, errors.New("unclosed [")
}

type refMatcher struct {
	sep []rune
}

func (m refMatcher) isSep(r rune) bool {
	for _, s := range m.sep {
		if s == r {
			return true
		}
	}
	return false
}

// match reports whether seq matches the whole of s.
func (m refMatcher) match(seq []refNode, s []rune) bool {
	return m.at(seq, 0, s, 0)
}

func (m refMatcher) anyAlt(alts [][]refNode, s []rune) bool {
	for _, a := range alts {
		if m.match(a, s) {
			return true
		}
	}
	return false
}

// at reports whether seq[k:] matches s[pos:] entirely.
func (m refMatcher) at(seq []refNode, k int, s []rune, pos int) bool {
	if k == len(seq) {
		return pos == len(s)
	}
	n := seq[k]
	switch n.kind {
	case refLit:
		return pos < len(s) && s[pos] == n.r && m.at(seq, k+1, s, pos+1)
	case refSingle:
		return pos < len(s) && !m.isSep(s[pos]) && m.at(seq, k+1, s, pos+1)
	case refClass:
		if pos >= len(s) {
			return false
		}
		var in bool
		if n.isRange {
			in = n.lo <= s[pos] && s[pos] <= n.hi
		} else {
			in = n.set[s[pos]]
		}
		return in != n.not && m.at(seq, k+1, s, pos+1)
	case refAny:
		for e := pos; e <= len(s); e++ {
			if e > pos && m.isSep(s[e-1]) {
				break
			}
			if m.at(seq, k+1, s, e) {
				return true
			}
		}
		return false
	case refSuper:
		for e := pos; e <= len(s); e++ {
			if m.at(seq, k+1, s, e) {
				return true
			}
		}
		return false
	case refAlt:
		return m.oneOf(n.alts, seq, k, s, pos)
	case refGroup:
		switch n.op {
		case '@':
			return m.oneOf(n.alts, seq, k, s, pos)
		case '?':
			return m.at(seq, k+1, s, pos) || m.oneOf(n.alts, seq, k, s, pos)
		case '*', '+':
			// The positions reachable from pos by whole iterations.
			reach := make([]bool, len(s)+1)
			reach[pos] = true
			for p := pos; p <= len(s); p++ {
				if !reach[p] {
					continue
				}
				for e := p + 1; e <= len(s); e++ {
					if !reach[e] && m.anyAlt(n.alts, s[p:e]) {
						reach[e] = true
					}
				}
			}
			for e := pos; e <= len(s); e++ {
				if !reach[e] {
					continue
				}
				if e == pos && n.op == '+' && !m.anyAlt(n.alts, nil) {
					continue
				}
				if m.at(seq, k+1, s, e) {
					return true
				}
			}
			return false
		case '!':
			for e := pos; e <= len(s); e++ {
				if e > pos && m.isSep(s[e-1]) {
					break
				}
				if !m.anyAlt(n.alts, s[pos:e]) && m.at(seq, k+1, s, e) {
					return true
				}
			}
			return false
		}
	}
	panic("unreachable")
}

// oneOf reports whether one of the alts matches s[pos:e] exactly and seq[k+1:]
// matches the rest, for some e.
func (m refMatcher) oneOf(alts [][]refNode, seq []refNode, k int, s []rune, pos int) bool {
	for e := pos; e <= len(s); e++ {
		if m.anyAlt(alts, s[pos:e]) && m.at(seq, k+1, s, e) {
			return true
		}
	}
	return false
}

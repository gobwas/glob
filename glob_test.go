package glob

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCompileErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		pat    string
		offset int
		reason string
	}{
		{
			pat:    "{a,b",
			offset: 4,
			reason: "unclosed `{`",
		},
		{
			pat:    "[abc",
			offset: 4,
			reason: "unexpected end of input",
		},
		{
			pat:    "[a-cx]",
			offset: 5,
			reason: "expected close range character",
		},
		{
			// The class is lexed as a whole: the offset is past its `]`.
			pat:    "[c-a]",
			offset: 5,
			reason: "range hi character is less than lo",
		},
		{
			name:   "trailing backslash",
			pat:    `a\`,
			offset: 2,
			reason: "trailing backslash",
		},
		{
			name:   "invalid utf8",
			pat:    "a\xffb",
			offset: 1,
			reason: "invalid UTF-8 sequence",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := Compile(test.pat)
			var se *SyntaxError
			if !errors.As(err, &se) {
				t.Fatalf("Compile(%q) = %v; want *SyntaxError", test.pat, err)
			}
			if se.Offset != test.offset || se.Reason != test.reason {
				t.Errorf(
					"Compile(%q) = (%d, %q); want (%d, %q)",
					test.pat, se.Offset, se.Reason, test.offset, test.reason,
				)
			}
		})
	}
}

func TestPatternMatch(t *testing.T) {
	for i, test := range []struct {
		sep []rune
		pat string
		str string
		exp bool
	}{
		{
			pat: "a*a*a*a*b",
			str: strings.Repeat("a", 100),
			exp: false,
		},
		{
			pat: "a*a*a*a*b",
			str: strings.Repeat("a", 100) + "b",
			exp: true,
		},
		{
			pat: "{a,ab}c",
			str: "abc",
			exp: true,
		},
		{
			pat: "* ?at * eyes",
			str: "my cat has very bright eyes",
			exp: true,
		},
		{
			pat: "",
			str: "",
			exp: true,
		},
		{
			pat: "",
			str: "b",
			exp: false,
		},
		{
			pat: "*ä",
			str: "åä",
			exp: true,
		},
		{
			pat: "abc",
			str: "abc",
			exp: true,
		},
		{
			pat: "a*c",
			str: "abc",
			exp: true,
		},
		{
			pat: "a*c",
			str: "a12345c",
			exp: true,
		},
		{
			pat: "a?c",
			str: "a1c",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "a.b",
			str: "a.b",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "a.*",
			str: "a.b",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "a.**",
			str: "a.b.c",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "a.?.c",
			str: "a.b.c",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "a.?.?",
			str: "a.b.c",
			exp: true,
		},
		{
			pat: "?at",
			str: "cat",
			exp: true,
		},
		{
			pat: "?at",
			str: "fat",
			exp: true,
		},
		{
			pat: "*",
			str: "abc",
			exp: true,
		},
		{
			pat: "*a*",
			str: "a",
			exp: true,
		},
		{
			pat: "\\*",
			str: "*",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "**",
			str: "a.b.c",
			exp: true,
		},
		{
			pat: "?at",
			str: "at",
			exp: false,
		},
		{
			sep: []rune{'f'},
			pat: "?at",
			str: "fat",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "a.*",
			str: "a.b.c",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "a.?.c",
			str: "a.bb.c",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "*",
			str: "a.b.c",
			exp: false,
		},
		{
			pat: "*test",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "this*",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "*is *",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "*is*a*",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "**test**",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "**is**a***test*",
			str: "this is a test",
			exp: true,
		},
		{
			pat: "*test*",
			str: "test",
			exp: true,
		},
		{
			pat: "*is",
			str: "this is a test",
			exp: false,
		},
		{
			pat: "*no*",
			str: "this is a test",
			exp: false,
		},
		{
			pat: "[!a]*",
			str: "this is a test3",
			exp: true,
		},
		{
			pat: "*abc",
			str: "abcabc",
			exp: true,
		},
		{
			pat: "**abc",
			str: "abcabc",
			exp: true,
		},
		{
			pat: "???",
			str: "abc",
			exp: true,
		},
		{
			pat: "?*?",
			str: "abc",
			exp: true,
		},
		{
			pat: "?*?",
			str: "ac",
			exp: true,
		},
		{
			pat: "sta",
			str: "stagnation",
			exp: false,
		},
		{
			pat: "sta*",
			str: "stagnation",
			exp: true,
		},
		{
			pat: "sta?",
			str: "stagnation",
			exp: false,
		},
		{
			pat: "sta?n",
			str: "stagnation",
			exp: false,
		},
		{
			pat: "{abc,def}ghi",
			str: "defghi",
			exp: true,
		},
		{
			pat: "{abc,abcd}a",
			str: "abcda",
			exp: true,
		},
		{
			pat: "{,a}",
			str: "",
			exp: true,
		},
		{
			pat: "{a,}",
			str: "",
			exp: true,
		},
		{
			pat: "{a,ab}{bc,f}",
			str: "abc",
			exp: true,
		},
		{
			pat: "{*,**}{a,b}",
			str: "ab",
			exp: true,
		},
		{
			pat: "{*,**}{a,b}",
			str: "ac",
			exp: false,
		},
		{
			pat: "/{rate,[a-z][a-z][a-z]}*",
			str: "/rate",
			exp: true,
		},
		{
			pat: "/{rate,[0-9][0-9][0-9]}*",
			str: "/rate",
			exp: true,
		},
		{
			pat: "/{rate,[a-z][a-z][a-z]}*",
			str: "/usd",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,*.yandex.*}",
			str: "www.google.com",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,*.yandex.*}",
			str: "www.yandex.com",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,*.yandex.*}",
			str: "yandex.com",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,*.yandex.*}",
			str: "google.com",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,yandex.*}",
			str: "www.google.com",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,yandex.*}",
			str: "yandex.com",
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,yandex.*}",
			str: "www.yandex.com",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "{*.google.*,yandex.*}",
			str: "google.com",
			exp: false,
		},
		{
			pat: "*//{,*.}example.com",
			str: "https://www.example.com",
			exp: true,
		},
		{
			pat: "*//{,*.}example.com",
			str: "http://example.com",
			exp: true,
		},
		{
			pat: "*//{,*.}example.com",
			str: "http://example.com.net",
			exp: false,
		},
		{
			sep: []rune{'.'},
			pat: "{a*,b}c",
			str: "abc",
			exp: true,
		},
		{
			pat: "[a-z][!a-x]*cat*[h][!b]*eyes*",
			str: "my cat has very bright eyes",
			exp: true,
		},
		{
			pat: "[a-z][!a-x]*cat*[h][!b]*eyes*",
			str: "my dog has very bright eyes",
			exp: false,
		},

		{
			pat: "google.com",
			str: "google.com",
			exp: true,
		},
		{
			pat: "google.com",
			str: "gobwas.com",
			exp: false,
		},

		{
			pat: "https://*.google.*",
			str: "https://account.google.com",
			exp: true,
		},
		{
			pat: "https://*.google.*",
			str: "https://google.com",
			exp: false,
		},

		{
			pat: "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}",
			str: "http://yahoo.com",
			exp: true,
		},
		{
			pat: "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}",
			str: "http://google.com",
			exp: false,
		},

		{
			pat: "{https://*gobwas.com,http://exclude.gobwas.com}",
			str: "https://safe.gobwas.com",
			exp: true,
		},
		{
			pat: "{https://*gobwas.com,http://exclude.gobwas.com}",
			str: "http://safe.gobwas.com",
			exp: false,
		},
		{
			pat: "{https://*gobwas.com,http://exclude.gobwas.com}",
			str: "http://exclude.gobwas.com",
			exp: true,
		},

		{
			pat: "{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}",
			str: "abczqdef",
			exp: true,
		},

		{
			pat: "{abc*def,abc?def,abc[zte]def}",
			str: "abczdef",
			exp: true,
		},

		{
			pat: "abc*",
			str: "abcdef",
			exp: true,
		},
		{
			pat: "abc*",
			str: "af",
			exp: false,
		},

		{
			pat: "*def",
			str: "abcdef",
			exp: true,
		},
		{
			pat: "*def",
			str: "af",
			exp: false,
		},

		{
			pat: "ab*ef",
			str: "abcdef",
			exp: true,
		},
		{
			pat: "ab*ef",
			str: "af",
			exp: false,
		},

		{
			pat: "{a,ab,abc}",
			str: "ab",
			exp: true,
		},
		{
			pat: "{a,ab,abc}",
			str: "",
			exp: false,
		},
		{
			pat: "{a,ab,abc,abcd}{b,bc,bcd,bcde}{c,cd,cde,cdef}",
			str: "abcdbcdecdef",
			exp: true,
		},
		{
			pat: "{{a,b},c}",
			str: "a",
			exp: true,
		},
		{
			pat: "{{a,b},c}",
			str: "c",
			exp: true,
		},

		// A separator-limited `*` following a `**` may get stuck at a
		// separator; matching must then backtrack to the pending `**`
		// restart point. See the star checkpoints stack in Pattern.Match().
		{
			sep: []rune{'.'},
			pat: "**a*b",
			str: ".a.ab", // ** = ".a.", * = ""
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "**a*b",
			str: "axc.ab", // ** = "axc.", * = ""
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "**a*b",
			str: "a.axb", // ** = "a.", * = "x"
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "*a**b",
			str: "xa.b", // * = "x", ** = "."
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "**.x",
			str: "a.b.x", // ** = "a.b"
			exp: true,
		},
		{
			sep: []rune{'.'},
			pat: "*a**b",
			str: "x.ab", // `*` can not extend over the first separator.
			exp: false,
		},

		// A star inside an alternative must not discard the restart point
		// of a star outside of it: here the outer `*` must consume "1" and
		// the empty alternative must be taken. Found by FuzzMatchRegexp;
		// see the starsFloor handling in matchContext.storeStar().
		{
			pat: "*{*0,}",
			str: "1",
			exp: true,
		},

		// Braces without commas hold a single alternative: `{ab*}` is the
		// pattern `ab*`, not `{ab,*}`. See the opList handling in compile().
		{
			pat: "{ab*}x",
			str: "abzx",
			exp: true,
		},
		{
			pat: "{ab*}x",
			str: "zzx",
			exp: false,
		},
		{
			pat: "{ab*}",
			str: "ab",
			exp: true,
		},
		{
			pat: "{a}{b}{c}",
			str: "abc",
			exp: true,
		},
		{
			pat: "{**/daxing}/x",
			str: "a/daxing/x",
			exp: true,
		},
		{
			pat: "{**/daxing}/x",
			str: "zz/x",
			exp: false,
		},
		// U+FFFD is a character like any other, both in the pattern and in
		// the input: it must not be taken for an invalid byte.
		{
			pat: "[!a]",
			str: "�",
			exp: true,
		},
		{
			pat: "[!a-c]",
			str: "�",
			exp: true,
		},
		{
			pat: "[�]",
			str: "�",
			exp: true,
		},
		{
			pat: "a�b",
			str: "a�b",
			exp: true,
		},
		{
			pat: "?",
			str: "�",
			exp: true,
		},
	} {

		suffix := fmt.Sprintf("-%02d", i)
		t.Run("glob"+suffix, func(t *testing.T) {
			t.Logf("testing pattern=%#q", test.pat)
			var (
				pat     = MustCompile(test.pat, test.sep...)
				start   = time.Now()
				result  = pat.Match(test.str)
				latency = time.Since(start)
			)
			t.Logf(
				"delim=%#q pattern=%#q fixture=%#q result=%t want=%t latency=%.4fms",
				test.sep, test.pat, test.str, result, test.exp,
				latency.Seconds()*1000,
			)
			if result != test.exp {
				t.Errorf(
					"pattern %q matching %q should be %v but got %v",
					test.pat, test.str, test.exp, result,
				)
			}
			if latency > 5*time.Millisecond {
				t.Errorf("too slow: %.4fms", latency.Seconds()*1000)
			}
		})
		if len(test.sep) == 0 && !strings.ContainsAny(test.pat, "{}") {
			// NOTE: filepath.Match doesn't support `{}`.
			t.Run("filepath"+suffix, func(t *testing.T) {
				// NOTE: filepath.Match negates a character class with `^`
				// instead of `!`.
				pat := strings.ReplaceAll(test.pat, "[!", "[^")
				start := time.Now()
				result, err := filepath.Match(pat, test.str)
				latency := time.Since(start)
				t.Logf(
					"[filepath] pattern=%#q fixture=%#q result=%t want=%t latency=%.4fms",
					test.pat, test.str, result, test.exp,
					latency.Seconds()*1000,
				)
				if err != nil {
					t.Errorf("filepath.Match(%q) failed: %v", test.pat, err)
				}
				if result != test.exp {
					t.Errorf(
						"filepath.Match(%q, %q) = %t; test expects %t",
						test.pat, test.str, result, test.exp,
					)
				}
			})
		}
	}
}

func ExampleQuoteMeta() {
	s := "{foo*}"
	// Output: \{foo\*\}
	fmt.Println(QuoteMeta(s))
}

func TestQuoteMeta(t *testing.T) {
	for id, test := range []struct {
		in  string
		out string
	}{
		{
			in:  `[foo*]`,
			out: `\[foo\*\]`,
		},
		{
			in:  `{foo*}`,
			out: `\{foo\*\}`,
		},
		{
			in:  `*?\[]{}`,
			out: `\*\?\\\[\]\{\}`,
		},
		{
			in:  `some text and *?\[]{}`,
			out: `some text and \*\?\\\[\]\{\}`,
		},
	} {
		act := QuoteMeta(test.in)
		if act != test.out {
			t.Errorf("#%d QuoteMeta(%q) = %q; want %q", id, test.in, act, test.out)
		}
		if _, err := Compile(act); err != nil {
			t.Errorf("#%d _, err := Compile(QuoteMeta(%q) = %q); err = %q", id, test.in, act, err)
		}
	}
}

func BenchmarkPattern(b *testing.B) {
	for _, test := range []struct {
		name  string
		pat   string
		input map[string]string
	}{
		{
			name: "segments",
			pat:  `{a,ab,abc}`,
			input: map[string]string{
				"0":     "a",
				"1":     "ab",
				"2":     "abc",
				"empty": "",
			},
		},
		{
			name: "long_segments",
			pat:  `{a,ab,abc,abcd}{b,bc,bcd,bcde}{c,cd,cde,cdef}`,
			input: map[string]string{
				"long": "abcdbcdecdef",
			},
		},
		{
			name: "cat",
			pat:  `[a-z][!a-x]*cat*[h][!b]*eyes*`,
			input: map[string]string{
				"match":    "my cat has very bright eyes",
				"mismatch": "my dog has very bright eyes",
			},
		},
		{
			name: "wildcard",
			pat:  `https://*.google.*`,
			input: map[string]string{
				"match":    "https://account.google.com",
				"mismatch": "https://google.com",
			},
		},
		{
			name: "alternatives",
			pat:  `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`,
			input: map[string]string{
				"match":    "http://yahoo.com",
				"mismatch": "http://google.com",
			},
		},
		{
			name: "alternatives_suffix_first",
			pat:  `{https://*gobwas.com,http://exclude.gobwas.com}`,
			input: map[string]string{
				"match":    "https://safe.gobwas.com",
				"mismatch": "http://safe.gobwas.com",
			},
		},
		{
			name: "alternatives_suffix_second",
			pat:  `{https://*gobwas.com,http://exclude.gobwas.com}`,
			input: map[string]string{
				"match": "http://exclude.gobwas.com",
				//"mismatch": "",
			},
		},
		{
			name: "alternatives_combine_lite",
			pat:  `{abc*def,abc?def,abc[zte]def}`,
			input: map[string]string{
				"match": "abczdef",
				//"mismatch": "",
			},
		},
		{
			name: "alternatives_combine_hard",
			pat:  `{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}`,
			input: map[string]string{
				"match": "abczqdef",
				//"mismatch": "",
			},
		},
		{
			name: "plain",
			pat:  `google.com`,
			input: map[string]string{
				"match":    "google.com",
				"mismatch": "gobwas.com",
			},
		},
		{
			name: "prefix",
			pat:  `abc*`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
		{
			name: "suffix",
			pat:  `*def`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
		{
			name: "prefix_and_suffix",
			pat:  `ab*ef`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
	} {
		b.Run(test.name+"-compile", func(b *testing.B) {
			for b.Loop() {
				MustCompile(test.pat)
			}
		})
		pat := MustCompile(test.pat)
		for _, key := range slices.Sorted(maps.Keys(test.input)) {
			str := test.input[key]
			b.Run(test.name+"-match-"+key, func(b *testing.B) {
				for b.Loop() {
					pat.Match(str)
				}
			})
		}
	}
}

// BenchmarkCompareGlobAndRegexp is what the regexp table in the readme is
// made of. The regexps are the exact equivalents of the globs: anchored,
// with the meta characters escaped and, since `*` matches a newline as any
// other character, with the `s` flag (see issue #21).
func BenchmarkCompareGlobAndRegexp(b *testing.B) {
	for _, test := range []struct {
		name   string
		glob   string
		regexp string
		input  map[string]string
	}{
		{
			name:   "cat",
			glob:   `[a-z][!a-x]*cat*[h][!b]*eyes*`,
			regexp: `(?s)^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`,
			input: map[string]string{
				"match":    "my cat has very bright eyes",
				"mismatch": "my dog has very bright eyes",
			},
		},
		{
			name:   "wildcard",
			glob:   `https://*.google.*`,
			regexp: `(?s)^https://.*\.google\..*$`,
			input: map[string]string{
				"match":    "https://account.google.com",
				"mismatch": "https://google.com",
			},
		},
		{
			name:   "alternatives",
			glob:   `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`,
			regexp: `(?s)^(https://.*\.google\..*|.*yandex\..*|.*yahoo\..*|.*mail\.ru)$`,
			input: map[string]string{
				"match":    "http://yahoo.com",
				"mismatch": "http://google.com",
			},
		},
		{
			name:   "alternatives_suffix_first",
			glob:   `{https://*gobwas.com,http://exclude.gobwas.com}`,
			regexp: `(?s)^(https://.*gobwas\.com|http://exclude\.gobwas\.com)$`,
			input: map[string]string{
				"match":    "https://safe.gobwas.com",
				"mismatch": "http://safe.gobwas.com",
			},
		},
		{
			name:   "alternatives_suffix_second",
			glob:   `{https://*gobwas.com,http://exclude.gobwas.com}`,
			regexp: `(?s)^(https://.*gobwas\.com|http://exclude\.gobwas\.com)$`,
			input: map[string]string{
				"match": "http://exclude.gobwas.com",
				//"mismatch": "",
			},
		},
		{
			name:   "alternatives_combine_lite",
			glob:   `{abc*def,abc?def,abc[zte]def}`,
			regexp: `(?s)^(abc.*def|abc.def|abc[zte]def)$`,
			input: map[string]string{
				"match": "abczdef",
				//"mismatch": "",
			},
		},
		{
			name:   "alternatives_combine_hard",
			glob:   `{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}`,
			regexp: `(?s)^(abc.*[a-c]def|abc.[d-g]def|abc[zte].def)$`,
			input: map[string]string{
				"match": "abczqdef",
				//"mismatch": "",
			},
		},
		{
			name:   "plain",
			glob:   `google.com`,
			regexp: `^google\.com$`,
			input: map[string]string{
				"match":    "google.com",
				"mismatch": "gobwas.com",
			},
		},
		{
			name:   "prefix",
			glob:   `abc*`,
			regexp: `(?s)^abc.*$`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
		{
			name:   "suffix",
			glob:   `*def`,
			regexp: `(?s)^.*def$`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
		{
			name:   "prefix_and_suffix",
			glob:   `ab*ef`,
			regexp: `(?s)^ab.*ef$`,
			input: map[string]string{
				"match":    "abcdef",
				"mismatch": "af",
			},
		},
	} {
		b.Run(test.name+"-glob-compile", func(b *testing.B) {
			for b.Loop() {
				MustCompile(test.glob)
			}
		})
		b.Run(test.name+"-regexp-compile", func(b *testing.B) {
			for b.Loop() {
				regexp.MustCompile(test.regexp)
			}
		})
		var (
			pat = MustCompile(test.glob)
			exp = regexp.MustCompile(test.regexp)
		)
		for _, key := range slices.Sorted(maps.Keys(test.input)) {
			str := test.input[key]
			b.Run(test.name+"-glob-"+key, func(b *testing.B) {
				for b.Loop() {
					pat.Match(str)
				}
			})
			b.Run(test.name+"-regexp-"+key, func(b *testing.B) {
				for b.Loop() {
					exp.MatchString(str)
				}
			})
		}
	}
}

func TestPatternString(t *testing.T) {
	for _, test := range []struct {
		name string
		str  string
	}{
		{
			str: "",
		},
		{
			str: "foo",
		},
		{
			str: "*.github.com",
		},
		{
			str: "{cat,bat,[fr]at}",
		},
		{
			str: `\*escaped\?`,
		},
		{
			str: "ångstr[ö]m",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := Compile(test.str, '.', '/')
			if err != nil {
				t.Fatalf("Compile(%q): %v", test.str, err)
			}
			if act := p.String(); act != test.str {
				t.Errorf("String() = %q; want %q", act, test.str)
			}
			if act := fmt.Sprint(p); act != test.str {
				t.Errorf("fmt.Sprint() = %q; want %q", act, test.str)
			}
		})
	}
}

func TestPatternSeparators(t *testing.T) {
	for _, test := range []struct {
		name string
		sep  []rune
	}{
		{
			sep: nil,
		},
		{
			sep: []rune{'.'},
		},
		{
			sep: []rune{'.', '/'},
		},
		{
			sep: []rune{'ö', '/', '.'},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := MustCompile("*", test.sep...)
			act := p.Separators()
			if !slices.Equal(act, test.sep) {
				t.Errorf("Separators() = %v; want %v", act, test.sep)
			}
			if len(test.sep) == 0 && act != nil {
				t.Errorf("Separators() = %v; want nil", act)
			}
		})
	}
}

// TestPatternSeparatorsAliasing checks that Separators() returns the slice
// given to Compile as is, and that matching does not depend on it.
func TestPatternSeparatorsAliasing(t *testing.T) {
	sep := []rune{'.'}
	p := MustCompile("*", sep...)
	if got := p.Separators(); &got[0] != &sep[0] {
		t.Errorf(
			"Separators() does not share the backing array of the given slice",
		)
	}
	sep[0] = 'x'
	if got := p.Separators(); !slices.Equal(got, sep) {
		t.Errorf("Separators() = %v; want %v", got, sep)
	}
	if p.Match("a.b") {
		t.Errorf("Match() picked up the modification of the given slice")
	}
	if !p.Match("axb") {
		t.Errorf("Match() picked up the modification of the given slice")
	}
}

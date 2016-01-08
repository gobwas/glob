package glob

import (
	"github.com/gobwas/glob/match"
	"reflect"
	"testing"
)

const (
	pattern_all = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	fixture_all = "my cat has very bright eyes"

	pattern_plain = "google.com"
	fixture_plain = "google.com"

	pattern_multiple = "https://*.google.*"
	fixture_multiple = "https://account.google.com"

	pattern_alternatives = "{https://*.google.*,*yahoo.*}"
	fixture_alternatives = "http://yahoo.com"

	pattern_prefix        = "abc*"
	pattern_suffix        = "*def"
	pattern_prefix_suffix = "ab*ef"
	fixture_prefix_suffix = "abcdef"
)

type test struct {
	pattern, match string
	should         bool
	delimiters     []string
}

func glob(s bool, p, m string, d ...string) test {
	return test{p, m, s, d}
}

func TestCompilePattern(t *testing.T) {
	for id, test := range []struct {
		pattern string
		sep     string
		exp     match.Matcher
	}{
	//		{
	//			pattern: "{http://*yandex.ru,b}",
	//			exp:     match.Raw{"t"},
	//		},
	} {
		glob, err := Compile(test.pattern, test.sep)
		if err != nil {
			t.Errorf("#%d compile pattern error: %s", id, err)
			continue
		}

		matcher := glob.(match.Matcher)

		if !reflect.DeepEqual(test.exp, matcher) {
			t.Errorf("#%d unexpected compilation:\nexp: %s\nact: %s", id, test.exp, matcher)
			continue
		}
	}
}

func TestIndexByteNonEscaped(t *testing.T) {
	for _, test := range []struct {
		s    string
		n, e byte
		i    int
	}{
		{
			"\\n_n",
			'n',
			'\\',
			3,
		},
		{
			"ab",
			'a',
			'\\',
			0,
		},
		{
			"ab",
			'b',
			'\\',
			1,
		},
		{
			"",
			'b',
			'\\',
			-1,
		},
		{
			"\\b",
			'b',
			'\\',
			-1,
		},
	} {
		i := indexByteNonEscaped(test.s, test.n, test.e, 0)
		if i != test.i {
			t.Errorf("unexpeted index: expected %v, got %v", test.i, i)
		}
	}
}

func TestGlob(t *testing.T) {
	for _, test := range []test{
		glob(true, "* ?at * eyes", "my cat has very bright eyes"),

		glob(true, "abc", "abc"),
		glob(true, "a*c", "abc"),
		glob(true, "a*c", "a12345c"),
		glob(true, "a?c", "a1c"),
		glob(true, "a.b", "a.b", "."),
		glob(true, "a.*", "a.b", "."),
		glob(true, "a.**", "a.b.c", "."),
		glob(true, "a.?.c", "a.b.c", "."),
		glob(true, "a.?.?", "a.b.c", "."),
		glob(true, "?at", "cat"),
		glob(true, "?at", "fat"),
		glob(true, "*", "abc"),
		glob(true, `\*`, "*"),
		glob(true, "**", "a.b.c", "."),

		glob(false, "?at", "at"),
		glob(false, "?at", "fat", "f"),
		glob(false, "a.*", "a.b.c", "."),
		glob(false, "a.?.c", "a.bb.c", "."),
		glob(false, "*", "a.b.c", "."),

		glob(true, "*test", "this is a test"),
		glob(true, "this*", "this is a test"),
		glob(true, "*is *", "this is a test"),
		glob(true, "*is*a*", "this is a test"),
		glob(true, "**test**", "this is a test"),
		glob(true, "**is**a***test*", "this is a test"),

		glob(false, "*is", "this is a test"),
		glob(false, "*no*", "this is a test"),
		glob(true, "[!a]*", "this is a test3"),

		glob(true, "*abc", "abcabc"),
		glob(true, "**abc", "abcabc"),
		glob(true, "???", "abc"),
		glob(true, "?*?", "abc"),
		glob(true, "?*?", "ac"),

		glob(true, pattern_all, fixture_all),
		glob(true, pattern_plain, fixture_plain),
		glob(true, pattern_multiple, fixture_multiple),
		glob(true, pattern_alternatives, fixture_alternatives),
		glob(true, pattern_prefix, fixture_prefix_suffix),
		glob(true, pattern_suffix, fixture_prefix_suffix),
		glob(true, pattern_prefix_suffix, fixture_prefix_suffix),
	} {
		g, err := Compile(test.pattern, test.delimiters...)
		if err != nil {
			t.Errorf("parsing pattern %q error: %s", test.pattern, err)
			continue
		}

		result := g.Match(test.match)
		if result != test.should {
			t.Errorf("pattern %q matching %q should be %v but got %v", test.pattern, test.match, test.should, result)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Compile(pattern_all)
	}
}

func BenchmarkAll(b *testing.B) {
	m, _ := Compile(pattern_all)
	//	fmt.Println("tree all:")
	//	fmt.Println(m)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all)
	}
}

func BenchmarkMultiple(b *testing.B) {
	m, _ := Compile(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple)
	}
}
func BenchmarkAlternatives(b *testing.B) {
	m, _ := Compile(pattern_alternatives)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives)
	}
}
func BenchmarkPlain(b *testing.B) {
	m, _ := Compile(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain)
	}
}
func BenchmarkPrefix(b *testing.B) {
	m, _ := Compile(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}
func BenchmarkSuffix(b *testing.B) {
	m, _ := Compile(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}
func BenchmarkPrefixSuffix(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}

//BenchmarkParse-8          500000              2235 ns/op
//BenchmarkAll-8          20000000                73.1 ns/op
//BenchmarkMultiple-8     10000000               130 ns/op
//BenchmarkPlain-8        200000000                6.70 ns/op
//BenchmarkPrefix-8       200000000                8.36 ns/op
//BenchmarkSuffix-8       200000000                8.35 ns/op
//BenchmarkPrefixSuffix-8 100000000               13.6 ns/op

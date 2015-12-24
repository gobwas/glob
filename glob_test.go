package glob

import (
	"testing"
)

const (
	pattern_all = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	fixture_all = "my cat has very bright eyes"

	pattern_plain = "google.com"
	fixture_plain = "google.com"

	pattern_multiple = "https://*.google.*"
	fixture_multiple = "https://account.google.com"

	pattern_prefix = "abc*"
	pattern_suffix = "*def"
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

func TestIndexByteNonEscaped(t *testing.T) {
	for _, test := range []struct {
		s string
		n, e byte
		i int
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

		glob(true, "* ?at * eyes", "my cat has very bright eyes"),

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
		glob(true, "[!a]*", "this is a test"),

		glob(true, pattern_all, fixture_all),
		glob(true, pattern_plain, fixture_plain),
		glob(true, pattern_multiple, fixture_multiple),
		glob(true, pattern_prefix, fixture_prefix_suffix),
		glob(true, pattern_suffix, fixture_prefix_suffix),
		glob(true, pattern_prefix_suffix, fixture_prefix_suffix),
	} {
		g, err := New(test.pattern, test.delimiters...)
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
		New(pattern_all)
	}
}

func BenchmarkAll(b *testing.B) {
	m, _ := New(pattern_all)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all)
	}
}

func BenchmarkMultiple(b *testing.B) {
	m, _ := New(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple)
	}
}
func BenchmarkPlain(b *testing.B) {
	m, _ := New(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain)
	}
}
func BenchmarkPrefix(b *testing.B) {
	m, _ := New(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}
func BenchmarkSuffix(b *testing.B) {
	m, _ := New(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}
func BenchmarkPrefixSuffix(b *testing.B) {
	m, _ := New(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix)
	}
}
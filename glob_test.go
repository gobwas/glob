package glob

import (
	rGlob "github.com/ryanuber/go-glob"
	"regexp"
	"testing"
)

type test struct {
	pattern, match string
	should         bool
	delimiters     []string
}

func glob(s bool, p, m string, d ...string) test {
	return test{p, m, s, d}
}

func TestIndexOfNonEscaped(t *testing.T) {
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
	} {
		g, err := New(test.pattern, test.delimiters...)
		if err != nil {
			t.Error(err)
			continue
		}

		result := g.Match(test.match)
		if result != test.should {
			t.Errorf("pattern %q matching %q should be %v but got %v", test.pattern, test.match, test.should, result)
		}
	}
}

const Pattern = "*cat*eyes*"
const ExpPattern = ".*cat.*eyes.*"
const String = "my cat has very bright eyes"

const ProfPattern = "* ?at * eyes"
const ProfString = "my cat has very bright eyes"

//const Pattern = "*.google.com"
//const ExpPattern = ".*google\\.com"
//const String = "mail.google.com"
const PlainPattern = "google.com"
const PlainExpPattern = "google\\.com"
const PlainString = "google.com"

const PSPattern = "https://*.google.com"
const PSExpPattern = `https:\/\/[a-z]+\.google\\.com`
const PSString = "https://account.google.com"

func BenchmarkProf(b *testing.B) {
	m, _ := New(Pattern)

	for i := 0; i < b.N; i++ {
		_ = m.Match(String)
	}
}

func BenchmarkGobwas(b *testing.B) {
	m, _ := New(Pattern)

	for i := 0; i < b.N; i++ {
		_ = m.Match(String)
	}
}
func BenchmarkGobwasPlain(b *testing.B) {
	m, _ := New(PlainPattern)

	for i := 0; i < b.N; i++ {
		_ = m.Match(PlainString)
	}
}
func BenchmarkGobwasPrefix(b *testing.B) {
	m, _ := New("abc*")

	for i := 0; i < b.N; i++ {
		_ = m.Match("abcdef")
	}
}
func BenchmarkGobwasSuffix(b *testing.B) {
	m, _ := New("*def")

	for i := 0; i < b.N; i++ {
		_ = m.Match("abcdef")
	}
}
func BenchmarkGobwasPrefixSuffix(b *testing.B) {
	m, _ := New("ab*ef")

	for i := 0; i < b.N; i++ {
		_ = m.Match("abcdef")
	}
}

func BenchmarkRyanuber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rGlob.Glob(Pattern, String)
	}
}
func BenchmarkRyanuberPlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rGlob.Glob(PlainPattern, PlainString)
	}
}
func BenchmarkRyanuberPrefixSuffix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rGlob.Glob(PSPattern, PSString)
	}
}


func BenchmarkRegExp(b *testing.B) {
	r := regexp.MustCompile(ExpPattern)
	for i := 0; i < b.N; i++ {
		_ = r.Match([]byte(String))
	}
}
func BenchmarkRegExpPrefixSuffix(b *testing.B) {
	r := regexp.MustCompile(PSExpPattern)
	for i := 0; i < b.N; i++ {
		_ = r.Match([]byte(PSString))
	}
}
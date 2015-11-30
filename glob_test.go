package glob
import (
	"testing"
	rGlob "github.com/ryanuber/go-glob"
	"regexp"
	"strings"
)


type test struct {
	pattern, match string
	should         bool
	delimiters     []string
}

func glob(s bool, p, m string, d ...string) test {
	return test{p, m, s, d}
}

func TestFirstIndexOfChars(t *testing.T) {
	for _, test := range []struct{
		s string
		c []string
		i int
		r string
	}{
		{
			"**",
			[]string{"**", "*"},
			0,
			"**",
		},
		{
			"**",
			[]string{"*", "**"},
			0,
			"**",
		},
	}{
		i, r := firstIndexOfChars(test.s, test.c)
		if i != test.i || r != test.r {
			t.Errorf("unexpeted index: expected %q at %v, got %q at %v", test.r, test.i, r, i)
		}
	}
}

func TestGlob(t *testing.T) {
	for _, test := range []test {
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
		glob(true, "*", "abc"),
		glob(true, "**", "a.b.c", "."),

		glob(false, "?at", "at"),
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
	}{
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

func BenchmarkGobwas(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m, err := New(Pattern)
		if err != nil {
			b.Fatal(err)
		}

		_ = m.Match(String)
	}
}

func BenchmarkRyanuber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rGlob.Glob(Pattern, String)
	}
}
func BenchmarkRegExp(b *testing.B) {
	r := regexp.MustCompile(ExpPattern)
	for i := 0; i < b.N; i++ {
		_ = r.Match([]byte(String))
	}
}

var ALPHABET_S = []string{"a", "b", "c"}
const ALPHABET = "abc"
const STR = "faafsdfcsdffc"


func BenchmarkIndexOfAny(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.IndexAny(STR, ALPHABET)
	}
}
func BenchmarkFirstIndexOfChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		firstIndexOfChars(STR, ALPHABET_S)
	}
}
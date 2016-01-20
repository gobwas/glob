package glob

import (
	"bytes"
	"fmt"
	"github.com/gobwas/glob/match"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

const (
	pattern_all          = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	regexp_all           = `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`
	fixture_all_match    = "my cat has very bright eyes"
	fixture_all_mismatch = "my dog has very bright eyes"

	pattern_plain          = "google.com"
	regexp_plain           = `^google\.com$`
	fixture_plain_match    = "google.com"
	fixture_plain_mismatch = "gobwas.com"

	pattern_multiple          = "https://*.google.*"
	regexp_multiple           = `^https:\/\/.*\.google\..*$`
	fixture_multiple_match    = "https://account.google.com"
	fixture_multiple_mismatch = "https://google.com"

	pattern_alternatives          = "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}"
	regexp_alternatives           = `^(https:\/\/.*\.google\..*|.*yandex\..*|.*yahoo\..*|.*mail\.ru)$`
	fixture_alternatives_match    = "http://yahoo.com"
	fixture_alternatives_mismatch = "http://google.com"

	pattern_alternatives_suffix                = "{https://*gobwas.com,http://exclude.gobwas.com}"
	regexp_alternatives_suffix                 = `^(https:\/\/.*gobwas\.com|http://exclude.gobwas.com)$`
	fixture_alternatives_suffix_first_match    = "https://safe.gobwas.com"
	fixture_alternatives_suffix_first_mismatch = "http://safe.gobwas.com"
	fixture_alternatives_suffix_second         = "http://exclude.gobwas.com"

	pattern_prefix                 = "abc*"
	regexp_prefix                  = `^abc.*$`
	pattern_suffix                 = "*def"
	regexp_suffix                  = `^.*def$`
	pattern_prefix_suffix          = "ab*ef"
	regexp_prefix_suffix           = `^ab.*ef$`
	fixture_prefix_suffix_match    = "abcdef"
	fixture_prefix_suffix_mismatch = "af"

	pattern_alternatives_combine_lite = "{abc*def,abc?def,abc[zte]def}"
	regexp_alternatives_combine_lite  = `^(abc.*def|abc.def|abc[zte]def)$`
	fixture_alternatives_combine_lite = "abczdef"

	pattern_alternatives_combine_hard = "{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}"
	regexp_alternatives_combine_hard  = `^(abc.*[a-c]def|abc.[d-g]def|abc[zte].def)$`
	fixture_alternatives_combine_hard = "abczqdef"
)

type test struct {
	pattern, match string
	should         bool
	delimiters     []string
}

func glob(s bool, p, m string, d ...string) test {
	return test{p, m, s, d}
}

func draw(pattern string, m match.Matcher) string {
	return fmt.Sprintf(`digraph G {graph[label="%s"];%s}`, pattern, graphviz(m, fmt.Sprintf("%x", rand.Int63())))
}

func graphviz(m match.Matcher, id string) string {
	buf := &bytes.Buffer{}

	switch matcher := m.(type) {
	case match.BTree:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, matcher.Value.String())
		for _, m := range []match.Matcher{matcher.Left, matcher.Right} {
			switch n := m.(type) {
			case nil:
				rnd := rand.Int63()
				fmt.Fprintf(buf, `"%x"[label="<nil>"];`, rnd)
				fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)

			default:
				sub := fmt.Sprintf("%x", rand.Int63())
				fmt.Fprintf(buf, `"%s"->"%s";`, id, sub)
				fmt.Fprintf(buf, graphviz(n, sub))
			}
		}

	case match.AnyOf:
		fmt.Fprintf(buf, `"%s"[label="AnyOf"];`, id)
		for _, m := range matcher.Matchers {
			rnd := rand.Int63()
			fmt.Fprintf(buf, graphviz(m, fmt.Sprintf("%x", rnd)))
			fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
		}

	case match.EveryOf:
		fmt.Fprintf(buf, `"%s"[label="EveryOf"];`, id)
		for _, m := range matcher.Matchers {
			rnd := rand.Int63()
			fmt.Fprintf(buf, graphviz(m, fmt.Sprintf("%x", rnd)))
			fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
		}

	default:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, m.String())
	}

	return buf.String()
}

func DrawPatterns(t *testing.T) {
	for id, test := range []struct {
		pattern string
		sep     string
	}{
		//		{
		//			pattern: pattern_all,
		//		},
		//		{
		//			pattern: pattern_alternatives_suffix,
		//			sep:     separators,
		//		},
		//		{
		//			pattern: pattern_alternatives_combine_lite,
		//		},
		//		{
		//			pattern: pattern_alternatives_combine_hard,
		//		},
		{
			pattern: pattern_alternatives_suffix,
		},
		//		{
		//			pattern: "{https://*.mail.ru,*my.mail.ru,*my.myalpha*.i.mail.ru}",
		//		},
	} {
		glob, err := Compile(test.pattern, test.sep)
		if err != nil {
			t.Errorf("#%d compile pattern error: %s", id, err)
			continue
		}

		matcher := glob.(match.Matcher)
		fmt.Println(test.pattern)
		fmt.Println(strings.Repeat("=", len(test.pattern)))
		fmt.Println(draw(test.pattern, matcher))
		fmt.Println()
		fmt.Println(matcher.String())
		fmt.Println()
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

		glob(true, "{abc,def}ghi", "defghi"),
		glob(true, "{abc,abcd}a", "abcda"),
		glob(true, "{a,ab}{bc,f}", "abc"),
		glob(true, "{*,**}{a,b}", "ab"),
		glob(false, "{*,**}{a,b}", "ac"),

		glob(true, pattern_all, fixture_all_match),
		glob(false, pattern_all, fixture_all_mismatch),

		glob(true, pattern_plain, fixture_plain_match),
		glob(false, pattern_plain, fixture_plain_mismatch),

		glob(true, pattern_multiple, fixture_multiple_match),
		glob(false, pattern_multiple, fixture_multiple_mismatch),

		glob(true, pattern_alternatives, fixture_alternatives_match),
		glob(false, pattern_alternatives, fixture_alternatives_mismatch),

		glob(true, pattern_alternatives_suffix, fixture_alternatives_suffix_first_match),
		glob(false, pattern_alternatives_suffix, fixture_alternatives_suffix_first_mismatch),
		glob(true, pattern_alternatives_suffix, fixture_alternatives_suffix_second),

		glob(true, pattern_alternatives_combine_hard, fixture_alternatives_combine_hard),

		glob(true, pattern_alternatives_combine_lite, fixture_alternatives_combine_lite),

		glob(true, pattern_prefix, fixture_prefix_suffix_match),
		glob(false, pattern_prefix, fixture_prefix_suffix_mismatch),

		glob(true, pattern_suffix, fixture_prefix_suffix_match),
		glob(false, pattern_suffix, fixture_prefix_suffix_mismatch),

		glob(true, pattern_prefix_suffix, fixture_prefix_suffix_match),
		glob(false, pattern_prefix_suffix, fixture_prefix_suffix_mismatch),
	} {
		g, err := Compile(test.pattern, test.delimiters...)
		if err != nil {
			t.Errorf("parsing pattern %q error: %s", test.pattern, err)
			continue
		}

		result := g.Match(test.match)
		if result != test.should {
			t.Errorf("pattern %q matching %q should be %v but got %v\n%s", test.pattern, test.match, test.should, result, g)
		}

		//		if len(test.delimiters) == 0 || (len(test.delimiters) == 1 && test.delimiters[0] == string(filepath.Separator)) {
		//			result, err = filepath.Match(test.pattern, test.match)
		//			if err != nil {
		//				t.Errorf("[filepath] error matching pattern %q: %s", test.pattern, err)
		//				continue
		//			}
		//			if result != test.should {
		//				t.Errorf("[filepath] pattern %q matching %q should be %v but got %v\n%s", test.pattern, test.match, test.should, result, g)
		//			}
		//		}
	}
}

func BenchmarkParseGlob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Compile(pattern_all)
	}
}
func BenchmarkParseRegexp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		regexp.MustCompile(regexp_all)
	}
}

func BenchmarkAllGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_all)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all_match)
	}
}
func BenchmarkAllRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAllGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_all)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all_mismatch)
	}
}
func BenchmarkAllRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkMultipleGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple_match)
	}
}
func BenchmarkMultipleRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkMultipleGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple_mismatch)
	}
}
func BenchmarkMultipleRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_match)
	}
}
func BenchmarkAlternativesGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_mismatch)
	}
}
func BenchmarkAlternativesRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAlternativesRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixFirstGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_first_match)
	}
}
func BenchmarkAlternativesSuffixFirstGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_first_mismatch)
	}
}
func BenchmarkAlternativesSuffixSecondGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_second)
	}
}
func BenchmarkAlternativesCombineLiteGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_lite)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_combine_lite)
	}
}
func BenchmarkAlternativesCombineHardGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_hard)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_combine_hard)
	}
}
func BenchmarkAlternativesSuffixFirstRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAlternativesSuffixFirstRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAlternativesSuffixSecondRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_second)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAlternativesCombineLiteRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_lite)
	f := []byte(fixture_alternatives_combine_lite)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkAlternativesCombineHardRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_hard)
	f := []byte(fixture_alternatives_combine_hard)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPlainGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain_match)
	}
}
func BenchmarkPlainRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkPlainGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain_mismatch)
	}
}
func BenchmarkPlainRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}
func BenchmarkPrefixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkPrefixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}
func BenchmarkPrefixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}
func BenchmarkSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}
func BenchmarkSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}
func BenchmarkPrefixSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}
func BenchmarkPrefixSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}
func BenchmarkPrefixSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

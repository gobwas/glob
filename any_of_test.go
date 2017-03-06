package glob

import "testing"

func TestAnyOfMatch(t *testing.T) {
	for id, test := range []struct {
		globs   Globs
		fixture string
		match   bool
	}{
		{
			Globs{},
			"abcd",
			false,
		},
		{
			Globs{
				MustCompile("a*"),
				MustCompile("ab*"),
				MustCompile("abc*"),
			},
			"abcd",
			true,
		},
		{
			Globs{
				MustCompile("xa*"),
				MustCompile("xab*"),
				MustCompile("xabc*"),
			},
			"abcd",
			false,
		},
		{
			Globs{
				MustCompile("xa*"),
				MustCompile("xab*"),
				MustCompile("abc*"),
			},
			"abcd",
			true,
		},
	} {
		anyOf := NewAnyOf(test.globs...)
		match := anyOf.Match(test.fixture)
		if match != test.match {
			t.Errorf("#%d unexpected index: exp: %t, act: %t", id, test.match, match)
		}
	}
}

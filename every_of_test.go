package glob

import "testing"

func TestEveryOfMatch(t *testing.T) {
	for id, test := range []struct {
		globs   Globs
		fixture string
		match   bool
	}{
		{
			Globs{},
			"abcd",
			true,
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
				MustCompile("a*"),
				MustCompile("axb*"),
				MustCompile("abc*"),
			},
			"abcd",
			false,
		},
		{
			Globs{
				MustCompile("a*"),
				MustCompile("ab*"),
				MustCompile("abcx*"),
			},
			"abcd",
			false,
		},
	} {
		everyOf := NewEveryOf(test.globs...)
		match := everyOf.Match(test.fixture)
		if match != test.match {
			t.Errorf("#%d unexpected index: exp: %t, act: %t", id, test.match, match)
		}
	}
}

package glob

import (
	"testing"
)

func TestAltMatcher(t *testing.T) {
	for _, test := range []struct {
		name string
		str  string
	}{
		{
			str: "foo",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// {x,f,fo,{f,fo,foo},x}
			m := altMatcher{
				//&textMatcher{Text: "x"},
				//&textMatcher{Text: "f"},
				//&textMatcher{Text: "fo"},
				multiMatcher{
					&textMatcher{Text: "f"},
					altMatcher{
						&textMatcher{Text: "a"},
						&textMatcher{Text: "b"},
						&textMatcher{Text: "c"},
						&textMatcher{Text: "o"},
					},
					//&textMatcher{Text: "o"},
					&starMatcher{},
				},
				//altMatcher{
				//	&textMatcher{Text: "f"},
				//	&textMatcher{Text: "fo"},
				//	&textMatcher{Text: "foo"},
				//},
				//&textMatcher{Text: "foo"},
				//&textMatcher{Text: "x"},
			}
			p := Pattern{
				m:     m,
				state: needsState(m),
			}
			t.Logf("Pattern<%s>.Match(%q) = %t", m, test.str, p.Match(test.str))
		})
	}
}

type matcherFunc func(matchContext, string) (int, bool)

func (f matcherFunc) Match(x matchContext, s string) (int, bool) {
	return f(x, s)
}

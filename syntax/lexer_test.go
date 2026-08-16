package syntax

import (
	"testing"
)

func TestLexGood(t *testing.T) {
	for id, test := range []struct {
		pattern string
		items   []Token
	}{
		{
			pattern: "",
			items: []Token{
				{EOF, ""},
			},
		},
		{
			pattern: "hello",
			items: []Token{
				{Text, "hello"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{rate,[0-9]]}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "rate"},
				{TermSeparator, ","},
				{RangeOpen, "["},
				{RangeLo, "0"},
				{RangeBetween, "-"},
				{RangeHi, "9"},
				{RangeClose, "]"},
				{Text, "]"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\{world",
			items: []Token{
				{Text, "hello{world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello?",
			items: []Token{
				{Text, "hello"},
				{Single, "?"},
				{EOF, ""},
			},
		},
		{
			pattern: "hellof*",
			items: []Token{
				{Text, "hellof"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello**",
			items: []Token{
				{Text, "hello"},
				{Super, "**"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日-語]",
			items: []Token{
				{RangeOpen, "["},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "{a,b}",
			items: []Token{
				{TermsOpen, "{"},
				{Text, "a"},
				{TermSeparator, ","},
				{Text, "b"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{z,ab}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "z"},
				{TermSeparator, ","},
				{Text, "ab"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "{[!日-語],*,?,{a,b,\\c}}",
			items: []Token{
				{TermsOpen, "{"},
				{RangeOpen, "["},
				{Not, "!"},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{TermSeparator, ","},
				{Any, "*"},
				{TermSeparator, ","},
				{Single, "?"},
				{TermSeparator, ","},
				{TermsOpen, "{"},
				{Text, "a"},
				{TermSeparator, ","},
				{Text, "b"},
				{TermSeparator, ","},
				{Text, "c"},
				{TermsClose, "}"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
	} {
		lexer := NewLexer(test.pattern)
		for i, exp := range test.items {
			act := lexer.Next()
			if act.Type != exp.Type {
				t.Errorf("#%d %q: wrong %d-th item type: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Type, act.Type, exp, act)
			}
			if act.Data != exp.Data {
				t.Errorf("#%d %q: wrong %d-th item contents: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Data, act.Data, exp, act)
			}
		}
	}
}

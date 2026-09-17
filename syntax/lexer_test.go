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
		{
			pattern: "@(a|b)",
			items: []Token{
				{GroupOpen, "@("},
				{Text, "a"},
				{GroupSeparator, "|"},
				{Text, "b"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			pattern: "*(a)?(b)+(c)!(d)",
			items: []Token{
				{GroupOpen, "*("},
				{Text, "a"},
				{GroupClose, ")"},
				{GroupOpen, "?("},
				{Text, "b"},
				{GroupClose, ")"},
				{GroupOpen, "+("},
				{Text, "c"},
				{GroupClose, ")"},
				{GroupOpen, "!("},
				{Text, "d"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			// Outside of a group `(`, `)` and `|` are plain text, and so is
			// an operator character with no `(` right after it.
			pattern: "a|b)c",
			items: []Token{
				{Text, "a|b)c"},
				{EOF, ""},
			},
		},
		{
			pattern: "(a)",
			items: []Token{
				{Text, "(a)"},
				{EOF, ""},
			},
		},
		{
			pattern: "!@(a)",
			items: []Token{
				{Text, "!"},
				{GroupOpen, "@("},
				{Text, "a"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			// The `(` binds to the character right before it.
			pattern: "**(a)",
			items: []Token{
				{Any, "*"},
				{GroupOpen, "*("},
				{Text, "a"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			pattern: "***(a)",
			items: []Token{
				{Super, "**"},
				{GroupOpen, "*("},
				{Text, "a"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			pattern: "\\@(a)",
			items: []Token{
				{Text, "@(a)"},
				{EOF, ""},
			},
		},
		{
			pattern: "@(a\\|b)",
			items: []Token{
				{GroupOpen, "@("},
				{Text, "a|b"},
				{GroupClose, ")"},
				{EOF, ""},
			},
		},
		{
			// The innermost open construct decides which characters are
			// special: `,` and `}` inside a `{...}`, `|` and `)` inside a
			// group.
			pattern: "{@(a,b),c}",
			items: []Token{
				{TermsOpen, "{"},
				{GroupOpen, "@("},
				{Text, "a,b"},
				{GroupClose, ")"},
				{TermSeparator, ","},
				{Text, "c"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
		{
			pattern: "@(a|{b,c|d})",
			items: []Token{
				{GroupOpen, "@("},
				{Text, "a"},
				{GroupSeparator, "|"},
				{TermsOpen, "{"},
				{Text, "b"},
				{TermSeparator, ","},
				{Text, "c|d"},
				{TermsClose, "}"},
				{GroupClose, ")"},
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

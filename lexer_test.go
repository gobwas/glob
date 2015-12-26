package glob

import (
	"testing"
)

func TestLexGood(t *testing.T) {
	for _, test := range []struct {
		pattern string
		items   []item
	}{
		{
			pattern: "hello",
			items: []item{
				item{item_text, "hello"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "hello?",
			items: []item{
				item{item_text, "hello"},
				item{item_single, "?"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "hello*",
			items: []item{
				item{item_text, "hello"},
				item{item_any, "*"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "hello**",
			items: []item{
				item{item_text, "hello"},
				item{item_any, "*"},
				item{item_any, "*"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "[日-語]",
			items: []item{
				item{item_range_open, "["},
				item{item_range_lo, "日"},
				item{item_range_minus, "-"},
				item{item_range_hi, "語"},
				item{item_range_close, "]"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []item{
				item{item_range_open, "["},
				item{item_range_not, "!"},
				item{item_range_lo, "日"},
				item{item_range_minus, "-"},
				item{item_range_hi, "語"},
				item{item_range_close, "]"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []item{
				item{item_range_open, "["},
				item{item_range_chars, "日本語"},
				item{item_range_close, "]"},
				item{item_eof, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []item{
				item{item_range_open, "["},
				item{item_range_not, "!"},
				item{item_range_chars, "日本語"},
				item{item_range_close, "]"},
				item{item_eof, ""},
			},
		},
	} {
		lexer := newLexer(test.pattern)
		for _, exp := range test.items {
			act := lexer.nextItem()
			if act.t != exp.t {
				t.Errorf("wrong item type: exp: %v; act: %v (%s vs %s)", exp.t, act.t, exp, act)
				break
			}
			if act.s != exp.s {
				t.Errorf("wrong item contents: exp: %q; act: %q (%s vs %s)", exp.s, act.s, exp, act)
				break
			}
		}
	}
}

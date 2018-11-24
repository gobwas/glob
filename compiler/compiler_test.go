package compiler

import (
	"reflect"
	"testing"

	"github.com/gobwas/glob/match"
	"github.com/gobwas/glob/syntax/ast"
)

var separators = []rune{'.'}

func TestCompiler(t *testing.T) {
	for id, test := range []struct {
		ast    *ast.Node
		result match.Matcher
		sep    []rune
	}{
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			sep:    separators,
			result: match.NewAny(separators),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewSuper(),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
			),
			result: match.NewSuper(),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSingle, nil),
			),
			sep:    separators,
			result: match.NewSingle(separators),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			result: match.NewRange('a', 'z', true),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindList, ast.List{
					Chars: "abc",
					Not:   true,
				}),
			),
			result: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(separators),
			}),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			result: match.NewMin(3),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewAny(separators),
				nil,
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"/"}),
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindText, ast.Text{"z"}),
					ast.NewNode(ast.KindText, ast.Text{"ab"}),
				),
				ast.NewNode(ast.KindSuper, nil),
			),
			sep: separators,
			result: match.NewTree(
				match.NewText("/"),
				nil,
				match.NewTree(
					match.MustIndexedAnyOf(
						match.NewText("z"),
						match.NewText("ab"),
					),
					nil,
					match.NewSuper(),
				),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewSingle(separators),
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewSuper(),
				nil,
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewSuffix("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewPrefix("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"def"}),
			),
			result: match.NewPrefixSuffix("abc", "def"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewContains("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			sep: separators,
			result: match.NewTree(
				match.NewText("abc"),
				match.NewAny(separators),
				match.NewAny(separators),
			),
		},
		{
			// TODO: THIS!
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			result: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindAnyOf, nil,
							ast.NewNode(ast.KindPattern, nil,
								ast.NewNode(ast.KindText, ast.Text{"abc"}),
							),
						),
					),
				),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindSingle, nil),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "def"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
					),
				),
			),
			result: match.NewTree(
				match.NewText("abc"),
				nil,
				match.NewAnyOf(
					match.NewSingle(nil),
					match.NewList([]rune{'d', 'e', 'f'}, false),
					match.NewNothing(),
				),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'z'}),
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'x', Not: true}),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewRange('a', 'z', false),
					match.NewRange('a', 'x', true),
				}),
				nil,
				match.NewSuper(),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "abc"}),
						ast.NewNode(ast.KindText, ast.Text{"ghi"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "def"}),
						ast.NewNode(ast.KindText, ast.Text{"ghi"}),
					),
				),
			),
			result: match.NewRow([]match.MatchIndexSizer{
				match.NewText("abc"),
				match.MustIndexedSizedAnyOf(
					match.NewList([]rune{'a', 'b', 'c'}, false),
					match.NewList([]rune{'d', 'e', 'f'}, false),
				),
				match.NewText("ghi"),
			}),
		},
	} {
		m, err := Compile(test.ast, test.sep)
		if err != nil {
			t.Errorf("compilation error: %s", err)
			continue
		}

		if !reflect.DeepEqual(m, test.result) {
			t.Errorf("[%d] Compile():\nexp: %#v\nact: %#v\n\ngraphviz:\nexp:\n%s\nact:\n%s\n", id, test.result, m, match.Graphviz("", test.result.(match.Matcher)), match.Graphviz("", m.(match.Matcher)))
			continue
		}
	}
}

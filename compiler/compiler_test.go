package compiler

import (
	"reflect"
	"testing"

	"github.com/gobwas/glob/match"
	"github.com/gobwas/glob/syntax/ast"
)

var separators = []rune{'.'}

func TestCompiler(t *testing.T) {
	for _, test := range []struct {
		name string
		ast  *ast.Node
		exp  match.Matcher
		sep  []rune
	}{
		{
			// #0
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #1
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			sep: separators,
			exp: match.NewAny(separators),
		},
		{
			// #2
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #3
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #4
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewSingle(separators),
		},
		{
			// #5
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			exp: match.NewRange('a', 'z', true),
		},
		{
			// #6
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindList, ast.List{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			// #7
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(separators),
			}),
		},
		{
			// #8
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			exp: match.NewMin(3),
		},
		{
			// #9
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewAny(separators),
				match.Nothing{},
			),
		},
		{
			// #10
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"/"}),
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindText, ast.Text{"z"}),
					ast.NewNode(ast.KindText, ast.Text{"ab"}),
				),
				ast.NewNode(ast.KindSuper, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewText("/"),
				match.Nothing{},
				match.NewTree(
					match.MustIndexedAnyOf(
						match.NewText("z"),
						match.NewText("ab"),
					),
					match.Nothing{},
					match.NewSuper(),
				),
			),
		},
		{
			// #11
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewSingle(separators),
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewSuper(),
				match.Nothing{},
			),
		},
		{
			// #12
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			exp: match.NewSuffix("abc"),
		},
		{
			// #13
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
			),
			exp: match.NewPrefix("abc"),
		},
		{
			// #14
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"def"}),
			),
			exp: match.NewPrefixSuffix("abc", "def"),
		},
		{
			// #15
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			exp: match.NewContains("abc"),
		},
		{
			// #16
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewAny(separators),
				match.NewAny(separators),
			),
		},
		{
			// #17
			// pattern: "**?abc**?"
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			// #18
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #19
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
			exp: match.NewText("abc"),
		},
		{
			// #20
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
			exp: match.NewTree(
				match.NewText("abc"),
				match.Nothing{},
				match.NewAnyOf(
					match.NewSingle(nil),
					match.NewList([]rune{'d', 'e', 'f'}, false),
					match.NewNothing(),
				),
			),
		},
		{
			// #21
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'z'}),
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'x', Not: true}),
				ast.NewNode(ast.KindAny, nil),
			),
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewRange('a', 'z', false),
					match.NewRange('a', 'x', true),
				}),
				match.Nothing{},
				match.NewSuper(),
			),
		},
		{
			// #22
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
			exp: match.NewRow([]match.MatchIndexSizer{
				match.NewText("abc"),
				match.MustIndexedSizedAnyOf(
					match.NewList([]rune{'a', 'b', 'c'}, false),
					match.NewList([]rune{'d', 'e', 'f'}, false),
				),
				match.NewText("ghi"),
			}),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			act, err := Compile(test.ast, test.sep)
			if err != nil {
				t.Fatalf("compilation error: %s", err)
			}
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf(
					"Compile():\nact: %#v\nexp: %#v\n\ngraphviz:\n%s\n%s\n",
					act, test.exp,
					match.Graphviz("act", act.(match.Matcher)),
					match.Graphviz("exp", test.exp.(match.Matcher)),
				)
			}
		})
	}
}

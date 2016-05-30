package ast

import (
	"github.com/gobwas/glob/syntax"
	"reflect"
	"testing"
)

type stubLexer struct {
	tokens []syntax.Token
	pos    int
}

func (s *stubLexer) Next() (ret syntax.Token) {
	if s.pos == len(s.tokens) {
		return syntax.Token{syntax.EOF, ""}
	}
	ret = s.tokens[s.pos]
	s.pos++
	return
}

func TestParseString(t *testing.T) {
	for id, test := range []struct {
		tokens []syntax.Token
		tree   Node
	}{
		{
			//pattern: "abc",
			tokens: []syntax.Token{
				syntax.Token{syntax.Text, "abc"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, &Text{Text: "abc"}),
			),
		},
		{
			//pattern: "a*c",
			tokens: []syntax.Token{
				syntax.Token{syntax.Text, "a"},
				syntax.Token{syntax.Any, "*"},
				syntax.Token{syntax.Text, "c"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, &Text{Text: "a"}),
				NewNode(KindAny, nil),
				NewNode(KindText, &Text{Text: "c"}),
			),
		},
		{
			//pattern: "a**c",
			tokens: []syntax.Token{
				syntax.Token{syntax.Text, "a"},
				syntax.Token{syntax.Super, "**"},
				syntax.Token{syntax.Text, "c"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, &Text{Text: "a"}),
				NewNode(KindSuper, nil),
				NewNode(KindText, &Text{Text: "c"}),
			),
		},
		{
			//pattern: "a?c",
			tokens: []syntax.Token{
				syntax.Token{syntax.Text, "a"},
				syntax.Token{syntax.Single, "?"},
				syntax.Token{syntax.Text, "c"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, &Text{Text: "a"}),
				NewNode(KindSingle, nil),
				NewNode(KindText, &Text{Text: "c"}),
			),
		},
		{
			//pattern: "[!a-z]",
			tokens: []syntax.Token{
				syntax.Token{syntax.RangeOpen, "["},
				syntax.Token{syntax.Not, "!"},
				syntax.Token{syntax.RangeLo, "a"},
				syntax.Token{syntax.RangeBetween, "-"},
				syntax.Token{syntax.RangeHi, "z"},
				syntax.Token{syntax.RangeClose, "]"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindRange, &Range{Lo: 'a', Hi: 'z', Not: true}),
			),
		},
		{
			//pattern: "[az]",
			tokens: []syntax.Token{
				syntax.Token{syntax.RangeOpen, "["},
				syntax.Token{syntax.Text, "az"},
				syntax.Token{syntax.RangeClose, "]"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindList, &List{Chars: "az"}),
			),
		},
		{
			//pattern: "{a,z}",
			tokens: []syntax.Token{
				syntax.Token{syntax.TermsOpen, "{"},
				syntax.Token{syntax.Text, "a"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.Text, "z"},
				syntax.Token{syntax.TermsClose, "}"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, &Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, &Text{Text: "z"}),
					),
				),
			),
		},
		{
			//pattern: "/{z,ab}*",
			tokens: []syntax.Token{
				syntax.Token{syntax.Text, "/"},
				syntax.Token{syntax.TermsOpen, "{"},
				syntax.Token{syntax.Text, "z"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.Text, "ab"},
				syntax.Token{syntax.TermsClose, "}"},
				syntax.Token{syntax.Any, "*"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, &Text{Text: "/"}),
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, &Text{Text: "z"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, &Text{Text: "ab"}),
					),
				),
				NewNode(KindAny, nil),
			),
		},
		{
			//pattern: "{a,{x,y},?,[a-z],[!qwe]}",
			tokens: []syntax.Token{
				syntax.Token{syntax.TermsOpen, "{"},
				syntax.Token{syntax.Text, "a"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.TermsOpen, "{"},
				syntax.Token{syntax.Text, "x"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.Text, "y"},
				syntax.Token{syntax.TermsClose, "}"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.Single, "?"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.RangeOpen, "["},
				syntax.Token{syntax.RangeLo, "a"},
				syntax.Token{syntax.RangeBetween, "-"},
				syntax.Token{syntax.RangeHi, "z"},
				syntax.Token{syntax.RangeClose, "]"},
				syntax.Token{syntax.Separator, ","},
				syntax.Token{syntax.RangeOpen, "["},
				syntax.Token{syntax.Not, "!"},
				syntax.Token{syntax.Text, "qwe"},
				syntax.Token{syntax.RangeClose, "]"},
				syntax.Token{syntax.TermsClose, "}"},
				syntax.Token{syntax.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, &Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindAnyOf, nil,
							NewNode(KindPattern, nil,
								NewNode(KindText, &Text{Text: "x"}),
							),
							NewNode(KindPattern, nil,
								NewNode(KindText, &Text{Text: "y"}),
							),
						),
					),
					NewNode(KindPattern, nil,
						NewNode(KindSingle, nil),
					),
					NewNode(KindPattern, nil,
						NewNode(KindRange, &Range{Lo: 'a', Hi: 'z', Not: false}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindList, &List{Chars: "qwe", Not: true}),
					),
				),
			),
		},
	} {
		lexer := &stubLexer{tokens: test.tokens}
		result, err := Parse(lexer)
		if err != nil {
			t.Errorf("[%d] unexpected error: %s", id, err)
		}
		if !reflect.DeepEqual(test.tree, result) {
			t.Errorf("[%d] Parse():\nact:\t%s\nexp:\t%s\n", id, result, test.tree)
		}
	}
}

type kv struct {
	kind  Kind
	value interface{}
}

type visitor struct {
	visited []kv
}

func (v *visitor) Visit(n Node) Visitor {
	v.visited = append(v.visited, kv{n.Kind(), n.Value()})
	return v
}

func TestWalkTree(t *testing.T) {

	for i, test := range []struct {
		tree    *Node
		visited []kv
	}{
		{
			tree: NewNode(KindPattern, nil,
				NewNode(KindSingle, nil),
			),
			visited: []kv{
				kv{KindPattern, nil},
				kv{KindSingle, nil},
			},
		},
	} {
		v := &visitor{}
		Walk(v, test.tree)

		if !reflect.DeepEqual(test.visited, v.visited) {
			t.Errorf("[%d] unexpected result of Walk():\nvisited:\t%v\nwant:\t\t%v", i, v.visited, test.visited)
		}
	}
}

package parser

import (
	"fmt"
	"github.com/gobwas/glob/lexer"
	"reflect"
	"testing"
)

type stubLexer struct {
	tokens []lexer.Token
	pos    int
}

func (s *stubLexer) Next() (ret lexer.Token) {
	if s.pos == len(s.tokens) {
		return lexer.Token{lexer.EOF, ""}
	}
	ret = s.tokens[s.pos]
	s.pos++
	return
}

func TestParseString(t *testing.T) {
	for id, test := range []struct {
		tokens []lexer.Token
		tree   Node
	}{
		{
			//pattern: "abc",
			tokens: []lexer.Token{
				lexer.Token{lexer.Text, "abc"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&TextNode{Text: "abc"},
					},
				},
			},
		},
		{
			//pattern: "a*c",
			tokens: []lexer.Token{
				lexer.Token{lexer.Text, "a"},
				lexer.Token{lexer.Any, "*"},
				lexer.Token{lexer.Text, "c"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&TextNode{Text: "a"},
						&AnyNode{},
						&TextNode{Text: "c"},
					},
				},
			},
		},
		{
			//pattern: "a**c",
			tokens: []lexer.Token{
				lexer.Token{lexer.Text, "a"},
				lexer.Token{lexer.Super, "**"},
				lexer.Token{lexer.Text, "c"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&TextNode{Text: "a"},
						&SuperNode{},
						&TextNode{Text: "c"},
					},
				},
			},
		},
		{
			//pattern: "a?c",
			tokens: []lexer.Token{
				lexer.Token{lexer.Text, "a"},
				lexer.Token{lexer.Single, "?"},
				lexer.Token{lexer.Text, "c"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&TextNode{Text: "a"},
						&SingleNode{},
						&TextNode{Text: "c"},
					},
				},
			},
		},
		{
			//pattern: "[!a-z]",
			tokens: []lexer.Token{
				lexer.Token{lexer.RangeOpen, "["},
				lexer.Token{lexer.Not, "!"},
				lexer.Token{lexer.RangeLo, "a"},
				lexer.Token{lexer.RangeBetween, "-"},
				lexer.Token{lexer.RangeHi, "z"},
				lexer.Token{lexer.RangeClose, "]"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&RangeNode{Lo: 'a', Hi: 'z', Not: true},
					},
				},
			},
		},
		{
			//pattern: "[az]",
			tokens: []lexer.Token{
				lexer.Token{lexer.RangeOpen, "["},
				lexer.Token{lexer.Text, "az"},
				lexer.Token{lexer.RangeClose, "]"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&ListNode{Chars: "az"},
					},
				},
			},
		},
		{
			//pattern: "{a,z}",
			tokens: []lexer.Token{
				lexer.Token{lexer.TermsOpen, "{"},
				lexer.Token{lexer.Text, "a"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.Text, "z"},
				lexer.Token{lexer.TermsClose, "}"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&AnyOfNode{node: node{children: []Node{
							&PatternNode{
								node: node{children: []Node{
									&TextNode{Text: "a"},
								}},
							},
							&PatternNode{
								node: node{children: []Node{
									&TextNode{Text: "z"},
								}},
							},
						}}},
					},
				},
			},
		},
		{
			//pattern: "/{z,ab}*",
			tokens: []lexer.Token{
				lexer.Token{lexer.Text, "/"},
				lexer.Token{lexer.TermsOpen, "{"},
				lexer.Token{lexer.Text, "z"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.Text, "ab"},
				lexer.Token{lexer.TermsClose, "}"},
				lexer.Token{lexer.Any, "*"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&TextNode{Text: "/"},
						&AnyOfNode{node: node{children: []Node{
							&PatternNode{
								node: node{children: []Node{
									&TextNode{Text: "z"},
								}},
							},
							&PatternNode{
								node: node{children: []Node{
									&TextNode{Text: "ab"},
								}},
							},
						}}},
						&AnyNode{},
					},
				},
			},
		},
		{
			//pattern: "{a,{x,y},?,[a-z],[!qwe]}",
			tokens: []lexer.Token{
				lexer.Token{lexer.TermsOpen, "{"},
				lexer.Token{lexer.Text, "a"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.TermsOpen, "{"},
				lexer.Token{lexer.Text, "x"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.Text, "y"},
				lexer.Token{lexer.TermsClose, "}"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.Single, "?"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.RangeOpen, "["},
				lexer.Token{lexer.RangeLo, "a"},
				lexer.Token{lexer.RangeBetween, "-"},
				lexer.Token{lexer.RangeHi, "z"},
				lexer.Token{lexer.RangeClose, "]"},
				lexer.Token{lexer.Separator, ","},
				lexer.Token{lexer.RangeOpen, "["},
				lexer.Token{lexer.Not, "!"},
				lexer.Token{lexer.Text, "qwe"},
				lexer.Token{lexer.RangeClose, "]"},
				lexer.Token{lexer.TermsClose, "}"},
				lexer.Token{lexer.EOF, ""},
			},
			tree: &PatternNode{
				node: node{
					children: []Node{
						&AnyOfNode{node: node{children: []Node{
							&PatternNode{
								node: node{children: []Node{
									&TextNode{Text: "a"},
								}},
							},
							&PatternNode{
								node: node{children: []Node{
									&AnyOfNode{node: node{children: []Node{
										&PatternNode{
											node: node{children: []Node{
												&TextNode{Text: "x"},
											}},
										},
										&PatternNode{
											node: node{children: []Node{
												&TextNode{Text: "y"},
											}},
										},
									}}},
								}},
							},
							&PatternNode{
								node: node{children: []Node{
									&SingleNode{},
								}},
							},
							&PatternNode{
								node: node{
									children: []Node{
										&RangeNode{Lo: 'a', Hi: 'z', Not: false},
									},
								},
							},
							&PatternNode{
								node: node{
									children: []Node{
										&ListNode{Chars: "qwe", Not: true},
									},
								},
							},
						}}},
					},
				},
			},
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

const abstractNodeImpl = "nodeImpl"

func nodeEqual(a, b Node) error {
	if (a == nil || b == nil) && a != b {
		return fmt.Errorf("nodes are not equal: exp %s, act %s", a, b)
	}

	aValue, bValue := reflect.Indirect(reflect.ValueOf(a)), reflect.Indirect(reflect.ValueOf(b))
	aType, bType := aValue.Type(), bValue.Type()
	if aType != bType {
		return fmt.Errorf("nodes are not equal: exp %s, act %s", aValue.Type(), bValue.Type())
	}

	for i := 0; i < aType.NumField(); i++ {
		var eq bool

		f := aType.Field(i).Name
		if f == abstractNodeImpl {
			continue
		}

		af, bf := aValue.FieldByName(f), bValue.FieldByName(f)

		switch af.Kind() {
		case reflect.String:
			eq = af.String() == bf.String()
		case reflect.Bool:
			eq = af.Bool() == bf.Bool()
		default:
			eq = fmt.Sprint(af) == fmt.Sprint(bf)
		}

		if !eq {
			return fmt.Errorf("nodes<%s> %q fields are not equal: exp %q, act %q", aType, f, af, bf)
		}
	}

	for i, aDesc := range a.Children() {
		if len(b.Children())-1 < i {
			return fmt.Errorf("node does not have enough children (got %d children, wanted %d-th token)", len(b.Children()), i)
		}

		bDesc := b.Children()[i]

		if err := nodeEqual(aDesc, bDesc); err != nil {
			return err
		}
	}

	return nil
}

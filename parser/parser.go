package parser

import (
	"errors"
	"fmt"
	"github.com/gobwas/glob/lexer"
	"unicode/utf8"
)

type Lexer interface {
	Next() lexer.Token
}

type parseFn func(Node, Lexer) (parseFn, Node, error)

func Parse(lexer Lexer) (*PatternNode, error) {
	var parser parseFn

	root := &PatternNode{}

	var (
		tree Node
		err  error
	)
	for parser, tree = parserMain, root; parser != nil; {
		parser, tree, err = parser(tree, lexer)
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

func parserMain(tree Node, lex Lexer) (parseFn, Node, error) {
	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, nil

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Text:
			return parserMain, tree.append(&TextNode{Text: token.Raw}), nil

		case lexer.Any:
			return parserMain, tree.append(&AnyNode{}), nil

		case lexer.Super:
			return parserMain, tree.append(&SuperNode{}), nil

		case lexer.Single:
			return parserMain, tree.append(&SingleNode{}), nil

		case lexer.RangeOpen:
			return parserRange, tree, nil

		case lexer.TermsOpen:
			return parserMain, tree.append(&AnyOfNode{}).append(&PatternNode{}), nil

		case lexer.Separator:
			return parserMain, tree.Parent().append(&PatternNode{}), nil

		case lexer.TermsClose:
			return parserMain, tree.Parent().Parent(), nil

		default:
			return nil, tree, fmt.Errorf("unexpected token: %s", token)
		}
	}
	return nil, tree, fmt.Errorf("unknown error")
}

func parserRange(tree Node, lex Lexer) (parseFn, Node, error) {
	var (
		not   bool
		lo    rune
		hi    rune
		chars string
	)
	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, errors.New("unexpected end")

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Not:
			not = true

		case lexer.RangeLo:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, fmt.Errorf("unexpected length of lo character")
			}
			lo = r

		case lexer.RangeBetween:
			//

		case lexer.RangeHi:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, fmt.Errorf("unexpected length of lo character")
			}

			hi = r

			if hi < lo {
				return nil, tree, fmt.Errorf("hi character '%s' should be greater than lo '%s'", string(hi), string(lo))
			}

		case lexer.Text:
			chars = token.Raw

		case lexer.RangeClose:
			isRange := lo != 0 && hi != 0
			isChars := chars != ""

			if isChars == isRange {
				return nil, tree, fmt.Errorf("could not parse range")
			}

			if isRange {
				tree = tree.append(&RangeNode{
					Lo:  lo,
					Hi:  hi,
					Not: not,
				})
			} else {
				tree = tree.append(&ListNode{
					Chars: chars,
					Not:   not,
				})
			}

			return parserMain, tree, nil
		}
	}
}

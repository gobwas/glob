package glob

import (
	"errors"
	"fmt"
	"github.com/gobwas/glob/match"
)

func parseAll(source, separators string) ([]token, error) {
	lexer := newLexer(source)

	var tokens []token
	for parser := parserMain; parser != nil; {
		var err error
		tokens, parser, err = parser(lexer, separators)
		if err != nil {
			return nil, err
		}
	}

	return tokens, nil
}

type parseFn func(*lexer, string) ([]token, parseFn, error)

func parserMain(lexer *lexer, separators string) ([]token, parseFn, error) {
	var (
		prev   *token
		tokens []token
	)

	for item := lexer.nextItem(); ; {
		var t token

		if item.t == item_eof {
			break
		}

		switch item.t {
		case item_eof:
			return tokens, nil, nil

		case item_error:
			return nil, nil, errors.New(item.s)

		case item_text:
			t = token{match.Raw{item.s}, item.s}

		case item_any:
			if prev != nil && prev.matcher.Kind() == match.KindMultipleSeparated {
				// remove simple any and replace it with super_any
				tokens = tokens[:len(tokens)-1]
				t = token{match.Any{""}, item.s}
			} else {
				t = token{match.Any{separators}, item.s}
			}

		case item_single:
			t = token{match.Single{separators}, item.s}

		case item_range_open:
			return tokens, parserRange, nil
		}

		prev = &t
	}

	return tokens, nil, nil
}

func parserRange(lexer *lexer, separators string) ([]token, parseFn, error) {
	var (
		not   bool
		lo    rune
		hi    rune
		chars string
	)

	for item := lexer.nextItem(); ; {
		switch item.t {
		case item_eof:
			return nil, nil, errors.New("unexpected end")

		case item_error:
			return nil, nil, errors.New(item.s)

		case item_range_not:
			not = true

		case item_range_lo:
			r := []rune(item.s)
			if len(r) != 1 {
				return nil, nil, fmt.Errorf("unexpected length of lo character")
			}

			lo = r[0]

		case item_range_minus:
			//

		case item_range_hi:
			r := []rune(item.s)
			if len(r) != 1 {
				return nil, nil, fmt.Errorf("unexpected length of hi character")
			}

			if hi < lo {
				return nil, nil, fmt.Errorf("hi character should be greater than lo")
			}

			hi = r[0]

		case item_range_chars:
			chars = item.s

		case item_range_close:
			isRange := lo != 0 && hi != 0
			isChars := chars == ""

			if !(isChars != isRange) {
				return nil, nil, fmt.Errorf("parse error: unexpected lo, hi, chars in range")
			}

			if isRange {
				return []token{token{match.Between{lo, hi, not}, ""}}, parserMain, nil
			} else {
				if len(chars) == 0 {
					return nil, nil, fmt.Errorf("chars range should not be empty")
				}

				return []token{token{match.RangeList{chars, not}, ""}}, parserMain, nil
			}
		}
	}
}

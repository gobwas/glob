package glob

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gobwas/glob/runes"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	char_any           = '*'
	char_comma         = ','
	char_single        = '?'
	char_escape        = '\\'
	char_range_open    = '['
	char_range_close   = ']'
	char_terms_open    = '{'
	char_terms_close   = '}'
	char_range_not     = '!'
	char_range_between = '-'
)

var specials = []byte{
	char_any,
	char_single,
	char_escape,
	char_range_open,
	char_range_close,
	char_terms_open,
	char_terms_close,
}

func special(c byte) bool {
	return bytes.IndexByte(specials, c) != -1
}

var eof rune = 0

type itemType int

const (
	item_eof itemType = iota
	item_error
	item_text
	item_char
	item_any
	item_super
	item_single
	item_not
	item_separator
	item_range_open
	item_range_close
	item_range_lo
	item_range_hi
	item_range_between
	item_terms_open
	item_terms_close
)

func (i itemType) String() string {
	switch i {
	case item_eof:
		return "eof"

	case item_error:
		return "error"

	case item_text:
		return "text"

	case item_char:
		return "char"

	case item_any:
		return "any"

	case item_super:
		return "super"

	case item_single:
		return "single"

	case item_not:
		return "not"

	case item_separator:
		return "separator"

	case item_range_open:
		return "range_open"

	case item_range_close:
		return "range_close"

	case item_range_lo:
		return "range_lo"

	case item_range_hi:
		return "range_hi"

	case item_range_between:
		return "range_between"

	case item_terms_open:
		return "terms_open"

	case item_terms_close:
		return "terms_close"

	default:
		return "undef"
	}
}

type item struct {
	t itemType
	s string
}

func (i item) String() string {
	return fmt.Sprintf("%v<%s>", i.t, i.s)
}

type stubLexer struct {
	Items []item
	pos   int
}

func (s *stubLexer) nextItem() (ret item) {
	if s.pos == len(s.Items) {
		return item{item_eof, ""}
	}
	ret = s.Items[s.pos]
	s.pos++
	return
}

type lexer struct {
	data       string
	start      int
	pos        int
	current    rune
	items      []item
	termsLevel int
	r          *bufio.Reader
}

func newLexer(source string) *lexer {
	l := &lexer{
		r:    bufio.NewReader(strings.NewReader(source)),
		data: source,
	}
	return l
}

func (l *lexer) shiftItem() (ret item) {
	ret, l.items = l.items[0], l.items[1:]
	return
}

func (l *lexer) pushItem(i item) {
	l.items = append(l.items, i)
}

func (l *lexer) hasItem() bool {
	return len(l.items) > 0
}

func (l *lexer) peekRune() rune {
	r, _ := utf8.DecodeRuneInString(l.data[l.start:])
	return r
}

func (l *lexer) inTerms() bool {
	return l.termsLevel > 0
}

func (l *lexer) termsEnter() {
	l.termsLevel++
}

func (l *lexer) termsLeave() {
	l.termsLevel--
}

func (l *lexer) nextItem() item {
	if l.hasItem() {
		return l.shiftItem()
	}

	r, _, err := l.r.ReadRune()
	if err != nil {
		switch err {
		case io.EOF:
			return item{item_eof, ""}
		default:
			return item{item_error, err.Error()}
		}
	}

	switch r {
	case char_terms_open:
		l.termsEnter()
		return item{item_terms_open, string(r)}

	case char_comma:
		if l.inTerms() {
			return item{item_separator, string(r)}
		}

	case char_terms_close:
		if l.inTerms() {
			l.termsLeave()
			return item{item_terms_close, string(r)}
		}

	case char_range_open:
		l.fetchRange()
		return item{item_range_open, string(r)}

	case char_single:
		return item{item_single, string(r)}

	case char_any:
		b, err := l.r.Peek(1)
		if err == nil && b[0] == char_any {
			l.r.ReadRune()
			return item{item_super, string(r) + string(r)}
		}
		return item{item_any, string(r)}
	}

	l.r.UnreadRune()
	breakers := []rune{char_single, char_any, char_range_open, char_terms_open}
	if l.inTerms() {
		breakers = append(breakers, char_terms_close, char_comma)
	}
	l.fetchText(breakers)

	return l.nextItem()
}

func (l *lexer) fetchRange() {
	var wantHi bool
	var wantClose bool
	var seenNot bool
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			l.pushItem(item{item_error, err.Error()})
			return
		}

		if wantClose {
			if r != char_range_close {
				l.pushItem(item{item_error, "expecting close range character"})
			} else {
				l.pushItem(item{item_range_close, string(r)})
			}
			return
		}

		if wantHi {
			l.pushItem(item{item_range_hi, string(r)})
			wantClose = true
			continue
		}

		if !seenNot && r == char_range_not {
			l.pushItem(item{item_not, string(r)})
			seenNot = true
			continue
		}

		b, err := l.r.Peek(1)
		if err == nil && b[0] == char_range_between {
			l.pushItem(item{item_range_lo, string(r)})
			l.r.ReadRune()
			l.pushItem(item{item_range_between, string(char_range_between)})
			wantHi = true
			continue
		}

		l.r.UnreadRune()
		l.fetchText([]rune{char_range_close})
		wantClose = true
	}
}

func (l *lexer) fetchText(breakers []rune) {
	var data []rune
	var escaped bool

reading:
	for {
		r, _, err := l.r.ReadRune()
		if err != nil {
			break
		}

		if !escaped {
			if r == char_escape {
				escaped = true
				continue
			}

			if runes.IndexRune(breakers, r) != -1 {
				l.r.UnreadRune()
				break reading
			}
		}

		escaped = false
		data = append(data, r)
	}

	if len(data) > 0 {
		l.pushItem(item{item_text, string(data)})
	}
}

package glob

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var eof int = 0

type stateFn func(*lexer) stateFn

type itemType int

const (
	item_eof itemType = iota
	item_error
	item_text
	item_any
	item_single
	item_range_open
	item_range_not
	item_range_minus
	item_range_close
)

type item struct {
	t itemType
	s string
}

type lexer struct {
	input string
	start int
	pos   int
	width int
	runes int
	items chan item
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) read() (rune int) {
	if l.pos >= len(l.input) {
		return eof
	}

	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	l.runes++

	return
}

func (l *lexer) unread() {
	l.pos -= l.width
	l.runes--
}

func (l *lexer) ignore() {
	l.start = l.pos
	l.runes = 0
}

func (l *lexer) lookahead() int {
	r := l.read()
	l.unread()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.read()) != -1 {
		return true
	}
	l.unread()
	return false
}

func (l *lexer) acceptAll(valid string) {
	for strings.IndexRune(valid, l.read()) != -1 {
	}
	l.unread()
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
	l.runes = 0
	l.width = 0
}

func (l *lexer) flush(t itemType) {
	if l.pos > l.start {
		l.emit(t)
	}
}

func (l *lexer) errorf(format string, args ...interface{}) {
	l.emit(item{item_error, fmt.Sprintf(format, args...)})
}

func lex(source string) *lexer {
	l := &lexer{
		input: strings.NewReader(source),
		items: make(chan item),
	}

	go l.run()

	return l
}

func lexText(l *lexer) stateFn {
	for {
		switch l.input[l.pos] {
		case escape:
			if l.read() == eof {
				l.errorf("unclosed '%s' character", string(escape))
				return nil
			}
		case single:
			l.flush(item_text)
			return lexSingle
		case any:
			l.flush(item_text)
			return lexAny
		case range_open:
			l.flush(item_text)
			return lexRangeOpen
		}

		if l.read() == eof {
			break
		}
	}

	if l.pos > l.start {
		l.emit(item_text)
	}

	l.emit(item_eof)

	return nil
}

func lexRangeOpen(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_range_open)
	return lexInsideRange
}

func lexInsideRange(l *lexer) stateFn {
	for {
		switch l.input[l.pos] {

		case inside_range_not:
			// only first char makes sense
			if l.pos == l.start {
				l.emit(item_range_not)
			}

		case inside_range_minus:
			if len(l.runes) != 1 {
				l.errorf("unexpected character '%s'", string(inside_range_minus))
				return nil
			}

			l.emit(item_text)

			l.pos += 1
			l.emit(item_range_minus)

			switch l.input[l.pos] {
			case eof, range_close:
				l.errorf("unexpected end of range: character is expected")
				return nil
			default:
				l.read()
				l.emit(item_text)
			}

			return lexText

		case range_close:
			l.flush(item_text)
			return lexRangeClose
		}

		if l.read() == eof {
			l.errorf("unclosed range construction")
			return nil
		}
	}
}

func lexAny(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_any)
	return lexText
}

func lexRangeHiLo(l *lexer) stateFn {

	l.emit(item_text)
	return lexRangeMinus

	l.pos += 1
	l.emit(item_range_minus)
	return lexInsideRange
}

func lexSingle(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_single)
	return lexText
}

func lexRangeClose(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_range_close)
	return lexText
}

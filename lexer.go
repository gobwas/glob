package glob

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var eof rune = 0

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
	item_range_lo
	item_range_minus
	item_range_hi
	item_range_chars
	item_range_close
)

func (i itemType) String() string {
	switch i {
	case item_eof:
		return "eof"

	case item_error:
		return "error"

	case item_text:
		return "text"

	case item_any:
		return "any"

	case item_single:
		return "single"

	case item_range_open:
		return "range_open"

	case item_range_not:
		return "range_not"

	case item_range_lo:
		return "range_lo"

	case item_range_minus:
		return "range_minus"

	case item_range_hi:
		return "range_hi"

	case item_range_chars:
		return "range_chars"

	case item_range_close:
		return "range_close"

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

type lexer struct {
	input string
	start int
	pos   int
	width int
	runes int
	state stateFn
	items chan item
}

func newLexer(source string) *lexer {
	l := &lexer{
		input: source,
		state: lexText,
		items: make(chan item, 5),
	}
	return l
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) read() (r rune) {
	if l.pos >= len(l.input) {
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	l.runes++

	return
}

func (l *lexer) unread() {
	l.pos -= l.width
	l.runes--
}

func (l *lexer) reset() {
	l.pos = l.start
	l.runes = 0
}

func (l *lexer) ignore() {
	l.start = l.pos
	l.runes = 0
}

func (l *lexer) lookahead() rune {
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

func (l *lexer) emitMaybe(t itemType) {
	if l.pos > l.start {
		l.emit(t)
	}
}

func (l *lexer) errorf(format string, args ...interface{}) {
	l.items <- item{item_error, fmt.Sprintf(format, args...)}
}

func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			if l.state == nil {
				return item{t: item_eof}
			}

			l.state = l.state(l)
		}
	}

	panic("something went wrong")
}

func lexText(l *lexer) stateFn {
	for {
		c := l.read()
		if c == eof {
			break
		}

		switch c {
		case escape:
			if l.read() == eof {
				l.errorf("unclosed '%s' character", string(escape))
				return nil
			}
		case single:
			l.unread()
			l.emitMaybe(item_text)
			return lexSingle
		case any:
			l.unread()
			l.emitMaybe(item_text)
			return lexAny
		case range_open:
			l.unread()
			l.emitMaybe(item_text)
			return lexRangeOpen
		}

	}

	if l.pos > l.start {
		l.emit(item_text)
	}

	l.emit(item_eof)

	return nil
}

func lexInsideRange(l *lexer) stateFn {
	for {
		c := l.read()
		if c == eof {
			l.errorf("unclosed range construction")
			return nil
		}

		switch c {
		case inside_range_not:
			// only first char makes sense
			if l.pos-l.width == l.start {
				l.emit(item_range_not)
			}

		case inside_range_minus:
			if l.runes != 2 {
				l.errorf("unexpected length of lo char inside range")
				return nil
			}

			l.reset()
			return lexRangeHiLo

		case range_close:
			l.unread()
			l.emitMaybe(item_range_chars)
			return lexRangeClose
		}
	}
}

func lexAny(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_any)
	return lexText
}

func lexRangeHiLo(l *lexer) stateFn {
	start := l.start

	for {
		c := l.read()
		if c == eof {
			l.errorf("unexpected end of input")
			return nil
		}

		switch c {
		case inside_range_minus:
			if l.runes != 1 {
				l.errorf("unexpected length of range: single character expected before minus")
				return nil
			}

			l.emit(item_range_minus)

		case range_close:
			l.unread()

			if l.runes != 1 {
				l.errorf("unexpected length of range: single character expected before close")
				return nil
			}

			l.emit(item_range_hi)
			return lexRangeClose

		default:
			if start != l.start {
				continue
			}

			if l.runes != 1 {
				l.errorf("unexpected length of range: single character expected at the begining")
				return nil
			}

			l.emit(item_range_lo)
		}
	}
}

func lexSingle(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_single)
	return lexText
}

func lexRangeOpen(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_range_open)
	return lexInsideRange
}

func lexRangeClose(l *lexer) stateFn {
	l.pos += 1
	l.emit(item_range_close)
	return lexText
}

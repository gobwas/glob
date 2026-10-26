// Package syntax implements the lexer of the glob pattern syntax. The parser
// lives in package glob; the syntax itself is described at [glob.Compile].
package syntax

import (
	"bytes"
	"fmt"
	"slices"
	"unicode/utf8"
)

// TokenType tells the kind of a [Token].
type TokenType int

const (
	// EOF marks the end of the input; the lexer returns it repeatedly.
	EOF TokenType = iota
	// Error carries an error message in Token.Data; the lexer keeps
	// returning it once it happened. Note that the lexer catches only the
	// errors local to a token (an invalid UTF-8 sequence, a malformed
	// character class): the structural ones, like an unclosed `{`, are for
	// the parser to detect.
	Error
	// Text is a run of literal characters, with the escapes resolved.
	Text
	// Any is the `*` wildcard.
	Any
	// Super is the `**` wildcard.
	Super
	// Single is the `?` wildcard.
	Single
	// Not is the `!` right after the `[` of a character class.
	Not
	// TermSeparator is the `,` between the alternatives of a `{...}` group.
	// Outside of a group a comma is a plain Text character.
	TermSeparator
	// RangeOpen and RangeClose are the `[` and `]` of a character class.
	// Between them the lexer produces either a Text token (a set of
	// characters, `[abc]`) or a RangeLo, RangeBetween, RangeHi triple (a
	// range, `[a-c]`), possibly preceded by Not.
	RangeOpen
	RangeClose
	RangeLo
	RangeHi
	RangeBetween
	// TermsOpen and TermsClose are the `{` and `}` of an alternatives group.
	TermsOpen
	TermsClose
	// GroupOpen is the opening of an extended group: the operator character
	// together with the `(` right after it, one of `?(`, `*(`, `+(`, `@(`
	// and `!(`; Token.Data carries both characters. GroupSeparator is the
	// `|` between the alternatives of a group and GroupClose is its `)`.
	// Outside of a group `|` and `)` are plain Text characters, and so is a
	// `(` not preceded by an operator character.
	GroupOpen
	GroupSeparator
	GroupClose
)

func (tt TokenType) String() string {
	switch tt {
	case EOF:
		return "eof"
	case Error:
		return "error"
	case Text:
		return "text"
	case Any:
		return "any"
	case Super:
		return "super"
	case Single:
		return "single"
	case Not:
		return "not"
	case TermSeparator:
		return "separator"
	case RangeOpen:
		return "range_open"
	case RangeClose:
		return "range_close"
	case RangeLo:
		return "range_lo"
	case RangeHi:
		return "range_hi"
	case RangeBetween:
		return "range_between"
	case TermsOpen:
		return "terms_open"
	case TermsClose:
		return "terms_close"
	case GroupOpen:
		return "group_open"
	case GroupSeparator:
		return "group_separator"
	case GroupClose:
		return "group_close"
	default:
		return "<unknown token type>"
	}
}

// Token is a lexeme of the pattern: its kind and the source text it was
// read from (or the error message for Error, the literal characters with
// the escapes resolved for Text).
type Token struct {
	Type TokenType
	Data string
}

func (t Token) String() string {
	return fmt.Sprintf("%v<%q>", t.Type, t.Data)
}

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
	char_group_open    = '('
	char_group_close   = ')'
	char_group_sep     = '|'
	char_group_plus    = '+'
	char_group_at      = '@'
	char_group_not     = '!'
)

var specials = []byte{
	char_any,
	char_single,
	char_escape,
	char_range_open,
	char_range_close,
	char_terms_open,
	char_terms_close,
	char_group_open,
	char_group_close,
}

// IsSpecial reports whether c is a glob meta character, that is, one that
// [glob.QuoteMeta] escapes. Note that `,`, `|`, `!` and `-` are not among
// them: they are special only inside `{...}`, `(...)` and `[...]`
// respectively, which are. Likewise the operator characters of the extended
// groups are not: `+`, `@` and `!` open a group only together with a `(`
// right after them, and escaping the `(` is enough to break the pair.
func IsSpecial(c byte) bool {
	return bytes.IndexByte(specials, c) != -1
}

type tokens []Token

func (i *tokens) shift() (ret Token) {
	ret = (*i)[0]
	copy(*i, (*i)[1:])
	*i = (*i)[:len(*i)-1]
	return
}

func (i *tokens) push(v Token) {
	*i = append(*i, v)
}

func (i *tokens) empty() bool {
	return len(*i) == 0
}

// eof is the end-of-input sentinel. It must not collide with any rune that
// can appear in a valid pattern -- note that U+0000 can.
const eof rune = -1

// Lexer splits a pattern into tokens; see [Lexer.Next].
type Lexer struct {
	data string
	pos  int
	err  error

	tokens tokens
	// nesting is the stack of the open groups the lexer is inside of, the
	// innermost last: char_terms_open for a `{...}` and char_group_open for
	// an extended `(...)` group. The innermost one tells which characters
	// are special: `,` and `}` inside a `{...}`, `|` and `)` inside a
	// `(...)`; the others are plain text there.
	nesting []rune

	lastRune     rune
	lastRuneSize int
	hasRune      bool
}

// NewLexer returns a lexer over the source pattern.
func NewLexer(source string) *Lexer {
	l := &Lexer{
		data:   source,
		tokens: tokens(make([]Token, 0, 4)),
	}
	return l
}

// Offset returns the byte offset in the source the lexer stopped at, that
// is, the position right after the most recently returned token.
func (l *Lexer) Offset() int {
	return l.pos
}

// Next returns the next token. Once the input is over it returns EOF, and
// once an error happened it returns that Error, repeatedly.
func (l *Lexer) Next() Token {
	if l.err != nil {
		return Token{Error, l.err.Error()}
	}
	if !l.tokens.empty() {
		return l.tokens.shift()
	}

	l.fetchItem()
	return l.Next()
}

func (l *Lexer) peek() (r rune, w int) {
	if l.pos == len(l.data) {
		return eof, 0
	}

	r, w = utf8.DecodeRuneInString(l.data[l.pos:])
	if r == utf8.RuneError && w == 1 {
		// An invalid encoding: a valid U+FFFD decodes at its width of 3.
		l.errorf("invalid UTF-8 sequence")
		r = eof
		w = 0
	}

	return
}

func (l *Lexer) read() rune {
	if l.hasRune {
		l.hasRune = false
		l.seek(l.lastRuneSize)
		return l.lastRune
	}

	r, s := l.peek()
	l.seek(s)

	l.lastRune = r
	l.lastRuneSize = s

	return r
}

func (l *Lexer) seek(w int) {
	l.pos += w
}

func (l *Lexer) unread() {
	if l.hasRune {
		l.errorf("could not unread rune")
		return
	}
	l.seek(-l.lastRuneSize)
	l.hasRune = true
}

func (l *Lexer) errorf(f string, v ...any) {
	l.err = fmt.Errorf(f, v...)
}

// innermost returns the kind of the innermost open group, or eof when the
// lexer is at the top level.
func (l *Lexer) innermost() rune {
	if n := len(l.nesting); n > 0 {
		return l.nesting[n-1]
	}
	return eof
}

func (l *Lexer) inTerms() bool {
	return l.innermost() == char_terms_open
}

func (l *Lexer) inGroup() bool {
	return l.innermost() == char_group_open
}

func (l *Lexer) enter(kind rune) {
	l.nesting = append(l.nesting, kind)
}

func (l *Lexer) leave() {
	l.nesting = l.nesting[:len(l.nesting)-1]
}

// isGroupOperator reports whether r followed by a `(` opens an extended
// group; see [GroupOpen].
func isGroupOperator(r rune) bool {
	switch r {
	case char_single, char_any, char_group_plus, char_group_at, char_group_not:
		return true
	}
	return false
}

var inTextBreakers = []rune{char_single, char_any, char_range_open, char_terms_open}
var inTermsBreakers = append(inTextBreakers, char_terms_close, char_comma)
var inGroupBreakers = append(inTextBreakers, char_group_close, char_group_sep)

func (l *Lexer) fetchItem() {
	r := l.read()
	switch {
	case r == eof:
		l.tokens.push(Token{EOF, ""})

	case r == char_terms_open:
		l.enter(char_terms_open)
		l.tokens.push(Token{TermsOpen, string(r)})

	case r == char_comma && l.inTerms():
		l.tokens.push(Token{TermSeparator, string(r)})

	case r == char_terms_close && l.inTerms():
		l.tokens.push(Token{TermsClose, string(r)})
		l.leave()

	case r == char_group_sep && l.inGroup():
		l.tokens.push(Token{GroupSeparator, string(r)})

	case r == char_group_close && l.inGroup():
		l.tokens.push(Token{GroupClose, string(r)})
		l.leave()

	case r == char_range_open:
		l.tokens.push(Token{RangeOpen, string(r)})
		l.fetchRange()

	case isGroupOperator(r) && l.peekIs(char_group_open):
		l.read()
		l.enter(char_group_open)
		l.tokens.push(Token{GroupOpen, string(r) + string(char_group_open)})

	case r == char_single:
		l.tokens.push(Token{Single, string(r)})

	case r == char_any:
		// A `(` binds to the `*` right before it, so `**(` is a `*`
		// followed by the `*(` group opener, not a `**` and a literal `(`.
		if l.read() == char_any && !l.peekIs(char_group_open) {
			l.tokens.push(Token{Super, string(r) + string(r)})
		} else {
			l.unread()
			l.tokens.push(Token{Any, string(r)})
		}

	default:
		l.unread()

		var breakers []rune
		switch l.innermost() {
		case char_terms_open:
			breakers = inTermsBreakers
		case char_group_open:
			breakers = inGroupBreakers
		default:
			breakers = inTextBreakers
		}
		l.fetchText(breakers)
	}
}

// peekIs reports whether the next rune is r, without consuming it.
func (l *Lexer) peekIs(r rune) bool {
	n, _ := l.peek()
	return n == r
}

func (l *Lexer) fetchRange() {
	var wantHi bool
	var wantClose bool
	var seenNot bool
	for {
		r := l.read()
		if r == eof {
			l.errorf("unexpected end of input")
			return
		}

		if wantClose {
			if r != char_range_close {
				l.errorf("expected close range character")
			} else {
				l.tokens.push(Token{RangeClose, string(r)})
			}
			return
		}

		if wantHi {
			l.tokens.push(Token{RangeHi, string(r)})
			wantClose = true
			continue
		}

		if !seenNot && r == char_range_not {
			l.tokens.push(Token{Not, string(r)})
			seenNot = true
			continue
		}

		if n, w := l.peek(); n == char_range_between {
			l.seek(w)
			l.tokens.push(Token{RangeLo, string(r)})
			l.tokens.push(Token{RangeBetween, string(n)})
			wantHi = true
			continue
		}

		l.unread() // unread first peek and fetch as text
		l.fetchText([]rune{char_range_close})
		wantClose = true
	}
}

func (l *Lexer) fetchText(breakers []rune) {
	var data []rune
	var escaped bool

reading:
	for {
		r := l.read()
		if r == eof {
			if escaped {
				l.errorf("trailing backslash")
			}
			break
		}

		if !escaped {
			if r == char_escape {
				escaped = true
				continue
			}
			if slices.Contains(breakers, r) {
				l.unread()
				break reading
			}
			// The `+`, `@` and `!` are plain text unless a `(` follows:
			// then they open an extended group.
			if isGroupOperator(r) && l.peekIs(char_group_open) {
				l.unread()
				break reading
			}
		}

		escaped = false
		data = append(data, r)
	}

	if len(data) > 0 {
		l.tokens.push(Token{Text, string(data)})
	}
}

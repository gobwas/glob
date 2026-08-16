package glob

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/gobwas/glob/internal/debug"
	"github.com/gobwas/glob/syntax"
)

// SyntaxError is returned by Compile() when the given pattern can not be
// parsed. Offset points at the place in the pattern the error was detected
// at, so the tooling can do things like:
//
//	{a,b
//	----^ unclosed `{`
type SyntaxError struct {
	// Offset is a byte offset in the pattern.
	Offset int
	// Reason describes the error.
	Reason string
}

func (s *SyntaxError) Error() string {
	return fmt.Sprintf("glob: syntax error at %d: %s", s.Offset, s.Reason)
}

// Pattern represents a compiled glob pattern.
//
// A pattern is compiled into a tree of matchers:
//
//	`a`      => "a"
//	`a*`     => ["a"·*]
//	`{a*,b}` => {["a"·*]|"b"}
//
// Matching is a backtracking walk over that tree; see Pattern.Match().
type Pattern struct {
	// str is the pattern text the Pattern was compiled from; see String().
	str string

	// sep are the separators the Pattern was compiled with; see Separators().
	sep []rune

	m matcher

	// state tells whether matching m needs the backtracking state, that
	// is, whether it may save checkpoints; see needsState(). A pattern
	// without them is matched with a plain call chain.
	state bool

	// The match preconditions: every matching string is at least minLen
	// bytes and ends with suffix. They fail the obvious mismatches in O(1)
	// instead of a backtracking walk -- e.g. `a*a*a*b` requires the
	// trailing `b`, no matter how the stars go.
	//
	// A precondition pays off only when it catches a mismatch earlier than
	// the walk would, which is why:
	//
	//   - there is no required prefix: the walk is left-to-right, so a
	//     leading literal is the first thing checked anyway, while a bad
	//     suffix or length is discovered last, after the whole
	//     backtracking exploration;
	//
	//   - they are computed for the stateful patterns only: a stateless
	//     pattern is a plain call chain whose matchers perform these very
	//     checks themselves (e.g. suffixMatcher is a HasSuffix), so the
	//     precondition would only duplicate them.
	minLen int
	suffix string
}

// The shaped matchers below are the compile-time rewrites of the common
// terminal sub-sequences; see specialize(). Nothing may follow them in the
// pattern, so each one either consumes the whole remainder of the input or
// fails -- deterministically, storing no checkpoints.

// prefixMatcher is a terminal `abc*`.
type prefixMatcher struct {
	Text string
	Sep  string
}

func (m *prefixMatcher) String() string {
	return "prefix(" + strconv.Quote(m.Text) + ")"
}

func (m *prefixMatcher) Match(_ matchContext, s string) (int, bool) {
	if strings.HasPrefix(s, m.Text) && noSep(s[len(m.Text):], m.Sep) {
		return len(s), true
	}
	return 0, false
}

// suffixMatcher is a terminal `*abc`.
type suffixMatcher struct {
	Text string
	Sep  string
}

func (m *suffixMatcher) String() string {
	return "suffix(" + strconv.Quote(m.Text) + ")"
}

func (m *suffixMatcher) Match(_ matchContext, s string) (int, bool) {
	if strings.HasSuffix(s, m.Text) && noSep(s[:len(s)-len(m.Text)], m.Sep) {
		return len(s), true
	}
	return 0, false
}

// prefixSuffixMatcher is a terminal `abc*def`.
type prefixSuffixMatcher struct {
	Prefix string
	Suffix string
	Sep    string
}

func (m *prefixSuffixMatcher) String() string {
	return "prefix_suffix(" + strconv.Quote(m.Prefix) + "," + strconv.Quote(m.Suffix) + ")"
}

func (m *prefixSuffixMatcher) Match(_ matchContext, s string) (int, bool) {
	// The length check keeps the prefix and the suffix from overlapping:
	// `a*ant` must not match `ant`.
	if len(s) >= len(m.Prefix)+len(m.Suffix) &&
		strings.HasPrefix(s, m.Prefix) &&
		strings.HasSuffix(s, m.Suffix) &&
		noSep(s[len(m.Prefix):len(s)-len(m.Suffix)], m.Sep) {
		return len(s), true
	}
	return 0, false
}

// containsMatcher is a terminal `*abc*` with the separator-free stars.
type containsMatcher struct {
	Text string
}

func (m *containsMatcher) String() string {
	return "contains(" + strconv.Quote(m.Text) + ")"
}

func (m *containsMatcher) Match(_ matchContext, s string) (int, bool) {
	if strings.Contains(s, m.Text) {
		return len(s), true
	}
	return 0, false
}

// specialize rewrites the terminal sub-sequences of the simplified matcher
// tree into the shaped matchers above. The tail flag tells whether nothing
// follows m in the pattern; only there the rewrites apply, since a shaped
// matcher consumes the whole remainder of the input.
func specialize(m matcher, tail bool) matcher {
	switch v := m.(type) {
	case altMatcher:
		// Every alternative ends where the alt ends.
		for i, c := range v {
			v[i] = specialize(c, tail)
		}
		return v

	case multiMatcher:
		for i, c := range v {
			v[i] = specialize(c, tail && i == len(v)-1)
		}
		if !tail {
			return v
		}
		ms := foldTail([]matcher(v))
		if len(ms) == 1 {
			return ms[0]
		}
		return multiMatcher(ms)
	}
	return m
}

// foldTail repeatedly folds the two trailing matchers of the terminal
// sequence ms into a shaped one, while possible:
//
//	[..·"abc"·*]              => [..·prefix("abc")]
//	[..·*·"abc"]              => [..·suffix("abc")]
//	[..·"abc"·prefix("def")]  => [..·prefix("abcdef")]
//	[..·*·prefix("abc")]      => [..·contains("abc")]      (separator-free)
//	[..·"abc"·suffix("def")]  => [..·prefix_suffix("abc","def")]
//	[..·*·contains("abc")]    => [..·contains("abc")]      (separator-free)
func foldTail(ms []matcher) []matcher {
	for len(ms) >= 2 {
		var (
			prev   = ms[len(ms)-2]
			folded matcher
		)
		switch last := ms[len(ms)-1].(type) {
		case *starMatcher:
			if t, ok := prev.(*textMatcher); ok {
				folded = &prefixMatcher{Text: t.Text, Sep: last.SepStr}
			}

		case *textMatcher:
			if star, ok := prev.(*starMatcher); ok {
				folded = &suffixMatcher{Text: last.Text, Sep: star.SepStr}
			}

		case *prefixMatcher:
			switch p := prev.(type) {
			case *textMatcher:
				folded = &prefixMatcher{Text: p.Text + last.Text, Sep: last.Sep}
			case *starMatcher:
				if p.SepStr == "" && last.Sep == "" {
					folded = &containsMatcher{Text: last.Text}
				}
			}

		case *suffixMatcher:
			if t, ok := prev.(*textMatcher); ok {
				folded = &prefixSuffixMatcher{
					Prefix: t.Text,
					Suffix: last.Text,
					Sep:    last.Sep,
				}
			}

		case *containsMatcher:
			if star, ok := prev.(*starMatcher); ok && star.SepStr == "" {
				folded = last
			}
		}
		if folded == nil {
			break
		}
		ms = ms[:len(ms)-1]
		ms[len(ms)-1] = folded
	}
	return ms
}

// needsState reports whether matching m may save a checkpoint. Only the
// alts and the non-terminal stars do; a pattern without them is matched
// with a plain call chain -- see Pattern.Match().
// minLength returns the minimum length in bytes of a string m can match.
func minLength(m matcher) (n int) {
	switch v := m.(type) {
	case *textMatcher:
		return len(v.Text)
	case *charMatcher, *runeRangeMatcher, *runeSetMatcher:
		return 1
	case *prefixMatcher:
		return len(v.Text)
	case *suffixMatcher:
		return len(v.Text)
	case *prefixSuffixMatcher:
		return len(v.Prefix) + len(v.Suffix)
	case *containsMatcher:
		return len(v.Text)
	case multiMatcher:
		for _, c := range v {
			n += minLength(c)
		}
		return n
	case altMatcher:
		n = minLength(v[0])
		for _, c := range v[1:] {
			n = min(n, minLength(c))
		}
		return n
	}
	return 0 // A star or a void.
}

// requiredSuffix returns the literal every string m matches must end with.
func requiredSuffix(m matcher) string {
	switch v := m.(type) {
	case *textMatcher:
		return v.Text
	case *suffixMatcher:
		return v.Text
	case *prefixSuffixMatcher:
		return v.Suffix
	case multiMatcher:
		return requiredSuffix(v[len(v)-1])
	case altMatcher:
		s := requiredSuffix(v[0])
		for _, c := range v[1:] {
			s = commonSuffix(s, requiredSuffix(c))
			if s == "" {
				break
			}
		}
		return s
	}
	return "" // A star, a single-character matcher or a void.
}

// commonPrefix returns the longest common prefix of a and b, never
// splitting a multi-byte rune.
func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) {
		ra, wa := utf8.DecodeRuneInString(a[i:])
		rb, wb := utf8.DecodeRuneInString(b[i:])
		if ra != rb || wa != wb {
			break
		}
		i += wa
	}
	return a[:i]
}

// commonSuffix is the mirror of commonPrefix.
func commonSuffix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) {
		ra, wa := utf8.DecodeLastRuneInString(a[:len(a)-i])
		rb, wb := utf8.DecodeLastRuneInString(b[:len(b)-i])
		if ra != rb || wa != wb {
			break
		}
		i += wa
	}
	return a[len(a)-i:]
}

func needsState(m matcher) bool {
	switch v := m.(type) {
	case altMatcher:
		return true
	case multiMatcher:
		for _, c := range v {
			if needsState(c) {
				return true
			}
		}
	case *starMatcher:
		return !v.Terminal
	}
	return false
}

// noSep reports whether s contains none of the separators.
func noSep(s, sep string) bool {
	return sep == "" || !strings.ContainsAny(s, sep)
}

// matchState holds the backtracking state of a single Match() call. The
// states are pooled globally: the buffers keep their grown capacity between
// the matches, so a steady-state Match() does not allocate them.
type matchState struct {
	stars []checkpoint
	stack []checkpoint
	// arena is the buffer the checkpoint paths are allocated from; see
	// allocPath(). It is bulk-freed when the match ends, which spares the
	// per-path lifetime reasoning: a path may be shared between the current
	// frame and several checkpoints.
	arena []int
}

// allocPath returns a zeroed []int of length n allocated from the state's
// arena. When the arena runs out of capacity, a fresh chunk is started; the
// paths allocated from the previous chunks stay valid, since the chunks are
// kept alive by the paths referencing them.
func (st *matchState) allocPath(n int) []int {
	if cap(st.arena)-len(st.arena) < n {
		st.arena = make([]int, 0, max(2*cap(st.arena), n, 32))
	}
	p := st.arena[len(st.arena) : len(st.arena)+n : len(st.arena)+n]
	st.arena = st.arena[:len(st.arena)+n]
	clear(p)
	return p
}

var statePool sync.Pool // Pool[*matchState]

func acquireState() *matchState {
	if st, _ := statePool.Get().(*matchState); st != nil {
		return st
	}
	return &matchState{}
}

func releaseState(st *matchState) {
	resetCheckpoints(&st.stars)
	resetCheckpoints(&st.stack)
	// The arena holds no references; keep the (largest) chunk as is.
	st.arena = st.arena[:0]
	statePool.Put(st)
}

// resetCheckpoints empties s keeping its capacity. The whole backing array
// is zeroed (not only the live part) to drop the references to the
// checkpoint paths popped during the match.
func resetCheckpoints(s *[]checkpoint) {
	full := (*s)[:cap(*s)]
	clear(full)
	*s = full[:0]
}

// String returns the source text used to compile the pattern, the same way
// regexp.Regexp.String() does.
//
// Note that separators are not part of String: they are given to Compile
// alongside the pattern text.
func (p *Pattern) String() string {
	return p.str
}

// Separators returns the separators the pattern was compiled with, in the
// order they were given to Compile; nil when there are none.
//
// The returned slice is the very one given to Compile, sharing its backing
// array: it is not copied on the way in or out. Matching does not use it.
func (p *Pattern) Separators() []rune {
	return p.sep
}

// Match reports whether s matches the pattern.
func (p *Pattern) Match(s string) bool {
	var x matchContext
	if p.state {
		if len(s) < p.minLen || !strings.HasSuffix(s, p.suffix) {
			return false
		}
		state := acquireState()
		defer releaseState(state)
		x.state = state
	}
	for {
		n, match := p.m.Match(x, s[x.offset:])
		// Note: debug.Enabled is a build-tag constant; when it is false the
		// whole block (including the argument evaluation) is compiled away.
		if debug.Enabled && x.state != nil {
			debug.Printf("stack: %s\n", formatStack(x.state.stack))
			debug.Printf("stars: %s\n", formatStack(x.state.stars))
		}
		if match && n == len(s[x.offset:]) {
			if debug.Enabled {
				debug.Printf("match!\n")
			}
			return true
		}
		if x.state == nil {
			// The pattern never saves checkpoints (see needsState()):
			// nothing to backtrack to.
			return false
		}
		var (
			c checkpoint
			k checkpointKind
		)
		switch {
		case len(x.state.stars) > 0:
			if debug.Enabled {
				debug.Printf("has star\n")
			}
			c = popLast(&x.state.stars)
			k = checkpointStars

		case len(x.state.stack) > 0:
			if debug.Enabled {
				debug.Printf("has stack\n")
			}
			c = popLast(&x.state.stack)
			k = checkpointStack

		default:
			if debug.Enabled {
				debug.Printf("no match\n")
			}
			return false
		}
		x.offset = c.offset
		x.kind = k
		x.frame = frame{
			ptr: c.ptr,
			pos: 0,
		}
	}

}

type matcher interface {
	Match(matchContext, string) (n int, matched bool)
	String() string
}

type frame struct {
	ptr []int
	pos int
}

func (v frame) String() string {
	var sb strings.Builder
	for i, p := range v.ptr {
		if i > 0 {
			sb.WriteByte(',')
		}
		if i == v.pos {
			sb.WriteByte('[')
		}
		sb.WriteString(strconv.Itoa(p))
		if i == v.pos {
			sb.WriteByte(']')
		}
	}
	return sb.String()
}

func (v frame) index() int {
	if v.pos >= len(v.ptr) {
		return 0
	}
	return v.ptr[v.pos]
}

type checkpoint struct {
	offset int
	ptr    []int
}

// checkpointKind tells which pile a checkpoint was taken from during the
// backtracking in Pattern.Match().
type checkpointKind int

const (
	// checkpointStack is an alternative checkpoint saved by altMatcher.
	checkpointStack checkpointKind = iota
	// checkpointStars is a star restart point saved by starMatcher.
	checkpointStars
)

type matchContext struct {
	offset int
	frame  frame
	// state holds the checkpoint piles and the path arena shared by the
	// whole walk.
	state *matchState
	// kind tells which pile the checkpoint being resumed was taken from.
	// See altMatcher.Match() for its use.
	kind checkpointKind
	// starsFloor is the number of star restart points that existed when
	// the walk entered the current alternative. The entries below it were
	// born outside of the alternative and must not be discarded by the
	// stars inside it; see storeStar().
	starsFloor int
}

func (x matchContext) push(f frame) {
	x.state.stack = append(x.state.stack, checkpoint{
		offset: x.offset,
		ptr:    f.ptr,
	})
	if debug.Enabled {
		debug.Printf(
			"checkpoint offset=%d ptr=%v\n",
			x.offset, f.ptr,
		)
	}
}

func (x matchContext) storeStar(offset int, reset bool) {
	ptr := x.frame.ptr
	if d := x.frame.pos; len(ptr) < d {
		// The walk records an index in ptr only when it turns to a child
		// other than the first one; levels entered at child #0 are implicit.
		// Store the path at its full length (the missing entries are always
		// zeros) so that the alts above can tell this checkpoint from their
		// own. See altMatcher.Match().
		p := x.state.allocPath(d)
		copy(p, ptr)
		ptr = p
	}
	if reset {
		// This star can extend over anything the pending restart points
		// could reach -- they are redundant, discard them.
		// See research.swtch.com/glob.
		//
		// However, only the restart points born inside the current
		// alternative may be discarded. An outer star, when resumed,
		// re-enters the enclosing alt and may pick another alternative --
		// something this star, locked inside its own alternative, can not
		// absorb. See the `*{*0,}` test: the outer star must survive the
		// inner one to reach the empty alternative.
		x.state.stars = x.state.stars[:x.starsFloor]
	}
	x.state.stars = append(x.state.stars, checkpoint{
		offset: x.offset + offset,
		ptr:    ptr,
	})
	if debug.Enabled {
		debug.Printf(
			"star offset=%d ptr=%v reset=%t\n",
			x.offset+offset, ptr, reset,
		)
	}
}

func (x matchContext) next(offset int) matchContext {
	x.offset = x.offset + offset
	x.frame.pos += 1
	return x
}

// branch returns a copy of the current frame with its path turned to child
// i at the current level, discarding the deeper levels. The new path is
// allocated from the state's arena.
func (x matchContext) branch(i int) frame {
	f := x.frame
	ptr := x.state.allocPath(f.pos + 1)
	copy(ptr, f.ptr)
	ptr[f.pos] = i
	f.ptr = ptr
	return f
}

// annotateStars computes the compile-time hints for the star matchers, in
// order to keep the number of restart points they store at match time low:
//
//   - a star directly followed by a literal jumps between the literal
//     occurrences instead of retrying at every rune (see storeSkip());
//
//   - a star with nothing after it anywhere in the pattern (tail is true
//     for m and the star closes it) consumes its whole reach at once and
//     stores no restart points at all.
func annotateStars(m matcher, tail bool) {
	switch v := m.(type) {
	case multiMatcher:
		for i, c := range v {
			last := i == len(v)-1
			star, ok := c.(*starMatcher)
			if !ok {
				annotateStars(c, tail && last)
				continue
			}
			star.Terminal = tail && last
			if !last {
				star.Next = leadingLiteral(v[i+1])
			}
		}
	case altMatcher:
		for _, c := range v {
			annotateStars(c, tail)
		}
	case *starMatcher:
		v.Terminal = tail
	}
}

// leadingLiteral returns the literal the given matcher is guaranteed to
// begin its match with, if any.
func leadingLiteral(m matcher) string {
	switch v := m.(type) {
	case *textMatcher:
		return v.Text
	case *prefixMatcher:
		return v.Text
	case *prefixSuffixMatcher:
		return v.Prefix
	}
	return ""
}

func simplify(m matcher) matcher {
	var (
		ms      []matcher
		isMulti bool
	)
	switch v := m.(type) {
	case multiMatcher:
		ms, isMulti = v, true
	case altMatcher:
		ms = v
	default:
		return m
	}
	for i, m := range ms {
		ms[i] = simplify(m)
	}
	if isMulti {
		ms = normalizeSequence(ms)
	}
	switch len(ms) {
	case 0:
		return &voidMatcher{}
	case 1:
		return ms[0]
	}
	if isMulti {
		return multiMatcher(ms)
	}
	return altMatcher(ms)
}

// normalizeSequence rewrites a sequence of (already simplified) matchers
// into a simpler equivalent one:
//
//	["a"·["b"·"c"]·"d"] => ["a"·"b"·"c"·"d"]  inline the nested sequences
//	["a"·void]          => ["a"]              drop the void matchers
//	["a"·"b"]           => ["ab"]             merge the adjacent literals
//	[*·**]              => [**]               coalesce the adjacent stars
//
// Longer literals also make better star jumps; see annotateStars().
func normalizeSequence(ms []matcher) []matcher {
	if !needsNormalize(ms) {
		// The common case: nothing to rewrite, no copy needed.
		return ms
	}
	out := make([]matcher, 0, len(ms))
	var push func(m matcher)
	push = func(m matcher) {
		switch v := m.(type) {
		case multiMatcher:
			for _, c := range v {
				push(c)
			}
			return
		case *voidMatcher:
			return
		case *textMatcher:
			if len(out) > 0 {
				if prev, ok := out[len(out)-1].(*textMatcher); ok {
					out[len(out)-1] = &textMatcher{Text: prev.Text + v.Text}
					return
				}
			}
		case *starMatcher:
			if len(out) > 0 {
				if prev, ok := out[len(out)-1].(*starMatcher); ok {
					// Adjacent stars are equivalent to the most general
					// of them: the one not limited by separators, if any.
					if len(prev.Sep) > 0 && len(v.Sep) == 0 {
						out[len(out)-1] = v
					}
					return
				}
			}
		}
		out = append(out, m)
	}
	for _, m := range ms {
		push(m)
	}
	return out
}

func needsNormalize(ms []matcher) bool {
	for i, m := range ms {
		switch m.(type) {
		case multiMatcher, *voidMatcher:
			return true
		case *textMatcher:
			if i > 0 {
				if _, ok := ms[i-1].(*textMatcher); ok {
					return true
				}
			}
		case *starMatcher:
			if i > 0 {
				if _, ok := ms[i-1].(*starMatcher); ok {
					return true
				}
			}
		}
	}
	return false
}

type multiMatcher []matcher

func (ms multiMatcher) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, m := range ms {
		if i > 0 {
			sb.WriteString("·")
		}
		sb.WriteString(m.String())
	}
	sb.WriteByte(']')
	return sb.String()
}

func (ms multiMatcher) Match(x matchContext, s string) (n int, ok bool) {
	for i := x.frame.index(); i < len(ms); i++ {
		if i != x.frame.index() && x.state != nil {
			// The path is recorded for the checkpoints the descendants may
			// save; in a stateless walk (see needsState()) there are none
			// and nobody would ever read it.
			x.frame = x.branch(i)
		}
		k, ok := ms[i].Match(x.next(n), s[n:])
		if debug.Enabled {
			debug.Printf(
				"[%T@%p] #%d match %#q against %[5]T<%[5]s> => %d %t\n",
				ms, unsafe.SliceData(ms), i, s[n:], ms[i], k, ok,
			)
		}
		if !ok {
			return 0, false
		}
		n += k
	}
	return n, true
}

type altMatcher []matcher

func (ms altMatcher) String() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, m := range ms {
		if i > 0 {
			sb.WriteString("|")
		}
		sb.WriteString(m.String())
	}
	sb.WriteByte('}')
	return sb.String()
}

func (ms altMatcher) Match(x matchContext, s string) (int, bool) {
	i := x.frame.index()
	// Save a checkpoint for the next alternative to consider it in case of
	// a mismatch later (if any). This must be done only when:
	//
	//   - the alt is entered for the first time (the resume path ends above
	//     this level, or there is none);
	//
	//   - the resume path ends exactly at this level with an alternative
	//     checkpoint -- its job is "try alternative #i", so the one for the
	//     next alternative must be saved now. Note that a star restart point
	//     may end at this level too (a star being a direct child of the alt,
	//     as in `{*,b}`) -- it must not trigger a save.
	//
	// Otherwise the walk is merely passing through this alt on its way to
	// resume a deeper checkpoint -- the one for the next alternative was
	// already saved when the alt was entered for the first time, and saving
	// it again on every star restart would blow the stack up exponentially.
	// See the "alternatives" tests.
	if next := i + 1; next < len(ms) {
		d := len(x.frame.ptr)
		if x.frame.pos >= d || (x.frame.pos == d-1 && x.kind == checkpointStack) {
			x.push(x.branch(next))
		}
	}
	// The stars below this level may only discard the restart points born
	// inside the same alternative; see storeStar().
	x.starsFloor = len(x.state.stars)
	n, match := ms[i].Match(x.next(0), s)
	if debug.Enabled {
		debug.Printf(
			"[%T@%p] #%d match %#q against %[5]T<%[5]s> => %d %t\n",
			ms, unsafe.SliceData(ms), i, ms[i], s, n, match,
		)
	}
	return n, match
}

type textMatcher struct {
	Text string
}

func (m *textMatcher) String() string {
	return strconv.Quote(m.Text)
}

func (m *textMatcher) Match(_ matchContext, s string) (int, bool) {
	if strings.HasPrefix(s, m.Text) {
		return len(m.Text), true
	}
	return 0, false
}

type charMatcher struct {
	Sep []rune
}

func (m *charMatcher) String() string {
	var sb strings.Builder
	sb.WriteByte('?')
	if len(m.Sep) > 0 {
		sb.WriteByte('(')
		formatRunes(&sb, m.Sep)
		sb.WriteByte(')')
	}
	return sb.String()
}

func (m *charMatcher) Match(_ matchContext, s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	r, n := utf8.DecodeRuneInString(s)
	if slices.Contains(m.Sep, r) {
		return 0, false
	}
	return n, true
}

type starMatcher struct {
	Sep []rune
	// SepStr is Sep as a string, for the byte-wise scans below.
	SepStr string

	// Next is the literal the matcher right after this star begins with
	// (when it is a textMatcher), set by annotateStars(). A restart point
	// at a position where the literal does not occur is a guaranteed
	// mismatch, so the star jumps between its occurrences instead of
	// retrying at every rune.
	Next string
	// Terminal is set by annotateStars() when nothing follows this star
	// anywhere in the pattern: the star then consumes everything in its
	// reach at once, and no restart point can change the outcome.
	Terminal bool
}

// reach returns the length of the prefix of s the star may extend over:
// everything up to the nearest separator.
func (m *starMatcher) reach(s string) int {
	if m.SepStr == "" {
		return len(s)
	}
	if e := strings.IndexAny(s, m.SepStr); e >= 0 {
		return e
	}
	return len(s)
}

// storeSkip stores the restart point at the next occurrence of the m.Next
// literal instead of the next rune.
func (m *starMatcher) storeSkip(x matchContext, s string) {
	reach := m.reach(s)
	// Look for the occurrences starting within the star's reach; the
	// literal itself may extend past it (it may contain the separators).
	// Note that a valid UTF-8 literal can not match at a mid-rune
	// position, so the one-byte skip below is rune-safe.
	end := min(reach+len(m.Next), len(s))
	j := strings.Index(s[1:end], m.Next)
	if j < 0 || 1+j > reach {
		return
	}
	x.storeStar(1+j, len(m.Sep) == 0)
}

func (m *starMatcher) String() string {
	var sb strings.Builder
	sb.WriteByte('*')
	if len(m.Sep) > 0 {
		sb.WriteByte('(')
		formatRunes(&sb, m.Sep)
		sb.WriteByte(')')
	}
	return sb.String()
}

func (m *starMatcher) Match(x matchContext, s string) (int, bool) {
	if m.Terminal {
		// Nothing follows this star in the pattern: either it consumes
		// the whole remainder within its reach, or the match fails.
		return m.reach(s), true
	}
	if len(s) == 0 {
		return 0, true
	}
	if m.Next != "" {
		m.storeSkip(x, s)
		return 0, true
	}
	r, n := utf8.DecodeRuneInString(s)
	if !slices.Contains(m.Sep, r) {
		// Can extend only over a non-separator chars.
		// Merge with previous stars only if this star has no separator (a
		// super match).
		x.storeStar(n, len(m.Sep) == 0)
	}
	return 0, true
}

type runeRangeMatcher struct {
	Lo  rune
	Hi  rune
	Not bool
}

func (m *runeRangeMatcher) String() string {
	var sb strings.Builder
	if m.Not {
		sb.WriteByte('!')
	}
	sb.WriteByte('[')
	sb.WriteRune(m.Lo)
	sb.WriteByte('-')
	sb.WriteRune(m.Hi)
	sb.WriteByte(']')
	return sb.String()
}

func (m *runeRangeMatcher) Match(_ matchContext, s string) (int, bool) {
	r, n := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0, false
	}
	ok := m.Lo <= r && r <= m.Hi
	if ok != m.Not {
		return n, true
	}
	return 0, false
}

type runeSetMatcher struct {
	Set map[rune]struct{}
	Not bool
}

func formatRunes(sb *strings.Builder, rs []rune) {
	slices.Sort(rs)
	for i, r := range rs {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(r)
	}
}

func (m *runeSetMatcher) String() string {
	rs := make([]rune, 0, len(m.Set))
	for r := range m.Set {
		rs = append(rs, r)
	}
	var sb strings.Builder
	if m.Not {
		sb.WriteByte('!')
	}
	sb.WriteByte('[')
	formatRunes(&sb, rs)
	sb.WriteByte(']')
	return sb.String()
}

func (m *runeSetMatcher) Match(_ matchContext, s string) (int, bool) {
	r, n := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0, false
	}
	if _, has := m.Set[r]; has != m.Not {
		return n, true
	}
	return 0, false
}

type voidMatcher struct{}

func (*voidMatcher) String() string {
	return "void"
}

func (*voidMatcher) Match(matchContext, string) (int, bool) {
	return 0, true
}

// TODO(optimisations):
// {a,ab} -> a{,b}
// {ab,b} -> {a,}b
//
// *a* -> contains("a")
// *a  -> suffix("a")
// a*  -> prefix("a")
func compile(str string, sep []rune) (*Pattern, error) {
	if debug.Enabled {
		debug.Printf("compiling %#q\n", str)
	}
	// The matchers keep sep and read it while matching, and the variadic slice
	// may alias an array owned by the caller: give them a copy of their own.
	//
	// The pattern itself keeps the slice as given, to return it from
	// Separators() without cloning.
	var (
		sepCopy = slices.Clone(sep)
		sepStr  = string(sep)
	)

	lex := syntax.NewLexer(str)

	type operator struct {
		kind  int
		index int
	}
	const (
		opTerms = iota
		opList
	)
	/*
		Stack-based parsing is a technique used to evaluate mathematical
		expressions by leveraging the properties of the LIFO (Last-In,
		First-Out) data structure, the stack. It involves using two stacks: one
		for operands (numbers) and one for operators. By processing the
		expression from left to right and strategically pushing and popping
		elements from the stacks, the expression can be effectively evaluated.

		https://cp-algorithms.com/string/expression_parsing.html
	*/
	var stack []matcher
	var operators []operator
parsing:
	for {
		token := lex.Next()
		if debug.Enabled {
			debug.Printf("token: %s\n", token)
		}
		switch token.Type {
		case syntax.EOF:
			break parsing

		case syntax.Error:
			return nil, &SyntaxError{
				Offset: lex.Offset(),
				Reason: token.Data,
			}

		case syntax.Single:
			stack = append(stack, &charMatcher{
				Sep: sepCopy,
			})

		case syntax.Text:
			stack = append(stack, &textMatcher{
				Text: token.Data,
			})

		case syntax.RangeOpen:
			m, err := parseRange(lex)
			if err != nil {
				return nil, err
			}
			stack = append(stack, m)

		case syntax.Any:
			stack = append(stack, &starMatcher{
				Sep:    sepCopy,
				SepStr: sepStr,
			})

		case syntax.Super:
			stack = append(stack, &starMatcher{
				Sep: nil,
			})

		case syntax.TermsOpen:
			// Note that the `{` opens both the group and its first
			// alternative: every alternative is delimited by an opList
			// operator. This way TermsClose always collapses the trailing
			// alternative into a single matcher first, even when the group
			// has no commas at all, e.g. `{ab*}`.
			operators = append(operators,
				operator{kind: opTerms, index: len(stack)},
				operator{kind: opList, index: len(stack)},
			)
			if debug.Enabled {
				debug.Printf("terms enter: %d\n", len(stack))
			}

		case syntax.TermSeparator:
			k := len(operators) - 1
			if k < 0 {
				return nil, &SyntaxError{
					Offset: lex.Offset(),
					Reason: "unexpected `,`",
				}
			}
			x := operators[k]
			if x.kind == opList {
				// Remove the most recent "comma" operator.
				// Note that the previous one is the terms operator.
				operators = operators[:k]
			}
			i := x.index
			// Handle the `{,a}` case.
			if i == len(stack) {
				// Empty matchers.
				stack = append(stack, &voidMatcher{})
			} else {
				stack[i] = multiMatcher(slices.Clone(stack[i:]))
				stack = stack[:i+1]
				if debug.Enabled {
					debug.Printf("terms next: %d: %s\n", i, stack[i])
				}
			}
			operators = append(operators, operator{
				kind:  opList,
				index: len(stack),
			})
			if debug.Enabled {
				debug.Printf("terms separator: %d\n", len(stack))
			}

		case syntax.TermsClose:
			for {
				k := len(operators) - 1
				if k < 0 {
					return nil, &SyntaxError{
						Offset: lex.Offset(),
						Reason: "unexpected `}`",
					}
				}
				x := operators[k]
				operators = operators[:k]

				i := x.index
				c := slices.Clone(stack[i:])
				var m matcher
				switch x.kind {
				case opTerms:
					m = altMatcher(c)
				case opList:
					m = multiMatcher(c)
				}
				// Handle the `{a,}` case.
				if i == len(stack) {
					stack = append(stack, m)
				} else {
					stack = stack[:i+1]
					stack[i] = m
				}

				if debug.Enabled {
					debug.Printf(
						"terms leave(%d): %d: %s\n",
						x.kind, i, stack[i],
					)
				}
				if x.kind == opTerms {
					break
				}
			}

		default:
			return nil, &SyntaxError{
				Offset: lex.Offset(),
				Reason: "unexpected token " + token.String(),
			}
		}
	}
	if len(operators) != 0 {
		return nil, &SyntaxError{
			Offset: lex.Offset(),
			Reason: "unclosed `{`",
		}
	}
	m := simplify(multiMatcher(stack))
	m = specialize(m, true)
	annotateStars(m, true)
	if debug.Enabled {
		debug.Printf("compiled %#q: %s\n", str, m)
	}
	p := &Pattern{
		str:   str,
		sep:   sep,
		m:     m,
		state: needsState(m),
	}
	if p.state {
		p.minLen = minLength(m)
		p.suffix = requiredSuffix(m)
	}
	return p, nil
}

func parseRange(lex *syntax.Lexer) (matcher, error) {
	// -1 marks a range boundary as unset: any decoded rune, including
	// U+0000, is non-negative.
	var (
		not    bool
		lo, hi rune = -1, -1
		chars  map[rune]struct{}
	)
	for {
		token := lex.Next()
		switch token.Type {
		case syntax.EOF:
			return nil, &SyntaxError{
				Offset: lex.Offset(),
				Reason: "unclosed `[`",
			}

		case syntax.Error:
			return nil, &SyntaxError{
				Offset: lex.Offset(),
				Reason: token.Data,
			}

		case syntax.Not:
			not = true

		case syntax.RangeLo:
			r, w := utf8.DecodeRuneInString(token.Data)
			if len(token.Data) > w {
				return nil, &SyntaxError{
					Offset: lex.Offset(),
					Reason: "unexpected length of range lo character",
				}
			}
			lo = r

		case syntax.RangeBetween:
			//

		case syntax.RangeHi:
			r, w := utf8.DecodeRuneInString(token.Data)
			if len(token.Data) > w {
				return nil, &SyntaxError{
					Offset: lex.Offset(),
					Reason: "unexpected length of range hi character",
				}
			}
			hi = r

			if hi < lo {
				return nil, &SyntaxError{
					Offset: lex.Offset(),
					Reason: "range hi character is less than lo",
				}
			}

		case syntax.Text:
			chars = make(map[rune]struct{})
			for _, r := range token.Data {
				chars[r] = struct{}{}
			}

		case syntax.RangeClose:
			isRange := lo >= 0 && hi >= 0
			isChars := chars != nil

			if isChars == isRange {
				return nil, &SyntaxError{
					Offset: lex.Offset(),
					Reason: "could not parse range",
				}
			}
			if isRange {
				return &runeRangeMatcher{
					Lo:  lo,
					Hi:  hi,
					Not: not,
				}, nil
			}
			return &runeSetMatcher{
				Set: chars,
				Not: not,
			}, nil
		}
	}
}

func formatCheckpoint(c checkpoint) string {
	return fmt.Sprintf("[%d:]<%v>", c.offset, c.ptr)
}

func formatStack(s []checkpoint) string {
	var sb strings.Builder
	for i, c := range s {
		if i > 0 {
			sb.WriteString(" -> ")
		}
		sb.WriteString(formatCheckpoint(c))
	}
	return sb.String()
}

// popLast panics if s is empty.
func popLast[T any, E ~[]T](s *E) T {
	n := len(*s)
	r := (*s)[n-1]
	*s = (*s)[:n-1]
	return r
}

package match

import (
	"fmt"

	"github.com/gobwas/glob/internal/debug"
	"github.com/gobwas/glob/util/runes"
)

func Optimize(m Matcher) (opt Matcher) {
	if debug.Enabled {
		defer func() {
			a := fmt.Sprintf("%s", m)
			b := fmt.Sprintf("%s", opt)
			if a != b {
				debug.EnterPrefix("optimized %s: -> %s", a, b)
				debug.LeavePrefix()
			}
		}()
	}
	switch v := m.(type) {
	case Any:
		if len(v.sep) == 0 {
			return NewSuper()
		}

	case List:
		if v.not == false && len(v.rs) == 1 {
			return NewText(string(v.rs))
		}
		return m

	case Tree:
		v.left = Optimize(v.left)
		v.right = Optimize(v.right)

		txt, ok := v.value.(Text)
		if !ok {
			return m
		}

		var (
			leftNil  = v.left == nil
			rightNil = v.right == nil
		)
		if leftNil && rightNil {
			return NewText(txt.s)
		}

		_, leftSuper := v.left.(Super)
		lp, leftPrefix := v.left.(Prefix)
		la, leftAny := v.left.(Any)

		_, rightSuper := v.right.(Super)
		rs, rightSuffix := v.right.(Suffix)
		ra, rightAny := v.right.(Any)

		switch {
		case leftSuper && rightSuper:
			return NewContains(txt.s)

		case leftSuper && rightNil:
			return NewSuffix(txt.s)

		case rightSuper && leftNil:
			return NewPrefix(txt.s)

		case leftNil && rightSuffix:
			return NewPrefixSuffix(txt.s, rs.s)

		case rightNil && leftPrefix:
			return NewPrefixSuffix(lp.s, txt.s)

		case rightNil && leftAny:
			return NewSuffixAny(txt.s, la.sep)

		case leftNil && rightAny:
			return NewPrefixAny(txt.s, ra.sep)
		}

	case Container:
		var (
			first Matcher
			n     int
		)
		v.Content(func(m Matcher) {
			first = m
			n++
		})
		if n == 1 {
			return first
		}
		return m
	}

	return m
}

func Compile(ms []Matcher) (m Matcher, err error) {
	if debug.Enabled {
		debug.EnterPrefix("compiling %s", ms)
		defer func() {
			debug.Logf("-> %s, %v", m, err)
			debug.LeavePrefix()
		}()
	}
	if len(ms) == 0 {
		return nil, fmt.Errorf("compile error: need at least one matcher")
	}
	if len(ms) == 1 {
		return ms[0], nil
	}
	if m := glueMatchers(ms); m != nil {
		return m, nil
	}

	var (
		x   = -1
		max = -2

		wantText bool
		indexer  MatchIndexer
	)
	for i, m := range ms {
		mx, ok := m.(MatchIndexer)
		if !ok {
			continue
		}
		_, isText := m.(Text)
		if wantText && !isText {
			continue
		}
		n := m.MinLen()
		if (!wantText && isText) || n > max {
			max = n
			x = i
			indexer = mx
			wantText = isText
		}
	}
	if indexer == nil {
		return nil, fmt.Errorf("can not index on matchers")
	}

	left := ms[:x]
	var right []Matcher
	if len(ms) > x+1 {
		right = ms[x+1:]
	}

	var (
		l Matcher = Nothing{}
		r Matcher = Nothing{}
	)
	if len(left) > 0 {
		l, err = Compile(left)
		if err != nil {
			return nil, err
		}
	}
	if len(right) > 0 {
		r, err = Compile(right)
		if err != nil {
			return nil, err
		}
	}
	return NewTree(indexer, l, r), nil
}

func glueMatchers(ms []Matcher) Matcher {
	if m := glueMatchersAsEvery(ms); m != nil {
		return m
	}
	if m := glueMatchersAsRow(ms); m != nil {
		return m
	}
	return nil
}

func glueMatchersAsRow(ms []Matcher) Matcher {
	if len(ms) <= 1 {
		return nil
	}
	var s []MatchIndexSizer
	for _, m := range ms {
		rsz, ok := m.(MatchIndexSizer)
		if !ok {
			return nil
		}
		s = append(s, rsz)
	}
	return NewRow(s)
}

func glueMatchersAsEvery(ms []Matcher) Matcher {
	if len(ms) <= 1 {
		return nil
	}

	var (
		hasAny    bool
		hasSuper  bool
		hasSingle bool
		min       int
		separator []rune
	)

	for i, matcher := range ms {
		var sep []rune

		switch m := matcher.(type) {
		case Super:
			sep = []rune{}
			hasSuper = true

		case Any:
			sep = m.sep
			hasAny = true

		case Single:
			sep = m.sep
			hasSingle = true
			min++

		case List:
			if !m.not {
				return nil
			}
			sep = m.rs
			hasSingle = true
			min++

		default:
			return nil
		}

		// initialize
		if i == 0 {
			separator = sep
		}

		if runes.Equal(sep, separator) {
			continue
		}

		return nil
	}

	if hasSuper && !hasAny && !hasSingle {
		return NewSuper()
	}

	if hasAny && !hasSuper && !hasSingle {
		return NewAny(separator)
	}

	if (hasAny || hasSuper) && min > 0 && len(separator) == 0 {
		return NewMin(min)
	}

	var every []Matcher
	if min > 0 {
		every = append(every, NewMin(min))
		if !hasAny && !hasSuper {
			every = append(every, NewMax(min))
		}
	}
	if len(separator) > 0 {
		every = append(every, NewAny(separator))
	}

	return NewEveryOf(every)
}

type result struct {
	ms       []Matcher
	matchers int
	minLen   int
	nesting  int
}

func compareResult(a, b result) int {
	if x := b.minLen - a.minLen; x != 0 {
		return x
	}
	if x := a.matchers - b.matchers; x != 0 {
		return x
	}
	if x := a.nesting - b.nesting; x != 0 {
		return x
	}
	if x := len(a.ms) - len(b.ms); x != 0 {
		return x
	}
	return 0
}

func collapse(ms []Matcher, x Matcher, i, j int) (cp []Matcher) {
	cp = make([]Matcher, len(ms)-(j-i)+1)
	copy(cp[0:i], ms[0:i])
	copy(cp[i+1:], ms[j:])
	cp[i] = x
	return cp
}

func matchersCount(ms []Matcher) (n int) {
	n = len(ms)
	for _, m := range ms {
		n += countNestedMatchers(m)
	}
	return n
}

func countNestedMatchers(m Matcher) (n int) {
	if c, _ := m.(Container); c != nil {
		c.Content(func(m Matcher) {
			n += 1 + countNestedMatchers(m)
		})
	}
	return n
}

func nestingDepth(m Matcher) (depth int) {
	c, ok := m.(Container)
	if !ok {
		return 0
	}
	var max int
	c.Content(func(m Matcher) {
		if d := nestingDepth(m); d > max {
			max = d
		}
	})
	return max + 1
}

func maxMinLen(ms []Matcher) (max int) {
	for _, m := range ms {
		if n := m.MinLen(); n > max {
			max = n
		}
	}
	return max
}

func maxNestingDepth(ms []Matcher) (max int) {
	for _, m := range ms {
		if n := nestingDepth(m); n > max {
			max = n
		}
	}
	return
}

func minimize(ms []Matcher, i, j int, best *result) *result {
	if j > len(ms) {
		j = 0
		i++
	}
	if i > len(ms)-2 {
		return best
	}
	if j == 0 {
		j = i + 2
	}
	if g := glueMatchers(ms[i:j]); g != nil {
		cp := collapse(ms, g, i, j)
		r := result{
			ms:       cp,
			matchers: matchersCount(cp),
			minLen:   maxMinLen(cp),
			nesting:  maxNestingDepth(cp),
		}
		if debug.Enabled {
			debug.EnterPrefix(
				"intermediate: %s (matchers:%d, minlen:%d, nesting:%d)",
				cp, r.matchers, r.minLen, r.nesting,
			)
		}
		if best == nil {
			best = new(result)
		}
		if best.ms == nil || compareResult(r, *best) < 0 {
			*best = r
			if debug.Enabled {
				debug.Logf("new best result")
			}
		}
		best = minimize(cp, 0, 0, best)
		if debug.Enabled {
			debug.LeavePrefix()
		}
	}
	return minimize(ms, i, j+1, best)
}

func Minimize(ms []Matcher) (m []Matcher) {
	if debug.Enabled {
		debug.EnterPrefix("minimizing %s", ms)
		defer func() {
			debug.Logf("-> %s", m)
			debug.LeavePrefix()
		}()
	}
	best := minimize(ms, 0, 0, nil)
	if best == nil {
		return ms
	}
	return best.ms
}

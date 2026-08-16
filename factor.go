package glob

// This file is parked research and is not wired into the release code path.
//
// factor() rewrites the alts by pulling the common prefixes and the common
// suffixes of their alternatives out:
//
//	{a,ab}    => a{,b}
//	{ab,b}    => {a,}b
//	{ax*,bx*} => {a,b}x*
//
// The shared parts are then matched once instead of per alternative, and
// the residual alternatives get smaller.
//
// Lessons learned while benchmarking it (see the v1 rewrite history):
//
//   - A tail alt (one with nothing after it in the pattern) must never be
//     factored: its alternatives specialize into the O(1) shaped matchers
//     (see specialize()), which beats sharing a factored-out part by far.
//     Naive factoring regressed `{https://*gobwas.com,...}` 10x by pulling
//     the alt out of its terminal position.
//
//   - For the non-tail alts factoring measurably helps the chained groups
//     (`{a,ab,abc,abcd}{b,bc,bcd,bcde}...` matched ~15% faster) and should
//     shine on machine-generated alternative lists sharing long prefixes
//     (the way Telegraf-like tools join pattern lists, see issue #50) --
//     but that workload needs real-world profiling to justify carrying the
//     complexity in the release.
//
// To wire it in, put
//
//	m = simplify(factor(m, true))
//
// between the simplify() and specialize() calls in compile(). The pass is
// differentially guarded by the independent FuzzMatchRegexp oracle.

import (
	"maps"
	"slices"
)

// factor rewrites the alts of the simplified matcher tree by factoring the
// common parts of their alternatives out. The tail flag tells whether
// nothing follows m in the pattern; a tail alt is left for specialize().
func factor(m matcher, tail bool) matcher {
	switch v := m.(type) {
	case multiMatcher:
		for i, c := range v {
			v[i] = factor(c, tail && i == len(v)-1)
		}
	case altMatcher:
		for i, c := range v {
			v[i] = factor(c, tail)
		}
		if tail {
			return v
		}
		return factorAlt(v)
	}
	return m
}

func factorAlt(v altMatcher) matcher {
	// The alternatives as sequences; empty for the void ones.
	seqs := make([][]matcher, len(v))
	for i, c := range v {
		switch m := c.(type) {
		case multiMatcher:
			seqs[i] = m
		case *voidMatcher:
			//
		default:
			seqs[i] = []matcher{c}
		}
	}
	prefix := factorPrefix(seqs)
	suffix := factorSuffix(seqs)
	if len(prefix) == 0 && len(suffix) == 0 {
		return v
	}
	// Rebuild the residual alternatives.
	var residual bool
	for i, s := range seqs {
		switch len(s) {
		case 0:
			v[i] = &voidMatcher{}
		case 1:
			v[i] = s[0]
			residual = true
		default:
			v[i] = multiMatcher(s)
			residual = true
		}
	}
	var out []matcher
	out = append(out, prefix...)
	if residual {
		// Note: with no residuals left every alternative was consumed
		// whole (all of them were equal) and the alt is dropped.
		out = append(out, v)
	}
	out = append(out, suffix...)
	if len(out) == 1 {
		return out[0]
	}
	return multiMatcher(out)
}

// factorPrefix trims the longest common prefix off the sequences and
// returns it: whole structurally-equal matchers first, then a common text
// prefix of the leading literals, if any.
func factorPrefix(seqs [][]matcher) (prefix []matcher) {
	for {
		for _, s := range seqs {
			if len(s) == 0 {
				return prefix
			}
		}
		head := seqs[0][0]
		same := true
		for _, s := range seqs[1:] {
			if !matcherEqual(head, s[0]) {
				same = false
				break
			}
		}
		if same {
			prefix = append(prefix, head)
			for i := range seqs {
				seqs[i] = seqs[i][1:]
			}
			continue
		}
		// The head matchers differ; the only option left is a common text
		// prefix of the head literals.
		common := ""
		if t, ok := head.(*textMatcher); ok {
			common = t.Text
			for _, s := range seqs[1:] {
				t, ok := s[0].(*textMatcher)
				if !ok {
					common = ""
					break
				}
				common = commonPrefix(common, t.Text)
				if common == "" {
					break
				}
			}
		}
		if common == "" {
			return prefix
		}
		prefix = append(prefix, &textMatcher{Text: common})
		for i := range seqs {
			t := seqs[i][0].(*textMatcher)
			if rest := t.Text[len(common):]; rest != "" {
				seqs[i][0] = &textMatcher{Text: rest}
			} else {
				seqs[i] = seqs[i][1:]
			}
		}
	}
}

// factorSuffix is the mirror of factorPrefix for the sequence tails.
func factorSuffix(seqs [][]matcher) (suffix []matcher) {
	for {
		for _, s := range seqs {
			if len(s) == 0 {
				slices.Reverse(suffix)
				return suffix
			}
		}
		last := func(s []matcher) matcher { return s[len(s)-1] }
		tail := last(seqs[0])
		same := true
		for _, s := range seqs[1:] {
			if !matcherEqual(tail, last(s)) {
				same = false
				break
			}
		}
		if same {
			suffix = append(suffix, tail)
			for i := range seqs {
				seqs[i] = seqs[i][:len(seqs[i])-1]
			}
			continue
		}
		common := ""
		if t, ok := tail.(*textMatcher); ok {
			common = t.Text
			for _, s := range seqs[1:] {
				t, ok := last(s).(*textMatcher)
				if !ok {
					common = ""
					break
				}
				common = commonSuffix(common, t.Text)
				if common == "" {
					break
				}
			}
		}
		if common == "" {
			slices.Reverse(suffix)
			return suffix
		}
		suffix = append(suffix, &textMatcher{Text: common})
		for i := range seqs {
			s := seqs[i]
			t := last(s).(*textMatcher)
			if rest := t.Text[:len(t.Text)-len(common)]; rest != "" {
				s[len(s)-1] = &textMatcher{Text: rest}
			} else {
				seqs[i] = s[:len(s)-1]
			}
		}
	}
}

// matcherEqual reports whether the two matchers are structurally identical.
func matcherEqual(a, b matcher) bool {
	switch x := a.(type) {
	case *voidMatcher:
		_, ok := b.(*voidMatcher)
		return ok
	case *textMatcher:
		y, ok := b.(*textMatcher)
		return ok && x.Text == y.Text
	case *charMatcher:
		y, ok := b.(*charMatcher)
		return ok && slices.Equal(x.Sep, y.Sep)
	case *starMatcher:
		y, ok := b.(*starMatcher)
		return ok && slices.Equal(x.Sep, y.Sep)
	case *runeRangeMatcher:
		y, ok := b.(*runeRangeMatcher)
		return ok && *x == *y
	case *runeSetMatcher:
		y, ok := b.(*runeSetMatcher)
		return ok && x.Not == y.Not && maps.Equal(x.Set, y.Set)
	case multiMatcher:
		y, ok := b.(multiMatcher)
		return ok && slices.EqualFunc(x, y, matcherEqual)
	case altMatcher:
		y, ok := b.(altMatcher)
		return ok && slices.EqualFunc(x, y, matcherEqual)
	}
	return false
}

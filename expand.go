package glob

// This file is parked research and is not used by the release code path.
//
// It explores an alternative approach to handling the brace alternatives:
// instead of keeping them in the matcher tree and backtracking over the
// alternative checkpoints at match time (what Pattern.Match() does now), the
// braces could be expanded at compile time into a cartesian product of flat
// patterns:
//
//	`{a,b}c`      => [a·c], [b·c]
//	`{a*,b}{c,d}` => [a·*·c], [a·*·d], [b·c], [b·d]
//
// Matching would then be a simple star-only backtracking loop over a flat
// list of matchers (see research.swtch.com/glob) -- no tree walk, no
// checkpoint paths, no alternative checkpoints. It would also allow
// per-pattern shortcuts like literal prefix/suffix checks.
//
// The downsides are the exponential number of flat patterns for sequential
// and nested brace groups, and the duplicated work of matching the patterns
// sharing a common prefix. So if this is ever wired in, it should be a
// compile-time optimization for patterns with a small expansion, not the
// general engine.
//
// expandBracesRec and expandBracesStack implement the same expansion
// recursively and iteratively (with an explicit stack).
//
// Known bug: expandBracesRec fails on the empty alternatives as in `{,a}`
// and `{a,}`. TODO: add tests if this is ever wired in.

import (
	"container/list"
	"slices"
)

func expandBracesRec(m matcher) [][]matcher {
	var f func([][]matcher, matcher) [][]matcher
	f = func(dst [][]matcher, m matcher) (ret [][]matcher) {
		switch ms := m.(type) {
		case altMatcher:
			ret = make([][]matcher, 0, len(dst)*len(ms))
			for _, m := range ms {
				for _, ds := range dst {
					for _, r := range f([][]matcher{slices.Clone(ds)}, m) {
						ret = append(ret, r)
					}
				}
			}

		case multiMatcher:
			ret = grow(slices.Clone(dst), 1)
			for _, m := range ms {
				ret = f(ret, m)
			}

		default:
			dst = grow(dst, 1)
			ret = make([][]matcher, 0, len(dst))
			for _, d := range dst {
				ret = append(ret, append(slices.Clone(d), m))
			}
		}
		return ret
	}
	return f(nil, m)
}

func expandBracesStack(m matcher) [][]matcher {
	type elem struct {
		ret *[][]matcher
		exp *[][]matcher
		m   matcher
	}
	var exp [][]matcher
	var ptr [][]matcher
	stack := list.New()
	stack.PushBack(elem{
		exp: &ptr,
		m:   m,
	})
	stack.PushBack(elem{
		ret: &exp,
		exp: &ptr,
	})
	for stack.Len() > 0 {
		el := stack.Remove(stack.Front()).(elem)
		switch v := el.m.(type) {
		case nil:
			*el.ret = append(*el.ret, *el.exp...)

		case altMatcher:
			// Need to do reverse to PushFront() to the stack.
			for i := len(v) - 1; i >= 0; i-- {
				cp := slices.Clone(*el.exp)
				for i, ms := range cp {
					cp[i] = slices.Clone(ms)
				}
				stack.PushFront(elem{
					ret: el.exp,
					exp: &cp,
					m:   nil,
				})
				stack.PushFront(elem{
					exp: &cp,
					m:   v[i],
				})
			}
			*el.exp = (*el.exp)[:0]

		case multiMatcher:
			// Need to do reverse to PushFront() to the stack.
			for i := len(v) - 1; i >= 0; i-- {
				stack.PushFront(elem{
					exp: el.exp,
					m:   v[i],
				})
			}

		default:
			// An ordinary (not a container) matcher.
			*el.exp = grow(*el.exp, 1)
			for i, ms := range *el.exp {
				(*el.exp)[i] = append(ms, el.m)
			}
		}
	}
	return exp
}

func grow[T any](s []T, n int) []T {
	if len(s) >= n {
		return s
	}
	return slices.Grow(s, n)[:n]
}

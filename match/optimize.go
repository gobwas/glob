package match

import (
	"fmt"

	"gopkg.in/readline.v1/runes"
)

func Optimize(m Matcher) Matcher {
	switch v := m.(type) {
	case Any:
		if len(v.sep) == 0 {
			return NewSuper()
		}

	case Container:
		ms := v.Content()
		if len(ms) == 1 {
			return ms[0]
		}
		return m

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
	}

	return m
}

func Compile(ms []Matcher) (Matcher, error) {
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
		idx     = -1
		maxLen  = -2
		indexer MatchIndexer
	)
	for i, m := range ms {
		mi, ok := m.(MatchIndexer)
		if !ok {
			continue
		}
		if n := m.MinLen(); n > maxLen {
			maxLen = n
			idx = i
			indexer = mi
		}
	}
	if indexer == nil {
		return nil, fmt.Errorf("can not index on matchers")
	}

	left := ms[:idx]
	var right []Matcher
	if len(ms) > idx+1 {
		right = ms[idx+1:]
	}

	var l, r Matcher
	var err error
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

func Minimize(ms []Matcher) []Matcher {
	var (
		result Matcher
		left   int
		right  int
		count  int
	)
	for l := 0; l < len(ms); l++ {
		for r := len(ms); r > l; r-- {
			if glued := glueMatchers(ms[l:r]); glued != nil {
				var swap bool
				if result == nil {
					swap = true
				} else {
					swap = glued.MinLen() > result.MinLen() || count < r-l
				}
				if swap {
					result = glued
					left = l
					right = r
					count = r - l
				}
			}
		}
	}
	if result == nil {
		return ms
	}
	next := append(append([]Matcher{}, ms[:left]...), result)
	if right < len(ms) {
		next = append(next, ms[right:]...)
	}
	if len(next) == len(ms) {
		return next
	}
	return Minimize(next)
}

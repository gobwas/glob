package glob

import (
	"fmt"
	"github.com/gobwas/glob/match"
)

func optimize(matcher match.Matcher) match.Matcher {
	switch m := matcher.(type) {

	case match.Any:
		if m.Separators == "" {
			return match.Super{}
		}

	case match.BTree:
		m.Left = optimize(m.Left)
		m.Right = optimize(m.Right)

		r, ok := m.Value.(match.Raw)
		if !ok {
			return m
		}

		leftNil := m.Left == nil
		rightNil := m.Right == nil

		if leftNil && rightNil {
			return match.Raw{r.Str}
		}

		_, leftSuper := m.Left.(match.Super)
		lp, leftPrefix := m.Left.(match.Prefix)

		_, rightSuper := m.Right.(match.Super)
		rs, rightSuffix := m.Right.(match.Suffix)

		if leftSuper && rightSuper {
			return match.Contains{r.Str, false}
		}

		if leftSuper && rightNil {
			return match.Suffix{r.Str}
		}

		if rightSuper && leftNil {
			return match.Prefix{r.Str}
		}

		if leftNil && rightSuffix {
			return match.PrefixSuffix{Prefix: r.Str, Suffix: rs.Suffix}
		}

		if rightNil && leftPrefix {
			return match.PrefixSuffix{Prefix: lp.Prefix, Suffix: r.Str}
		}

		return m
	}

	return matcher
}

func glueMatchers(matchers []match.Matcher) match.Matcher {
	var (
		glued  []match.Matcher
		winner match.Matcher
	)
	maxLen := -1

	if m := glueAsEvery(matchers); m != nil {
		glued = append(glued, m)
		return m
	}

	if m := glueAsRow(matchers); m != nil {
		glued = append(glued, m)
		return m
	}

	for _, g := range glued {
		if l := g.Len(); l > maxLen {
			maxLen = l
			winner = g
		}
	}

	return winner
}

func glueAsRow(matchers []match.Matcher) match.Matcher {
	if len(matchers) <= 1 {
		return nil
	}

	row := match.Row{}
	for _, matcher := range matchers {
		err := row.Add(matcher)
		if err != nil {
			return nil
		}
	}

	return row
}

func glueAsEvery(matchers []match.Matcher) match.Matcher {
	if len(matchers) <= 1 {
		return nil
	}

	var (
		hasAny    bool
		hasSuper  bool
		hasSingle bool
		min       int
		separator string
	)

	for i, matcher := range matchers {
		var sep string
		switch m := matcher.(type) {

		case match.Super:
			sep = ""
			hasSuper = true

		case match.Any:
			sep = m.Separators
			hasAny = true

		case match.Single:
			sep = m.Separators
			hasSingle = true
			min++

		case match.List:
			if !m.Not {
				return nil
			}
			sep = m.List
			hasSingle = true
			min++

		default:
			return nil
		}

		// initialize
		if i == 0 {
			separator = sep
		}

		if sep == separator {
			continue
		}

		return nil
	}

	if hasSuper && !hasAny && !hasSingle {
		return match.Super{}
	}

	if hasAny && !hasSuper && !hasSingle {
		return match.Any{separator}
	}

	if (hasAny || hasSuper) && min > 0 && separator == "" {
		return match.Min{min}
	}

	every := match.EveryOf{}

	if min > 0 {
		every.Add(match.Min{min})

		if !hasAny && !hasSuper {
			every.Add(match.Max{min})
		}
	}

	if separator != "" {
		every.Add(match.Contains{separator, true})
	}

	return every
}

func convertMatchers(matchers []match.Matcher) []match.Matcher {
	var done match.Matcher
	var left, right, count int

	for l := 0; l < len(matchers); l++ {
		for r := len(matchers); r > l; r-- {
			if glued := glueMatchers(matchers[l:r]); glued != nil {
				var swap bool

				if done == nil {
					swap = true
				} else {
					cl, gl := done.Len(), glued.Len()
					swap = cl > -1 && gl > -1 && gl > cl

					swap = swap || count < r-l
				}

				if swap {
					done = glued
					left = l
					right = r
					count = r - l
				}
			}
		}
	}

	if done == nil {
		return matchers
	}

	next := append(append([]match.Matcher{}, matchers[:left]...), done)
	if right < len(matchers) {
		next = append(next, matchers[right:]...)
	}

	if len(next) == len(matchers) {
		return next
	}

	return convertMatchers(next)
}

func compileMatchers(matchers []match.Matcher) (match.Matcher, error) {
	if len(matchers) == 0 {
		return nil, fmt.Errorf("compile error: need at least one matcher")
	}

	if len(matchers) == 1 {
		return matchers[0], nil
	}

	if m := glueMatchers(matchers); m != nil {
		return m, nil
	}

	var (
		val match.Matcher
		idx int
	)
	maxLen := -1
	for i, matcher := range matchers {
		l := matcher.Len()
		if l >= maxLen {
			maxLen = l
			idx = i
			val = matcher
		}
	}

	//	_, ok := val.(match.BTree)
	//	fmt.Println("a tree", ok)

	left := matchers[:idx]
	var right []match.Matcher
	if len(matchers) > idx+1 {
		right = matchers[idx+1:]
	}

	tree := match.BTree{Value: val}

	if len(left) > 0 {
		l, err := compileMatchers(left)
		if err != nil {
			return nil, err
		}

		tree.Left = l
	}

	if len(right) > 0 {
		r, err := compileMatchers(right)
		if err != nil {
			return nil, err
		}

		tree.Right = r
	}

	return tree, nil
}

func do(node node, s string) (m match.Matcher, err error) {
	switch n := node.(type) {

	case *nodePattern, *nodeAnyOf:
		var matchers []match.Matcher
		for _, desc := range node.children() {
			m, err := do(desc, s)
			if err != nil {
				return nil, err
			}
			matchers = append(matchers, optimize(m))
		}

		if _, ok := node.(*nodeAnyOf); ok {
			m = match.AnyOf{matchers}
		} else {
			m, err = compileMatchers(convertMatchers(matchers))
			if err != nil {
				return nil, err
			}
		}

	case *nodeList:
		m = match.List{n.chars, n.not}

	case *nodeRange:
		m = match.Range{n.lo, n.hi, n.not}

	case *nodeAny:
		m = match.Any{s}

	case *nodeSuper:
		m = match.Super{}

	case *nodeSingle:
		m = match.Single{s}

	case *nodeText:
		m = match.Raw{n.text}

	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type")
	}

	return optimize(m), nil
}

func do2(node node, s string) ([]match.Matcher, error) {
	var result []match.Matcher

	switch n := node.(type) {

	case *nodePattern:
		ways := [][]match.Matcher{[]match.Matcher{}}

		for _, desc := range node.children() {
			variants, err := do2(desc, s)
			if err != nil {
				return nil, err
			}

			fmt.Println("variants pat", variants)

			for i, l := 0, len(ways); i < l; i++ {
				for i := 0; i < len(variants); i++ {
					o := optimize(variants[i])
					if i == len(variants)-1 {
						ways[i] = append(ways[i], o)
					} else {
						var w []match.Matcher
						copy(w, ways[i])
						ways = append(ways, append(w, o))
					}
				}
			}

			fmt.Println("ways pat", ways)
		}

		for _, matchers := range ways {
			c, err := compileMatchers(convertMatchers(matchers))
			if err != nil {
				return nil, err
			}
			result = append(result, c)
		}

	case *nodeAnyOf:
		ways := make([][]match.Matcher, len(node.children()))
		for _, desc := range node.children() {
			variants, err := do2(desc, s)
			if err != nil {
				return nil, err
			}

			fmt.Println("variants any", variants)

			for x, l := 0, len(ways); x < l; x++ {
				for i := 0; i < len(variants); i++ {
					o := optimize(variants[i])
					if i == len(variants)-1 {
						ways[x] = append(ways[x], o)
					} else {
						var w []match.Matcher
						copy(w, ways[x])
						ways = append(ways, append(w, o))
					}
				}
			}

			fmt.Println("ways any", ways)
		}

		for _, matchers := range ways {
			c, err := compileMatchers(convertMatchers(matchers))
			if err != nil {
				return nil, err
			}
			result = append(result, c)
		}

	case *nodeList:
		result = append(result, match.List{n.chars, n.not})

	case *nodeRange:
		result = append(result, match.Range{n.lo, n.hi, n.not})

	case *nodeAny:
		result = append(result, match.Any{s})

	case *nodeSuper:
		result = append(result, match.Super{})

	case *nodeSingle:
		result = append(result, match.Single{s})

	case *nodeText:
		result = append(result, match.Raw{n.text})

	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type")
	}

	for i, m := range result {
		result[i] = optimize(m)
	}

	return result, nil
}

func compile(ast *nodePattern, s string) (Glob, error) {
	//	ms, err := do2(ast, s)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if len(ms) == 1 {
	//		return ms[0], nil
	//	} else {
	//		return match.AnyOf{ms}, nil
	//	}

	g, err := do(ast, s)
	if err != nil {
		return nil, err
	}

	return g, nil
}

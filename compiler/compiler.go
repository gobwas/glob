package compiler

// TODO use constructor with all matchers, and to their structs private
// TODO glue multiple Text nodes (like after QuoteMeta)

import (
	"fmt"

	"github.com/gobwas/glob/internal/debug"
	"github.com/gobwas/glob/match"
	"github.com/gobwas/glob/syntax/ast"
)

func Compile(tree *ast.Node, sep []rune) (match.Matcher, error) {
	m, err := compile(tree, sep)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func compileNodes(ns []*ast.Node, sep []rune) ([]match.Matcher, error) {
	var matchers []match.Matcher
	for _, n := range ns {
		m, err := compile(n, sep)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}

func compile(node *ast.Node, sep []rune) (m match.Matcher, err error) {
	if debug.Enabled {
		debug.EnterPrefix("compiler: compiling %s", node)
		defer func() {
			if err != nil {
				debug.Logf("->! %v", err)
			} else {
				debug.Logf("-> %s", m)
			}
			debug.LeavePrefix()
		}()
	}

	// todo this could be faster on pattern_alternatives_combine_lite (see glob_test.go)
	if n := ast.Minimize(node); n != nil {
		debug.Logf("minimized tree -> %s", node, n)
		r, err := compile(n, sep)
		if debug.Enabled {
			if err != nil {
				debug.Logf("compiler: compile minimized tree failed: %v", err)
			} else {
				debug.Logf("compiler: minimized tree")
				debug.Logf("compiler: \t%s", node)
				debug.Logf("compiler: \t%s", n)
			}
		}
		if err == nil {
			return r, nil
		}
	}

	switch node.Kind {
	case ast.KindAnyOf:
		matchers, err := compileNodes(node.Children, sep)
		if err != nil {
			return nil, err
		}
		return match.NewAnyOf(matchers...), nil

	case ast.KindPattern:
		if len(node.Children) == 0 {
			return match.NewNothing(), nil
		}
		matchers, err := compileNodes(node.Children, sep)
		if err != nil {
			return nil, err
		}
		m, err = match.Compile(match.Minimize(matchers))
		if err != nil {
			return nil, err
		}

	case ast.KindAny:
		m = match.NewAny(sep)

	case ast.KindSuper:
		m = match.NewSuper()

	case ast.KindSingle:
		m = match.NewSingle(sep)

	case ast.KindNothing:
		m = match.NewNothing()

	case ast.KindList:
		l := node.Value.(ast.List)
		m = match.NewList([]rune(l.Chars), l.Not)

	case ast.KindRange:
		r := node.Value.(ast.Range)
		m = match.NewRange(r.Lo, r.Hi, r.Not)

	case ast.KindText:
		t := node.Value.(ast.Text)
		m = match.NewText(t.Text)

	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type")
	}

	return match.Optimize(m), nil
}

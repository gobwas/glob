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

func compile(tree *ast.Node, sep []rune) (m match.Matcher, err error) {
	if debug.Enabled {
		debug.Enter()
		debug.Logf("compiler: compiling %s", tree)
		defer func() {
			debug.Logf("compiler: result %s", m)
			debug.Leave()
		}()
	}

	// todo this could be faster on pattern_alternatives_combine_lite (see glob_test.go)
	if n := ast.Minimize(tree); n != nil {
		r, err := compile(n, sep)
		if debug.Enabled {
			if err != nil {
				debug.Logf("compiler: compile minimized tree failed: %v", err)
			} else {
				debug.Logf("compiler: minimized tree")
				debug.Logf("compiler: \t%s", tree)
				debug.Logf("compiler: \t%s", n)
			}
		}
		if err == nil {
			return r, nil
		}
	}

	switch tree.Kind {
	case ast.KindAnyOf:
		matchers, err := compileNodes(tree.Children, sep)
		if err != nil {
			return nil, err
		}
		return match.NewAnyOf(matchers...), nil

	case ast.KindPattern:
		if len(tree.Children) == 0 {
			return match.NewNothing(), nil
		}
		matchers, err := compileNodes(tree.Children, sep)
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
		l := tree.Value.(ast.List)
		m = match.NewList([]rune(l.Chars), l.Not)

	case ast.KindRange:
		r := tree.Value.(ast.Range)
		m = match.NewRange(r.Lo, r.Hi, r.Not)

	case ast.KindText:
		t := tree.Value.(ast.Text)
		m = match.NewText(t.Text)

	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type")
	}

	return match.Optimize(m), nil
}

package glob

import (
	"github.com/gobwas/glob/match"
	"reflect"
	"testing"
)

const separators = "."

func TestGlueMatchers(t *testing.T) {
	for id, test := range []struct {
		in  []match.Matcher
		exp match.Matcher
	}{
		{
			[]match.Matcher{
				match.Super{},
				match.Single{},
			},
			match.Min{1},
		},
		{
			[]match.Matcher{
				match.Any{separators},
				match.Single{separators},
			},
			match.EveryOf{match.Matchers{
				match.Min{1},
				match.Contains{separators, true},
			}},
		},
		{
			[]match.Matcher{
				match.Single{},
				match.Single{},
				match.Single{},
			},
			match.EveryOf{match.Matchers{
				match.Min{3},
				match.Max{3},
			}},
		},
		{
			[]match.Matcher{
				match.List{"a", true},
				match.Any{"a"},
			},
			match.EveryOf{match.Matchers{
				match.Min{1},
				match.Contains{"a", true},
			}},
		},
	} {
		act, err := compileMatchers(test.in)
		if err != nil {
			t.Errorf("#%d convert matchers error: %s", id, err)
			continue
		}

		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers result:\nact: %s;\nexp: %s", id, act, test.exp)
			continue
		}
	}
}

func TestCompileMatchers(t *testing.T) {
	for id, test := range []struct {
		in  []match.Matcher
		exp match.Matcher
	}{
		{
			[]match.Matcher{
				match.Super{},
				match.Single{separators},
				match.Raw{"c"},
			},
			match.BTree{
				Left: match.BTree{
					Left:  match.Super{},
					Value: match.Single{separators},
				},
				Value: match.Raw{"c"},
			},
		},
		{
			[]match.Matcher{
				match.Any{},
				match.Raw{"c"},
				match.Any{},
			},
			match.BTree{
				Left:  match.Any{},
				Value: match.Raw{"c"},
				Right: match.Any{},
			},
		},
		{
			[]match.Matcher{
				match.Range{'a', 'c', true},
				match.List{"zte", false},
				match.Raw{"c"},
				match.Single{},
			},
			match.Row{Matchers: match.Matchers{
				match.Range{'a', 'c', true},
				match.List{"zte", false},
				match.Raw{"c"},
				match.Single{},
			}},
		},
	} {
		act, err := compileMatchers(test.in)
		if err != nil {
			t.Errorf("#%d convert matchers error: %s", id, err)
			continue
		}

		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers result:\nact: %s;\nexp: %s", id, act, test.exp)
			continue
		}
	}
}

func TestConvertMatchers(t *testing.T) {
	for id, test := range []struct {
		in, exp []match.Matcher
	}{
		{
			[]match.Matcher{
				match.Range{'a', 'c', true},
				match.List{"zte", false},
				match.Raw{"c"},
				match.Single{},
				match.Any{},
			},
			[]match.Matcher{
				match.Row{Matchers: match.Matchers{
					match.Range{'a', 'c', true},
					match.List{"zte", false},
					match.Raw{"c"},
					match.Single{},
				}},
				match.Any{},
			},
		},
		{
			[]match.Matcher{
				match.Range{'a', 'c', true},
				match.List{"zte", false},
				match.Raw{"c"},
				match.Single{},
				match.Any{},
				match.Single{},
				match.Single{},
				match.Any{},
			},
			[]match.Matcher{
				match.Row{Matchers: match.Matchers{
					match.Range{'a', 'c', true},
					match.List{"zte", false},
					match.Raw{"c"},
					match.Single{},
				}},
				match.Min{2},
			},
		},
	} {
		act := convertMatchers(test.in, nil)
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers 2 result:\nact: %s;\nexp: %s", id, act, test.exp)
			continue
		}
	}
}

func pattern(nodes ...node) *nodePattern {
	return &nodePattern{
		nodeImpl: nodeImpl{
			desc: nodes,
		},
	}
}
func anyOf(nodes ...node) *nodeAnyOf {
	return &nodeAnyOf{
		nodeImpl: nodeImpl{
			desc: nodes,
		},
	}
}
func TestCompiler(t *testing.T) {
	for id, test := range []struct {
		ast    *nodePattern
		result Glob
		sep    string
	}{
		{
			ast:    pattern(&nodeText{text: "abc"}),
			result: match.Raw{"abc"},
		},
		{
			ast:    pattern(&nodeAny{}),
			sep:    separators,
			result: match.Any{separators},
		},
		{
			ast:    pattern(&nodeAny{}),
			result: match.Super{},
		},
		{
			ast:    pattern(&nodeSuper{}),
			result: match.Super{},
		},
		{
			ast:    pattern(&nodeSingle{}),
			sep:    separators,
			result: match.Single{separators},
		},
		{
			ast: pattern(&nodeRange{
				lo:  'a',
				hi:  'z',
				not: true,
			}),
			result: match.Range{'a', 'z', true},
		},
		{
			ast: pattern(&nodeList{
				chars: "abc",
				not:   true,
			}),
			result: match.List{"abc", true},
		},
		{
			ast: pattern(&nodeAny{}, &nodeSingle{}, &nodeSingle{}, &nodeSingle{}),
			sep: separators,
			result: match.EveryOf{Matchers: match.Matchers{
				match.Min{3},
				match.Contains{separators, true},
			}},
		},
		{
			ast:    pattern(&nodeAny{}, &nodeSingle{}, &nodeSingle{}, &nodeSingle{}),
			result: match.Min{3},
		},
		{
			ast: pattern(&nodeAny{}, &nodeText{text: "abc"}, &nodeSingle{}),
			sep: separators,
			result: match.BTree{
				Left: match.Any{separators},
				Value: match.Row{Matchers: match.Matchers{
					match.Raw{"abc"},
					match.Single{separators},
				}},
			},
		},
		{
			ast: pattern(&nodeSuper{}, &nodeSingle{}, &nodeText{text: "abc"}, &nodeSingle{}),
			sep: separators,
			result: match.BTree{
				Left: match.Super{},
				Value: match.Row{Matchers: match.Matchers{
					match.Single{separators},
					match.Raw{"abc"},
					match.Single{separators},
				}},
			},
		},
		{
			ast:    pattern(&nodeAny{}, &nodeText{text: "abc"}),
			result: match.Suffix{"abc"},
		},
		{
			ast:    pattern(&nodeText{text: "abc"}, &nodeAny{}),
			result: match.Prefix{"abc"},
		},
		{
			ast:    pattern(&nodeText{text: "abc"}, &nodeAny{}, &nodeText{text: "def"}),
			result: match.PrefixSuffix{"abc", "def"},
		},
		{
			ast:    pattern(&nodeAny{}, &nodeAny{}, &nodeAny{}, &nodeText{text: "abc"}, &nodeAny{}, &nodeAny{}),
			result: match.Contains{"abc", false},
		},
		{
			ast:    pattern(&nodeAny{}, &nodeAny{}, &nodeAny{}, &nodeText{text: "abc"}, &nodeAny{}, &nodeAny{}),
			sep:    separators,
			result: match.BTree{Left: match.Any{separators}, Value: match.Raw{"abc"}, Right: match.Any{separators}},
		},
		{
			ast: pattern(&nodeSuper{}, &nodeSingle{}, &nodeText{text: "abc"}, &nodeSuper{}, &nodeSingle{}),
			result: match.BTree{
				Left:  match.Min{1},
				Value: match.Raw{"abc"},
				Right: match.Min{1},
			},
		},
		{
			ast: pattern(anyOf(&nodeText{text: "abc"})),
			result: match.AnyOf{match.Matchers{
				match.Raw{"abc"},
			}},
		},
		{
			ast: pattern(anyOf(pattern(anyOf(pattern(&nodeText{text: "abc"}))))),
			result: match.AnyOf{match.Matchers{
				match.AnyOf{match.Matchers{
					match.Raw{"abc"},
				}},
			}},
		},
		{
			ast: pattern(
				&nodeRange{lo: 'a', hi: 'z'},
				&nodeRange{lo: 'a', hi: 'x', not: true},
				&nodeAny{},
			),
			result: match.BTree{
				Value: match.Row{Matchers: match.Matchers{
					match.Range{Lo: 'a', Hi: 'z'},
					match.Range{Lo: 'a', Hi: 'x', Not: true},
				}},
				Right: match.Super{},
			},
		},
		//		{
		//			ast: pattern(
		//				anyOf(&nodeText{text: "a"}, &nodeText{text: "b"}),
		//				anyOf(&nodeText{text: "c"}, &nodeText{text: "d"}),
		//			),
		//			result: match.AnyOf{Matchers: match.Matchers{
		//				match.Row{Matchers: match.Matchers{match.Raw{"a"}, match.Raw{"c"}}},
		//				match.Row{Matchers: match.Matchers{match.Raw{"a"}, match.Raw{"d"}}},
		//				match.Row{Matchers: match.Matchers{match.Raw{"b"}, match.Raw{"c"}}},
		//				match.Row{Matchers: match.Matchers{match.Raw{"b"}, match.Raw{"d"}}},
		//			}},
		//		},
	} {
		prog, err := compile(test.ast, test.sep)
		if err != nil {
			t.Errorf("compilation error: %s", err)
			continue
		}

		if !reflect.DeepEqual(prog, test.result) {
			t.Errorf("#%d results are not equal:\nexp: %s,\nact: %s", id, test.result, prog)
			continue
		}
	}
}

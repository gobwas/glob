package match

import (
	"reflect"
	"testing"

	"github.com/gobwas/glob/match"
)

func TestCompile(t *testing.T) {
	for id, test := range []struct {
		in  []Matcher
		exp Matcher
	}{
		{
			[]Matcher{
				NewSuper(),
				NewSingle(nil),
			},
			NewMin(1),
		},
		{
			[]Matcher{
				NewAny(separators),
				NewSingle(separators),
			},
			NewEveryOf([]Matcher{
				NewMin(1),
				NewContains(string(separators)),
			}),
		},
		{
			[]Matcher{
				NewSingle(nil),
				NewSingle(nil),
				NewSingle(nil),
			},
			NewEveryOf([]Matcher{
				NewMin(3),
				NewMax(3),
			}),
		},
		{
			[]Matcher{
				NewList([]rune{'a'}, true),
				NewAny([]rune{'a'}),
			},
			NewEveryOf([]Matcher{
				NewMin(1),
				NewContains("a"),
			}),
		},
		{
			[]Matcher{
				NewSuper(),
				NewSingle(separators),
				NewText("c"),
			},
			NewTree(
				NewText("c"),
				NewBTree(
					NewSingle(separators),
					NewSuper(),
					nil,
				),
				nil,
			),
		},
		{
			[]Matcher{
				NewAny(nil),
				NewText("c"),
				NewAny(nil),
			},
			NewTree(
				NewText("c"),
				NewAny(nil),
				NewAny(nil),
			),
		},
		{
			[]Matcher{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
			},
			NewRow([]MatchIndexSizer{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
			}),
		},
	} {
		act, err := Compile(test.in)
		if err != nil {
			t.Errorf("#%d compile matchers error: %s", id, err)
			continue
		}
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected compile matchers result:\nact: %#v;\nexp: %#v", id, act, test.exp)
			continue
		}
	}
}

func TestMinimize(t *testing.T) {
	for id, test := range []struct {
		in, exp []match.Matcher
	}{
		{
			[]match.Matcher{
				match.NewRange('a', 'c', true),
				match.NewList([]rune{'z', 't', 'e'}, false),
				match.NewText("c"),
				match.NewSingle(nil),
				match.NewAny(nil),
			},
			[]match.Matcher{
				match.NewRow(
					4,
					[]match.Matcher{
						match.NewRange('a', 'c', true),
						match.NewList([]rune{'z', 't', 'e'}, false),
						match.NewText("c"),
						match.NewSingle(nil),
					}...,
				),
				match.NewAny(nil),
			},
		},
		{
			[]match.Matcher{
				match.NewRange('a', 'c', true),
				match.NewList([]rune{'z', 't', 'e'}, false),
				match.NewText("c"),
				match.NewSingle(nil),
				match.NewAny(nil),
				match.NewSingle(nil),
				match.NewSingle(nil),
				match.NewAny(nil),
			},
			[]match.Matcher{
				match.NewRow(
					3,
					match.Matchers{
						match.NewRange('a', 'c', true),
						match.NewList([]rune{'z', 't', 'e'}, false),
						match.NewText("c"),
					}...,
				),
				match.NewMin(3),
			},
		},
	} {
		act := minimizeMatchers(test.in)
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers 2 result:\nact: %#v\nexp: %#v", id, act, test.exp)
			continue
		}
	}
}

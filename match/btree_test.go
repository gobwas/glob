package match

import (
	"testing"
)

func TestBTree(t *testing.T) {
	for id, test := range []struct {
		tree BTree
		str  string
		exp  bool
	}{
		{
			BTree{Value: Raw{"abc"}, Left: Super{}, Right: Super{}},
			"abc",
			true,
		},
		{
			BTree{Value: Raw{"a"}, Left: Single{}, Right: Single{}},
			"aaa",
			true,
		},
		{
			BTree{Value: Raw{"b"}, Left: Single{}},
			"bbb",
			false,
		},
		{
			BTree{
				Left: BTree{
					Left:  Super{},
					Value: Single{},
				},
				Value: Raw{"c"},
			},
			"abc",
			true,
		},
	} {
		act := test.tree.Match(test.str)
		if act != test.exp {
			t.Errorf("#%d match %q error: act: %t; exp: %t", id, test.str, act, test.exp)
			continue
		}
	}
}

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
			BTree{Value: Text{"abc"}, Left: Super{}, Right: Super{}},
			"abc",
			true,
		},
		{
			BTree{Value: Text{"a"}, Left: Single{}, Right: Single{}},
			"aaa",
			true,
		},
		{
			BTree{Value: Text{"b"}, Left: Single{}},
			"bbb",
			false,
		},
		{
			BTree{
				Left: BTree{
					Left:  Super{},
					Value: Single{},
				},
				Value: Text{"c"},
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

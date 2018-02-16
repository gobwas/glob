package match

import (
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {
	for _, test := range []struct {
		tree Matcher
		str  string
		exp  bool
	}{
		{
			NewTree(NewText("abc"), NewSuper(), NewSuper()),
			"abc",
			true,
		},
		{
			NewTree(NewText("a"), NewSingle(nil), NewSingle(nil)),
			"aaa",
			true,
		},
		{
			NewTree(NewText("b"), NewSingle(nil), nil),
			"bbb",
			false,
		},
		{
			NewTree(
				NewText("c"),
				NewTree(
					NewSingle(nil),
					NewSuper(),
					nil,
				),
				nil,
			),
			"abc",
			true,
		},
	} {
		t.Run("", func(t *testing.T) {
			act := test.tree.Match(test.str)
			if act != test.exp {
				fmt.Println(Graphviz("NIL", test.tree))
				t.Errorf("match %q error: act: %t; exp: %t", test.str, act, test.exp)
			}
		})
	}
}

type fakeMatcher struct {
	len  int
	segn int
	name string
}

func (f *fakeMatcher) Match(string) bool {
	return true
}

func (f *fakeMatcher) Index(s string) (int, []int) {
	seg := make([]int, 0, f.segn)
	for x := 0; x < f.segn; x++ {
		seg = append(seg, f.segn)
	}
	return 0, seg
}

func (f *fakeMatcher) MinLen() int {
	return f.len
}

func (f *fakeMatcher) String() string {
	return f.name
}

func BenchmarkMatchTree(b *testing.B) {
	l := &fakeMatcher{4, 3, "left_fake"}
	r := &fakeMatcher{4, 3, "right_fake"}
	v := &fakeMatcher{2, 3, "value_fake"}

	// must be <= len(l + r + v)
	fixture := "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"

	bt := NewTree(v, l, r)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bt.Match(fixture)
		}
	})
}

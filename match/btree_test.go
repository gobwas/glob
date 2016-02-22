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
			NewBTree(NewText("abc"), Super{}, Super{}),
			"abc",
			true,
		},
		{
			NewBTree(NewText("a"), Single{}, Single{}),
			"aaa",
			true,
		},
		{
			NewBTree(NewText("b"), Single{}, nil),
			"bbb",
			false,
		},
		{
			NewBTree(
				NewText("c"),
				NewBTree(
					Single{},
					Super{},
					nil,
				),
				nil,
			),
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

//type Matcher interface {
//	Match(string) bool
//	Index(string, []int) (int, []int)
//	Len() int
//	String() string
//}

type fakeMatcher struct {
	len  int
	name string
}

func (f *fakeMatcher) Match(string) bool {
	return true
}
func (f *fakeMatcher) Index(s string, seg []int) (int, []int) {
	return 0, append(seg, 1)
}
func (f *fakeMatcher) Len() int {
	return f.len
}
func (f *fakeMatcher) String() string {
	return f.name
}

func BenchmarkMatchBTree(b *testing.B) {
	l := &fakeMatcher{4, "left_fake"}
	r := &fakeMatcher{4, "right_fake"}
	v := &fakeMatcher{2, "value_fake"}

	// must be <= len(l + r + v)
	fixture := "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"

	bt := NewBTree(v, l, r)

	b.SetParallelism(1)
	for i := 0; i < b.N; i++ {
		bt.Match(fixture)
	}
}

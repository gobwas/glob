package match

import (
	"reflect"
	"testing"
)

func TestListIndex(t *testing.T) {
	for id, test := range []struct {
		list     []rune
		not      bool
		fixture  string
		index    int
		segments []int
	}{
		{
			[]rune("ab"),
			false,
			"abc",
			0,
			[]int{1},
		},
		{
			[]rune("ab"),
			true,
			"fffabfff",
			0,
			[]int{1},
		},
	} {
		p := List{test.list, test.not}
		index, segments := p.Index(test.fixture, []int{})
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func BenchmarkIndexList(b *testing.B) {
	m := List{[]rune("def"), false}
	in := make([]int, 0, len(bench_pattern))

	for i := 0; i < b.N; i++ {
		_, in = m.Index(bench_pattern, in[:0])
	}
}

func BenchmarkIndexListParallel(b *testing.B) {
	m := List{[]rune("def"), false}
	in := make([]int, 0, len(bench_pattern))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, in = m.Index(bench_pattern, in[:0])
		}
	})
}

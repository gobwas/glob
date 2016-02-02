package match

import (
	"reflect"
	"testing"
)

func TestRangeIndex(t *testing.T) {
	for id, test := range []struct {
		lo, hi   rune
		not      bool
		fixture  string
		index    int
		segments []int
	}{
		{
			'a', 'z',
			false,
			"abc",
			0,
			[]int{1},
		},
		{
			'a', 'c',
			false,
			"abcd",
			0,
			[]int{1},
		},
		{
			'a', 'c',
			true,
			"abcd",
			3,
			[]int{1},
		},
	} {
		m := Range{test.lo, test.hi, test.not}
		index, segments := m.Index(test.fixture, []int{})
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func BenchmarkIndexRange(b *testing.B) {
	m := Range{'0', '9', false}
	in := acquireSegments(len(bench_pattern))

	for i := 0; i < b.N; i++ {
		m.Index(bench_pattern, in[:0])
	}
}

func BenchmarkIndexRangeParallel(b *testing.B) {
	m := Range{'0', '9', false}
	in := acquireSegments(len(bench_pattern))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Index(bench_pattern, in[:0])
		}
	})
}

package match

import (
	"reflect"
	"testing"
)

func TestRowIndex(t *testing.T) {
	for id, test := range []struct {
		matchers []MatchIndexSizer
		fixture  string
		index    int
		segments []int
	}{
		{
			[]MatchIndexSizer{
				NewText("abc"),
				NewText("def"),
				NewSingle(nil),
			},
			"qweabcdefghij",
			3,
			[]int{7},
		},
		{
			[]MatchIndexSizer{
				NewText("abc"),
				NewText("def"),
				NewSingle(nil),
			},
			"abcd",
			-1,
			nil,
		},
	} {
		p := NewRow(test.matchers)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func BenchmarkRowIndex(b *testing.B) {
	m := NewRow([]MatchIndexSizer{
		NewText("abc"),
		NewText("def"),
		NewSingle(nil),
	})
	for i := 0; i < b.N; i++ {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexRowParallel(b *testing.B) {
	m := NewRow([]MatchIndexSizer{
		NewText("abc"),
		NewText("def"),
		NewSingle(nil),
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

package match

import (
	"reflect"
	"testing"
)

func TestIndexedAnyOf(t *testing.T) {
	for id, test := range []struct {
		matchers Matchers
		fixture  string
		index    int
		segments []int
	}{
		{
			Matchers{
				NewAny(nil),
				NewText("b"),
				NewText("c"),
			},
			"abc",
			0,
			[]int{0, 1, 2, 3},
		},
		{
			Matchers{
				NewPrefix("b"),
				NewSuffix("c"),
			},
			"abc",
			0,
			[]int{3},
		},
		{
			Matchers{
				NewList([]rune("[def]"), false),
				NewList([]rune("[abc]"), false),
			},
			"abcdef",
			0,
			[]int{1},
		},
	} {
		t.Run("", func(t *testing.T) {
			a := NewAnyOf(test.matchers...).(Indexer)
			index, segments := a.Index(test.fixture)
			if index != test.index {
				t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
			}
			if !reflect.DeepEqual(segments, test.segments) {
				t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
			}
		})
	}
}

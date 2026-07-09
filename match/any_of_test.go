package match

import (
	"reflect"
	"testing"
)

func TestAnyOfLen(t *testing.T) {
	for id, test := range []struct {
		matchers Matchers
		want     int
	}{
		{
			// all matchers have the same known length
			Matchers{NewText("abc"), NewText("xyz")},
			3,
		},
		{
			// matchers have different known lengths
			Matchers{NewText("ab"), NewText("xyz")},
			-1,
		},
		{
			// first matcher has unknown length, second has known length
			Matchers{NewSuffix("/daxing"), NewText("daxing")},
			-1,
		},
		{
			// first matcher has known length, second has unknown length
			Matchers{NewText("daxing"), NewSuffix("/daxing")},
			-1,
		},
		{
			// all matchers have unknown length
			Matchers{NewSuffix("/a"), NewPrefix("b/")},
			-1,
		},
		{
			// single matcher with known length
			Matchers{NewText("hello")},
			5,
		},
		{
			// single matcher with unknown length
			Matchers{NewSuffix("hello")},
			-1,
		},
	} {
		anyOf := NewAnyOf(test.matchers...)
		got := anyOf.Len()
		if got != test.want {
			t.Errorf("#%d AnyOf.Len() = %d, want %d", id, got, test.want)
		}
	}
}

func TestAnyOfIndex(t *testing.T) {
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
		everyOf := NewAnyOf(test.matchers...)
		index, segments := everyOf.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

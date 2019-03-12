package ast

import (
	"testing"
)

func TestCommonChildren(t *testing.T) {
	for _, test := range []struct {
		nodes []*Node
		left  []*Node
		right []*Node
	}{
		{
			nodes: []*Node{
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"z"}),
					NewNode(KindText, Text{"c"}),
				),
			},
		},
		{
			nodes: []*Node{
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"z"}),
					NewNode(KindText, Text{"c"}),
				),
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"c"}),
				),
			},
			left: []*Node{
				NewNode(KindText, Text{"a"}),
			},
			right: []*Node{
				NewNode(KindText, Text{"c"}),
			},
		},
		{
			nodes: []*Node{
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"c"}),
					NewNode(KindText, Text{"d"}),
				),
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"c"}),
					NewNode(KindText, Text{"c"}),
					NewNode(KindText, Text{"d"}),
				),
			},
			left: []*Node{
				NewNode(KindText, Text{"a"}),
				NewNode(KindText, Text{"b"}),
			},
			right: []*Node{
				NewNode(KindText, Text{"c"}),
				NewNode(KindText, Text{"d"}),
			},
		},
		{
			nodes: []*Node{
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"c"}),
				),
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"b"}),
					NewNode(KindText, Text{"c"}),
				),
			},
			left: []*Node{
				NewNode(KindText, Text{"a"}),
				NewNode(KindText, Text{"b"}),
			},
			right: []*Node{
				NewNode(KindText, Text{"c"}),
			},
		},
		{
			nodes: []*Node{
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"d"}),
				),
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"d"}),
				),
				NewNode(KindNothing, nil,
					NewNode(KindText, Text{"a"}),
					NewNode(KindText, Text{"e"}),
				),
			},
			left: []*Node{
				NewNode(KindText, Text{"a"}),
			},
			right: []*Node{},
		},
	} {
		t.Run("", func(t *testing.T) {
			left, right := CommonChildren(test.nodes)
			if !Equal(left, test.left) {
				t.Errorf(
					"left, right := commonChildren(); left = %v; want %v",
					left, test.left,
				)
			}
			if !Equal(right, test.right) {
				t.Errorf(
					"left, right := commonChildren(); right = %v; want %v",
					right, test.right,
				)
			}
		})
	}
}

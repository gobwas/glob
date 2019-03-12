package ast

import (
	"reflect"
)

// Minimize tries to apply some heuristics to minimize number of nodes in given
// t
func Minimize(t *Node) *Node {
	switch t.Kind {
	case KindAnyOf:
		return minimizeAnyOf(t)
	default:
		return nil
	}
}

// minimizeAnyOf tries to find common children of given node of AnyOf pattern
// it searches for common children from left and from right
// if any common children are found – then it returns new optimized ast t
// else it returns nil
func minimizeAnyOf(t *Node) *Node {
	if !SameKind(t.Children, KindPattern) {
		return nil
	}

	commonLeft, commonRight := CommonChildren(t.Children)
	commonLeftCount, commonRightCount := len(commonLeft), len(commonRight)
	if commonLeftCount == 0 && commonRightCount == 0 { // there are no common parts
		return nil
	}

	var result []*Node
	if commonLeftCount > 0 {
		result = append(result, NewNode(KindPattern, nil, commonLeft...))
	}

	var anyOf []*Node
	for _, child := range t.Children {
		reuse := child.Children[commonLeftCount : len(child.Children)-commonRightCount]
		var node *Node
		if len(reuse) == 0 {
			// this pattern is completely reduced by commonLeft and commonRight patterns
			// so it become nothing
			node = NewNode(KindNothing, nil)
		} else {
			node = NewNode(KindPattern, nil, reuse...)
		}
		anyOf = AppendUnique(anyOf, node)
	}
	switch {
	case len(anyOf) == 1 && anyOf[0].Kind != KindNothing:
		result = append(result, anyOf[0])
	case len(anyOf) > 1:
		result = append(result, NewNode(KindAnyOf, nil, anyOf...))
	}

	if commonRightCount > 0 {
		result = append(result, NewNode(KindPattern, nil, commonRight...))
	}

	return NewNode(KindPattern, nil, result...)
}

func CommonChildren(nodes []*Node) (commonLeft, commonRight []*Node) {
	if len(nodes) <= 1 {
		return
	}

	// find node that has least number of children
	idx := OneWithLeastChildren(nodes)
	if idx == -1 {
		return
	}
	tree := nodes[idx]
	treeLength := len(tree.Children)

	// allocate max able size for rightCommon slice
	// to get ability insert elements in reverse order (from end to start)
	// without sorting
	commonRight = make([]*Node, treeLength)
	lastRight := treeLength // will use this to get results as commonRight[lastRight:]

	var (
		breakLeft   bool
		breakRight  bool
		commonTotal int
	)
	for i, j := 0, treeLength-1; commonTotal < treeLength && j >= 0 && !(breakLeft && breakRight); i, j = i+1, j-1 {
		treeLeft := tree.Children[i]
		treeRight := tree.Children[j]

		for k := 0; k < len(nodes) && !(breakLeft && breakRight); k++ {
			// skip least children node
			if k == idx {
				continue
			}

			restLeft := nodes[k].Children[i]
			restRight := nodes[k].Children[j+len(nodes[k].Children)-treeLength]

			breakLeft = breakLeft || !treeLeft.Equal(restLeft)

			// disable searching for right common parts, if left part is already overlapping
			breakRight = breakRight || (!breakLeft && j <= i)
			breakRight = breakRight || !treeRight.Equal(restRight)
		}

		if !breakLeft {
			commonTotal++
			commonLeft = append(commonLeft, treeLeft)
		}
		if !breakRight {
			commonTotal++
			lastRight = j
			commonRight[j] = treeRight
		}
	}

	commonRight = commonRight[lastRight:]

	return
}

func AppendUnique(target []*Node, val *Node) []*Node {
	for _, n := range target {
		if reflect.DeepEqual(n, val) {
			return target
		}
	}
	return append(target, val)
}

func SameKind(nodes []*Node, kind Kind) bool {
	for _, n := range nodes {
		if n.Kind != kind {
			return false
		}
	}
	return true
}

func OneWithLeastChildren(nodes []*Node) int {
	min := -1
	idx := -1
	for i, n := range nodes {
		if idx == -1 || (len(n.Children) < min) {
			min = len(n.Children)
			idx = i
		}
	}
	return idx
}

func Equal(a, b []*Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if !av.Equal(b[i]) {
			return false
		}
	}
	return true
}

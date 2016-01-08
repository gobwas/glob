package match

import (
	"fmt"
)

type BTree struct {
	Value       Primitive
	Left, Right Matcher
}

func (self BTree) Kind() Kind {
	return KindBTree
}

func (self BTree) Match(s string) bool {
	runes := []rune(s)
	inputLen := len(runes)

	for offset := 0; offset < inputLen; {
		index, min, max := self.Value.Index(string(runes[offset:]))

		if index == -1 {
			return false
		}

		for length := min; length <= max; length++ {
			var left, right bool

			l := string(runes[:offset+index])
			if self.Left != nil {
				left = self.Left.Match(l)
			} else {
				left = l == ""
			}

			if !left {
				break
			}

			var r string
			// if there is no string for the right branch
			if inputLen <= offset+index+length {
				r = ""
			} else {
				r = string(runes[offset+index+length:])
			}

			if self.Right != nil {
				right = self.Right.Match(r)
			} else {
				right = r == ""
			}

			if left && right {
				return true
			}
		}

		offset += index + 1
	}

	return false
}

func (self BTree) String() string {
	return fmt.Sprintf("[btree:%s<-%s->%s]", self.Left, self.Value, self.Right)
}

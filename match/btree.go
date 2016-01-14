package match

import (
	"fmt"
	"unicode/utf8"
)

type BTree struct {
	Value, Left, Right Matcher
	VLen, LLen, RLen   int
	Length             int
}

func NewBTree(Value, Left, Right Matcher) (tree BTree) {
	tree.Value = Value
	tree.Left = Left
	tree.Right = Right

	lenOk := true
	if tree.VLen = Value.Len(); tree.VLen == -1 {
		lenOk = false
	}

	if Left != nil {
		if tree.LLen = Left.Len(); tree.LLen == -1 {
			lenOk = false
		}
	}

	if Right != nil {
		if tree.RLen = Right.Len(); tree.RLen == -1 {
			lenOk = false
		}
	}

	if lenOk {
		tree.Length = tree.LLen + tree.VLen + tree.RLen
	} else {
		tree.Length = -1
	}

	return tree
}

func (self BTree) Kind() Kind {
	return KindBTree
}

func (self BTree) Len() int {
	return self.Length
}

// todo?
func (self BTree) Index(s string) (int, []int) {
	return -1, nil
}

func (self BTree) Match(s string) bool {
	inputLen := len(s)

	if self.Length != -1 && self.Length > inputLen {
		return false
	}

	var offset, limit int
	if self.LLen >= 0 {
		offset = self.LLen
	}
	if self.RLen >= 0 {
		limit = inputLen - self.RLen
	} else {
		limit = inputLen
	}

	for offset < limit {
		index, segments := self.Value.Index(s[offset:limit])
		if index == -1 {
			return false
		}

		l := string(s[:offset+index])
		var left bool
		if self.Left != nil {
			left = self.Left.Match(l)
		} else {
			left = l == ""
		}

		if left {
			for i := len(segments) - 1; i >= 0; i-- {
				length := segments[i]

				if self.RLen >= 0 && inputLen-(offset+index+length) != self.RLen {
					continue
				}

				var right bool

				var r string
				// if there is no string for the right branch
				if inputLen <= offset+index+length {
					r = ""
				} else {
					r = s[offset+index+length:]
				}

				if self.Right != nil {
					right = self.Right.Match(r)
				} else {
					right = r == ""
				}

				if right {
					return true
				}
			}
		}

		_, step := utf8.DecodeRuneInString(s[offset+index:])
		offset += index + step
	}

	return false
}

func (self BTree) String() string {
	return fmt.Sprintf("<btree:[%s<-%s->%s]>", self.Left, self.Value, self.Right)
}

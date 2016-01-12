package match

import (
	"fmt"
	"unicode/utf8"
)

type BTree struct {
	Value, Left, Right Matcher
}

func (self BTree) Kind() Kind {
	return KindBTree
}

func (self BTree) len() (l, v, r int, ok bool) {
	v = self.Value.Len()

	if self.Left != nil {
		l = self.Left.Len()
	}

	if self.Right != nil {
		r = self.Right.Len()
	}

	ok = l > -1 && v > -1 && r > -1

	return
}

func (self BTree) Len() int {
	l, v, r, ok := self.len()
	if ok {
		return l + v + r
	}

	return -1
}

// todo
func (self BTree) Index(s string) (int, []int) {
	return -1, nil
}

func (self BTree) Match(s string) bool {
	inputLen := len(s)

	lLen, vLen, rLen, ok := self.len()
	if ok && lLen+vLen+rLen > inputLen {
		return false
	}

	var offset, limit int
	if lLen >= 0 {
		offset = lLen
	}
	if rLen >= 0 {
		limit = inputLen - rLen
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

				if rLen >= 0 && inputLen-(offset+index+length) != rLen {
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

const tpl = `
"%p"[label="%s"]
"%p"[label="%s"]
"%p"[label="%s"]
"%p"->"%p"
"%p"->"%p"
`

func (self BTree) String() string {
	//	return fmt.Sprintf("[btree:%s<-%s->%s]", self.Left, self.Value, self.Right)

	l, r := "nil", "nil"
	if self.Left != nil {
		l = self.Left.String()
	}
	if self.Right != nil {
		r = self.Right.String()
	}
	return fmt.Sprintf(tpl, &self, self.Value, &l, l, &r, r, &self, &l, &self, &r)
}

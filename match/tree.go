package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

type Tree struct {
	value MatchIndexer
	left  Matcher
	right Matcher

	minLen int

	runes  int
	vrunes int
	lrunes int
	rrunes int
}

type SizedTree struct {
	Tree
}

func (st SizedTree) RunesCount() int {
	return st.Tree.runes
}

func NewTree(v MatchIndexer, l, r Matcher) Matcher {
	tree := Tree{
		value: v,
		left:  l,
		right: r,
	}
	tree.minLen = v.MinLen()
	if l != nil {
		tree.minLen += l.MinLen()
	}
	if r != nil {
		tree.minLen += r.MinLen()
	}
	var (
		ls, lsz = l.(Sizer)
		rs, rsz = r.(Sizer)
		vs, vsz = v.(Sizer)
	)
	if lsz {
		tree.lrunes = ls.RunesCount()
	} else {
		tree.lrunes = -1
	}
	if rsz {
		tree.rrunes = rs.RunesCount()
	} else {
		tree.rrunes = -1
	}
	if vsz {
		tree.vrunes = vs.RunesCount()
	} else {
		tree.vrunes = -1
	}
	if vsz && lsz && rsz {
		tree.runes = tree.vrunes + tree.lrunes + tree.rrunes
		return SizedTree{tree}
	}
	tree.runes = -1
	return tree
}

func (t Tree) MinLen() int {
	return t.minLen
}

func (t Tree) Match(s string) (ok bool) {
	enter()
	logf("matching %q: %v", s, t)
	defer func(s string) {
		logf("result: %q -> %v", s, ok)
		leave()
	}(s)

	offset, limit := t.offsetLimit(s)
	q := s[offset : len(s)-limit]
	logf("OFFSET/LIMIT: %d/%d %q of %q", offset, limit, q, s)

	for len(q) >= t.vrunes {
		// search for matching part in substring
		index, segments := t.value.Index(q)
		logf("INDEX #%d %q (%v)", index, q, t.value)
		if index == -1 {
			releaseSegments(segments)
			return false
		}

		l := s[:offset+index]
		var left bool
		if t.left != nil {
			left = t.left.Match(l)
		} else {
			left = l == ""
		}
		logf("LEFT %q %v", l, left)
		if left {
			for _, seg := range segments {
				var (
					right bool
				)
				r := s[offset+index+seg:]
				if t.right != nil {
					right = t.right.Match(r)
				} else {
					right = r == ""
				}
				logf("RIGHT %q %v", r, right)
				if right {
					releaseSegments(segments)
					return true
				}
			}
		}

		_, x := utf8.DecodeRuneInString(q[index:])
		releaseSegments(segments)
		q = q[x:]
		offset += x
		logf("SLICED TO %q", q)
	}

	return false
}

// Retuns substring and offset/limit pair in bytes.
func (t Tree) offsetLimit(s string) (offset, limit int) {
	n := utf8.RuneCountInString(s)
	if t.runes > n {
		return 0, 0
	}
	if n := t.lrunes; n > 0 {
		offset = len(runes.Head(s, n))
	}
	if n := t.rrunes; n > 0 {
		limit = len(runes.Tail(s, n))
	}
	return
}

func (t Tree) String() string {
	return fmt.Sprintf(
		"<btree:[%v<-%s->%v]>",
		t.left, t.value, t.right,
	)
}

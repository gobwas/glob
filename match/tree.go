package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/gobwas/glob/internal/debug"
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

type IndexedTree struct {
	value MatchIndexer
	left  MatchIndexer
	right MatchIndexer
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
	}
	if rsz {
		tree.rrunes = rs.RunesCount()
	}
	if vsz {
		tree.vrunes = vs.RunesCount()
	}
	//li, lix := l.(MatchIndexer)
	//ri, rix := r.(MatchIndexer)
	if vsz && lsz && rsz {
		tree.runes = tree.vrunes + tree.lrunes + tree.rrunes
		return SizedTree{tree}
	}
	return tree
}

func (t Tree) MinLen() int {
	return t.minLen
}

func (t Tree) Content(cb func(Matcher)) {
	if t.left != nil {
		cb(t.left)
	}
	cb(t.value)
	if t.right != nil {
		cb(t.right)
	}
}

func (t Tree) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("tree", s)
		defer func() { done(ok) }()
	}

	n := len(s)
	offset, limit := t.offsetLimit(s)

	for len(s)-offset-limit >= t.vrunes {
		if debug.Enabled {
			debug.Logf(
				"value %s indexing: %q (offset=%d; limit=%d)",
				t.value, s[offset:n-limit], offset, limit,
			)
		}
		index, segments := t.value.Index(s[offset : n-limit])
		if debug.Enabled {
			debug.Logf(
				"value %s index: %d; %v",
				t.value, index, segments,
			)
		}
		if index == -1 {
			releaseSegments(segments)
			return false
		}

		if debug.Enabled {
			debug.Logf("matching left: %q", s[:offset+index])
		}
		left := t.left.Match(s[:offset+index])
		if debug.Enabled {
			debug.Logf("matching left: -> %t", left)
		}

		if left {
			for _, seg := range segments {
				if debug.Enabled {
					debug.Logf("matching right: %q", s[offset+index+seg:])
				}
				right := t.right.Match(s[offset+index+seg:])
				if debug.Enabled {
					debug.Logf("matching right: -> %t", right)
				}
				if right {
					releaseSegments(segments)
					return true
				}
			}
		}

		releaseSegments(segments)

		_, x := utf8.DecodeRuneInString(s[offset+index:])
		if x == 0 {
			// No progress.
			break
		}
		offset = offset + index + x
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

package match

import (
	"fmt"

	"github.com/gobwas/glob/util/runes"
)

type Any struct {
	sep []rune
}

func NewAny(s []rune) Any {
	return Any{s}
}

func (a Any) Match(s string) bool {
	return runes.IndexAnyRune(s, a.sep) == -1
}

func (a Any) Index(s string) (int, []int) {
	found := runes.IndexAnyRune(s, a.sep)
	switch found {
	case -1:
	case 0:
		return 0, segments0
	default:
		s = s[:found]
	}

	segments := acquireSegments(len(s))
	for i := range s {
		segments = append(segments, i)
	}
	segments = append(segments, len(s))

	return 0, segments
}

func (a Any) MinLen() int {
	return 0
}

func (a Any) String() string {
	return fmt.Sprintf("<any:![%s]>", string(a.sep))
}

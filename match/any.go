package match

import (
	"fmt"
	"github.com/gobwas/glob/strings"
)

type Any struct {
	Separators []rune
}

func (self Any) Match(s string) bool {
	return strings.IndexAnyRunes(s, self.Separators) == -1
}

func (self Any) Index(s string) (int, []int) {
	found := strings.IndexAnyRunes(s, self.Separators)
	switch found {
	case -1:
	case 0:
		return 0, []int{0}
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

func (self Any) Len() int {
	return lenNo
}

func (self Any) String() string {
	return fmt.Sprintf("<any:![%s]>", self.Separators)
}

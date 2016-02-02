package match

import (
	"fmt"
	"strings"
)

type PrefixSuffix struct {
	Prefix, Suffix string
}

func (self PrefixSuffix) Index(s string, segments []int) (int, []int) {
	prefixIdx := strings.Index(s, self.Prefix)
	if prefixIdx == -1 {
		return -1, nil
	}

	suffixLen := len(self.Suffix)

	if suffixLen > 0 {
		for sub := s[prefixIdx:]; ; {
			suffixIdx := strings.LastIndex(sub, self.Suffix)
			if suffixIdx == -1 {
				break
			}

			segments = append(segments, suffixIdx+suffixLen)
			sub = sub[:suffixIdx]
		}

		if len(segments) == 0 {
			return -1, nil
		}

		reverseSegments(segments)
	} else {
		segments = append(segments, len(s)-prefixIdx)
	}

	return prefixIdx, segments
}

func (self PrefixSuffix) Len() int {
	return lenNo
}

func (self PrefixSuffix) Match(s string) bool {
	return strings.HasPrefix(s, self.Prefix) && strings.HasSuffix(s, self.Suffix)
}

func (self PrefixSuffix) String() string {
	return fmt.Sprintf("<prefix_suffix:[%s,%s]>", self.Prefix, self.Suffix)
}

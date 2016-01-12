package match

import (
	"fmt"
	"strings"
)

type PrefixSuffix struct {
	Prefix, Suffix string
}

func (self PrefixSuffix) Kind() Kind {
	return KindPrefixSuffix
}

func (self PrefixSuffix) Index(s string) (int, []int) {
	prefixIdx := strings.Index(s, self.Prefix)
	if prefixIdx == -1 {
		return -1, nil
	}

	var segments []int
	for sub := s[prefixIdx:]; ; {
		suffixIdx := strings.LastIndex(sub, self.Suffix)
		if suffixIdx == -1 {
			break
		}

		segments = append(segments, suffixIdx+len(self.Suffix))
		sub = s[:suffixIdx]
	}

	segLen := len(segments)
	if segLen == 0 {
		return -1, nil
	}

	resp := make([]int, segLen)
	for i, s := range segments {
		resp[segLen-i-1] = s
	}

	return prefixIdx, resp
}

func (self PrefixSuffix) Len() int {
	return -1
}

func (self PrefixSuffix) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self PrefixSuffix) Match(s string) bool {
	return strings.HasPrefix(s, self.Prefix) && strings.HasSuffix(s, self.Suffix)
}

func (self PrefixSuffix) String() string {
	return fmt.Sprintf("<prefix_suffix:[%s,%s]>", self.Prefix, self.Suffix)
}

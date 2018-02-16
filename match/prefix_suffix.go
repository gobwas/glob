package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type PrefixSuffix struct {
	p, s   string
	minLen int
}

func NewPrefixSuffix(p, s string) PrefixSuffix {
	pn := utf8.RuneCountInString(p)
	sn := utf8.RuneCountInString(s)
	return PrefixSuffix{p, s, pn + sn}
}

func (ps PrefixSuffix) Index(s string) (int, []int) {
	prefixIdx := strings.Index(s, ps.p)
	if prefixIdx == -1 {
		return -1, nil
	}

	suffixLen := len(ps.s)
	if suffixLen <= 0 {
		return prefixIdx, []int{len(s) - prefixIdx}
	}

	if (len(s) - prefixIdx) <= 0 {
		return -1, nil
	}

	segments := acquireSegments(len(s) - prefixIdx)
	for sub := s[prefixIdx:]; ; {
		suffixIdx := strings.LastIndex(sub, ps.s)
		if suffixIdx == -1 {
			break
		}

		segments = append(segments, suffixIdx+suffixLen)
		sub = sub[:suffixIdx]
	}

	if len(segments) == 0 {
		releaseSegments(segments)
		return -1, nil
	}

	reverseSegments(segments)

	return prefixIdx, segments
}

func (ps PrefixSuffix) Match(s string) bool {
	return strings.HasPrefix(s, ps.p) && strings.HasSuffix(s, ps.s)
}

func (ps PrefixSuffix) MinLen() int {
	return ps.minLen
}

func (ps PrefixSuffix) String() string {
	return fmt.Sprintf("<prefix_suffix:[%s,%s]>", ps.p, ps.s)
}

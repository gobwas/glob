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
	return fmt.Sprintf("[prefix_suffix:%s-%s]", self.Prefix, self.Suffix)
}

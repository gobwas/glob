package match

import (
	"fmt"
	"strings"
)

type Suffix struct {
	Suffix string
}

func (self Suffix) Kind() Kind {
	return KindSuffix
}

func (self Suffix) Index(s string) (int, []int) {
	idx := strings.Index(s, self.Suffix)
	if idx == -1 {
		return -1, nil
	}

	return 0, []int{idx + len(self.Suffix)}
}

func (self Suffix) Len() int {
	return lenNo
}

func (self Suffix) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self Suffix) Match(s string) bool {
	return strings.HasSuffix(s, self.Suffix)
}

func (self Suffix) String() string {
	return fmt.Sprintf("<suffix:%s>", self.Suffix)
}

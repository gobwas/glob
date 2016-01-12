package match

import (
	"fmt"
	"unicode/utf8"
)

type Min struct {
	Limit int
}

func (self Min) Match(s string) bool {
	return utf8.RuneCountInString(s) >= self.Limit
}

func (self Min) Index(s string) (int, []int) {
	var count int

	c := utf8.RuneCountInString(s)
	if c < self.Limit {
		return -1, nil
	}

	segments := make([]int, 0, c-self.Limit+1)
	for i, r := range s {
		count++
		if count >= self.Limit {
			segments = append(segments, i+utf8.RuneLen(r))
		}
	}

	return 0, segments
}

func (self Min) Len() int {
	return -1
}

func (self Min) Search(s string) (int, int, bool) {
	return 0, 0, false
}

func (self Min) Kind() Kind {
	return KindMin
}

func (self Min) String() string {
	return fmt.Sprintf("[min:%d]", self.Limit)
}

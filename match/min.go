package match

import (
	"fmt"
	"unicode/utf8"
)

type Min struct {
	Limit int
}

func (self Min) Match(s string) bool {
	var l int
	for range s {
		l += 1
		if l >= self.Limit {
			return true
		}
	}

	return false
}

func (self Min) Index(s string, segments []int) (int, []int) {
	var count int
	var found bool

	for i, r := range s {
		count++
		if count >= self.Limit {
			found = true
			segments = append(segments, i+utf8.RuneLen(r))
		}
	}

	if !found {
		return -1, nil
	}

	return 0, segments
}

func (self Min) Len() int {
	return lenNo
}

func (self Min) String() string {
	return fmt.Sprintf("<min:%d>", self.Limit)
}

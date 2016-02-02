package match

import (
	"fmt"
)

type Nothing struct{}

func (self Nothing) Match(s string) bool {
	return len(s) == 0
}

func (self Nothing) Index(s string, segments []int) (int, []int) {
	return 0, append(segments, 0)
}

func (self Nothing) Len() int {
	return lenZero
}

func (self Nothing) String() string {
	return fmt.Sprintf("<nothing>")
}

package match

import (
	"fmt"
)

type Super struct{}

func NewSuper() Super {
	return Super{}
}

func (s Super) Match(_ string) bool {
	return true
}

func (s Super) MinLen() int {
	return 0
}

func (s Super) Index(v string) (int, []int) {
	seg := acquireSegments(len(v) + 1)
	for i := range v {
		seg = append(seg, i)
	}
	seg = append(seg, len(v))
	return 0, seg
}

func (s Super) String() string {
	return fmt.Sprintf("<super>")
}

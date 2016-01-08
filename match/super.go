package match

import (
	"fmt"
)

type Super struct{}

func (self Super) Match(s string) bool {
	return true
}

func (self Super) Index(s string) (index, min, max int) {
	return 0, 0, len([]rune(s))
}

func (self Super) Kind() Kind {
	return KindSuper
}

func (self Super) String() string {
	return fmt.Sprintf("[super]")
}

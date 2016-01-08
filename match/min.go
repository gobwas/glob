package match

import "fmt"

type Min struct {
	Limit int
}

func (self Min) Match(s string) bool {
	return len([]rune(s)) >= self.Limit
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

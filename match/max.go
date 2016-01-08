package match

import "fmt"

type Max struct {
	Limit int
}

func (self Max) Match(s string) bool {
	return len([]rune(s)) <= self.Limit
}

func (self Max) Search(s string) (int, int, bool) {
	return 0, 0, false
}

func (self Max) Kind() Kind {
	return KindMax
}

func (self Max) String() string {
	return fmt.Sprintf("[max:%d]", self.Limit)
}

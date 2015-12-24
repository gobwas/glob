package match


import (
	"strings"
	"fmt"
)


type RangeList struct {
	List string
	Not  bool
}

func (self RangeList) Kind() Kind {
	return KindRangeList
}

func (self RangeList) Search(s string) (i int, l int, ok bool) {
	if self.Match(s) {
		return 0, len(s), true
	}

	return
}

func (self RangeList) Match(s string) bool {
	r := []rune(s)

	if (len(r) != 1) {
		return false
	}

	inList := strings.IndexRune(self.List, r[0]) >= 0

	return inList == !self.Not
}

func (self RangeList) String() string {
	return fmt.Sprintf("[range_list:%s]", self.List)
}

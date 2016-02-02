package match

import (
	"fmt"
	"strings"
	"sync"
)

const lenOne = 1
const lenZero = 0
const lenNo = -1

type Matcher interface {
	Match(string) bool
	Index(string, []int) (int, []int)
	Len() int
	String() string
}

type Matchers []Matcher

func (m Matchers) String() string {
	var s []string
	for _, matcher := range m {
		s = append(s, fmt.Sprint(matcher))
	}

	return fmt.Sprintf("%s", strings.Join(s, ","))
}

var segmentsPools [1024]sync.Pool

func toPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++

	return v
}

const (
	minSegment         = 32
	minSegmentMinusOne = 31
	maxSegment         = 1024
	maxSegmentMinusOne = 1023
)

func init() {
	for i := maxSegment; i >= minSegment; i >>= 1 {
		func(i int) {
			segmentsPools[i-1] = sync.Pool{
				New: func() interface{} {
					return make([]int, 0, i)
				},
			}
		}(i)
	}
}

func getIdx(c int) int {
	p := toPowerOfTwo(c)
	switch {
	case p >= maxSegment:
		return maxSegmentMinusOne
	case p <= minSegment:
		return minSegmentMinusOne
	default:
		return p - 1
	}
}

func acquireSegments(c int) []int {
	//	fmt.Println("GET", getIdx(c))
	return segmentsPools[getIdx(c)].Get().([]int)[:0]
}

func releaseSegments(s []int) {
	//	fmt.Println("PUT", getIdx(cap(s)))
	segmentsPools[getIdx(cap(s))].Put(s)
}

// appendMerge merges and sorts given already SORTED and UNIQUE segments.
func appendMerge(target, sub []int) []int {
	lt, ls := len(target), len(sub)
	out := acquireSegments(lt + ls)

	for x, y := 0, 0; x < lt || y < ls; {
		if x >= lt {
			out = append(out, sub[y:]...)
			break
		}

		if y >= ls {
			out = append(out, target[x:]...)
			break
		}

		xValue := target[x]
		yValue := sub[y]

		switch {

		case xValue == yValue:
			out = append(out, xValue)
			x++
			y++

		case xValue < yValue:
			out = append(out, xValue)
			x++

		case yValue < xValue:
			out = append(out, yValue)
			y++

		}
	}

	target = append(target[:0], out...)
	releaseSegments(out)

	return target
}

func reverseSegments(input []int) {
	l := len(input)
	m := l / 2

	for i := 0; i < m; i++ {
		input[i], input[l-i-1] = input[l-i-1], input[i]
	}
}

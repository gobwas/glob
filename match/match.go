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

func init() {
	for i := 1024; i >= 1; i >>= 1 {
		func(i int) {
			segmentsPools[i-1] = sync.Pool{
				New: func() interface{} {
					return make([]int, 0, i)
				},
			}
		}(i)
	}
}

var segmentsPool = sync.Pool{
	New: func() interface{} {
		return make([]int, 0, 64)
	},
}

func getIdx(c int) int {
	p := toPowerOfTwo(c)
	switch {
	case p >= 1024:
		return 1023
	case p < 1:
		return 0
	default:
		return p - 1
	}
}

func acquireSegments(c int) []int {
	return segmentsPools[getIdx(c)].Get().([]int)[:0]
}

func releaseSegments(s []int) {
	segmentsPools[getIdx(cap(s))].Put(s)
}

func appendIfNotAsPrevious(target []int, val int) []int {
	l := len(target)
	if l != 0 && target[l-1] == val {
		return target
	}

	return append(target, val)
}

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

// mergeSegments merges and sorts given already SORTED and UNIQUE segments.
func mergeSegments(list [][]int, out []int) []int {
	var current []int
	switch len(list) {
	case 0:
		return out
	case 1:
		return list[0]
	default:
		current = acquireSegments(len(list[0]))
		current = append(current, list[0]...)
		//		releaseSegments(list[0])
	}

	for _, s := range list[1:] {
		next := acquireSegments(len(current) + len(s))
		for x, y := 0, 0; x < len(current) || y < len(s); {
			if x >= len(current) {
				next = append(next, s[y:]...)
				break
			}

			if y >= len(s) {
				next = append(next, current[x:]...)
				break
			}

			xValue := current[x]
			yValue := s[y]

			switch {

			case xValue == yValue:
				x++
				y++
				next = appendIfNotAsPrevious(next, xValue)

			case xValue < yValue:
				next = appendIfNotAsPrevious(next, xValue)
				x++

			case yValue < xValue:
				next = appendIfNotAsPrevious(next, yValue)
				y++

			}
		}

		releaseSegments(current)
		current = next
	}

	out = append(out, current...)
	releaseSegments(current)

	return out
}

func reverseSegments(input []int) {
	l := len(input)
	m := l / 2

	for i := 0; i < m; i++ {
		input[i], input[l-i-1] = input[l-i-1], input[i]
	}
}

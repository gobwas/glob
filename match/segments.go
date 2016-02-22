package match

import "sync"

var segmentsPools [1024]*PoolSequenced

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
			//			pool := sync.Pool{
			//				New: func() interface{} {
			//					//					fmt.Printf("N%d;", i)
			//					return make([]int, 0, i)
			//				},
			//			}

			pool := NewPoolSequenced(64, func() []int {
				return make([]int, 0, i)
			})

			segmentsPools[i-1] = pool
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

//var p = make([]int, 0, 128)

func acquireSegments(c int) []int {
	//	return p
	//	fmt.Printf("a%d;", getIdx(c))
	return segmentsPools[getIdx(c)].Get()
}

func releaseSegments(s []int) {
	//	p = s
	//		fmt.Printf("r%d;", getIdx(cap(s)))
	segmentsPools[getIdx(cap(s))].Put(s)
}

type newSegmentsFunc func() []int

// Pool holds Clients.
type PoolSequenced struct {
	new  newSegmentsFunc
	pool chan []int
}

// NewPool creates a new pool of Clients.
func NewPoolSequenced(size int, f newSegmentsFunc) *PoolSequenced {
	return &PoolSequenced{
		new:  f,
		pool: make(chan []int, size),
	}
}

// Borrow a Client from the pool.
func (p *PoolSequenced) Get() []int {
	var s []int
	select {
	case s = <-p.pool:
	default:
		s = p.new()
	}

	return s[:0]
}

// Return returns a Client to the pool.
func (p *PoolSequenced) Put(s []int) {
	select {
	case p.pool <- s:
	default:
		// let it go, let it go...
	}
}

type PoolSynced struct {
	size int
	mu   sync.Mutex
	list [][]int
}

func NewPoolSynced(size int) *PoolSynced {
	return &PoolSynced{
		size: size,
	}
}

func (p *PoolSynced) Get() []int {
	var s []int

	p.mu.Lock()
	ll := len(p.list)
	if ll > 0 {
		s, p.list = p.list[ll-1], p.list[:ll-1]
	}
	p.mu.Unlock()

	if s == nil {
		return make([]int, 0, p.size)
	}

	return s[:0]
}

func (p *PoolSynced) Put(s []int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.list = append(p.list, s)
}

type PoolNative struct {
	size int
	pool sync.Pool
}

func NewPoolNative(size int) *PoolNative {
	return &PoolNative{
		size: size,
	}
}

func (p *PoolNative) Get() []int {
	s := p.pool.Get()
	if s == nil {
		return make([]int, 0, p.size)
	}

	return s.([]int)
}

func (p *PoolNative) Put(s []int) {
	p.pool.Put(s)
}

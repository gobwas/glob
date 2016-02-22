package match

import (
	"sync"
	"sync/atomic"
)

var segmentsPools [1024]*PoolNative

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
	minSegment         = 4
	minSegmentMinusOne = 3
	maxSegment         = 1024
	maxSegmentMinusOne = 1023
)

func init() {
	for i := maxSegment; i >= minSegment; i >>= 1 {
		func(i int) {
			segmentsPools[i-1] = NewPoolNative(func() []int {
				return make([]int, 0, i)
			})
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
	return segmentsPools[getIdx(c)].Get()
}

func releaseSegments(s []int) {
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
	pool *sync.Pool
}

func NewPoolNative(f newSegmentsFunc) *PoolNative {
	return &PoolNative{
		pool: &sync.Pool{New: func() interface{} {
			return f()
		}},
	}
}

func (p *PoolNative) Get() []int {
	return p.pool.Get().([]int)[:0]
}

func (p *PoolNative) Put(s []int) {
	p.pool.Put(s)
}

type segments struct {
	data   []int
	locked int32
}

type PoolStatic struct {
	f    newSegmentsFunc
	pool []*segments
}

func NewPoolStatic(size int, f newSegmentsFunc) *PoolStatic {
	p := &PoolStatic{
		f:    f,
		pool: make([]*segments, 0, size),
	}

	for i := 0; i < size; i++ {
		p.pool = append(p.pool, &segments{
			data: f(),
		})
	}

	return p
}

func (p *PoolStatic) Get() (int, []int) {
	for i, s := range p.pool {
		if atomic.CompareAndSwapInt32(&s.locked, 0, 1) {
			return i, s.data
		}
	}

	return -1, p.f()
}

func (p *PoolStatic) Put(i int, s []int) {
	if i < 0 {
		return
	}

	p.pool[i].data = s
	atomic.CompareAndSwapInt32(&(p.pool[i].locked), 1, 0)
}

package match

import "sync"

// Pool holds Clients.
type PoolSequenced struct {
	size int
	pool chan []int
}

// NewPool creates a new pool of Clients.
func NewPoolSequenced(max, size int) *PoolSequenced {
	return &PoolSequenced{
		size: size,
		pool: make(chan []int, max),
	}
}

// Borrow a Client from the pool.
func (p *PoolSequenced) Get() []int {
	var s []int
	select {
	case s = <-p.pool:
	default:
		s = make([]int, 0, p.size)
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

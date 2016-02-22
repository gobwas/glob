package match

import (
	"testing"
)

func BenchmarkPerfPoolSequenced(b *testing.B) {
	pool := NewPoolSequenced(512, func() []int {
		return make([]int, 0, 16)
	})

	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get()
			pool.Put(s)
		}
	})
}

func BenchmarkPerfPoolSynced(b *testing.B) {
	pool := NewPoolSynced(32)

	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get()
			pool.Put(s)
		}
	})
}

func BenchmarkPerfPoolNative(b *testing.B) {
	pool := NewPoolNative(func() []int {
		return make([]int, 0, 16)
	})

	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get()
			pool.Put(s)
		}
	})
}

func BenchmarkPerfPoolStatic(b *testing.B) {
	pool := NewPoolStatic(32, func() []int {
		return make([]int, 0, 16)
	})

	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i, v := pool.Get()
			pool.Put(i, v)
		}
	})
}

func BenchmarkPerfMake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]int, 0, 32)
	}
}

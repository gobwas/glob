package match

import (
	"testing"
)

func BenchmarkPerfPoolSequenced(b *testing.B) {
	pool := NewPoolSequenced(32, 32)

	for i := 0; i < b.N; i++ {
		s := pool.Get()
		pool.Put(s)
	}
}

func BenchmarkPerfPoolSynced(b *testing.B) {
	pool := NewPoolSynced(32)

	for i := 0; i < b.N; i++ {
		s := pool.Get()
		pool.Put(s)
	}
}
func BenchmarkPerfPoolPoolNative(b *testing.B) {
	pool := NewPoolNative(32)

	for i := 0; i < b.N; i++ {
		s := pool.Get()
		pool.Put(s)
	}
}

func BenchmarkPerfMake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]int, 0, 32)
	}
}

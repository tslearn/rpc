package main

import (
	"testing"
)

func BenchmarkDebug(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		buf1 := make([]byte, 8192)
		buf2 := make([]byte, 8192)
		for pb.Next() {
			copy(buf1, buf2)
		}
	})
}

package router

import (
	"testing"
	"time"
)

func BenchmarkRouter_AddSlot(b *testing.B) {
	ch := make(chan bool)

	for i := 0; i < b.N; i++ {
		select {
		case <-ch:
		case <-time.After(10 * time.Microsecond):
		}
	}
}

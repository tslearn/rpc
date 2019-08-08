package core

import (
	"testing"
	"time"
)

func BenchmarkIndicator_count(b *testing.B) {
	in := newIndicator()
	in.setRecordOrigin(true)

	go func() {
		time.Sleep(500 * time.Millisecond)
		in.setRecordOrigin(false)
	}()

	for i := 0; i < b.N; i++ {
		in.count(time.Second, "$.nameService.sayHello", true)
	}
}

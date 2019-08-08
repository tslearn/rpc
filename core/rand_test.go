package core

import (
	"math/rand"
	"sync/atomic"
	"testing"
)

func TestGetRandUint32(t *testing.T) {
	assert := NewAssert(t)
	sum := int64(0)
	for i := 0; i < 100000; i++ {
		sum += int64(GetRandUint32())
	}
	delta := sum/100000 - 2147483648
	assert(delta > -20000000 && delta < 20000000).IsTrue()
}

func TestGetRandString(t *testing.T) {
	assert := NewAssert(t)
	assert(GetRandString(-1)).Equals("")
	for i := 0; i < 100; i++ {
		assert(len(GetRandString(i))).Equals(i)
	}
}

func BenchmarkGetRandUint32(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		GetRandString(1000)
	}
}

func BenchmarkGoroutineFixedRandom(b *testing.B) {
	n := uint32(10000)
	array := make([]uint64, n, n)

	//sum := 0

	//b.SetParallelism(100)

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			for m := 0; m < 100; m++ {
				pos := r.Uint32() % n
				// array[pos] += 1
				//  atomic.LoadInt32(&array[pos])
				atomic.AddUint64(&array[pos], 1)
			}
		}
	})

	//fmt.Println()
	//fmt.Println(array)
	//

}

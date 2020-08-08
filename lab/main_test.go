package lab

import (
	"fmt"
	"testing"
)

func TestDebug(t *testing.T) {
	buf := []int{1, 2, 3, 4, 5}

	buf1 := buf

	fmt.Println(buf, buf1)

	buf1[0] = 9

	fmt.Println(buf, buf1)
}

//
//func BenchmarkParallelDebug01(b *testing.B) {
//
//	arr := []int(nil)
//	arr = append(arr, 12)
//
//	fmt.Println(arr)
//
//	mp := &sync.Map{}
//
//	for i := 0; i < 1000000; i++ {
//		mp.Store(i, i)
//	}
//
//	sum := 0
//	b.ResetTimer()
//	for n := 0; n < b.N; n++ {
//		//mp.Store(10, 10)
//		if v, ok := mp.Load(10); ok {
//			sum += v.(int)
//		}
//	}
//
//	//b.ReportAllocs()
//	//b.RunParallel(func(pb *testing.PB) {
//	//  for pb.Next() {
//	//    mp.Store(10, 10)
//	//    if v, ok := mp.Load(10); ok {
//	//      sum += v.(int)
//	//    }
//	//  }
//	//})
//
//	fmt.Println(sum)
//}
//
//func BenchmarkParallelDebug02(b *testing.B) {
//	mp := make(map[int]int)
//
//	sum := 0
//	for n := 0; n < b.N; n++ {
//		mp[10] = 10
//		sum += mp[10]
//	}
//
//	//
//	//for i := 0; i < 100; i++ {
//	//  mp[i] = i
//	//}
//	//
//	//sum := 0
//	//b.ReportAllocs()
//	//b.RunParallel(func(pb *testing.PB) {
//	//  for pb.Next() {
//	//
//	//  }
//	//})
//
//	fmt.Println(sum)
//}
//
//func BenchmarkParallelDebug03(b *testing.B) {
//	mp := make(map[int]int)
//
//	mu := internal.NewLock()
//	sum := 0
//	b.ReportAllocs()
//	b.SetParallelism(1023)
//
//	b.RunParallel(func(pb *testing.PB) {
//		for pb.Next() {
//			mu.DoWithLock(func() {
//				mp[10] = 10
//				sum += mp[10]
//			})
//		}
//	})
//
//	//
//	//for i := 0; i < 100; i++ {
//	//  mp[i] = i
//	//}
//	//
//	//sum := 0
//	//b.ReportAllocs()
//	//b.RunParallel(func(pb *testing.PB) {
//	//  for pb.Next() {
//	//
//	//  }
//	//})
//
//	fmt.Println(sum)
//}

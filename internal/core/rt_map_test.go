package core

import (
	"fmt"
	"testing"
)

func TestRTMap_swapUint32(t *testing.T) {
	rtMap := streamTestRuntime.NewRTMap()

	rtMap.appendValue("v1", 34234)
	rtMap.appendValue("v8", 34234)
	rtMap.appendValue("v5", 34234)
	rtMap.appendValue("v2", 34234)
	rtMap.appendValue("v3", 34234)
	rtMap.appendValue("v7", 34234)
	rtMap.appendValue("v4", 34234)
	rtMap.appendValue("v6", 34234)
	fmt.Printf("0x%08x\n", getSort8(rtMap.items))
}

func BenchmarkMakeRequestStream(b *testing.B) {
	rtMap := streamTestRuntime.NewRTMap()
	rtMap.appendValue("v1", 34234)
	rtMap.appendValue("v8", 34234)
	rtMap.appendValue("v5", 34234)
	rtMap.appendValue("v2", 34234)
	rtMap.appendValue("v3", 34234)
	rtMap.appendValue("v7", 34234)
	rtMap.appendValue("v4", 34234)
	rtMap.appendValue("v6", 34234)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		getSort8(rtMap.items)
	}
}

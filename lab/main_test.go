package main

import (
	"runtime"
	"runtime/debug"
	"testing"
)

type memoryInner struct {
	a uint64
	b uint64
}

type memoryObject struct {
	a     int64
	items [64]*memoryInner
}

func newMemoryObject() *memoryObject {
	ret := &memoryObject{
		a: 10,
	}

	for i := 0; i < 64; i++ {
		ret.items[i] = &memoryInner{a: 5, b: 6}
	}

	return ret
}

func TestDebug(t *testing.T) {
	extra := make([]*memoryObject, 0)
	for i := 0; i < 10000; i++ {
		extra = append(extra, newMemoryObject())
	}

	colls := make([]*memoryObject, 0)

	for i := 0; i < 100; i++ {
		for j := 0; j < 10000; j++ {
			colls = append(colls, newMemoryObject())
		}
		colls = make([]*memoryObject, 0)
		debug.SetGCPercent(1)
		runtime.GC()
	}
}

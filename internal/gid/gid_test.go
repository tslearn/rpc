// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package gid

import (
	"sync"
	"testing"
)

func TestG(t *testing.T) {
	gp1 := getg()

	if gp1 == nil {
		t.Fatalf("fail to get G.")
	}

	t.Run("G in another goroutine", func(t *testing.T) {
		gp2 := getg()

		if gp2 == nil {
			t.Fatalf("fail to get G.")
		}

		if gp2 == gp1 {
			t.Fatalf("every living G must be different. [gp1:%p] [gp2:%p]", gp1, gp2)
		}
	})
}

func TestGetGidPos(t *testing.T) {
	if getGidPos() < 0 {
		t.Fatalf("getGidPos error")
	}

	// make fake error
	temp := goroutinePrefix
	defer func() {
		goroutinePrefix = temp
	}()
	goroutinePrefix = "fake "
	if getGidPos() != -1 {
		t.Fatalf("getGidPos error")
	}
}

func TestGidFast(t *testing.T) {
	idMap := make(map[int64]bool)
	waitCH := make(chan bool)
	testCount := 100000
	mu := &sync.Mutex{}
	for i := 0; i < testCount; i++ {
		go func() {
			id := Gid()
			if id <= 0 {
				t.Fatalf("Gid error")
			}
			mu.Lock()
			defer mu.Unlock()
			idMap[id] = true
			waitCH <- true
		}()
	}
	for i := 0; i < testCount; i++ {
		<-waitCH
	}
	if len(idMap) != testCount {
		t.Fatalf("Gid error")
	}
}

func TestGidSlow(t *testing.T) {
	temp := gidPos
	defer func() {
		gidPos = temp
	}()
	gidPos = -1

	idMap := make(map[int64]bool)
	waitCH := make(chan bool)
	testCount := 100000
	mu := &sync.Mutex{}
	for i := 0; i < testCount; i++ {
		go func() {
			id := Gid()
			if id <= 0 {
				t.Fatalf("Gid error")
			}
			mu.Lock()
			defer mu.Unlock()
			idMap[id] = true
			waitCH <- true
		}()
	}
	for i := 0; i < testCount; i++ {
		<-waitCH
	}
	if len(idMap) != testCount {
		t.Fatalf("Gid error")
	}
}

func BenchmarkG(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Gid()
	}
}

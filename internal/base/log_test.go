package base

import (
	"bytes"
	"io"
	"os"
	"sync/atomic"
	"testing"
)

func captureStdout(fn func()) string {
	oldStdout := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	func() {
		defer func() {
			_ = recover()
		}()
		fn()
	}()

	outCH := make(chan string)
	// copy the output in a separate goroutine so print can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCH <- buf.String()
	}()

	os.Stdout = oldStdout
	_ = w.Close()
	ret := <-outCH
	_ = r.Close()
	return ret
}

func TestLog(t *testing.T) {
	t.Run("helper is nil", func(t *testing.T) {
		assert := NewAssert(t)
		atomic.StorePointer(&logWriter, nil)
		defer SetLogWriter(os.Stdout)
		assert(captureStdout(func() {
			Log("message")
		})).Equal("message")
	})

	t.Run("helper is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(captureStdout(func() {
			// Stdout is fake
			SetLogWriter(os.Stdout)
			Log("message")
		})).Equal("message")
		// restore real Stdout
		SetLogWriter(os.Stdout)
	})
}

func TestSetLogWriter(t *testing.T) {
	t.Run("writer is nil", func(t *testing.T) {
		assert := NewAssert(t)
		curHelper := (*logHelper)(atomic.LoadPointer(&logWriter))
		SetLogWriter(nil)
		defer SetLogWriter(os.Stdout)
		assert((*logHelper)(atomic.LoadPointer(&logWriter))).Equal(curHelper)
	})

	t.Run("writer is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		r, w, _ := os.Pipe()
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		SetLogWriter(w)
		defer SetLogWriter(os.Stdout)
		assert((*logHelper)(atomic.LoadPointer(&logWriter)).writer).Equal(w)
	})
}

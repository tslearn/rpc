package base

import (
	"bytes"
	"io"
	"os"
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

func TestNewLogger(t *testing.T) {
	t.Run("open file error", func(t *testing.T) {
		assert := NewAssert(t)
		v, e := NewLogger(false, "/")
		assert(v).IsNil()
		assert(e.GetCode()).Equal(ErrLogOpenFile.GetCode())
	})

	t.Run("file is empty", func(t *testing.T) {
		assert := NewAssert(t)
		assert(NewLogger(false, "")).Equal(&Logger{
			isLogToScreen: false,
			file:          nil,
		}, nil)
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v, err := NewLogger(false, "test.log")
		assert(err).IsNil()
		assert(v.isLogToScreen).IsFalse()
		assert(v.file).IsNotNil()
		v.Close()
		_ = os.Remove("test.log")
	})
}

func TestLogger_Log(t *testing.T) {
	getFileContent := func(fileName string) string {
		buffer, _ := os.ReadFile(fileName)
		return string(buffer)
	}

	t.Run("isLogToScreen == true, file is empty", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(true, "")

		assert(captureStdout(func() {
			assert(v.Log("hello")).IsNil()
		})).Equal("hello")

		v.Close()
	})

	t.Run("isLogToScreen == false, file is not", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(false, "test.log")

		assert(captureStdout(func() {
			assert(v.Log("hello")).IsNil()
		})).Equal("")

		assert(getFileContent("test.log")).Equal("hello")

		v.Close()
		_ = os.Remove("test.log")
	})

	t.Run("stdout WriteString error", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")

		oldStdout := os.Stdout
		os.Stdout = v.file
		defer func() {
			os.Stdout = oldStdout
		}()

		v.file.Close()

		assert(v.Log("hello").GetCode()).Equal(ErrLogWriteFile.GetCode())

		v.Close()
		_ = os.Remove("test.log")
	})

	t.Run("file WriteString error", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")
		v.file.Close()

		assert(v.Log("hello").GetCode()).Equal(ErrLogWriteFile.GetCode())

		v.Close()
		_ = os.Remove("test.log")
	})
}

func TestLogger_Close(t *testing.T) {
	t.Run("ok, file is nil", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(true, "")
		assert(v.Close()).IsNil()
	})

	t.Run("ok, file is not nil", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")
		assert(v.Close()).IsNil()
		_ = os.Remove("test.log")
	})

	t.Run("file close error", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")
		v.file.Close()
		assert(v.Close().GetCode()).Equal(ErrLogCloseFile.GetCode())
		_ = os.Remove("test.log")
	})
}

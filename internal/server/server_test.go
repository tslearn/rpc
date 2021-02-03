package server

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

type testActionCache struct{}

func (p *testActionCache) Get(_ string) core.ActionCacheFunc {
	return nil
}

func TestServerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(defaultMaxNumOfThreads).Equal(1024 * 1024)
		assert(defaultThreadsPerCPU).Equal(16384)
		assert(defaultThreadBufferSize).Equal(2048)
		assert(defaultCloseTimeout).Equal(5 * time.Second)
		assert(defaultMaxNodeDepth).Equal(128)
		assert(defaultMaxCallDepth).Equal(128)
		assert(fnNumCPU()).Equal(runtime.NumCPU())

		stream := core.NewStream()
		stream.PutBytes([]byte{1, 2, 3, 4, 5, 6})
		assert(stream.GetWritePos() == core.StreamHeadSize).IsFalse()
		onReturnStream(stream)
		assert(stream.GetWritePos() == core.StreamHeadSize).IsTrue()
	})
}

func TestNewServer(t *testing.T) {
	t.Run("numOfThreads <= defaultMaxNumOfThreads", func(t *testing.T) {
		fnNumCPU = func() int {
			return 4
		}
		defer func() {
			fnNumCPU = runtime.NumCPU
		}()

		assert := base.NewAssert(t)
		v := NewServer()
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
		assert(v.router).IsNotNil()
		assert(v.gateway).IsNotNil()
		assert(v.numOfThreads).Equal(4 * 16384)
		assert(v.maxNodeDepth).Equal(int16(128))
		assert(v.maxCallDepth).Equal(int16(128))
		assert(v.threadBufferSize).Equal(uint32(2048))
		assert(v.actionCache).IsNil()
		assert(v.closeTimeout).Equal(5 * time.Second)
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
		assert(v.errorHandler).IsNil()
	})

	t.Run("numOfThreads > defaultMaxNumOfThreads", func(t *testing.T) {
		fnNumCPU = func() int {
			return 256
		}
		defer func() {
			fnNumCPU = runtime.NumCPU
		}()

		assert := base.NewAssert(t)
		v := NewServer()
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
		assert(v.router).IsNotNil()
		assert(v.gateway).IsNotNil()
		assert(v.numOfThreads).Equal(1024 * 1024)
		assert(v.maxNodeDepth).Equal(int16(128))
		assert(v.maxCallDepth).Equal(int16(128))
		assert(v.threadBufferSize).Equal(uint32(2048))
		assert(v.actionCache).IsNil()
		assert(v.closeTimeout).Equal(5 * time.Second)
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
		assert(v.errorHandler).IsNil()
	})
}

func TestServer_onError(t *testing.T) {
	t.Run("errorHandler == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		assert(base.RunWithCatchPanic(func() {
			v.onError(12, base.ErrStream)
		})).IsNil()
	})

	t.Run("errorHandler != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		v.onError(12, base.ErrStream)
		assert(errID).Equal(uint64(12))
		assert(err).Equal(base.ErrStream)
	})
}

func TestServer_Listen(t *testing.T) {
	t.Run("gateway is opened", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}

		v.Listen("tcp", "127.0.0.1:1234", nil)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(nil)
	})
}

func TestServer_SetNumOfThreads(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()

		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}

		v.isRunning = true
		_, source := v.SetNumOfThreads(1024), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
	})

	t.Run("numOfThreads == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()

		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}

		_, source := v.SetNumOfThreads(0), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrNumOfThreadsIsWrong.AddDebug(source))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.SetNumOfThreads(1024)
		assert(v.numOfThreads).Equal(1024)
	})
}

func TestServer_SetThreadBufferSize(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()

		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}

		v.isRunning = true
		_, source := v.SetThreadBufferSize(1024), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
	})

	t.Run("threadBufferSize == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()

		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}

		_, source := v.SetThreadBufferSize(0), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrThreadBufferSizeIsWrong.AddDebug(source))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.SetThreadBufferSize(1023)
		assert(v.threadBufferSize).Equal(uint32(1023))
	})
}

func TestServer_SetActionCache(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		ac := &testActionCache{}
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		v.isRunning = true
		_, source := v.SetActionCache(ac), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
		assert(v.actionCache).IsNil()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		ac := &testActionCache{}
		v := NewServer()
		v.SetActionCache(ac)
		assert(v.actionCache).Equal(ac)
	})
}

func TestServer_SetErrorHandler(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		handler := func(sessionID uint64, e *base.Error) {}
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		v.isRunning = true
		_, source := v.SetErrorHandler(handler), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		handler := func(sessionID uint64, e *base.Error) {}
		v := NewServer()
		v.SetErrorHandler(handler)
		assert(v.errorHandler).IsNotNil()
	})
}

func TestServer_AddService(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := core.NewService()
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		v.isRunning = true
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := core.NewService()
		v := NewServer()
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(v.mountServices[0]).Equal(core.NewServiceMeta(
			"t",
			service,
			source,
			nil,
		))
	})
}

func TestServer_BuildReplyCache(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("test ok", func(t *testing.T) {
		defer func() {
			_ = os.RemoveAll(path.Join(curDir, "cache"))
		}()
		assert := base.NewAssert(t)
		v := NewServer()
		assert(v.BuildReplyCache()).Equal(v)
	})

	t.Run("output file exists", func(t *testing.T) {
		defer func() {
			_ = os.RemoveAll(path.Join(curDir, "cache"))
		}()

		_ = os.MkdirAll(path.Join(curDir, "cache"), 0555)
		_ = ioutil.WriteFile(
			path.Join(curDir, "cache", "rpc_action_cache.go"),
			[]byte{1, 2, 3, 4},
			0555,
		)
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		assert(v.BuildReplyCache()).Equal(v)
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrCacheWriteFile)
	})
}

func TestServer_Open(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		v.isRunning = true
		isOpen, source := v.Open(), base.GetFileLine(0)
		assert(errID).Equal(uint64(0))
		assert(isOpen).Equal(false)
		assert(err).Equal(base.ErrServerAlreadyRunning.AddDebug(source))
	})

	t.Run("processor create error", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.numOfThreads = 0
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		assert(v.Open()).IsFalse()
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrNumOfThreadsIsWrong)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			isRunning := false
			for !isRunning {
				time.Sleep(10 * time.Millisecond)
				v.Lock()
				isRunning = v.isRunning
				v.Unlock()
			}
			time.Sleep(200 * time.Millisecond)
			v.Close()
		}()

		assert(v.Open()).IsTrue()
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("server is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		err := (*base.Error)(nil)
		errID := uint64(0)
		v := NewServer()
		v.errorHandler = func(sessionID uint64, e *base.Error) {
			errID = sessionID
			err = e
		}
		isSuccess, source := v.Close(), base.GetFileLine(0)
		assert(isSuccess).IsFalse()
		assert(errID).Equal(uint64(0))
		assert(err).Equal(base.ErrServerNotRunning.AddDebug(source))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			v.Open()
		}()

		isRunning := false
		for !isRunning {
			time.Sleep(10 * time.Millisecond)
			v.Lock()
			isRunning = v.isRunning
			v.Unlock()
		}
		time.Sleep(200 * time.Millisecond)
		assert(v.Close()).IsTrue()
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
	})
}

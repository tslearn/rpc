package server

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
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
		assert(v.logHub).IsNotNil()
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
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
		assert(v.logHub).IsNotNil()
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
	})
}

func TestServer_Listen(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.Listen("tcp", "127.0.0.1:1234", nil)
		assert(errorHub.GetStream()).IsNil()
	})
}

func TestServer_ListenWithDebug(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.ListenWithDebug("tcp", "127.0.0.1:1234", nil)
		assert(errorHub.GetStream()).IsNil()
	})
}
func TestServer_SetNumOfThreads(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		_, source := v.SetNumOfThreads(1024), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("numOfThreads == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		_, source := v.SetNumOfThreads(0), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrNumOfThreadsIsWrong.AddDebug(source).Standardize(),
		)
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
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		_, source := v.SetThreadBufferSize(1024), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("threadBufferSize == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		_, source := v.SetThreadBufferSize(0), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrThreadBufferSizeIsWrong.AddDebug(source).Standardize(),
		)
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
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		_, source := v.SetActionCache(&testActionCache{}), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
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

func TestServer_SetLogHub(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		setHub := core.NewTestStreamHub()
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		_, source := v.SetLogHub(setHub), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.SetLogHub(core.NewTestStreamHub())
		assert(v.logHub).IsNotNil()
	})
}

func TestServer_AddService(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := core.NewService()
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
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
		_ = os.MkdirAll(path.Join(curDir, "cache", "rpc_action_cache.go"), 0555)
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		assert(v.BuildReplyCache()).Equal(v)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrCacheWriteFile.Standardize(),
		)
	})
}

func TestServer_OnReceiveStream(t *testing.T) {
	t.Run("DataStreamInternalRequest", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer().
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		errorHub := core.NewTestStreamHub()
		v.logHub = errorHub
		go func() {
			v.Open()
		}()

		stream := core.NewStream()
		stream.SetKind(core.DataStreamInternalRequest)
		stream.SetDepth(0)
		stream.WriteString("#.test.Eval")
		stream.WriteString("@")

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}
		defer v.Close()

		v.OnReceiveStream(stream)

		assert(core.ParseResponseStream(errorHub.WaitStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})

	t.Run("DataStreamExternalRequest", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer().
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		errorHub := core.NewTestStreamHub()
		v.logHub = errorHub
		go func() {
			v.Open()
		}()

		stream := core.NewStream()
		stream.SetKind(core.DataStreamExternalRequest)
		stream.SetDepth(0)
		stream.WriteString("#.test.Eval")
		stream.WriteString("@")

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}
		defer v.Close()

		v.OnReceiveStream(stream)

		assert(core.ParseResponseStream(errorHub.WaitStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})

	t.Run("DataStreamResponseOK", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer().
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		errorHub := core.NewTestStreamHub()
		v.logHub = errorHub
		go func() {
			v.Open()
		}()

		stream := core.NewStream()
		stream.SetKind(core.DataStreamResponseOK)
		stream.Write(true)

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}
		defer v.Close()
		v.OnReceiveStream(stream)
		assert(core.ParseResponseStream(errorHub.WaitStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})

	t.Run("DataStreamResponseError", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer().
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		errorHub := core.NewTestStreamHub()
		v.logHub = errorHub
		go func() {
			v.Open()
		}()

		stream := core.NewStream()
		stream.SetKind(core.DataStreamResponseError)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}
		defer v.Close()
		v.OnReceiveStream(stream)
		assert(core.ParseResponseStream(errorHub.WaitStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})

	t.Run("DataStreamBoardCast", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer().
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		errorHub := core.NewTestStreamHub()
		v.logHub = errorHub
		go func() {
			v.Open()
		}()

		stream := core.NewStream()
		stream.SetKind(core.DataStreamBoardCast)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}
		defer v.Close()
		v.OnReceiveStream(stream)
		assert(core.ParseResponseStream(errorHub.WaitStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})

	t.Run("SystemStreamReportError log to screen", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()

		stream := core.NewStream()
		stream.SetKind(core.SystemStreamReportError)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())
		v.OnReceiveStream(stream)
		assert(stream.GetKind() != core.SystemStreamReportError).IsTrue()
	})

	t.Run("ControlStreamConnectResponse", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()

		stream := core.NewStream()
		stream.SetKind(core.ControlStreamConnectResponse)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())
		v.OnReceiveStream(stream)
		assert(stream.GetKind() != core.ControlStreamConnectResponse).IsTrue()
	})

}

func TestServer_Open(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		v.isRunning = true
		isOpen, source := v.Open(), base.GetFileLine(0)
		assert(isOpen).Equal(false)
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("processor create error", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.numOfThreads = 0
		v.logHub = errorHub
		assert(v.Open()).IsFalse()
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrNumOfThreadsIsWrong.Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			for !v.IsRunning() {
				time.Sleep(10 * time.Millisecond)
			}

			time.Sleep(200 * time.Millisecond)
			v.Close()
		}()

		assert(v.Open()).IsTrue()
	})
}

func TestServer_IsRunning(t *testing.T) {
	t.Run("not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		assert(v.isRunning).IsFalse()
	})

	t.Run("running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.isRunning = true
		assert(v.isRunning).IsTrue()
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("server is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		errorHub := core.NewTestStreamHub()
		v := NewServer()
		v.logHub = errorHub
		isSuccess, source := v.Close(), base.GetFileLine(0)
		assert(isSuccess).IsFalse()
		assert(core.ParseResponseStream(errorHub.GetStream())).Equal(
			nil, base.ErrServerNotRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer()
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			v.Open()
		}()

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)
		assert(v.Close()).IsTrue()
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
	})
}

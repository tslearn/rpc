package util

import (
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func runStdWriterLogger(onRun func(logger *Logger)) string {
	if r, w, err := os.Pipe(); err == nil {
		old := os.Stdout // keep backup of the real stdout
		os.Stdout = w
		defer func() {
			os.Stdout = old
		}()

		onRun(NewLogger(NewStdLogWriter()))

		retCH := make(chan string)

		go func() {
			buf := make([]byte, 10240)
			pos := 0
			for {
				if n, err := r.Read(buf[pos:]); err == io.EOF {
					break
				} else {
					pos += n
				}
			}
			retCH <- string(buf[:pos])
		}()
		_ = w.Close()

		return <-retCH
	}
	return "<error>"
}

func runCallbackWriterLogger(
	onRun func(logger *Logger),
) (isoTime string, tag string, msg string, extra string) {
	wait := make(chan bool, 1)
	onRun(NewLogger(NewCallbackLogWriter(
		func(_isoTime string, _tag string, _msg string, _extra string) {
			isoTime = _isoTime
			tag = _tag
			msg = _msg
			extra = _extra
			wait <- true
		},
	)))
	<-wait
	return
}

func TestRpcStdLogWriter_Write(t *testing.T) {
	assert := NewAssert(t)

	assert(strings.HasSuffix(
		runStdWriterLogger(func(logger *Logger) {
			logger.Info("message")
		}),
		" Info: message\n",
	)).IsTrue()

	assert(strings.HasSuffix(
		runStdWriterLogger(func(logger *Logger) {
			logger.InfoExtra("message", "extra")
		}),
		"(extra) Info: message\n",
	)).IsTrue()
}

func TestNewLogger(t *testing.T) {
	assert := NewAssert(t)
	logger1 := NewLogger(nil)
	assert(logger1.level).Equals(LogMaskAll)
	assert(logger1.writer).IsNotNil()

	logger2 := NewLogger(NewStdLogWriter())
	assert(logger2.level).Equals(LogMaskAll)
	assert(logger2.writer).IsNotNil()
}

func TestRpcLogger_SetLevel(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger(nil)

	assert(logger.SetLevel(LogMaskNone - 1)).IsFalse()
	assert(logger.level).Equals(LogMaskAll)

	assert(logger.SetLevel(LogMaskAll + 1)).IsFalse()
	assert(logger.level).Equals(LogMaskAll)

	assert(logger.SetLevel(LogMaskNone)).IsTrue()
	assert(logger.level).Equals(LogMaskNone)

	assert(logger.SetLevel(LogMaskAll)).IsTrue()
	assert(logger.level).Equals(LogMaskAll)

	// test all level and logs
	fnTestLogLevel := func(level int32) int32 {
		ret := int32(0)

		logger := NewLogger(NewCallbackLogWriter(
			func(_ string, tag string, msg string, _ string) {
				if msg == "message" {
					switch tag {
					case "Debug":
						atomic.AddInt32(&ret, LogMaskDebug)
					case "Info":
						atomic.AddInt32(&ret, LogMaskInfo)
					case "Warn":
						atomic.AddInt32(&ret, LogMaskWarn)
					case "Error":
						atomic.AddInt32(&ret, LogMaskError)
					case "Fatal":
						atomic.AddInt32(&ret, LogMaskFatal)
					}
				}
			},
		))
		logger.SetLevel(level)
		logger.Debug("message")
		logger.Info("message")
		logger.Warn("message")
		logger.Error("message")
		logger.Fatal("message")
		time.Sleep(30 * time.Millisecond)
		return atomic.LoadInt32(&ret)
	}

	assert(fnTestLogLevel(LogMaskNone - 1)).Equals(LogMaskAll)
	for i := int32(0); i < 32; i++ {
		assert(fnTestLogLevel(i)).Equals(i)
	}
	assert(fnTestLogLevel(LogMaskAll + 1)).Equals(LogMaskAll)
}

func TestRpcLogger_Debug(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Debug("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Debug")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestRpcLogger_DebugExtra(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.DebugExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Debug")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestRpcLogger_Info(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Info("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Info")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestRpcLogger_InfoExtra(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.InfoExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Info")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestRpcLogger_Warn(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Warn("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Warn")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestRpcLogger_WarnExtra(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.WarnExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Warn")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestRpcLogger_Error(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Error("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Error")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestRpcLogger_ErrorExtra(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.ErrorExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Error")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestRpcLogger_Fatal(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Fatal("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Fatal")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestRpcLogger_FatalExtra(t *testing.T) {
	assert := NewAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.FatalExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Fatal")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

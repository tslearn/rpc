package util

import (
	"fmt"
	"sync/atomic"
)

const (
	// LogMaskNone this level logs nothing
	LogMaskNone = int32(0)
	// LogMaskFatal this level logs Fatal
	LogMaskFatal = int32(1)
	// LogMaskError this level logs Error
	LogMaskError = int32(2)
	// LogMaskWarn this level logs Warn
	LogMaskWarn = int32(4)
	// LogMaskInfo this level logs Info
	LogMaskInfo = int32(8)
	// LogMaskDebug this level logs Debug
	LogMaskDebug = int32(16)
	// LogMaskAll this level logs Debug, Info, Warn, Error and Fatal
	LogMaskAll = LogMaskFatal |
		LogMaskError |
		LogMaskWarn |
		LogMaskInfo |
		LogMaskDebug
)

// LogWriter ...
type LogWriter interface {
	Write(isoTime string, tag string, msg string, extra string)
}

// StdLogWriter ...
type StdLogWriter = *rpcStdLogWriter
type rpcStdLogWriter struct{}

// NewStdLogWriter ...
func NewStdLogWriter() StdLogWriter {
	return &rpcStdLogWriter{}
}

func (p *rpcStdLogWriter) Write(
	isoTime string,
	tag string,
	msg string,
	extra string,
) {
	sb := NewStringBuilder()
	sb.AppendString(isoTime)
	if len(extra) > 0 {
		sb.AppendByte('(')
		sb.AppendString(extra)
		sb.AppendByte(')')
	}
	sb.AppendByte(' ')
	sb.AppendString(tag)
	sb.AppendByte(':')
	sb.AppendByte(' ')
	sb.AppendString(msg)
	sb.AppendByte('\n')
	fmt.Print(sb.String())
	sb.Release()
}

// CallbackLogWriter ...
type CallbackLogWriter = *rpcCallbackStdLogWriter
type rpcCallbackStdLogWriter struct {
	onWrite func(isoTime string, tag string, msg string, extra string)
}

// NewCallbackLogWriter ...
func NewCallbackLogWriter(
	onWrite func(isoTime string, tag string, msg string, extra string),
) CallbackLogWriter {
	return &rpcCallbackStdLogWriter{onWrite: onWrite}
}

// Write ...
func (p *rpcCallbackStdLogWriter) Write(
	isoTime string,
	tag string,
	msg string,
	extra string,
) {
	if p.onWrite != nil {
		p.onWrite(isoTime, tag, msg, extra)
	}
}

// Logger ...
type Logger = *rpcLogger
type rpcLogger struct {
	level   int32
	writers []LogWriter
	rpcAutoLock
}

// NewLogger ...
func NewLogger(writers []LogWriter) Logger {
	if writers == nil || len(writers) == 0 {
		return &rpcLogger{
			level:   LogMaskAll,
			writers: []LogWriter{NewStdLogWriter()},
		}
	}
	return &rpcLogger{
		level:   LogMaskAll,
		writers: writers,
	}
}

func (p *rpcLogger) SetLevel(level int32) bool {
	if level >= LogMaskNone && level <= LogMaskAll {
		atomic.StoreInt32(&p.level, level)
		return true
	}

	return false
}

func (p *rpcLogger) Debug(msg string) {
	p.DebugExtra(msg, "")
}

func (p *rpcLogger) DebugExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskDebug > 0 {
		isoTime := TimeNowISOString()
		for _, writer := range p.writers {
			writer.Write(isoTime, "Debug", msg, extra)
		}
	}
}

func (p *rpcLogger) Info(msg string) {
	p.InfoExtra(msg, "")
}

func (p *rpcLogger) InfoExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskInfo > 0 {
		isoTime := TimeNowISOString()
		for _, writer := range p.writers {
			writer.Write(isoTime, "Info", msg, extra)
		}
	}
}

func (p *rpcLogger) Warn(msg string) {
	p.WarnExtra(msg, "")
}

func (p *rpcLogger) WarnExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskWarn > 0 {
		isoTime := TimeNowISOString()
		for _, writer := range p.writers {
			writer.Write(isoTime, "Warn", msg, extra)
		}
	}
}

func (p *rpcLogger) Error(msg string) {
	p.ErrorExtra(msg, "")
}

func (p *rpcLogger) ErrorExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskError > 0 {
		isoTime := TimeNowISOString()
		for _, writer := range p.writers {
			writer.Write(isoTime, "Error", msg, extra)
		}
	}
}

func (p *rpcLogger) Fatal(msg string) {
	p.FatalExtra(msg, "")
}

func (p *rpcLogger) FatalExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskFatal > 0 {
		isoTime := TimeNowISOString()
		for _, writer := range p.writers {
			writer.Write(isoTime, "Fatal", msg, extra)
		}
	}
}

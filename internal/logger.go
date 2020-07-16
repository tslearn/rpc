package internal

import (
	"fmt"
	"sync/atomic"
)

const (
	// RPCLogTagDebug ...
	RPCLogTagDebug = "Debug"
	// RPCLogTagInfo ...
	RPCLogTagInfo = "Info"
	// RPCLogTagWarn ...
	RPCLogTagWarn = "Warn"
	// RPCLogTagError ...
	RPCLogTagError = "Error"
	// RPCLogTagFatal ...
	RPCLogTagFatal = "Fatal"
	// RPCLogMaskNone this level logs nothing
	RPCLogMaskNone = int32(0)
	// RPCLogMaskFatal this level logs Fatal
	RPCLogMaskFatal = int32(1 << 0)
	// RPCLogMaskError this level logs Error
	RPCLogMaskError = int32(1 << 1)
	// RPCLogMaskWarn this level logs Warn
	RPCLogMaskWarn = int32(1 << 2)
	// RPCLogMaskInfo this level logs Info
	RPCLogMaskInfo = int32(1 << 3)
	// RPCLogMaskDebug this level logs Debug
	RPCLogMaskDebug = int32(1 << 4)
	// RPCLogMaskAll this level logs Debug, Info, Warn, Error and Fatal
	RPCLogMaskAll = RPCLogMaskFatal |
		RPCLogMaskError |
		RPCLogMaskWarn |
		RPCLogMaskInfo |
		RPCLogMaskDebug
)

// RPCLogWriter ...
type RPCLogWriter interface {
	Write(isoTime string, tag string, msg string, extra string)
}

// RPCStdoutLogWriter ...
type RPCStdoutLogWriter struct{}

// NewRPCStdoutLogWriter ...
func NewRPCStdoutLogWriter() RPCLogWriter {
	return &RPCStdoutLogWriter{}
}

func (p *RPCStdoutLogWriter) Write(
	isoTime string,
	tag string,
	msg string,
	extra string,
) {
	sb := NewStringBuilder()
	defer sb.Release()
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
}

// RPCCallbackLogWriter ...
type RPCCallbackLogWriter struct {
	onWrite func(isoTime string, tag string, msg string, extra string)
}

// NewRPCCallbackLogWriter ...
func NewRPCCallbackLogWriter(
	onWrite func(isoTime string, tag string, msg string, extra string),
) *RPCCallbackLogWriter {
	return &RPCCallbackLogWriter{onWrite: onWrite}
}

// Write ...
func (p *RPCCallbackLogWriter) Write(
	isoTime string,
	tag string,
	msg string,
	extra string,
) {
	if p.onWrite != nil {
		p.onWrite(isoTime, tag, msg, extra)
	}
}

// RPCLogger ...
type RPCLogger struct {
	level  int32
	writer RPCLogWriter
	Lock
}

// NewRPCLogger ...
func NewRPCLogger(writer RPCLogWriter) *RPCLogger {
	if writer == nil {
		return &RPCLogger{
			level:  RPCLogMaskAll,
			writer: NewRPCStdoutLogWriter(),
		}
	}

	return &RPCLogger{
		level:  RPCLogMaskAll,
		writer: writer,
	}
}

// SetLevel ...
func (p *RPCLogger) SetLevel(level int32) bool {
	if level >= RPCLogMaskNone && level <= RPCLogMaskAll {
		atomic.StoreInt32(&p.level, level)
		return true
	}

	return false
}

// Debug ...
func (p *RPCLogger) Debug(msg string) {
	p.DebugExtra(msg, "")
}

// DebugExtra ...
func (p *RPCLogger) DebugExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&RPCLogMaskDebug > 0 {
		p.writer.Write(TimeNowISOString(), RPCLogTagDebug, msg, extra)
	}
}

// Info ...
func (p *RPCLogger) Info(msg string) {
	p.InfoExtra(msg, "")
}

// InfoExtra ...
func (p *RPCLogger) InfoExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&RPCLogMaskInfo > 0 {
		p.writer.Write(TimeNowISOString(), RPCLogTagInfo, msg, extra)
	}
}

// Warn ...
func (p *RPCLogger) Warn(msg string) {
	p.WarnExtra(msg, "")
}

// WarnExtra ...
func (p *RPCLogger) WarnExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&RPCLogMaskWarn > 0 {
		p.writer.Write(TimeNowISOString(), RPCLogTagWarn, msg, extra)
	}
}

// Error ...
func (p *RPCLogger) Error(msg string) {
	p.ErrorExtra(msg, "")
}

// ErrorExtra ...
func (p *RPCLogger) ErrorExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&RPCLogMaskError > 0 {
		p.writer.Write(TimeNowISOString(), RPCLogTagError, msg, extra)
	}
}

// Fatal ...
func (p *RPCLogger) Fatal(msg string) {
	p.FatalExtra(msg, "")
}

// FatalExtra ...
func (p *RPCLogger) FatalExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&RPCLogMaskFatal > 0 {
		p.writer.Write(TimeNowISOString(), RPCLogTagFatal, msg, extra)
	}
}

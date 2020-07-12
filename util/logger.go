package util

import (
	"fmt"
	"sync/atomic"
)

const (
	// LogTagDebug ...
	LogTagDebug = "Debug"
	// LogTagInfo ...
	LogTagInfo = "Info"
	// LogTagWarn ...
	LogTagWarn = "Warn"
	// LogTagError ...
	LogTagError = "Error"
	// LogTagFatal ...
	LogTagFatal = "Fatal"
	// LogMaskNone this level logs nothing
	LogMaskNone = int32(0)
	// LogMaskFatal this level logs Fatal
	LogMaskFatal = int32(1 << 0)
	// LogMaskError this level logs Error
	LogMaskError = int32(1 << 1)
	// LogMaskWarn this level logs Warn
	LogMaskWarn = int32(1 << 2)
	// LogMaskInfo this level logs Info
	LogMaskInfo = int32(1 << 3)
	// LogMaskDebug this level logs Debug
	LogMaskDebug = int32(1 << 4)
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
type StdLogWriter struct{}

// NewStdLogWriter ...
func NewStdLogWriter() LogWriter {
	return &StdLogWriter{}
}

func (p *StdLogWriter) Write(
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
type CallbackLogWriter struct {
	onWrite func(isoTime string, tag string, msg string, extra string)
}

// NewCallbackLogWriter ...
func NewCallbackLogWriter(
	onWrite func(isoTime string, tag string, msg string, extra string),
) *CallbackLogWriter {
	return &CallbackLogWriter{onWrite: onWrite}
}

// Write ...
func (p *CallbackLogWriter) Write(
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
type Logger struct {
	level  int32
	writer LogWriter
	AutoLock
}

// NewLogger ...
func NewLogger(writer LogWriter) *Logger {
	if writer == nil {
		return &Logger{
			level:  LogMaskAll,
			writer: NewStdLogWriter(),
		}
	}

	return &Logger{
		level:  LogMaskAll,
		writer: writer,
	}
}

// SetLevel ...
func (p *Logger) SetLevel(level int32) bool {
	if level >= LogMaskNone && level <= LogMaskAll {
		atomic.StoreInt32(&p.level, level)
		return true
	}

	return false
}

// Debug ...
func (p *Logger) Debug(msg string) {
	p.DebugExtra(msg, "")
}

// DebugExtra ...
func (p *Logger) DebugExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskDebug > 0 {
		p.writer.Write(TimeNowISOString(), LogTagDebug, msg, extra)
	}
}

// Info ...
func (p *Logger) Info(msg string) {
	p.InfoExtra(msg, "")
}

// InfoExtra ...
func (p *Logger) InfoExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskInfo > 0 {
		p.writer.Write(TimeNowISOString(), LogTagInfo, msg, extra)
	}
}

// Warn ...
func (p *Logger) Warn(msg string) {
	p.WarnExtra(msg, "")
}

// WarnExtra ...
func (p *Logger) WarnExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskWarn > 0 {
		p.writer.Write(TimeNowISOString(), LogTagWarn, msg, extra)
	}
}

// Error ...
func (p *Logger) Error(msg string) {
	p.ErrorExtra(msg, "")
}

// ErrorExtra ...
func (p *Logger) ErrorExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskError > 0 {
		p.writer.Write(TimeNowISOString(), LogTagError, msg, extra)
	}
}

// Fatal ...
func (p *Logger) Fatal(msg string) {
	p.FatalExtra(msg, "")
}

// FatalExtra ...
func (p *Logger) FatalExtra(msg string, extra string) {
	if atomic.LoadInt32(&p.level)&LogMaskFatal > 0 {
		p.writer.Write(TimeNowISOString(), LogTagFatal, msg, extra)
	}
}

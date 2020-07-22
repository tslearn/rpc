package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"strconv"
)

// LogWriter ...
type LogWriter interface {
	Write(sessionId uint64, err internal.Error)
}

// StdoutLogWriter ...
type StdoutLogWriter struct{}

// NewStdoutLogWriter ...
func NewStdoutLogWriter() LogWriter {
	return &StdoutLogWriter{}
}

func (p *StdoutLogWriter) Write(
	sessionId uint64,
	err internal.Error,
) {
	sb := internal.NewStringBuilder()
	defer sb.Release()
	sb.AppendString(internal.TimeNowISOString())
	if sessionId > 0 {
		sb.AppendByte('(')
		sb.AppendString(strconv.FormatUint(sessionId, 10))
		sb.AppendByte(')')
	}
	sb.AppendByte(' ')
	sb.AppendString(err.Error())
	sb.AppendByte('\n')
	fmt.Print(sb.String())
}

// CallbackLogWriter ...
type CallbackLogWriter struct {
	onWrite func(sessionId uint64, err internal.Error)
}

// NewCallbackLogWriter ...
func NewCallbackLogWriter(
	onWrite func(sessionId uint64, err internal.Error),
) LogWriter {
	return &CallbackLogWriter{onWrite: onWrite}
}

// Write ...
func (p *CallbackLogWriter) Write(
	sessionId uint64,
	err internal.Error,
) {
	if p.onWrite != nil {
		p.onWrite(sessionId, err)
	}
}

package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"strconv"
)

// LogWriter ...
type LogWriter interface {
	Write(sessionID uint64, err core.Error)
}

// StdoutLogWriter ...
type StdoutLogWriter struct{}

// NewStdoutLogWriter ...
func NewStdoutLogWriter() LogWriter {
	return &StdoutLogWriter{}
}

func (p *StdoutLogWriter) Write(
	sessionID uint64,
	err core.Error,
) {
	sb := base.NewStringBuilder()
	defer sb.Release()
	sb.AppendString(base.TimeNowISOString())
	if sessionID > 0 {
		sb.AppendByte('(')
		sb.AppendString(strconv.FormatUint(sessionID, 10))
		sb.AppendByte(')')
	}
	sb.AppendByte(' ')
	sb.AppendString(err.Error())
	sb.AppendByte('\n')
	fmt.Print(sb.String())
}

// CallbackLogWriter ...
type CallbackLogWriter struct {
	onWrite func(sessionId uint64, err core.Error)
}

// NewCallbackLogWriter ...
func NewCallbackLogWriter(
	onWrite func(sessionId uint64, err core.Error),
) LogWriter {
	return &CallbackLogWriter{onWrite: onWrite}
}

// Write ...
func (p *CallbackLogWriter) Write(
	sessionID uint64,
	err core.Error,
) {
	if p.onWrite != nil {
		p.onWrite(sessionID, err)
	}
}

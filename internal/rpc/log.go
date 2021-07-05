package rpc

import (
	"os"
	"strconv"
	"sync"

	"github.com/rpccloud/rpc/internal/base"
)

// ErrorLog ...
type ErrorLog struct {
	level         base.ErrorLevel
	isLogToScreen bool
	file          *os.File
	sync.Mutex
}

func getErrorString(
	machineID uint64,
	sessionID uint64,
	err *base.Error,
) string {
	if err == nil {
		return ""
	}

	machineString := ""
	if machineID != 0 {
		machineString = base.ConcatString(
			"<target:",
			strconv.FormatUint(machineID, 10),
			"> ",
		)
	}
	sessionString := ""
	if sessionID != 0 {
		sessionString = base.ConcatString(
			"<session:",
			strconv.FormatUint(sessionID, 10),
			"> ",
		)
	}

	return base.ConcatString(
		base.ConvertToIsoDateString(base.TimeNow()),
		" ",
		machineString,
		sessionString,
		err.Error(),
		"\n",
	)
}

// NewErrorLog ...
func NewErrorLog(
	isLogToScreen bool,
	outFile string,
	onLogToServer func(stream *Stream),
	level base.ErrorLevel,
) *ErrorLog {
	file, e := func() (*os.File, error) {
		if outFile == "" {
			return nil, nil
		}

		return os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}()

	ret := &ErrorLog{
		level:         level,
		isLogToScreen: isLogToScreen,
		file:          file,
	}

	if e != nil {
		_, _ = os.Stderr.WriteString(getErrorString(
			0,
			0,
			base.ErrLogOpenFile.AddDebug(e.Error()),
		))
		return nil
	}

	ret.Log(0, 0, base.ErrLogOpenFile.AddDebug(e.Error()))

	return ret
}

// Log ...
func (p *ErrorLog) Log(
	machineID uint64,
	sessionID uint64,
	err *base.Error,
) bool {
	if err == nil {
		return false
	}

	if err.GetLevel()&p.level == 0 {
		return false
	}

	logStr := getErrorString(machineID, sessionID, err)

	if p.isLogToScreen {
		_, _ = os.Stdout.WriteString(logStr)
	}

	if p.file != nil {
		_, _ = p.file.WriteString(logStr)
	}

	return true
}

// Close ...
func (p *ErrorLog) Close() bool {
	p.Lock()
	defer p.Unlock()

	if p.file != nil {
		if e := p.file.Close(); e != nil {
			p.Log(0, 0, base.ErrLogCloseFile.AddDebug(e.Error()))
			return false
		}

		p.file = nil
		return true
	}

	return true
}

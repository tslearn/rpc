package base

import (
	"os"
	"sync"
)

// Logger ...
type Logger struct {
	isLogToScreen bool
	file          *os.File
	sync.Mutex
}

// NewLogger ...
func NewLogger(isLogToScreen bool, outFile string) (*Logger, *Error) {
	file, e := func() (*os.File, error) {
		if outFile == "" {
			return nil, nil
		}

		return os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}()

	if e != nil {
		return nil, ErrLogOpenFile.AddDebug(e.Error())
	}

	return &Logger{
		isLogToScreen: isLogToScreen,
		file:          file,
	}, nil
}

// Log ...
func (p *Logger) Log(str string) *Error {
	if p.isLogToScreen {
		if _, e := os.Stdout.WriteString(str); e != nil {
			return ErrLogWriteFile.AddDebug(e.Error())
		}
	}

	if p.file != nil {
		if _, e := p.file.WriteString(str); e != nil {
			return ErrLogWriteFile.AddDebug(e.Error())
		}
	}

	return nil
}

// Close ...
func (p *Logger) Close() *Error {
	p.Lock()
	defer p.Unlock()

	if p.file != nil {
		if e := p.file.Close(); e != nil {
			return ErrLogCloseFile.AddDebug(e.Error())
		}

		p.file = nil
		return nil
	}

	return nil
}
